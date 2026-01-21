package translator

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectSharedReceiversFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *tmpl.CollectorConfig
		expected map[string]map[string][]string
	}{
		{
			name: "single receiver shared across multiple signal types",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":    []string{"otlp/receiver1"},
						"pipelines.metrics/def-456.receivers": []string{"otlp/receiver1"},
						"pipelines.traces/ghi-789.receivers":  []string{"otlp/receiver1"},
					},
				},
			},
			expected: map[string]map[string][]string{
				"otlp/receiver1": {
					"logs":    {"logs/abc-123"},
					"metrics": {"metrics/def-456"},
					"traces":  {"traces/ghi-789"},
				},
			},
		},
		{
			name: "unique receivers per pipeline - no sharing",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":    []string{"otlp/receiver1"},
						"pipelines.metrics/def-456.receivers": []string{"otlp/receiver2"},
						"pipelines.traces/ghi-789.receivers":  []string{"otlp/receiver3"},
					},
				},
			},
			expected: map[string]map[string][]string{}, // no shared receivers
		},
		{
			name: "one receiver shared, one unique",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":    []string{"otlp/shared"},
						"pipelines.metrics/def-456.receivers": []string{"otlp/shared"},
						"pipelines.traces/ghi-789.receivers":  []string{"otlp/unique"},
					},
				},
			},
			expected: map[string]map[string][]string{
				"otlp/shared": {
					"logs":    {"logs/abc-123"},
					"metrics": {"metrics/def-456"},
				},
			},
		},
		{
			name: "receiver shared within same signal type",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers": []string{"otlp/receiver1"},
						"pipelines.logs/def-456.receivers": []string{"otlp/receiver1"},
					},
				},
			},
			expected: map[string]map[string][]string{
				"otlp/receiver1": {
					// Order doesn't matter for this test, just that both are present
					"logs": {"logs/def-456", "logs/abc-123"},
				},
			},
		},
		{
			name: "empty config",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {},
				},
			},
			expected: map[string]map[string][]string{},
		},
		{
			name: "no service section",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Translator{}
			result := tr.detectSharedReceiversFromConfig(tt.config)

			// Check structure matches (keys and nested keys), but ignore slice order
			assert.Equal(t, len(tt.expected), len(result), "number of shared receivers")
			for receiver, expectedSignals := range tt.expected {
				assert.Contains(t, result, receiver, "receiver should exist")
				actualSignals := result[receiver]
				assert.Equal(t, len(expectedSignals), len(actualSignals), "number of signal types")
				for signal, expectedPipelines := range expectedSignals {
					assert.Contains(t, actualSignals, signal, "signal type should exist")
					assert.ElementsMatch(t, expectedPipelines, actualSignals[signal], "pipelines match regardless of order")
				}
			}
		})
	}
}

