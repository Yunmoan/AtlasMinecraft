package server

import (
	"syscall"
	"time"
	"unsafe"
)

var getProcessMemoryInfo = syscall.NewLazyDLL("kernel32.dll").NewProc("K32GetProcessMemoryInfo")

type processMemoryCounters struct {
	Size, PageFaultCount                                             uint32
	PeakWorkingSet, WorkingSet, QuotaPeakPagedPool, QuotaPagedPool   uintptr
	QuotaPeakNonPagedPool, QuotaNonPagedPool, Pagefile, PeakPagefile uintptr
}

func readProcessUsage() processUsage {
	var usage processUsage
	handle, err := syscall.GetCurrentProcess()
	if err != nil {
		return usage
	}
	var creation, exit, kernel, user syscall.Filetime
	if syscall.GetProcessTimes(handle, &creation, &exit, &kernel, &user) == nil {
		kernelTicks := uint64(kernel.HighDateTime)<<32 | uint64(kernel.LowDateTime)
		userTicks := uint64(user.HighDateTime)<<32 | uint64(user.LowDateTime)
		usage.cpu = time.Duration(kernelTicks+userTicks) * 100 * time.Nanosecond
		usage.cpuOK = true
	}
	if getProcessMemoryInfo.Find() == nil {
		var mem processMemoryCounters
		mem.Size = uint32(unsafe.Sizeof(mem))
		ok, _, _ := getProcessMemoryInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&mem)), uintptr(mem.Size))
		if ok != 0 {
			usage.resident, usage.residentOK = uint64(mem.WorkingSet), true
		}
	}
	return usage
}
