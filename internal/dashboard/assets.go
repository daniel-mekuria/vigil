package dashboard

// This file embeds the dashboard HTML and static image assets.

import (
	"crypto/sha256"
	"fmt"

	_ "embed"
)

//go:embed static/index.html
var Page string

//go:embed static/dashboard.js
var Script string

// ScriptPath is content-addressed so a binary upgrade can safely use immutable
// browser caching without leaving the dashboard paired with an old script.
func ScriptPath() string {
	sum := sha256.Sum256([]byte(Script))
	return fmt.Sprintf("/static/dashboard.%x.js", sum[:8])
}

//go:embed static/statlite-icon.png
var StatliteIconPNG []byte
