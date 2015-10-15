package runtime

import "unsafe"

var _atman_stack [0x10000]byte

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

// mem

func sysAlloc(n uintptr, sysStat *uint64) unsafe.Pointer                    { return nil }
func sysFree(v unsafe.Pointer, n uintptr, sysStat *uint64)                  {}
func sysMap(v unsafe.Pointer, n uintptr, reserved bool, sysStat *uint64)    {}
func sysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer { return nil }
func sysUnused(v unsafe.Pointer, n uintptr)                                 {}
func sysUsed(v unsafe.Pointer, n uintptr)                                   {}
func sysFault(v unsafe.Pointer, n uintptr)                                  {}

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
