package config

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadTemplateComponents(t *testing.T) {
	got, err := LoadTemplateComponents()
	if err != nil {
		t.Errorf("LoadTemplateComponents() error = '%v', want nil", err)
	}
	if len(got) == 0 {
		t.Errorf("LoadTemplateComponents() = %v, want non-empty", got)
	}
}

func TestTemplateComponents(t *testing.T) {
	components, err := LoadTemplateComponents()
	require.NoError(t, err)
	// for test component type
	tests := []struct {
		name       string
		kind       string
		cType      Type
		config     map[string]any
		wantOutput string
	}{
		{
			name:       "HoneycombExporter to refinery config",
			kind:       "HoneycombExporter",
			cType:      RefineryConfigType,
			config:     map[string]any{"APIKey": "test"},
			wantOutput: "HoneycombExporter_output_refinery_config.yaml",
		},
		{
			name:       "DeterministicSampler to refinery rules",
			kind:       "DeterministicSampler",
			cType:      RefineryRulesType,
			config:     map[string]any{"Environment": "staging", "SampleRate": 42},
			wantOutput: "DeterministicSampler_output_refinery_rules.yaml",
		},
		{
			name:       "EMAThroughputSampler to refinery rules",
			kind:       "EMAThroughput",
			cType:      RefineryRulesType,
			config:     map[string]any{"Environment": "staging", "GoalThroughput": 42, "AdjustmentInterval": 120, "FieldList": []string{"test"}},
			wantOutput: "EmaThroughput_output_refinery_rules.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, err := os.ReadFile(path.Join("testdata", tt.wantOutput))
			require.NoError(t, err)
			c, ok := components[tt.kind]
			require.True(t, ok)
			conf, err := c.GenerateConfig(tt.cType, tt.config)
			require.NoError(t, err)
			require.NotNil(t, conf)
			got, err := conf.RenderYAML()
			require.NoError(t, err)
			require.Equal(t, string(want), string(got))
		})
	}
}
