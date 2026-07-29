//go:build darwin

package collector

import (
	"testing"
	"time"
)

func TestDarwinHostSampling(t *testing.T) {
	sampler := NewHostSampler()
	if usage, err := sampler.cpuUsage(); err != nil || usage != nil {
		t.Fatalf("first cpuUsage() = %v, %v; want nil, nil", usage, err)
	}
	time.Sleep(20 * time.Millisecond)
	if usage, err := sampler.cpuUsage(); err != nil || usage == nil {
		t.Fatalf("second cpuUsage() = %v, %v; want a usage value and nil error", usage, err)
	}
	used, total, err := hostMemory()
	if err != nil {
		t.Fatalf("hostMemory() error = %v", err)
	}
	if used < 0 || total <= 0 || used > total {
		t.Fatalf("hostMemory() = used %v total %v, want valid totals", used, total)
	}
}

func TestDarwinCPUTimes(t *testing.T) {
	total, idle := darwinCPUTimes(darwinCPULoadInfo{
		User:   10,
		System: 20,
		Idle:   60,
		Nice:   5,
	})
	if total != 95 || idle != 60 {
		t.Fatalf("darwinCPUTimes() = total %d idle %d, want 95 and 60", total, idle)
	}
}

func TestDarwinAvailableMemory(t *testing.T) {
	available := darwinAvailableMemory(darwinVMStatistics{
		Free:        100,
		Inactive:    200,
		Speculative: 25,
	}, 4096)
	if available != 325*4096 {
		t.Fatalf("darwinAvailableMemory() = %d, want %d", available, 325*4096)
	}
}
