package translator

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectDuplicateDownstreams(t *testing.T) {
	tests := []struct {
		name     string
		config   *tmpl.CollectorConfig
		expected map[string][]string // signature -> []pipelineKeys
	}{
		{
			name: "detects two pipelines with same processors and exporters",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":  []string{"forward/otlp/receiver1"},
						"pipelines.logs/abc-123.processors": []string{"filter/MyFilter"},
						"pipelines.logs/abc-123.exporters":  []string{"honeycomb/MyExporter"},
						"pipelines.logs/def-456.receivers":  []string{"forward/otlp/receiver2"},
						"pipelines.logs/def-456.processors": []string{"filter/MyFilter"},
						"pipelines.logs/def-456.exporters":  []string{"honeycomb/MyExporter"},
					},
				},
			},
			expected: map[string][]string{
				"logs:filter/MyFilter:honeycomb/MyExporter": {"logs/abc-123", "logs/def-456"},
			},
		},
		{
			name: "does not detect pipelines with different processors",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":  []string{"forward/otlp/receiver1"},
						"pipelines.logs/abc-123.processors": []string{"filter/Filter1"},
						"pipelines.logs/abc-123.exporters":  []string{"honeycomb/MyExporter"},
						"pipelines.logs/def-456.receivers":  []string{"forward/otlp/receiver2"},
						"pipelines.logs/def-456.processors": []string{"filter/Filter2"},
						"pipelines.logs/def-456.exporters":  []string{"honeycomb/MyExporter"},
					},
				},
			},
			expected: map[string][]string{}, // no duplicates
		},
		{
			name: "ignores ingress pipelines",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/ingress_otlp/receiver1.receivers":  []string{"otlp/receiver1"},
						"pipelines.logs/ingress_otlp/receiver1.processors": []string{"memory_limiter/receiver1", "usage"},
						"pipelines.logs/ingress_otlp/receiver1.exporters":  []string{"forward/otlp/receiver1"},
						"pipelines.logs/ingress_otlp/receiver2.receivers":  []string{"otlp/receiver2"},
						"pipelines.logs/ingress_otlp/receiver2.processors": []string{"memory_limiter/receiver2", "usage"},
						"pipelines.logs/ingress_otlp/receiver2.exporters":  []string{"forward/otlp/receiver2"},
					},
				},
			},
			expected: map[string][]string{}, // no duplicates (ingress pipelines ignored)
		},
		{
			name: "detects duplicates with empty processors",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":  []string{"forward/otlp/receiver1"},
						"pipelines.logs/abc-123.processors": []string{},
						"pipelines.logs/abc-123.exporters":  []string{"s3/MyArchive"},
						"pipelines.logs/def-456.receivers":  []string{"forward/otlp/receiver2"},
						"pipelines.logs/def-456.processors": []string{},
						"pipelines.logs/def-456.exporters":  []string{"s3/MyArchive"},
					},
				},
			},
			expected: map[string][]string{
				"logs::s3/MyArchive": {"logs/abc-123", "logs/def-456"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Translator{}
			result := tr.detectDuplicateDownstreams(tt.config)

			assert.Equal(t, len(tt.expected), len(result), "number of duplicate groups")
			for signature, expectedPipelines := range tt.expected {
				assert.Contains(t, result, signature, "signature should exist")
				assert.ElementsMatch(t, expectedPipelines, result[signature], "pipelines match regardless of order")
			}
		})
	}
}

