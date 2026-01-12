package data

import "embed"

//go:embed components templates/*.yaml
var EmbeddedFS embed.FS
