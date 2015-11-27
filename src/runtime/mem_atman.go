package runtime

import "unsafe"

func sysFree(v unsafe.Pointer, n uintptr, sysStat *uint64) {}

func sysUnused(v unsafe.Pointer, n uintptr) {}
func sysUsed(v unsafe.Pointer, n uintptr)   {}
func sysFault(v unsafe.Pointer, n uintptr)  {}

// sysMap makes n bytes at v readable and writable and adjusts the stats.
func sysMap(v unsafe.Pointer, n uintptr, reserved bool, sysStat *uint64) {
	mSysStatInc(sysStat, n)
	p := memAlloc(v, n)
	if p != v {
		throw("runtime: cannot map pages in arena address space")
	}
}

// sysAlloc allocates n bytes, adjusts sysStat, and returns the address
// of the allocated bytes.
func sysAlloc(n uintptr, sysStat *uint64) unsafe.Pointer {
	p := memAlloc(nil, n)
	if p != nil {
		mSysStatInc(sysStat, n)
	}
	return p
}

// sysReserve reserves n bytes at v and updates reserved.
func sysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	*reserved = false
	return v
}

// memAlloc allocates n bytes of memory at address v
// and returns a pointer to the allocated memory.
// If v is nil, an address will be chosen.
func memAlloc(v unsafe.Pointer, n uintptr) unsafe.Pointer {
	requiredPages := uint64(round(n, _PAGESIZE) / _PAGESIZE)

	return _atman_mm.allocPages(v, requiredPages)
}

var _atman_mm = &atmanMemoryManager{}

type atmanMemoryManager struct {
	bootstrapPageTablePFN pfn // start of bootstrap page tables
	bootstrapStackPFN     pfn // start of bootstrap stack
	l4PFN                 pfn
	l3TempPFN             pfn // staging area for allocating page tables
	l2TempPFN             pfn
	l1TempPFN             pfn
	bootstrapEndPFN       pfn // end of bootstrap region

	nextPFN pfn // next free frame
	lastPFN pfn

	nextHeapPage vaddr

	l4     xenPageTable
	l3Temp xenPageTable
	l2Temp xenPageTable
	l1Temp xenPageTable
}

func (mm *atmanMemoryManager) init() {
	var (
		pageTableBase = _atman_start_info.PageTableBase
		ptStartPfn    = pageTableBase.pfn()
		ptEndPfn      = ptStartPfn.add(_atman_start_info.NrPageTableFrames)

		bootstrapStackPFN  = ptEndPfn.add(1)
		bootstrapStackAddr = bootstrapStackPFN.vaddr()

		bootstrapEnd = round(
			uintptr(bootstrapStackAddr)+0x80000, // minimum 512kB padding
			0x400000, // 4MB alignment
		)
		bootstrapEndPFN = vaddr(bootstrapEnd).pfn()
	)

	mm.bootstrapPageTablePFN = _atman_start_info.PageTableBase.pfn()
	mm.bootstrapStackPFN = bootstrapStackPFN
	mm.l4PFN = pageTableBase.pfn()
	mm.l3TempPFN = bootstrapEndPFN - 3
	mm.l2TempPFN = bootstrapEndPFN - 2
	mm.l1TempPFN = bootstrapEndPFN - 1
	mm.bootstrapEndPFN = bootstrapEndPFN
	mm.nextPFN = bootstrapEndPFN.add(1)
	mm.lastPFN = pfn(_atman_start_info.NrPages)

	mm.nextHeapPage = mm.nextPFN.vaddr()

	mm.l4 = newXenPageTable(mm.l4PFN.vaddr())
	mm.l3Temp = newXenPageTable(mm.l3TempPFN.vaddr())
	mm.l2Temp = newXenPageTable(mm.l2TempPFN.vaddr())
	mm.l1Temp = newXenPageTable(mm.l1TempPFN.vaddr())

	mm.zeroTempPageTables()
	mm.migrateBootstrapPageTables()
}

func (mm *atmanMemoryManager) zeroTempPageTables() {
	memclr(unsafe.Pointer(mm.l3TempPFN.vaddr()), _PAGESIZE)
	memclr(unsafe.Pointer(mm.l2TempPFN.vaddr()), _PAGESIZE)
	memclr(unsafe.Pointer(mm.l1TempPFN.vaddr()), _PAGESIZE)
}

func (mm *atmanMemoryManager) allocPages(v unsafe.Pointer, n uint64) unsafe.Pointer {
	if v == nil {
		v = mm.reserveHeapPages(n)
	}

	for page := vaddr(v); page < vaddr(v)+vaddr(n*_PAGESIZE); page += _PAGESIZE {
		mm.allocPage(page)
	}

	return v
}

