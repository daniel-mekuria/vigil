package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestStatliteMetricsClientFetchesCompleteV1Response(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/statlite/metrics" {
			t.Errorf("path = %q, want /statlite/metrics", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"schema": "statlite-metrics/v1",
			"status": "UP",
			"started_at": "2026-07-27T19:00:00Z",
			"metrics": {
				"requests_total": 1420,
				"responses_404_total": 18,
				"responses_4xx_total": 31,
				"responses_5xx_total": 4,
				"request_duration_seconds_total": 84.31,
				"request_duration_seconds_max": 1.42,
				"process_cpu_usage": 0.031,
				"host_cpu_usage": 0.4,
				"host_memory_used_bytes": 30,
				"host_memory_total_bytes": 100,
				"host_disk_used_bytes": 40,
				"host_disk_total_bytes": 200,
				"runtime_heap_used_bytes": 25165824,
				"uptime_seconds": 1820
			}
		}`))
	}))
	defer server.Close()

	client, err := NewStatliteMetricsClient(server.URL+"/statlite/metrics", 3*time.Second)
	if err != nil {
		t.Fatalf("NewStatliteMetricsClient() error = %v", err)
	}

	response, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("requests = %d, want 1", requests.Load())
	}
	if response.Schema != StatliteMetricsV1Schema {
		t.Fatalf("Schema = %q, want %q", response.Schema, StatliteMetricsV1Schema)
	}
	if response.Status != "UP" {
		t.Fatalf("Status = %q, want UP", response.Status)
	}
	if !response.StartedAt.present() {
		t.Fatal("StartedAt is absent, want present")
	}
	if response.Metrics == nil {
		t.Fatal("Metrics = nil, want complete metrics")
	}
	for name, field := range map[string]StatliteMetricsField{
		"requests_total":                 response.Metrics.RequestsTotal,
		"responses_404_total":            response.Metrics.Responses404Total,
		"responses_4xx_total":            response.Metrics.Responses4xxTotal,
		"responses_5xx_total":            response.Metrics.Responses5xxTotal,
		"request_duration_seconds_total": response.Metrics.RequestDurationSecondsTotal,
		"request_duration_seconds_max":   response.Metrics.RequestDurationSecondsMax,
		"process_cpu_usage":              response.Metrics.ProcessCPUUsage,
		"host_cpu_usage":                 response.Metrics.HostCPUUsage,
		"host_memory_used_bytes":         response.Metrics.HostMemoryUsedBytes,
		"host_memory_total_bytes":        response.Metrics.HostMemoryTotalBytes,
		"host_disk_used_bytes":           response.Metrics.HostDiskUsedBytes,
		"host_disk_total_bytes":          response.Metrics.HostDiskTotalBytes,
		"runtime_heap_used_bytes":        response.Metrics.RuntimeHeapUsedBytes,
		"uptime_seconds":                 response.Metrics.UptimeSeconds,
	} {
		if !field.present() {
			t.Errorf("%s is absent, want present", name)
		}
	}
	if len(response.Raw) == 0 {
		t.Fatal("Raw is empty, want response body")
	}
}

func TestStatliteMetricsClientAcceptsMinimalResponse(t *testing.T) {
	server := newStatliteMetricsServer(t, http.StatusOK, `{
		"schema": "statlite-metrics/v1",
		"status": "UP"
	}`)
	defer server.Close()

	client := newTestStatliteMetricsClient(t, server.URL)
	response, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if response.StartedAt.present() {
		t.Fatal("StartedAt is present, want absent")
	}
	if response.Metrics != nil {
		t.Fatalf("Metrics = %#v, want nil", response.Metrics)
	}
}

func TestStatliteMetricsClientIgnoresUnknownFields(t *testing.T) {
	server := newStatliteMetricsServer(t, http.StatusOK, `{
		"schema": "statlite-metrics/v1",
		"status": "UP",
		"future_top_level": {"enabled": true},
		"metrics": {
			"requests_total": 12,
			"future_metric": 99
		}
	}`)
	defer server.Close()

	client := newTestStatliteMetricsClient(t, server.URL)
	response, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if response.Metrics == nil || !response.Metrics.RequestsTotal.present() {
		t.Fatalf("Metrics = %#v, want requests_total", response.Metrics)
	}
}

func TestStatliteMetricsClientPreservesOptionalFieldRepresentations(t *testing.T) {
	server := newStatliteMetricsServer(t, http.StatusOK, `{
		"schema": "statlite-metrics/v1",
		"status": "UP",
		"started_at": 123,
		"metrics": {
			"requests_total": -1,
			"process_cpu_usage": 1e999,
			"runtime_heap_used_bytes": "invalid"
		}
	}`)
	defer server.Close()

	client := newTestStatliteMetricsClient(t, server.URL)
	response, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if got := string(response.StartedAt.raw); got != "123" {
		t.Fatalf("StartedAt raw = %q, want 123", got)
	}
	if got := string(response.Metrics.RequestsTotal.raw); got != "-1" {
		t.Fatalf("RequestsTotal raw = %q, want -1", got)
	}
	if got := string(response.Metrics.ProcessCPUUsage.raw); got != "1e999" {
		t.Fatalf("ProcessCPUUsage raw = %q, want 1e999", got)
	}
	if got := string(response.Metrics.RuntimeHeapUsedBytes.raw); got != `"invalid"` {
		t.Fatalf("RuntimeHeapUsedBytes raw = %q, want string", got)
	}
}

func TestStatliteMetricsClientRejectsMalformedJSON(t *testing.T) {
	for _, body := range []string{
		`{"schema":"statlite-metrics/v1","status":`,
		`{"schema":"statlite-metrics/v1","status":"UP","metrics":{"cpu_usage":NaN}}`,
	} {
		t.Run(body, func(t *testing.T) {
			server := newStatliteMetricsServer(t, http.StatusOK, body)
			defer server.Close()

			client := newTestStatliteMetricsClient(t, server.URL)
			_, err := client.Fetch(context.Background())
			if err == nil {
				t.Fatal("Fetch() error = nil, want malformed JSON error")
			}
			if !strings.Contains(err.Error(), "parsing statlite metrics response") {
				t.Fatalf("Fetch() error = %q, want parsing context", err)
			}
		})
	}
}

func TestStatliteMetricsClientRejectsNon2xxResponse(t *testing.T) {
	server := newStatliteMetricsServer(t, http.StatusServiceUnavailable, `unavailable`)
	defer server.Close()

	client := newTestStatliteMetricsClient(t, server.URL)
	_, err := client.Fetch(context.Background())
	if err == nil {
		t.Fatal("Fetch() error = nil, want HTTP error")
	}
	if !strings.Contains(err.Error(), "HTTP 503: unavailable") {
		t.Fatalf("Fetch() error = %q, want HTTP status and body", err)
	}
}

func TestStatliteMetricsClientResponseSizeLimit(t *testing.T) {
	const prefix = `{"schema":"statlite-metrics/v1","status":"UP","padding":"`
	const suffix = `"}`
	exact := prefix + strings.Repeat("x", statliteMetricsMaxResponseBytes-len(prefix)-len(suffix)) + suffix

	t.Run("exact limit accepted", func(t *testing.T) {
		server := newStatliteMetricsServer(t, http.StatusOK, exact)
		defer server.Close()

		client := newTestStatliteMetricsClient(t, server.URL)
		if _, err := client.Fetch(context.Background()); err != nil {
			t.Fatalf("Fetch() error = %v, want exact limit accepted", err)
		}
	})

	t.Run("over limit rejected", func(t *testing.T) {
		server := newStatliteMetricsServer(t, http.StatusOK, exact+" ")
		defer server.Close()

		client := newTestStatliteMetricsClient(t, server.URL)
		_, err := client.Fetch(context.Background())
		if err == nil {
			t.Fatal("Fetch() error = nil, want size limit error")
		}
		if !strings.Contains(err.Error(), "exceeds 1048576 byte limit") {
			t.Fatalf("Fetch() error = %q, want size limit context", err)
		}
	})
}

func TestStatliteMetricsClientValidatesRequiredTopLevelData(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "missing schema", body: `{"status":"UP"}`, want: "missing required schema"},
		{name: "unsupported schema", body: `{"schema":"statlite-metrics/v2","status":"UP"}`, want: `unsupported schema "statlite-metrics/v2"`},
		{name: "missing status", body: `{"schema":"statlite-metrics/v1"}`, want: "missing required status"},
		{name: "blank status", body: `{"schema":"statlite-metrics/v1","status":"  "}`, want: "missing required status"},
		{name: "null document", body: `null`, want: "missing required schema"},
		{name: "array document", body: `[]`, want: "parsing statlite metrics response"},
		{name: "invalid metrics shape", body: `{"schema":"statlite-metrics/v1","status":"UP","metrics":[]}`, want: "parsing statlite metrics response"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newStatliteMetricsServer(t, http.StatusOK, tt.body)
			defer server.Close()

			client := newTestStatliteMetricsClient(t, server.URL)
			_, err := client.Fetch(context.Background())
			if err == nil {
				t.Fatal("Fetch() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Fetch() error = %q, want %q", err, tt.want)
			}
		})
	}
}

