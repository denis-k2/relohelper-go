package uiassets

import (
	"embed"
	"io/fs"
)

//go:embed index.html app.js styles.css
var embeddedFiles embed.FS

func ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(embeddedFiles, name)
}
