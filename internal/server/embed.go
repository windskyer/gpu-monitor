package server

import (
	"embed"
	"io/fs"
)

// dist/ is populated by `cd web && npm run build` (Vite outDir → ../internal/server/dist)
//
//go:embed dist
var embeddedWeb embed.FS

// staticFS serves files rooted at dist/, so URL /index.html → dist/index.html.
var staticFS, _ = fs.Sub(embeddedWeb, "dist")
