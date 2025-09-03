package tmpl

// TemplateConfig is an interface for a configuration abstraction that can be rendered as a map or as YAML.
type TemplateConfig interface {
	RenderToMap(m map[string]any) map[string]any
	RenderYAML() ([]byte, error)
	Merge(other TemplateConfig) error
}
