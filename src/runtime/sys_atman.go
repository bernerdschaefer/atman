package runtime

import "unsafe"

const (
	_PAGESIZE = 0x1000
)

var (
	_atman_stack            [8 * _PAGESIZE]byte
	_atman_hypercall_page   [2 * _PAGESIZE]byte
	_atman_shared_info_page [2 * _PAGESIZE]byte

	_atman_start_info  = &xenStartInfo{}
	_atman_shared_info = &xenSharedInfo{}
)

//go:nosplit
func getRandomData(r []byte) {
	extendRandom(r, 0)
}

// lock

const (
	active_spin     = 4
	active_spin_cnt = 30
)

func lock(l *mutex)   {}
func unlock(l *mutex) {}

func noteclear(n *note)                  {}
func notewakeup(n *note)                 {}
func notesleep(n *note)                  {}
func notetsleep(n *note, ns int64) bool  { return false }
func notetsleepg(n *note, ns int64) bool { return false }

// env

func gogetenv(key string) string { return "" }

var _cgo_setenv unsafe.Pointer   // pointer to C function
var _cgo_unsetenv unsafe.Pointer // pointer to C function

// os

func sigpanic() {}
func crash()    {}
func goenvs()   {}

func newosproc(mp *m, stk unsafe.Pointer) {}

func resetcpuprofiler(hz int32) {}

func minit()         {}
func unminit()       {}
func mpreinit(mp *m) {}
func msigsave(mp *m) {}

//go:nosplit
func osyield() {}

func osinit() {}

// signals

const _NSIG = 0

func initsig()                 {}
func sigdisable(uint32)        {}
func sigenable(uint32)         {}
func sigignore(uint32)         {}
func raisebadsignal(sig int32) {}

// net

func netpoll(block bool) *g { return nil }
func netpollinited() bool   { return false }

type xenStartInfo struct {
	Magic          [32]byte
	NrPages        uint64
	SharedInfoAddr uintptr // machine address of share info struct
	SIFFlags       uint32
	StoreMfn       uint64 // machine page number of shared page
	StoreEventchn  uint32
	Console        struct {
		Mfn      uint64 // machine page number of console page
		Eventchn uint32 // event channel
	}
	PageTableBase     uint64 // virtual address of page directory
	NrPageTableFrames uint64
	MfnList           uintptr // virtual address of page-frame list
	ModStart          uintptr // virtual address of pre-loaded module
	ModLen            uint64  // size (bytes) of pre-loaded module
	CmdLine           [1024]byte

	// The pfn range here covers both page table and p->m table frames
	FirstP2mPfn uint64 // 1st pfn forming initial P->M table
	NrP2mFrames uint64 // # of pgns forming initial P->M table
}

type xenSharedInfo struct {
	VCPUInfo      [32]vcpuInfo
	EvtchnPending [64]uint64
	EvtchnMask    [64]uint64
	WcVersion     uint32
	WcSec         uint32
	WcNsec        uint32
	_             [4]byte
	Arch          archSharedInfo
}

type archSharedInfo struct {
	MaxPfn                uint64
	PfnToMfnFrameListList uint64
	NmiReason             uint64
	_                     [32]uint64
}

type archVCPUInfo struct {
	CR2 uint64
	_   uint64
}

type vcpuTimeInfo struct {
	Version        uint32
	_              uint32
	TscTimestamp   uint64
	SystemTime     uint64
	TscToSystemMul uint32
	TscShift       int8
	_              [3]int8
}

type vcpuInfo struct {
	UpcallPending uint8
	UpcallMask    uint8
	_             [6]byte
	PendingSel    uint64
	Arch          archVCPUInfo
	Time          vcpuTimeInfo
}

func atmaninit() {
	println("Atman OS")
	println("   start_info: ", _atman_start_info)
	println("        magic: ", string(_atman_start_info.Magic[:]))
	println("     nr_pages: ", _atman_start_info.NrPages)
	println("  shared_info: ", _atman_start_info.SharedInfoAddr)
	println("   siff_flags: ", _atman_start_info.SIFFlags)
	println("    store_mfn: ", _atman_start_info.StoreMfn)
	println("    store_evc: ", _atman_start_info.StoreEventchn)
	println("  console_mfn: ", _atman_start_info.Console.Mfn)
	println("  console_evc: ", _atman_start_info.Console.Eventchn)
	println("      pt_base: ", _atman_start_info.PageTableBase)
	println(" nr_pt_frames: ", _atman_start_info.NrPageTableFrames)
	println("     mfn_list: ", _atman_start_info.MfnList)
	println("    mod_start: ", _atman_start_info.ModStart)
	println("      mod_len: ", _atman_start_info.ModLen)
	println("     cmd_line: ", _atman_start_info.CmdLine[:])
	println("    first_pfn: ", _atman_start_info.FirstP2mPfn)
	println("nr_p2m_frames: ", _atman_start_info.NrP2mFrames)

	println("setting _atman_phys_to_machine_mapping")
	_atman_phys_to_machine_mapping = *(*[8192]uint64)(unsafe.Pointer(
		_atman_start_info.MfnList,
	))

	printPageDirectory()

	println("mapping _atman_start_info")
	mapSharedInfo(_atman_start_info.SharedInfoAddr, _atman_shared_info)

	// printPageDirectory()
}