func TestGenerateIngressPipelines(t *testing.T) {
	tests := []struct {
		name            string
		sharedReceivers map[string]map[string][]string
		verify          func(t *testing.T, cfg *tmpl.CollectorConfig)
	}{
		{
			name: "creates ingress pipelines for shared receiver",
			sharedReceivers: map[string]map[string][]string{
				"otlp/receiver1": {
					"logs":    {"logs/abc-123"},
					"metrics": {"metrics/def-456"},
				},
			},
			verify: func(t *testing.T, cfg *tmpl.CollectorConfig) {
				service := cfg.Sections["service"]
				connectors := cfg.Sections["connectors"]

				// Verify forward connector created for this receiver
				assert.NotNil(t, connectors)
				assert.Contains(t, connectors, "forward/otlp/receiver1")

				// Verify ingress pipelines created
				assert.Contains(t, service, "pipelines.logs/ingress_otlp/receiver1.receivers")
				assert.Equal(t, []string{"otlp/receiver1"}, service["pipelines.logs/ingress_otlp/receiver1.receivers"])
				assert.Contains(t, service, "pipelines.logs/ingress_otlp/receiver1.processors")
				assert.Equal(t, []string{"memory_limiter/receiver1", "usage"}, service["pipelines.logs/ingress_otlp/receiver1.processors"])
				assert.Contains(t, service, "pipelines.logs/ingress_otlp/receiver1.exporters")
				assert.Equal(t, []string{"forward/otlp/receiver1"}, service["pipelines.logs/ingress_otlp/receiver1.exporters"])

				assert.Contains(t, service, "pipelines.metrics/ingress_otlp/receiver1.receivers")
				assert.Equal(t, []string{"otlp/receiver1"}, service["pipelines.metrics/ingress_otlp/receiver1.receivers"])

				// Verify downstream pipelines rewired to use receiver-specific connector
				assert.Contains(t, service, "pipelines.logs/abc-123.receivers")
				assert.Equal(t, []string{"forward/otlp/receiver1"}, service["pipelines.logs/abc-123.receivers"])
				assert.Contains(t, service, "pipelines.metrics/def-456.receivers")
				assert.Equal(t, []string{"forward/otlp/receiver1"}, service["pipelines.metrics/def-456.receivers"])
			},
		},
		{
			name: "handles multiple shared receivers",
			sharedReceivers: map[string]map[string][]string{
				"otlp/receiver1": {
					"logs": {"logs/abc-123"},
				},
				"otlp/receiver2": {
					"traces": {"traces/def-456"},
				},
			},
			verify: func(t *testing.T, cfg *tmpl.CollectorConfig) {
				service := cfg.Sections["service"]
				connectors := cfg.Sections["connectors"]

				// Both ingress pipelines should exist
				assert.Contains(t, service, "pipelines.logs/ingress_otlp/receiver1.receivers")
				assert.Contains(t, service, "pipelines.traces/ingress_otlp/receiver2.receivers")

				// Each receiver should have its own connector
				assert.Contains(t, connectors, "forward/otlp/receiver1")
				assert.Contains(t, connectors, "forward/otlp/receiver2")

				// Each downstream pipeline uses its receiver's specific connector
				assert.Equal(t, []string{"forward/otlp/receiver1"}, service["pipelines.logs/abc-123.receivers"])
				assert.Equal(t, []string{"forward/otlp/receiver2"}, service["pipelines.traces/def-456.receivers"])
			},
		},
		{
			name:            "empty shared receivers - no changes",
			sharedReceivers: map[string]map[string][]string{},
			verify: func(t *testing.T, cfg *tmpl.CollectorConfig) {
				service := cfg.Sections["service"]

				// No ingress pipelines created
				for key := range service {
					assert.NotContains(t, key, "ingress")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tmpl.NewCollectorConfig()
			// Add some existing pipeline data for rewiring tests
			for receiverMap := range tt.sharedReceivers {
				for _, pipelines := range tt.sharedReceivers[receiverMap] {
					for _, pipeline := range pipelines {
						cfg.Set("service", pipeline+".receivers", []string{receiverMap})
					}
				}
			}

			tr := &Translator{}
			err := tr.generateIngressPipelines(cfg, tt.sharedReceivers)
			require.NoError(t, err)

			tt.verify(t, cfg)
		})
	}
}

func TestTransformToIngressPipelines_Integration(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func() *tmpl.CollectorConfig
		expectIngress  bool
		verifyPipeline string
	}{
		{
			name: "transforms config with shared receiver",
			setupConfig: func() *tmpl.CollectorConfig {
				cfg := tmpl.NewCollectorConfig()
				cfg.Set("service", "pipelines.logs/abc-123.receivers", []string{"otlp/receiver1"})
				cfg.Set("service", "pipelines.metrics/def-456.receivers", []string{"otlp/receiver1"})
				cfg.Set("service", "pipelines.traces/ghi-789.receivers", []string{"otlp/receiver1"})
				return cfg
			},
			expectIngress:  true,
			verifyPipeline: "pipelines.logs/ingress_otlp/receiver1.receivers",
		},
		{
			name: "does not transform config with unique receivers",
			setupConfig: func() *tmpl.CollectorConfig {
				cfg := tmpl.NewCollectorConfig()
				cfg.Set("service", "pipelines.logs/abc-123.receivers", []string{"otlp/receiver1"})
				cfg.Set("service", "pipelines.metrics/def-456.receivers", []string{"otlp/receiver2"})
				return cfg
			},
			expectIngress:  false,
			verifyPipeline: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			tr := &Translator{}

			result, err := tr.transformToIngressPipelines(cfg, nil, nil)
			require.NoError(t, err)

			resultCfg := result.(*tmpl.CollectorConfig)

			if tt.expectIngress {
				// Verify ingress pipeline exists
				assert.Contains(t, resultCfg.Sections["service"], tt.verifyPipeline)
				// Verify forward connector exists for this receiver
				assert.Contains(t, resultCfg.Sections["connectors"], "forward/otlp/receiver1")
			} else {
				// Verify no ingress pipelines
				for key := range resultCfg.Sections["service"] {
					assert.NotContains(t, key, "ingress")
				}
			}
		})
	}
}
