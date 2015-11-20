package runtime

import "unsafe"

func sysFree(v unsafe.Pointer, n uintptr, sysStat *uint64) {}

func sysUnused(v unsafe.Pointer, n uintptr) {}
func sysUsed(v unsafe.Pointer, n uintptr)   {}
func sysFault(v unsafe.Pointer, n uintptr)  {}

// sysMap makes n bytes at v readable and writable and adjusts the stats.
func sysMap(v unsafe.Pointer, n uintptr, reserved bool, sysStat *uint64) {
	println("sysMap(", v, ",", n, ",", reserved, ", ...)")

	mSysStatInc(sysStat, n)
	p := memAlloc(v, n)
	if p != v {
		throw("runtime: cannot map pages in arena address space")
	}
}

// sysAlloc allocates n bytes, adjusts sysStat, and returns the address
// of the allocated bytes.
func sysAlloc(n uintptr, sysStat *uint64) unsafe.Pointer {
	println("sysAlloc(", n, ",", ", ...)")

	p := memAlloc(nil, n)
	if p != nil {
		mSysStatInc(sysStat, n)
	}
	return p
}

// sysReserve reserves n bytes at v and updates reserved.
func sysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	println("sysReserve(", v, ",", n, ", ...)")

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

	nextHeapPage      vaddr
	nextPageTablePage vaddr

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
	mm.nextPageTablePage = vaddr(0xFFFF880000000000)

	mm.l4 = newXenPageTable(mm.l4PFN.vaddr())
	mm.l3Temp = newXenPageTable(mm.l3TempPFN.vaddr())
	mm.l2Temp = newXenPageTable(mm.l2TempPFN.vaddr())
	mm.l1Temp = newXenPageTable(mm.l1TempPFN.vaddr())

	mm.zeroTempPageTables()
}

func (mm *atmanMemoryManager) zeroTempPageTables() {
	memclr(unsafe.Pointer(mm.l3TempPFN.vaddr()), _PAGESIZE)
	memclr(unsafe.Pointer(mm.l2TempPFN.vaddr()), _PAGESIZE)
	memclr(unsafe.Pointer(mm.l1TempPFN.vaddr()), _PAGESIZE)
}

func (mm *atmanMemoryManager) allocPages(v unsafe.Pointer, n uint64) unsafe.Pointer {
	println("Requested allocation of", n, "pages at", v)
	return v
}
