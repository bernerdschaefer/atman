package runtime

import "unsafe"

func sysFree(v unsafe.Pointer, n uintptr, sysStat *uint64) {}

func sysUnused(v unsafe.Pointer, n uintptr) {}
func sysUsed(v unsafe.Pointer, n uintptr)   {}
func sysFault(v unsafe.Pointer, n uintptr)  {}

var _atman_mem_claimed uint64

// sysMap makes n bytes at v readable and writable and adjusts the stats.
func sysMap(v unsafe.Pointer, n uintptr, reserved bool, sysStat *uint64) {
	println("sysMap(", v, ",", n, ",", reserved, ", ...)")
}

// sysAlloc allocates n bytes, adjusts sysStat, and returns the address
// of the allocated bytes.
func sysAlloc(n uintptr, sysStat *uint64) unsafe.Pointer {
	println("sysAlloc(", n, ",", ", ...)")

	mSysStatInc(sysStat, n)
	return unsafe.Pointer(_atman_mem_start_addr)
}

// sysReserve reserves n bytes at v and updates reserved.
func sysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	println("sysReserve(", v, ",", n, ", ...)")

	*reserved = false
	return v
}
