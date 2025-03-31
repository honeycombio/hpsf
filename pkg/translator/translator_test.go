package translator

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestGenerateConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		inputTestData string
	}{
		{
			desc:          "OTLP GRPC & HTTP in, GRPC out",
			inputTestData: "simple_grpc.yaml",
		},
		{
			desc:          "GRPC in and out with headers",
			inputTestData: "simple_grpc_with_headers.yaml",
		},
		{
			desc:          "OTLP GRPC & HTTP in, HTTP out",
			inputTestData: "simple_http.yaml",
		},
		{
			desc:          "OTLP GRPC & HTTP in, HTTP out with headers",
			inputTestData: "simple_http_with_headers.yaml",
		},
		{
			desc:          "OTLP GRPC & HTTP in, HTTP out with headers",
			inputTestData: "simple_http_with_headers_insecure.yaml",
		},
		{
			desc:          "OTLP GRPC & HTTP in and a debug exporter",
			inputTestData: "otlp_with_debug_exporter.yaml",
		},
		{
			desc:          "Collector with filtering processor",
			inputTestData: "http_with_filtering.yaml",
		},
		{
			desc:          "Collector with log deduplication processor",
			inputTestData: "otlp_with_logdeduplication.yaml",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// test source config lives in testdata/hpsf
			b, err := os.ReadFile(path.Join("testdata", "hpsf", tC.inputTestData))
			require.NoError(t, err)
			var inputData = string(b)

			for _, configType := range []config.Type{config.CollectorConfigType} {
				b, err = os.ReadFile(path.Join("testdata", string(configType), tC.inputTestData))
				require.NoError(t, err)
				var expectedConfig = string(b)

				var hpsf *hpsf.HPSF
				dec := yamlv3.NewDecoder(strings.NewReader(inputData))
				err = dec.Decode(&hpsf)
				require.NoError(t, err)

				tlater := NewEmptyTranslator()
				comps, err := data.LoadEmbeddedComponents()
				require.NoError(t, err)
				tlater.InstallComponents(comps)
				require.Equal(t, comps, tlater.GetComponents())

				templates, err := data.LoadEmbeddedTemplates()
				require.NoError(t, err)
				tlater.InstallTemplates(templates)
				require.Equal(t, templates, tlater.GetTemplates())

				cfg, err := tlater.GenerateConfig(hpsf, configType, nil)
				require.NoError(t, err)

				got, err := cfg.RenderYAML()
				require.NoError(t, err)

				assert.Equal(t, expectedConfig, string(got))
			}
		})
	}
}

func TestDefaultHPSF(t *testing.T) {
	testCases := []struct {
		desc                   string
		ct                     config.Type
		inputHPSFTestData      string
		expectedConfigTestData string
	}{
		{
			desc:                   "Refinery Config",
			ct:                     config.RefineryConfigType,
			expectedConfigTestData: "testdata/default_refinery_config.yaml",
		},
		{
			desc:                   "Refinery Rules",
			ct:                     config.RefineryRulesType,
			expectedConfigTestData: "testdata/default_refinery_rules.yaml",
		},
		{
			desc:                   "Collector Config",
			ct:                     config.CollectorConfigType,
			expectedConfigTestData: "testdata/default_collector_config.yaml",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			b, err := os.ReadFile(tC.expectedConfigTestData)
			require.NoError(t, err)
			var expectedConfig = string(b)

			h, err := data.LoadEmbeddedDefaultTemplate()
			require.NoError(t, err)

			tlater := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			tlater.InstallComponents(comps)

			cfg, err := tlater.GenerateConfig(&h, tC.ct, nil)
			require.NoError(t, err)

			got, err := cfg.RenderYAML()
			require.NoError(t, err)

			assert.Equal(t, expectedConfig, string(got))
		})
	}
}

func TestHPSFWithoutSamplerComponentGeneratesValidRefineryRules(t *testing.T) {
	b, err := os.ReadFile("testdata/default_refinery_rules.yaml")
	require.NoError(t, err)
	var expectedConfig = string(b)

	hpsf := hpsf.HPSF{}
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)

	cfg, err := tlater.GenerateConfig(&hpsf, config.RefineryRulesType, nil)
	require.NoError(t, err)

	got, err := cfg.RenderYAML()
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, string(got))
}

func TestTranslatorValidation(t *testing.T) {
	// read all yaml files in testdata
	entries, err := os.ReadDir("testdata")
	require.NoError(t, err)
	// Filter for YAML files
	var yamlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			// Construct the full path to the file
			filePath := "testdata/" + entry.Name()
			yamlFiles = append(yamlFiles, filePath)
		}
	}

	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)

	// Iterate over each YAML file and test for validation
	for _, filePath := range yamlFiles {
		t.Run("Validate test for "+filePath, func(t *testing.T) {
			// Read the file content
			b, err := os.ReadFile(filePath)
			require.NoError(t, err)
			var inputData = string(b)

			var hpsf *hpsf.HPSF
			dec := yamlv3.NewDecoder(strings.NewReader(inputData))
			err = dec.Decode(&hpsf)
			require.NoError(t, err)

			err = tlater.ValidateConfig(hpsf)
			if err != nil {
				// If validation fails, check if it's a validator.Result
				// and ensure it contains errors
				if result, ok := err.(validator.Result); ok {
					// If it's a Result, we can check the details
					// This means the validation failed and we expect it to fail
					// We can log the error for debugging
					for _, err := range result.Unwrap() {
						t.Logf("Validation failed for %s: %v", filePath, err)
					}
				}
			}
			require.NoError(t, err)
		})
	}
}
