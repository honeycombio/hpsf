package tmpl

import (
	"strings"
	"testing"
)

func TestCollectorConfig_RenderYAML(t *testing.T) {
	cc := NewCollectorConfig()
	cc.Set("receivers", "otlp.Port", "4317")
	cc.Set("receivers", "otlp.Endpoint", "localhost")
	cc.Set("service", "pipelines.traces.receivers", []string{"otlp"})
	// NOTE: this "want" string is indented with spaces, not tabs; the YAML renderer uses spaces.
	want := `
receivers:
    otlp:
        endpoint: localhost
        port: 4317
service:
    pipelines:
        traces:
            receivers: [otlp]
            processors: []
            exporters: []
`
	got, err := cc.RenderYAML()
	if err != nil {
		t.Errorf("CollectorConfig.RenderYAML() error = %v, expected nil", err)
		return
	}
	if strings.TrimSpace(string(got)) != strings.TrimSpace(want) {
		t.Errorf("CollectorConfig.RenderYAML() got = \n%s, want \n%v", got, want)
	}
}
