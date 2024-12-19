package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/honeycombio/hpsf/pkg/yaml"
)

func TestCollectorBaseComponent(t *testing.T) {
	c := CollectorBaseComponent{}
	got, err := c.GenerateConfig(CollectorConfigType, map[string]any{})
	want := yaml.DottedConfig{
		"processors": map[string]interface{}{},
		"receivers":  map[string]interface{}{},
		"extensions": map[string]interface{}{},
		"service":    map[string]interface{}{},
	}
	require.NoError(t, err)
	require.Equal(t, want, got)
}
