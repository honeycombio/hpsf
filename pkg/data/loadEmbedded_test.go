package data

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/stretchr/testify/require"
	y "gopkg.in/yaml.v3"
)

func TestLoadEmbeddedComponents(t *testing.T) {
	got, err := LoadEmbeddedComponents()
	if err != nil {
		t.Errorf("LoadEmbeddedComponents() error = '%v', want nil", err)
	}
	if len(got) == 0 {
		t.Errorf("LoadEmbeddedComponents() = %v, want non-empty", got)
	}
	// we'll eventually move all of this to a validation library and use that; for now this is just a quick check
	for k, v := range got {
		switch v.Type {
		case config.ComponentTypeBase, config.ComponentTypeMeta, config.ComponentTypeTemplate:
			// ok
		default:
			t.Errorf("LoadEmbeddedComponents() %s style = %v, what's that?", k, v.Type)
		}
		switch v.Status {
		case config.ComponentStatusArchived, config.ComponentStatusDeprecated:
			t.Errorf("LoadEmbeddedComponents() %s status = %v, want something active", k, v.Status)
		case config.ComponentStatusAlpha, config.ComponentStatusDevelopment, config.ComponentStatusStable:
			// ok
		default:
			t.Errorf("LoadEmbeddedComponents() %s status = %v, what's that?", k, v.Status)
		}

		ym, err := v.AsYAML()
		require.NoError(t, err)
		require.NotEmpty(t, ym)
		var m map[string]any
		err = y.Unmarshal([]byte(ym), &m)
		require.NoError(t, err)
		require.NotEmpty(t, m)
		dc := tmpl.NewDottedConfig(m)
		require.NotEmpty(t, dc)
		mustHave := []string{"name", "kind", "type", "status", "style", "version"}
		for _, mh := range mustHave {
			v, ok := dc[mh]
			require.True(t, ok, fmt.Sprintf("missing %s in %s", mh, k))
			require.NotEmpty(t, v)
		}
	}
}

func TestTemplateComponents(t *testing.T) {
	components, err := LoadEmbeddedComponents()
	require.NoError(t, err)
	// for test component type
	tests := []struct {
		name       string
		kind       string
		cType      config.Type
		config     map[string]any
		wantOutput string
	}{
		{
			name:       "HoneycombExporter to refinery config",
			kind:       "HoneycombExporter",
			cType:      config.RefineryConfigType,
			config:     map[string]any{"APIKey": "test"},
			wantOutput: "HoneycombExporter_output_refinery_config.yaml",
		},
		{
			name:       "DeterministicSampler to refinery rules",
			kind:       "DeterministicSampler",
			cType:      config.RefineryRulesType,
			config:     map[string]any{"Environment": "staging", "SampleRate": 42},
			wantOutput: "DeterministicSampler_output_refinery_rules.yaml",
		},
		{
			name:  "EMAThroughputSampler to refinery rules",
			kind:  "EMAThroughput",
			cType: config.RefineryRulesType,
			config: map[string]any{
				"Environment":          "staging",
				"GoalThroughputPerSec": 42,
				"AdjustmentInterval":   "120s",
				"FieldList":            []string{"http.method", "http.status_code"},
			},
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

func TestLoadTemplates(t *testing.T) {
	templates, err := LoadEmbeddedTemplates()
	require.NoError(t, err)
	require.NotEmpty(t, templates)

	tests := []struct {
		name string
		kind string
	}{
		{
			name: "EMA Throughput Sampling",
			kind: "TemplateEMAThroughput",
		},
		{
			name: "Basic Proxy",
			kind: "TemplateProxy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, ok := templates[tt.kind]
			require.True(t, ok)
			require.NotNil(t, tmpl)
			require.Equal(t, tt.kind, tmpl.Kind)
			require.Equal(t, tt.name, tmpl.Name)
			require.NotEmpty(t, tmpl.Components)
			require.NotEmpty(t, tmpl.Version)
			require.NotEmpty(t, tmpl.Summary)
			require.NotEmpty(t, tmpl.Description)
			require.Empty(t, tmpl.Validate())
		})
	}
}

func TestDefaultConfigurationIsValidYAML(t *testing.T) {
	templates, err := LoadEmbeddedTemplates()
	require.NoError(t, err)
	template, ok := templates[DefaultConfigurationKind]
	require.True(t, ok)
	data, err := template.AsYAML()
	require.NoError(t, err)
	err = hpsf.EnsureHPSFYAML(data)
	require.NoError(t, err)
}
