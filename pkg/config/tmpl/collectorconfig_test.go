package tmpl

import (
	"strings"
	"testing"
)

func TestCollectorConfig_RenderYAML(t *testing.T) {
	cc := NewCollectorConfig()
	cc.Set("receivers", "otlp.port", "4317")
	cc.Set("receivers", "otlp.endpoint", "localhost")
	cc.Set("exporters", "otlphttp.endpoint", "http://localhost:4318")
	cc.Set("service", "pipelines.traces.receivers", []string{"otlp"})
	cc.Set("service", "pipelines.traces.processors", []string{})
	cc.Set("service", "pipelines.traces.exporters", []string{"otlphttp"})
	// NOTE: this "want" string is indented with spaces, not tabs; the YAML renderer uses spaces.
	want := `
receivers:
    otlp:
        endpoint: localhost
        port: "4317"
processors:
    usage: {}
exporters:
    otlphttp:
        endpoint: http://localhost:4318
extensions:
    honeycomb: {}
service:
    extensions: [honeycomb]
    pipelines:
        traces:
            receivers: [otlp]
            processors: [usage]
            exporters: [otlphttp]
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
	cc.Set("exporters", "otlphttp.endpoint", "http://localhost:4318")
	cc.Set("service", "pipelines.traces.receivers", []string{"otlp"})
	cc.Set("service", "pipelines.traces.processors", []string{"memory_limiter/otlp", "custom/foo"})
	cc.Set("service", "pipelines.traces.exporters", []string{"otlphttp"})
	// NOTE: this "want" string is indented with spaces, not tabs; the YAML renderer uses spaces.
	want := `
receivers:
    otlp:
        port: "4317"
processors:
    usage: {}
exporters:
    otlphttp:
        endpoint: http://localhost:4318
extensions:
    honeycomb: {}
service:
    extensions: [honeycomb]
    pipelines:
        traces:
            receivers: [otlp]
            processors: [memory_limiter/otlp, usage, custom/foo]
            exporters: [otlphttp]
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

func TestCollectorConfig_ValidatePipelinesWithMultipleErrors(t *testing.T) {
	cc := NewCollectorConfig()
	cc.Set("receivers", "otlp.port", "4317")
	cc.Set("exporters", "otlphttp.endpoint", "http://localhost:4318")
	// Create multiple pipelines with missing receivers and exporters
	cc.Set("service", "pipelines.traces.processors", []string{"custom/foo"})
	cc.Set("service", "pipelines.metrics.receivers", []string{"otlp"})
	cc.Set("service", "pipelines.logs.processors", []string{"custom/bar"})

	_, err := cc.RenderYAML()
	if err == nil {
		t.Error("CollectorConfig.RenderYAML() expected error, got nil")
		return
	}

	// Verify that the error contains messages for all invalid pipelines
	errStr := err.Error()
	expectedErrors := []string{
		"pipeline 'traces' must have at least one receiver",
		"pipeline 'traces' must have at least one exporter",
		"pipeline 'metrics' must have at least one exporter",
		"pipeline 'logs' must have at least one receiver",
		"pipeline 'logs' must have at least one exporter",
	}

	for _, expectedErr := range expectedErrors {
		if !strings.Contains(errStr, expectedErr) {
			t.Errorf("Expected error to contain %q, but got: %s", expectedErr, errStr)
		}
	}
}
