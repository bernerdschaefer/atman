package runtime

import "unsafe"

const (
	_PAGESIZE = 0x1000
)

var (
	_atman_stack          [8 * _PAGESIZE]byte
	_atman_start_info     = &xenStartInfo{}
	_atman_hypercall_page [2 * _PAGESIZE]byte

	_atman_phys_to_machine_mapping = [256]uint64{}
)

func HYPERVISOR_console_io(op uint64, size uint64, data uintptr)

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
	SharedInfoAddr uint64 // machine address of share info struct
	SIFFlags       uint32
	StoreMfn       uint64 // machine page number of shared page
	StoreEventchn  uint32
	Console        struct {
		Mfn      uint64 // machine page number of console page
		Eventchn uint32 // event channel
	}
	_                 uint64 // dom0 console
	PageTableBase     uint64 // virtual address of page directory
	NrPageTableFrames uint64
	MfnList           uint64 // virtual address of page-frame list
	ModStart          uint64 // virtual address of pre-loaded module
	ModLen            uint64 // size (bytes) of pre-loaded module
	CmdLine           [1024]byte

	// The pfn range here covers both page table and p->m table frames
	FirstP2mPfn uint64 // 1st pfn forming initial P->M table
	NrP2mFrames uint64 // # of pgns forming initial P->M table
}
