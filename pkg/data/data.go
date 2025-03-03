package data

import "embed"

//go:embed components/*.yaml templates/*.yaml
var EmbeddedFS embed.FS
