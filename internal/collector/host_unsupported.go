//go:build !darwin && !linux

package collector

import (
	"fmt"
	"runtime"
)

func (s *HostSampler) cpuUsage() (*float64, error) {
	return nil, fmt.Errorf("unsupported platform %s", runtime.GOOS)
}

func hostMemory() (float64, float64, error) {
	return 0, 0, fmt.Errorf("unsupported platform %s", runtime.GOOS)
}
