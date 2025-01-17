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
		wantOutput string
	}{
		{
			name:       "HoneycombExporter to refinery config",
			kind:       "HoneycombExporter",
			cType:      RefineryConfigType,
			wantOutput: "HoneycombExporter_output_refinery_config.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, err := os.ReadFile(path.Join("testdata", tt.wantOutput))
			require.NoError(t, err)
			c, ok := components[tt.kind]
			require.True(t, ok)
			conf, err := c.GenerateConfig(tt.cType, map[string]any{
				"APIKey": "test",
			})
			require.NoError(t, err)
			got, err := conf.RenderYAML()
			require.NoError(t, err)
			require.Equal(t, string(want), string(got))
		})
	}
}
