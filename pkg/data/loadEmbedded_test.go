package data

import (
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
		case config.ComponentStatusDevelopment, config.ComponentStatusAlpha, config.ComponentStatusStable, config.ComponentStatusArchived, config.ComponentStatusDeprecated:
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
			require.True(t, ok, "missing %s in %s", mh, k)
			require.NotEmpty(t, v)
		}
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
			require.NoError(t, tmpl.Validate())
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
