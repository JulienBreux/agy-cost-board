package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// FS returns an fs.FS sub-rooted at the compiled "dist" folder.
func FS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
