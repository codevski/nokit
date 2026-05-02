// Package assets embeds the built frontend into the binary.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// FS returns the embedded frontend filesystem rooted at dist/.
func FS() (fs.FS, error) {
	return fs.Sub(dist, "dist")
}