// allocPage makes page a writeable userspace page.
func (mm *atmanMemoryManager) allocPage(page vaddr) {
	var (
		l4offset = page.pageTableOffset(pageTableLevel4)
		l3offset = page.pageTableOffset(pageTableLevel3)
		l2offset = page.pageTableOffset(pageTableLevel2)
		l1offset = page.pageTableOffset(pageTableLevel1)

		l4 = mm.l4
	)

	l3pte := l4.Get(l4offset)

	if !l3pte.hasFlag(xenPageTablePresent) {
		l3pte = mm.allocPageTable(mm.l4PFN, l4offset)
	}

	l3 := newXenPageTable(mm.pageTableAddr(l3pte.pfn()))
	l2pte := l3.Get(l3offset)

	if !l2pte.hasFlag(xenPageTablePresent) {
		l2pte = mm.allocPageTable(l3pte.pfn(), l3offset)
	}

	l2 := newXenPageTable(mm.pageTableAddr(l2pte.pfn()))
	l1pte := l2.Get(l2offset)

	if !l1pte.hasFlag(xenPageTablePresent) {
		l1pte = mm.allocPageTable(l2pte.pfn(), l2offset)
	}

	pagepfn := mm.reservePFN()

	mm.clearPage(pagepfn)
	mm.writePte(l1pte.pfn(), l1offset, pagepfn, PTE_PAGE_FLAGS)
	*(*uintptr)(unsafe.Pointer(page)) = 0x0 // ensure page is writable
}

func (mm *atmanMemoryManager) pageTableWalk(addr vaddr) {
	var (
		l4offset = addr.pageTableOffset(pageTableLevel4)
		l3offset = addr.pageTableOffset(pageTableLevel3)
		l2offset = addr.pageTableOffset(pageTableLevel2)
		l1offset = addr.pageTableOffset(pageTableLevel1)

		l4 = mm.l4
	)

	println("page table walk from", unsafe.Pointer(addr))
	print("L4[")
	print(l4offset)
	print("] = ")

	l3pte := l4.Get(l4offset)
	l3pte.debug()

	if !l3pte.hasFlag(xenPageTablePresent) {
		return
	}

	l3 := newXenPageTable(mm.pageTableAddr(l3pte.pfn()))
	print("L3[")
	print(l3offset)
	print("] = ")

	l2pte := l3.Get(l3offset)
	l2pte.debug()

	if !l2pte.hasFlag(xenPageTablePresent) {
		return
	}

	l2 := newXenPageTable(mm.pageTableAddr(l2pte.pfn()))
	print("L2[")
	print(l2offset)
	print("] = ")

	l1pte := l2.Get(l2offset)
	l1pte.debug()

	if !l1pte.hasFlag(xenPageTablePresent) {
		return
	}

	l1 := newXenPageTable(mm.pageTableAddr(l1pte.pfn()))
	print("L1[")
	print(l1offset)
	print("] = ")

	l0pte := l1.Get(l1offset)
	l0pte.debug()

	if !l0pte.hasFlag(xenPageTablePresent) {
		return
	}
}

// allocPageTable allocates a new page table
// and installs it into previous page table.
func (mm *atmanMemoryManager) allocPageTable(prev pfn, prevOffset int) pageTableEntry {
	pagepfn := mm.allocPageTablePage()
	return mm.writePte(prev, prevOffset, pagepfn, PTE_PAGE_TABLE_FLAGS|xenPageTableWritable)
}

func (mm *atmanMemoryManager) reserveHeapPages(n uint64) unsafe.Pointer {
	var p vaddr
	p, mm.nextHeapPage = mm.nextHeapPage, mm.nextHeapPage+vaddr(n*_PAGESIZE)
	return unsafe.Pointer(p)
}

func (mm *atmanMemoryManager) reservePFN() pfn {
	var p pfn
	p, mm.nextPFN = mm.nextPFN, mm.nextPFN+1
	return p
}

// migrateBootstrapPageTables relocates the bootstrap page tables
// into the linear-mapped upper region of memory used by atman
// to store page tables.
func (mm *atmanMemoryManager) migrateBootstrapPageTables() {
	// step 1: make PML4 (page map level 4) reachable from high address
	var pml4addr = mm.pageTableAddr(mm.bootstrapPageTablePFN)

	mm.mapPageTablePage(mm.bootstrapPageTablePFN)
	mm.l4 = newXenPageTable(pml4addr)

	// now make the remaining page tables reachable
	for i := uint64(1); i < _atman_start_info.NrPageTableFrames; i++ {
		mm.mapPageTablePage(mm.bootstrapPageTablePFN.add(i))
	}

	// now unmap the old page table pages
	for i := uint64(1); i < _atman_start_info.NrPageTableFrames; i++ {
		addr := mm.bootstrapPageTablePFN.add(i).vaddr()
		HYPERVISOR_update_va_mapping(uintptr(addr), 0, 2)
	}
}

