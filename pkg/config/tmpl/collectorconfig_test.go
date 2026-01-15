package tmpl

import (
	"strings"
	"testing"
)

func TestCollectorConfig_RenderYAML(t *testing.T) {
	cc := NewCollectorConfig()
	cc.Set("receivers", "otlp.port", "4317")
	cc.Set("receivers", "otlp.endpoint", "localhost")
	cc.Set("service", "pipelines.traces.receivers", []string{"otlp"})
	cc.Set("service", "pipelines.traces.processors", []string{})
	// NOTE: this "want" string is indented with spaces, not tabs; the YAML renderer uses spaces.
	want := `
receivers:
    otlp:
        endpoint: localhost
        port: "4317"
processors:
    usage: {}
extensions:
    honeycomb: {}
service:
    extensions: [honeycomb]
    pipelines:
        traces:
            receivers: [otlp]
            processors: [usage]
            exporters: []
`
	got, err := cc.RenderYAML()
	if err != nil {
		t.Errorf("CollectorConfig.RenderYAML() error = %v, expected nil", err)
		return
	}
	x := strings.TrimSpace(string(got))
	if x != strings.TrimSpace(want) {
		t.Errorf("CollectorConfig.RenderYAML() got = \n%s, want \n%v", got, want)
	}
}

func TestCollectorConfig_ProcessorOrdering(t *testing.T) {
	cc := NewCollectorConfig()
	cc.Set("receivers", "otlp.port", "4317")
	cc.Set("service", "pipelines.traces.receivers", []string{"otlp"})
	cc.Set("service", "pipelines.traces.processors", []string{"memory_limiter/otlp", "custom/foo"})
	// NOTE: this "want" string is indented with spaces, not tabs; the YAML renderer uses spaces.
	want := `
receivers:
    otlp:
        port: "4317"
processors:
    usage: {}
extensions:
    honeycomb: {}
service:
    extensions: [honeycomb]
    pipelines:
        traces:
            receivers: [otlp]
            processors: [memory_limiter/otlp, usage, custom/foo]
            exporters: []
`
	got, err := cc.RenderYAML()
	if err != nil {
		t.Errorf("CollectorConfig.RenderYAML() error = %v, expected nil", err)
		return
	}
	x := strings.TrimSpace(string(got))
	if x != strings.TrimSpace(want) {
		t.Errorf("CollectorConfig.RenderYAML() got = \n%s, want \n%v", got, want)
	}
}
