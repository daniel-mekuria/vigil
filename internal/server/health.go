package server

// This file exposes lightweight StatLite readiness.

import (
	"encoding/json"
	"net/http"
	"syscall"
	"time"

	"github.com/pvrlabs/statlite/internal/version"
)

type HealthResponse struct {
	Status    string                `json:"status"`
	Version   string                `json:"version"`
	Timestamp string                `json:"timestamp"`
	Storage   StatliteStorageHealth `json:"storage"`
}

type StatliteStorageHealth struct {
	Status string `json:"status"`
}

// handleHealthz reports StatLite process health.
//
// Semantics:
//   - Top-level status/HTTP code reflect process readiness, not monitored-target health.
//   - SQLite storage check failure sets status to "error" and returns HTTP 503.
//   - When no monitor/manager is configured, storage is reported as "unavailable" and the process stays healthy.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	now := time.Now().UTC()

	storageStatus := "unavailable"
	if s.manager != nil {
		target := s.selectedTarget(r)
		if target.Monitor.StorageHealthy(r.Context()) {
			storageStatus = "ok"
		} else {
			storageStatus = "error"
		}
	}

	// Process health is independent of monitored-target poll success/failure.
	// A failing self-monitor first poll (server not ready yet) must not make
	// StatLite report itself as permanently unhealthy.
	processStatus := "ok"
	httpCode := http.StatusOK
	if storageStatus == "error" {
		processStatus = "error"
		httpCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    processStatus,
		Version:   version.Version,
		Timestamp: now.Format(time.RFC3339),
		Storage:   StatliteStorageHealth{Status: storageStatus},
	})
}

func (s *Server) processCPUUsage(now time.Time) float64 {
	cpuSeconds, err := processCPUSeconds()
	if err != nil {
		return 0
	}

	s.cpuMu.Lock()
	defer s.cpuMu.Unlock()

	if s.lastCPUAt.IsZero() {
		s.lastCPUAt = now
		s.lastCPUSeconds = cpuSeconds
		return 0
	}

	elapsed := now.Sub(s.lastCPUAt).Seconds()
	cpuDelta := cpuSeconds - s.lastCPUSeconds
	s.lastCPUAt = now
	s.lastCPUSeconds = cpuSeconds
	if elapsed <= 0 || cpuDelta < 0 {
		return s.lastCPUUsage
	}

	usage := cpuDelta / elapsed
	if usage < 0 {
		usage = 0
	}
	s.lastCPUUsage = usage
	return usage
}

func processCPUSeconds() (float64, error) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, err
	}
	return timevalSeconds(usage.Utime) + timevalSeconds(usage.Stime), nil
}

func timevalSeconds(value syscall.Timeval) float64 {
	return float64(value.Sec) + float64(value.Usec)/1_000_000
}
