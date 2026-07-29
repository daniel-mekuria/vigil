package collector

// This file collects lightweight, best-effort local host resource estimates.

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
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

type hostMetricsSampler interface {
	Sample(filesystemPath string) (HostMetrics, []error)
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

func (s *HostSampler) cpuUsage() (*float64, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}
	total, idle, err := parseCPUStat(strings.SplitN(string(data), "\n", 2)[0])
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.haveCPU {
		s.lastTotal, s.lastIdle, s.haveCPU = total, idle, true
		return nil, nil
	}
	totalDelta, idleDelta := total-s.lastTotal, idle-s.lastIdle
	s.lastTotal, s.lastIdle = total, idle
	if totalDelta == 0 || idleDelta > totalDelta {
		return nil, fmt.Errorf("invalid CPU interval")
	}
	usage := float64(totalDelta-idleDelta) / float64(totalDelta)
	return &usage, nil
}

func parseCPUStat(line string) (uint64, uint64, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, fmt.Errorf("missing aggregate CPU line")
	}
	var total, idle uint64
	for index, field := range fields[1:] {
		if index >= 8 { // guest and guest_nice are already included in user/nice.
			break
		}
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse CPU field %d: %w", index, err)
		}
		total += value
		if index == 3 || index == 4 { // idle and iowait
			idle += value
		}
	}
	return total, idle, nil
}

func hostMemory() (float64, float64, error) {
	if runtime.GOOS != "linux" {
		return 0, 0, fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	return parseHostMemory(file)
}

func parseHostMemory(reader io.Reader) (float64, float64, error) {
	values := make(map[string]uint64)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		values[strings.TrimSuffix(fields[0], ":")] = value * 1024
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}
	total, totalPresent := values["MemTotal"]
	available, availablePresent := values["MemAvailable"]
	if !totalPresent || !availablePresent || total == 0 || available > total {
		return 0, 0, fmt.Errorf("invalid memory totals")
	}
	return float64(total - available), float64(total), nil
}

type HostCollector struct {
	targetName     string
	filesystemPath string
	sampler        hostMetricsSampler
}

func NewHostCollector(targetName, filesystemPath string, sampler hostMetricsSampler) *HostCollector {
	if sampler == nil {
		sampler = NewHostSampler()
	}
	return &HostCollector{targetName: targetName, filesystemPath: filesystemPath, sampler: sampler}
}

func (c *HostCollector) Collect(_ context.Context) (*CollectionResult, error) {
	now := time.Now().UTC()
	result := &CollectionResult{TargetName: c.targetName, PollStartedAt: now, PollFinishedAt: now, HealthStatus: "UP"}
	metrics, warnings := c.sampler.Sample(c.filesystemPath)
	for _, warning := range warnings {
		result.addEvent(EventSeverityWarning, "host_metric_unavailable", "", warning.Error())
	}
	addHostMetricSample(result, "host_cpu_usage", metrics.CPUUsage, "ratio")
	addHostMetricSample(result, "host_memory_used_bytes", metrics.MemoryUsedBytes, "bytes")
	addHostMetricSample(result, "host_memory_total_bytes", metrics.MemoryTotalBytes, "bytes")
	addHostMetricSample(result, "host_disk_used_bytes", metrics.DiskUsedBytes, "bytes")
	addHostMetricSample(result, "host_disk_total_bytes", metrics.DiskTotalBytes, "bytes")
	return result, nil
}

func addHostMetricSample(result *CollectionResult, key string, value *float64, unit string) {
	if value != nil {
		result.addSample(key, MetricKindGauge, *value, unit)
	}
}
