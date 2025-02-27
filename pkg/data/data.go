package data

import "embed"

//go:embed components/*.yaml
var ComponentsFS embed.FS

//go:embed templates/*.yaml
var TemplatesFS embed.FS
