package translator

import (
	"os"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestGenerateConfig(t *testing.T) {
	b, err := os.ReadFile("testdata/simple_grpc_hpsf.yaml")
	require.NoError(t, err)
	var inputData = string(b)

	b, err = os.ReadFile("testdata/simple_grpc_collector_config.yaml")
	require.NoError(t, err)
	var expectedConfig = string(b)

	var hpsf *hpsf.HPSF
	dec := yamlv3.NewDecoder(strings.NewReader(inputData))
	err = dec.Decode(&hpsf)
	require.NoError(t, err)

	tlater, err := NewTranslator()
	require.NoError(t, err)

	cfg, err := tlater.GenerateConfig(hpsf, config.CollectorConfigType, nil)
	require.NoError(t, err)

	got, err := cfg.RenderYAML()
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, string(got))
}