func TestNewStatliteMetricsClientValidatesConfiguration(t *testing.T) {
	const timeout = 7 * time.Second
	client, err := NewStatliteMetricsClient("  https://example.com/statlite/metrics  ", timeout)
	if err != nil {
		t.Fatalf("NewStatliteMetricsClient() error = %v", err)
	}
	if client.url != "https://example.com/statlite/metrics" {
		t.Fatalf("url = %q, want trimmed URL", client.url)
	}
	if client.httpClient.Timeout != timeout {
		t.Fatalf("timeout = %s, want %s", client.httpClient.Timeout, timeout)
	}

	tests := []struct {
		name    string
		rawURL  string
		timeout time.Duration
		want    string
	}{
		{name: "zero timeout", rawURL: "https://example.com/metrics", timeout: 0, want: "timeout must be positive"},
		{name: "negative timeout", rawURL: "https://example.com/metrics", timeout: -time.Second, want: "timeout must be positive"},
		{name: "unsupported scheme", rawURL: "ftp://example.com/metrics", timeout: time.Second, want: "must use http or https"},
		{name: "missing host", rawURL: "/statlite/metrics", timeout: time.Second, want: "must use http or https"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStatliteMetricsClient(tt.rawURL, tt.timeout)
			if err == nil {
				t.Fatal("NewStatliteMetricsClient() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("NewStatliteMetricsClient() error = %q, want %q", err, tt.want)
			}
		})
	}
}

func newTestStatliteMetricsClient(t *testing.T, rawURL string) *StatliteMetricsClient {
	t.Helper()
	client, err := NewStatliteMetricsClient(rawURL, time.Second)
	if err != nil {
		t.Fatalf("NewStatliteMetricsClient() error = %v", err)
	}
	return client
}

func newStatliteMetricsServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}