func TestMergeDuplicateDownstreams(t *testing.T) {
	tests := []struct {
		name       string
		config     *tmpl.CollectorConfig
		duplicates map[string][]string
		verify     func(t *testing.T, cfg *tmpl.CollectorConfig)
	}{
		{
			name: "merges two pipelines with same config",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":  []string{"forward/otlp/receiver1"},
						"pipelines.logs/abc-123.processors": []string{"filter/MyFilter"},
						"pipelines.logs/abc-123.exporters":  []string{"honeycomb/MyExporter"},
						"pipelines.logs/def-456.receivers":  []string{"forward/otlp/receiver2"},
						"pipelines.logs/def-456.processors": []string{"filter/MyFilter"},
						"pipelines.logs/def-456.exporters":  []string{"honeycomb/MyExporter"},
					},
				},
			},
			duplicates: map[string][]string{
				"logs:filter/MyFilter:honeycomb/MyExporter": {"logs/abc-123", "logs/def-456"},
			},
			verify: func(t *testing.T, cfg *tmpl.CollectorConfig) {
				service := cfg.Sections["service"]

				// Canonical pipeline (abc-123) should have merged receivers
				assert.Contains(t, service, "pipelines.logs/abc-123.receivers")
				receivers := service["pipelines.logs/abc-123.receivers"].([]string)
				assert.ElementsMatch(t, []string{"forward/otlp/receiver1", "forward/otlp/receiver2"}, receivers)

				// Canonical pipeline should keep its processors and exporters
				assert.Contains(t, service, "pipelines.logs/abc-123.processors")
				assert.Contains(t, service, "pipelines.logs/abc-123.exporters")

				// Duplicate pipeline (def-456) should be deleted
				assert.NotContains(t, service, "pipelines.logs/def-456.receivers")
				assert.NotContains(t, service, "pipelines.logs/def-456.processors")
				assert.NotContains(t, service, "pipelines.logs/def-456.exporters")
			},
		},
		{
			name: "handles three duplicate pipelines",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":  []string{"forward/otlp/receiver1"},
						"pipelines.logs/abc-123.processors": []string{},
						"pipelines.logs/abc-123.exporters":  []string{"s3/Archive"},
						"pipelines.logs/def-456.receivers":  []string{"forward/otlp/receiver2"},
						"pipelines.logs/def-456.processors": []string{},
						"pipelines.logs/def-456.exporters":  []string{"s3/Archive"},
						"pipelines.logs/ghi-789.receivers":  []string{"forward/otlp/receiver3"},
						"pipelines.logs/ghi-789.processors": []string{},
						"pipelines.logs/ghi-789.exporters":  []string{"s3/Archive"},
					},
				},
			},
			duplicates: map[string][]string{
				"logs::s3/Archive": {"logs/abc-123", "logs/def-456", "logs/ghi-789"},
			},
			verify: func(t *testing.T, cfg *tmpl.CollectorConfig) {
				service := cfg.Sections["service"]

				// Canonical pipeline should have all three receivers
				receivers := service["pipelines.logs/abc-123.receivers"].([]string)
				assert.ElementsMatch(t, []string{"forward/otlp/receiver1", "forward/otlp/receiver2", "forward/otlp/receiver3"}, receivers)

				// Other two pipelines should be deleted
				assert.NotContains(t, service, "pipelines.logs/def-456.receivers")
				assert.NotContains(t, service, "pipelines.logs/ghi-789.receivers")
			},
		},
		{
			name: "no-op when no duplicates",
			config: &tmpl.CollectorConfig{
				Sections: map[string]tmpl.DottedConfig{
					"service": {
						"pipelines.logs/abc-123.receivers":  []string{"forward/otlp/receiver1"},
						"pipelines.logs/abc-123.processors": []string{"filter/Filter1"},
						"pipelines.logs/abc-123.exporters":  []string{"honeycomb/Exporter1"},
					},
				},
			},
			duplicates: map[string][]string{},
			verify: func(t *testing.T, cfg *tmpl.CollectorConfig) {
				service := cfg.Sections["service"]

				// Original pipeline unchanged
				assert.Contains(t, service, "pipelines.logs/abc-123.receivers")
				assert.Contains(t, service, "pipelines.logs/abc-123.processors")
				assert.Contains(t, service, "pipelines.logs/abc-123.exporters")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Translator{}
			err := tr.mergeDuplicateDownstreams(tt.config, tt.duplicates)
			require.NoError(t, err)

			tt.verify(t, tt.config)
		})
	}
}