func printPageDirectory() {
	var table xenPageDirectory
	table = *(*xenPageDirectory)(unsafe.Pointer(uintptr(
		_atman_start_info.PageTableBase,
	)))

	for _, x := range table {
		if x.hasFlag(xenPageTablePresent) {
			x.debug()
			println("pte2virt: ", pteToVirt(x))
		}
	}
}

func pteToVirt(e pageTableEntry) unsafe.Pointer {
	const (
		m2p xenMachineToPhysicalMap = 0xFFFF800000000000
	)

	return unsafe.Pointer(m2p.Get(e.Mfn()) << 12)
}

func mapSharedInfo(vaddr uintptr, i *xenSharedInfo) {
	pageAddr := roundUpPage(
		uintptr(unsafe.Pointer(&_atman_shared_info_page[0])),
	)

	ret := HYPERVISOR_update_va_mapping(
		pageAddr,
		vaddr|7,
		2, // UVMF_INVLPG: flush only one entry
	)

	if ret != 0 {
		println("HYPERVISOR_update_va_mapping returned ", ret)
		panic("HYPERVISOR_update_va_mapping failed")
	}

	*i = *(*xenSharedInfo)(unsafe.Pointer(pageAddr))
}

func roundUpPage(addr uintptr) uintptr {
	return (addr + _PAGESIZE - 1) &^ (_PAGESIZE - 1)
}

func hypercall(trap, a1, a2, a3 uintptr) uintptr

func HYPERVISOR_console_io(op uint64, size uint64, data uintptr) uintptr {
	const _HYPERVISOR_console_io = 18

	return hypercall(
		_HYPERVISOR_console_io,
		uintptr(op),
		uintptr(size),
		data,
	)
}

func HYPERVISOR_update_va_mapping(vaddr uintptr, val uintptr, flags uint64) uintptr {
	const _HYPERVISOR_update_va_mapping = 14

	return hypercall(
		_HYPERVISOR_update_va_mapping,
		vaddr,
		val,
		uintptr(flags),
	)
}

// memory management

var (
	// Map of (pseudo-)physical addresses to machine addresses.
	_atman_phys_to_machine_mapping [8192]uint64
)

// Entry in level 3, 2, or 1 page table.
//
// - 63 if set means No execute (NX)
// - 51-13 the machine frame number
// - 12 available for guest
// - 11 available for guest
// - 10 available for guest
// - 9 available for guest
// - 8 global
// - 7 PAT (PSE is disabled, must use hypercall to make 4MB or 2MB pages)
// - 6 dirty
// - 5 accessed
// - 4 page cached disabled
// - 3 page write through
// - 2 userspace accessible
// - 1 writeable
// - 0 present
type pageTableEntry uintptr

const (
	xenPageTablePresent = 1 << iota
	xenPageTableWritable
	xenPageTableUserspaceAccessible
	xenPageTablePageWriteThrough
	xenPageTablePageCacheDisabled
	xenPageTableAccessed
	xenPageTableDirty
	xenPageTablePAT
	xenPageTableGlobal
	xenPageTableGuest1
	xenPageTableGuest2
	xenPageTableGuest3
	xenPageTableGuest4
	xenPageTableNoExecute = 1 << 63

	xenPageAddrMask  = 1<<52 - 1
	xenPageMask      = 1<<12 - 1
	xenPageFlagShift = 12
)

func (e pageTableEntry) debug() {
	println(
		"PTE<", unsafe.Pointer(e), ">:",
		" MFN=", e.Mfn(),
		"  NX=", e.hasFlag(xenPageTableNoExecute),
		"   G=", e.hasFlag(xenPageTableGlobal),
		" PAT=", e.hasFlag(xenPageTablePAT),
		" DIR=", e.hasFlag(xenPageTableDirty),
		"   A=", e.hasFlag(xenPageTableAccessed),
		" PCD=", e.hasFlag(xenPageTablePageCacheDisabled),
		" PWT=", e.hasFlag(xenPageTablePageWriteThrough),
		"   U=", e.hasFlag(xenPageTableUserspaceAccessible),
		"   W=", e.hasFlag(xenPageTableWritable),
		"   P=", e.hasFlag(xenPageTablePresent),
	)
}

func (e pageTableEntry) hasFlag(f uintptr) bool {
	return uintptr(e)&f == f
}

func (e pageTableEntry) Mfn() uintptr {
	return (uintptr(e) & (xenPageAddrMask &^ xenPageMask)) >> xenPageFlagShift
}

type (
	xenPageDirectory [512]pageTableEntry
	xenPageTable     [512]pageTableEntry
)

type xenMachineToPhysicalMap uintptr

func (m xenMachineToPhysicalMap) Get(i uintptr) uintptr {
	return *(*uintptr)(add(unsafe.Pointer(m), i*ptrSize))
}
