package collector

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type testHostSampler struct {
	metrics  HostMetrics
	warnings []error
	path     string
}

func TestParseCPUStatExcludesGuestCounters(t *testing.T) {
	total, idle, err := parseCPUStat("cpu 100 20 30 400 10 5 5 5 500 500")
	if err != nil {
		t.Fatalf("parseCPUStat() error = %v", err)
	}
	if total != 575 || idle != 410 {
		t.Fatalf("parseCPUStat() = total %d idle %d, want total 575 idle 410", total, idle)
	}
}

func TestParseHostMemoryRequiresMemAvailable(t *testing.T) {
	for _, input := range []string{
		"MemTotal:       100 kB\n",
		"MemTotal:       100 kB\nMemAvailable: invalid kB\n",
	} {
		if _, _, err := parseHostMemory(strings.NewReader(input)); err == nil {
			t.Fatalf("parseHostMemory(%q) error = nil, want unavailable memory", input)
		}
	}
	used, total, err := parseHostMemory(strings.NewReader("MemTotal:       100 kB\nMemAvailable: 40 kB\n"))
	if err != nil || used != 60*1024 || total != 100*1024 {
		t.Fatalf("parseHostMemory() = used %v total %v err %v, want 61440 102400 nil", used, total, err)
	}
}

func (s *testHostSampler) Sample(path string) (HostMetrics, []error) {
	s.path = path
	return s.metrics, s.warnings
}

func TestHostCollectorUsesSharedNormalizedKeys(t *testing.T) {
	cpu, memoryUsed, memoryTotal := 0.75, 30.0, 100.0
	diskUsed, diskTotal := 40.0, 200.0
	sampler := &testHostSampler{metrics: HostMetrics{
		CPUUsage:         &cpu,
		MemoryUsedBytes:  &memoryUsed,
		MemoryTotalBytes: &memoryTotal,
		DiskUsedBytes:    &diskUsed,
		DiskTotalBytes:   &diskTotal,
	}}
	result, err := NewHostCollector("local", "/data/statlite.sqlite", sampler).Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if sampler.path != "/data/statlite.sqlite" {
		t.Fatalf("sampler path = %q, want configured SQLite path", sampler.path)
	}
	assertSample(t, result, "host_cpu_usage", MetricKindGauge, 0.75, "ratio")
	assertSample(t, result, "host_memory_used_bytes", MetricKindGauge, 30, "bytes")
	assertSample(t, result, "host_memory_total_bytes", MetricKindGauge, 100, "bytes")
	assertSample(t, result, "host_disk_used_bytes", MetricKindGauge, 40, "bytes")
	assertSample(t, result, "host_disk_total_bytes", MetricKindGauge, 200, "bytes")
	if result.HealthStatus != "UP" || len(result.Events) != 0 {
		t.Fatalf("result = %#v, want healthy result without warnings", result)
	}
}

func TestHostCollectorKeepsAvailableMetricsWhenSamplingIsPartial(t *testing.T) {
	diskUsed, diskTotal := 40.0, 200.0
	sampler := &testHostSampler{
		metrics:  HostMetrics{DiskUsedBytes: &diskUsed, DiskTotalBytes: &diskTotal},
		warnings: []error{errors.New("host CPU usage unavailable: first sample")},
	}
	result, err := NewHostCollector("local", "", sampler).Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	assertSample(t, result, "host_disk_used_bytes", MetricKindGauge, 40, "bytes")
	assertSample(t, result, "host_disk_total_bytes", MetricKindGauge, 200, "bytes")
	if len(result.Events) != 1 || result.Events[0].Type != "host_metric_unavailable" {
		t.Fatalf("Events = %#v, want one host availability warning", result.Events)
	}
}
