package tmpl

type TemplateConfig interface {
	Render() map[string]any
	RenderYAML() ([]byte, string, error)
	Merge(other TemplateConfig) TemplateConfig
}
