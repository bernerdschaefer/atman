package runtime

import "unsafe"

func osinit() {
	ncpu = 1
}

func sigpanic() {}

func crash() {
	*(*int32)(nil) = 0
}

func goenvs() {}

//go:nowritebarrier
func newosproc(mp *m, stk unsafe.Pointer) {
	mp.tls[0] = uintptr(mp.id) // so 386 asm can find it
	if true {
		print("newosproc stk=", stk, " m=", mp, " g=", mp.g0, " id=", mp.id, "/", mp.tls[0], " ostk=", &mp, "\n")
	}

}

func resetcpuprofiler(hz int32) {}

func minit() {
	println("minit()")
}

func unminit() {
	println("unminit()")
}

func mpreinit(mp *m) {
	println("mpreinit(m)")
}

func msigsave(mp *m) {
	println("msigsave(m)")
}

//go:nosplit
func osyield() {
	println("osyield()")
}

// Create a semaphore, which will be assigned to m->waitsema.
// The zero value is treated as absence of any semaphore,
// so be sure to return a non-zero value.
//go:nosplit
func semacreate() uintptr {
	println("semacreate()")
	return 1
}

// If ns < 0, acquire m->waitsema and return 0.
// If ns >= 0, try to acquire m->waitsema for at most ns nanoseconds.
// Return 0 if the semaphore was acquired, -1 if interrupted or timed out.
//go:nosplit
func semasleep(ns int64) int32 {
	return 0
}

// Wake up mp, which is or will soon be sleeping on mp->waitsema.
//go:nosplit
func semawakeup(mp *m) int32 {
	return 0
}
