package collector

// This file collects lightweight, best-effort local host resource estimates.

import (
	"fmt"
	"sync"
)

type HostSampler struct {
	mu        sync.Mutex
	lastTotal uint64
	lastIdle  uint64
	haveCPU   bool
}

type HostMetrics struct {
	CPUUsage         *float64
	MemoryUsedBytes  *float64
	MemoryTotalBytes *float64
	DiskUsedBytes    *float64
	DiskTotalBytes   *float64
}

func NewHostSampler() *HostSampler { return &HostSampler{} }

func (s *HostSampler) Sample(filesystemPath string) (HostMetrics, []error) {
	var metrics HostMetrics
	var warnings []error
	if value, err := s.cpuUsage(); err != nil {
		warnings = append(warnings, fmt.Errorf("host CPU usage unavailable: %w", err))
	} else {
		metrics.CPUUsage = value
	}
	if used, total, err := hostMemory(); err != nil {
		warnings = append(warnings, fmt.Errorf("host memory unavailable: %w", err))
	} else {
		metrics.MemoryUsedBytes = &used
		metrics.MemoryTotalBytes = &total
	}
	if used, total, err := diskUsage(filesystemPath); err != nil {
		warnings = append(warnings, fmt.Errorf("host disk usage unavailable: %w", err))
	} else {
		metrics.DiskUsedBytes = &used
		metrics.DiskTotalBytes = &total
	}
	return metrics, warnings
}
