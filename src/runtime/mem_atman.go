package runtime

import "unsafe"

func sysFree(v unsafe.Pointer, n uintptr, sysStat *uint64)               {}
func sysMap(v unsafe.Pointer, n uintptr, reserved bool, sysStat *uint64) {}
func sysUnused(v unsafe.Pointer, n uintptr)                              {}
func sysUsed(v unsafe.Pointer, n uintptr)                                {}
func sysFault(v unsafe.Pointer, n uintptr)                               {}

func sysAlloc(n uintptr, sysStat *uint64) unsafe.Pointer {
	mSysStatInc(sysStat, n)
	return unsafe.Pointer(n)
}

func sysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	*reserved = true
	return unsafe.Pointer(v)
}
