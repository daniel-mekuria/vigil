package storage

import (
	"testing"
	"time"

	"github.com/pvrlabs/statlite/internal/collector"
)

func TestBuildSeriesPointDerivesHostPercentagesOnlyFromValidPairs(t *testing.T) {
	poll := &pollSamples{timestamp: time.Now().UTC(), samples: map[string]sampleValue{
		"host_memory_used_bytes":  {kind: collector.MetricKindGauge, value: 30},
		"host_memory_total_bytes": {kind: collector.MetricKindGauge, value: 100},
		"host_disk_used_bytes":    {kind: collector.MetricKindGauge, value: 40},
		"host_disk_total_bytes":   {kind: collector.MetricKindGauge, value: 200},
	}}
	point, _ := buildSeriesPoint(poll, map[string]counterValue{}, time.Time{})
	assertFloatPtr(t, "host_memory_usage", point.HostMemoryUsage, 0.3)
	assertFloatPtr(t, "host_disk_usage", point.HostDiskUsage, 0.2)

	poll.samples["host_memory_used_bytes"] = sampleValue{kind: collector.MetricKindGauge, value: 101}
	poll.samples["host_disk_total_bytes"] = sampleValue{kind: collector.MetricKindGauge, value: 0}
	point, _ = buildSeriesPoint(poll, map[string]counterValue{}, time.Time{})
	if point.HostMemoryUsage != nil || point.HostDiskUsage != nil {
		t.Fatalf("usage = memory %v disk %v, want nil for invalid pairs", point.HostMemoryUsage, point.HostDiskUsage)
	}
}

func TestSetCurrentHostDiskUsesOnlyTheLatestPoint(t *testing.T) {
	series := &Series{Points: []SeriesPoint{
		{HostDiskUsedBytes: floatPtr(40), HostDiskTotalBytes: floatPtr(100), HostDiskUsage: floatPtr(0.4)},
		{HostMemoryUsedBytes: floatPtr(10)},
	}}
	setCurrentHostDisk(series)
	if series.CurrentHostDisk != nil {
		t.Fatalf("current host disk = %#v, want nil when latest point omits disk", series.CurrentHostDisk)
	}

	series.Points = series.Points[:1]
	setCurrentHostDisk(series)
	if series.CurrentHostDisk == nil || series.CurrentHostDisk.UsedBytes != 40 || series.CurrentHostDisk.TotalBytes != 100 || series.CurrentHostDisk.Usage != 0.4 {
		t.Fatalf("current host disk = %#v, want latest complete observation", series.CurrentHostDisk)
	}
}

func floatPtr(value float64) *float64 { return &value }
