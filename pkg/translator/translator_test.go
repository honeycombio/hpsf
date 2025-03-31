package translator

import (
	"os"
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
			desc:                   "GRPC in and out with headers",
			inputHPSFTestData:      "testdata/simple_grpc_hpsf_with_headers.yaml",
			expectedConfigTestData: "testdata/simple_grpc_collector_config_with_headers.yaml",
		},
		{
			desc:                   "OTLP GRPC & HTTP in, HTTP out",
			inputHPSFTestData:      "testdata/simple_http_hpsf.yaml",
			expectedConfigTestData: "testdata/simple_http_collector_config.yaml",
		},
		{
			desc:                   "OTLP GRPC & HTTP in, HTTP out with headers",
			inputHPSFTestData:      "testdata/simple_http_hpsf_with_headers.yaml",
			expectedConfigTestData: "testdata/simple_http_collector_config_with_headers.yaml",
		},
		{
			desc:                   "OTLP GRPC & HTTP in, HTTP out with headers",
			inputHPSFTestData:      "testdata/simple_http_hpsf_with_headers_insecure.yaml",
			expectedConfigTestData: "testdata/simple_http_collector_config_with_headers_insecure.yaml",
		},
		{
			desc:                   "OTLP GRPC & HTTP in and a debug exporter",
			inputHPSFTestData:      "testdata/otlp_with_debug_exporter_hpsf.yaml",
			expectedConfigTestData: "testdata/otlp_with_debug_exporter_collector_config.yaml",
		},
		{
			desc:                   "Collector with filtering processor",
			inputHPSFTestData:      "testdata/http_hpsf_with_filtering.yaml",
			expectedConfigTestData: "testdata/http_collector_config_with_filter_processor.yaml",
		},
		{
			desc:                   "Collector with log deduplication processor",
			inputHPSFTestData:      "testdata/otlp_with_logdeduplication_hpsf.yaml",
			expectedConfigTestData: "testdata/otlp_with_logdeduplication_collector_config.yaml",
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

			tlater := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			tlater.InstallComponents(comps)
			require.Equal(t, comps, tlater.GetComponents())

			templates, err := data.LoadEmbeddedTemplates()
			require.NoError(t, err)
			tlater.InstallTemplates(templates)
			require.Equal(t, templates, tlater.GetTemplates())

			cfg, err := tlater.GenerateConfig(hpsf, config.CollectorConfigType, nil)
			require.NoError(t, err)

			got, err := cfg.RenderYAML()
			require.NoError(t, err)

			assert.Equal(t, expectedConfig, string(got))
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
	t.Skip(t, "Skipping validation test since all files in testdata are valid.")
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
