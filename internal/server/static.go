package server

// This file serves embedded static assets used by the dashboard.

import (
	"net/http"

	"github.com/pvrlabs/statlite/internal/dashboard"
)

func (s *Server) handleDashboardScript(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != dashboard.ScriptPath() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write([]byte(dashboard.Script))
}

func (s *Server) handleStatliteIcon(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/static/statlite-icon.png" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.Write(dashboard.StatliteIconPNG)
}

func (s *Server) handleDashboardVendor(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case dashboard.ChartJSPath:
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(dashboard.ChartJS)
	case dashboard.OrbitronFontPath:
		w.Header().Set("Content-Type", "font/woff2")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(dashboard.OrbitronFontWoff2)
	default:
		http.NotFound(w, r)
	}
}
