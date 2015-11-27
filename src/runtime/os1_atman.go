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

func newosproc(mp *m, stk unsafe.Pointer) {
	println("newosproc(m, stk)")
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
