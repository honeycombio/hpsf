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
	testCases := []struct {
		desc                   string
		inputHPSFTestData      string
		expectedConfigTestData string
	}{
		{
			desc:                   "OTLP GRPC & HTTP in, GRPC out",
			inputHPSFTestData:      "testdata/simple_grpc_hpsf.yaml",
			expectedConfigTestData: "testdata/simple_grpc_collector_config.yaml",
		},
		{
			desc:                   "OTLP GRPC & HTTP in, HTTP out",
			inputHPSFTestData:      "testdata/simple_http_hpsf.yaml",
			expectedConfigTestData: "testdata/simple_http_collector_config.yaml",
		},
		{
			desc:                   "OTLP GRPC & HTTP in and a debug exporter",
			inputHPSFTestData:      "testdata/otlp_with_debug_exporter_hpsf.yaml",
			expectedConfigTestData: "testdata/otlp_with_debug_exporter_collector_config.yaml",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			b, err := os.ReadFile(tC.inputHPSFTestData)
			require.NoError(t, err)
			var inputData = string(b)

			b, err = os.ReadFile(tC.expectedConfigTestData)
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
		})
	}
}
