package models

// #include "../../c/cpu_timer.h"
import (
	"runtime"

	"github.com/markpotocki/health/internal/status"
)

type HealthStatus struct {
	CPU     HealthStatusCpu     `json:"cpu"`
	Memory  HealthStatusMem     `json:"mem"`
	Network HealthStatusNetwork `json:"network"`
}

type HealthStatusMem struct {
	ProcUsed  uint64 `json:"proc_used"`
	ProcTotal uint64 `json:"proc_total"`
	SysTotal  uint64 `json:"sys_total"`
}

type HealthStatusCpu struct {
	Cores       int `json:"cores"`
	Utilization int `json:"use"`
}

type HealthStatusNetwork struct {
	AverageTime float64 `json:"avg_response"`
}

func MakeHealthStatus() HealthStatus {
	// MEMORY
	runtimeMemory := runtime.MemStats{}
	runtime.ReadMemStats(&runtimeMemory)

	heapUsedMem := runtimeMemory.Alloc
	heapTotalMem := runtimeMemory.TotalAlloc
	heapSysTotalMem := runtimeMemory.Sys

	// CPU
	cpuCores := runtime.NumCPU()
	cpuUtil := status.CPUUtilization()

	// NETWORK
	averageResponse := status.GlobalNetworkInformation.Average()

	hs := HealthStatus{
		CPU: HealthStatusCpu{
			Cores:       cpuCores,
			Utilization: int(cpuUtil.Total),
		},
		Memory: HealthStatusMem{
			ProcUsed:  heapUsedMem,
			ProcTotal: heapTotalMem,
			SysTotal:  heapSysTotalMem,
		},
		Network: HealthStatusNetwork{
			AverageTime: averageResponse,
		},
	}

	return hs
}