func (mm *atmanMemoryManager) clearPage(pfn pfn) {
	mm.mmuExtOp([]mmuExtOp{
		{
			cmd:  16, // MMUEXT_CLEAR_PAGE
			arg1: uint64(pfn.mfn()),
		},
	})
}

func (mm *atmanMemoryManager) mmuExtOp(ops []mmuExtOp) {
	ret := HYPERVISOR_mmuext_op(ops, DOMID_SELF)

	if ret != 0 {
		println("HYPERVISOR_mmuext_op returned", ret)
	}
}

func (mm *atmanMemoryManager) writePte(table pfn, offset int, value pfn, flags uintptr) pageTableEntry {
	newpte := pageTableEntry(value.mfn() << xenPageFlagShift)
	newpte.setFlag(flags)

	updates := []mmuUpdate{
		{
			ptr: uintptr((table.mfn() << xenPageFlagShift)) + uintptr(offset*ptrSize),
			val: uintptr(newpte),
		},
	}
	ret := HYPERVISOR_mmu_update(updates, DOMID_SELF)

	if ret != 0 {
		println("writePte: HYPERVISOR_mmu_update returned", ret)
	}

	return newpte
}

func (mm *atmanMemoryManager) pageTableAddr(pfn pfn) vaddr {
	const pageTableVaddrOffset = vaddr(0xFFFF880000000000)

	return pageTableVaddrOffset + pfn.vaddr()
}

// allocPageTablePage maps a new page in the page table address space
// and makes it read-only.
func (mm *atmanMemoryManager) allocPageTablePage() pfn {
	pfn := mm.reservePFN()
	mm.mapPageTablePage(pfn)
	return pfn
}

func (mm *atmanMemoryManager) mapPageTablePage(pagepfn pfn) {
	var (
		page = mm.pageTableAddr(pagepfn)

		l4offset = page.pageTableOffset(pageTableLevel4)
		l3offset = page.pageTableOffset(pageTableLevel3)
		l2offset = page.pageTableOffset(pageTableLevel2)
		l1offset = page.pageTableOffset(pageTableLevel1)

		l4pfn pfn = mm.l4PFN

		l4 xenPageTable = mm.l4
		l3 xenPageTable
		l2 xenPageTable
	)

	var (
		l3pte = l4.Get(l4offset)
		l3pfn = l3pte.pfn()
	)

	if !l3pte.hasFlag(xenPageTablePresent) {
		l3pfn = mm.reservePFN()

		mm.clearPage(l3pfn)
		mm.writePte(l4pfn, l4offset, l3pfn, PTE_PAGE_TABLE_FLAGS|PTE_TEMP)
		mm.mapPageTablePage(l3pfn)
		mm.writePte(l4pfn, l4offset, l3pfn, PTE_PAGE_TABLE_FLAGS)
	}

	if l3pte.hasFlag(xenPageTableGuest1) {
		// we're in the process of mapping this page
		HYPERVISOR_update_va_mapping(uintptr(mm.l3TempPFN.vaddr()), uintptr(l3pte), 2)
		l3 = mm.l3Temp
	} else {
		l3 = newXenPageTable(mm.pageTableAddr(l3pfn))
	}

	var (
		l2pte = l3.Get(l3offset)
		l2pfn = l2pte.pfn()
	)

	if !l2pte.hasFlag(xenPageTablePresent) {
		l2pfn = mm.reservePFN()

		mm.clearPage(l2pfn)
		mm.writePte(l3pfn, l3offset, l2pfn, PTE_PAGE_TABLE_FLAGS|PTE_TEMP)
		mm.mapPageTablePage(l2pfn)
		mm.writePte(l3pfn, l3offset, l2pfn, PTE_PAGE_TABLE_FLAGS)
	}

	if l2pte.hasFlag(xenPageTableGuest1) {
		// we're in the process of mapping this page
		HYPERVISOR_update_va_mapping(uintptr(mm.l2TempPFN.vaddr()), uintptr(l2pte), 2)
		l2 = mm.l2Temp
	} else {
		l2 = newXenPageTable(mm.pageTableAddr(l2pfn))
	}

	var (
		l1pte = l2.Get(l2offset)
		l1pfn = l1pte.pfn()
	)

	if !l1pte.hasFlag(xenPageTablePresent) {
		l1pfn = mm.reservePFN()

		mm.clearPage(l1pfn)
		mm.writePte(l2pfn, l2offset, l1pfn, PTE_PAGE_TABLE_FLAGS|PTE_TEMP)
		mm.mapPageTablePage(l1pfn)
		mm.writePte(l2pfn, l2offset, l1pfn, PTE_PAGE_TABLE_FLAGS)
	}

	mm.writePte(l1pfn, l1offset, pagepfn, PTE_PAGE_TABLE_FLAGS)
}