func TestDownstreamDeduplication_Integration(t *testing.T) {
	tests := []struct {
		name          string
		setupConfig   func() *tmpl.CollectorConfig
		expectMerged  bool
		verifyMerged  string
		verifyDeleted []string
	}{
		{
			name: "merges downstream pipelines with same config after ingress generation",
			setupConfig: func() *tmpl.CollectorConfig {
				cfg := tmpl.NewCollectorConfig()
				// Shared receiver used by two different pipelines (logs and metrics)
				// Both downstream paths have same processors+exporters
				cfg.Set("service", "pipelines.logs/abc-123.receivers", []string{"otlp/receiver1"})
				cfg.Set("service", "pipelines.logs/abc-123.processors", []string{"filter/MyFilter"})
				cfg.Set("service", "pipelines.logs/abc-123.exporters", []string{"honeycomb/MyExporter"})
				cfg.Set("service", "pipelines.metrics/def-456.receivers", []string{"otlp/receiver1"}) // same receiver
				cfg.Set("service", "pipelines.metrics/def-456.processors", []string{})
				cfg.Set("service", "pipelines.metrics/def-456.exporters", []string{"honeycomb/MetricsExporter"})
				// Second shared receiver with two pipelines that have identical downstream config
				cfg.Set("service", "pipelines.logs/ghi-789.receivers", []string{"otlp/receiver2"})
				cfg.Set("service", "pipelines.logs/ghi-789.processors", []string{"filter/MyFilter"})
				cfg.Set("service", "pipelines.logs/ghi-789.exporters", []string{"honeycomb/MyExporter"})
				cfg.Set("service", "pipelines.metrics/jkl-012.receivers", []string{"otlp/receiver2"}) // same receiver
				cfg.Set("service", "pipelines.metrics/jkl-012.processors", []string{})
				cfg.Set("service", "pipelines.metrics/jkl-012.exporters", []string{"honeycomb/MetricsExporter"})
				return cfg
			},
			expectMerged:  true,
			verifyMerged:  "pipelines.logs/abc-123.receivers",
			verifyDeleted: []string{"pipelines.logs/ghi-789.receivers"},
		},
		{
			name: "does not merge pipelines with different configs",
			setupConfig: func() *tmpl.CollectorConfig {
				cfg := tmpl.NewCollectorConfig()
				// Shared receiver but different downstream configs
				cfg.Set("service", "pipelines.logs/abc-123.receivers", []string{"otlp/receiver1"})
				cfg.Set("service", "pipelines.logs/abc-123.processors", []string{"filter/Filter1"})
				cfg.Set("service", "pipelines.logs/abc-123.exporters", []string{"honeycomb/Exporter1"})
				cfg.Set("service", "pipelines.metrics/def-456.receivers", []string{"otlp/receiver1"}) // same receiver
				cfg.Set("service", "pipelines.metrics/def-456.processors", []string{"filter/Filter2"})
				cfg.Set("service", "pipelines.metrics/def-456.exporters", []string{"honeycomb/Exporter2"})
				return cfg
			},
			expectMerged:  false,
			verifyMerged:  "",
			verifyDeleted: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			tr := &Translator{}

			// Run full transformation (ingress + deduplication)
			result, err := tr.transformToIngressPipelines(cfg, nil, nil)
			require.NoError(t, err)

			resultCfg := result.(*tmpl.CollectorConfig)
			service := resultCfg.Sections["service"]

			if tt.expectMerged {
				// Verify that duplicate logs pipelines with same config got merged
				// After transform:
				// - logs/abc-123 and logs/ghi-789 both have same processors+exporters
				// - One should remain with both forward connectors, one should be deleted

				// Check which one remains (first in map iteration order)
				hasAbc := service["pipelines.logs/abc-123.receivers"] != nil
				hasGhi := service["pipelines.logs/ghi-789.receivers"] != nil

				// Exactly one should remain
				assert.True(t, (hasAbc && !hasGhi) || (!hasAbc && hasGhi), "Exactly one of the duplicate pipelines should remain")

				var mergedReceivers []string
				if hasAbc {
					mergedReceivers = service["pipelines.logs/abc-123.receivers"].([]string)
				} else {
					mergedReceivers = service["pipelines.logs/ghi-789.receivers"].([]string)
				}

				// Should have 2 forward connectors after merging
				assert.Len(t, mergedReceivers, 2, "Expected merged pipeline to have 2 receivers (forward connectors): %v", mergedReceivers)
				assert.ElementsMatch(t, []string{"forward/otlp/receiver1", "forward/otlp/receiver2"}, mergedReceivers)
			} else {
				// Verify both downstream pipelines still exist (not merged due to different configs)
				assert.Contains(t, service, "pipelines.logs/abc-123.receivers")
				assert.Contains(t, service, "pipelines.metrics/def-456.receivers")
			}
		})
	}
}
