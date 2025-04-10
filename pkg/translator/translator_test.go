package translator

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestGenerateConfigForAllComponents(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	for _, component := range comps {
		for _, properties := range []string{"all", "defaults"} {
			testData := fmt.Sprintf("%s_%s.yaml", strings.ToLower(component.Kind), properties)
			t.Run(testData, func(t *testing.T) {
				// test source config lives in testdata/hpsf
				b, err := os.ReadFile(path.Join("testdata", "hpsf", testData))
				require.NoError(t, err)
				var inputData = string(b)

				for _, template := range component.Templates {
					configType := config.Type(template.Kind)
					b, err = os.ReadFile(path.Join("testdata", string(configType), testData))
					require.NoError(t, err)
					var expectedConfig = string(b)

					var hpsf *hpsf.HPSF
					dec := yamlv3.NewDecoder(strings.NewReader(inputData))
					err = dec.Decode(&hpsf)
					require.NoError(t, err)

					cfg, err := tlater.GenerateConfig(hpsf, configType, nil)
					require.NoError(t, err)

					got, err := cfg.RenderYAML()
					require.NoError(t, err)

					assert.Equal(t, expectedConfig, string(got))
				}
			})
		}
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
			expectedConfigTestData: "testdata/refinery_config/default.yaml",
		},
		{
			desc:                   "Refinery Rules",
			ct:                     config.RefineryRulesType,
			expectedConfigTestData: "testdata/refinery_rules/default.yaml",
		},
		{
			desc:                   "Collector Config",
			ct:                     config.CollectorConfigType,
			expectedConfigTestData: "testdata/collector_config/default.yaml",
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
	b, err := os.ReadFile("testdata/refinery_rules/default.yaml")
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
	entries, err := os.ReadDir(path.Join("testdata", "hpsf"))
	require.NoError(t, err)
	// Filter for YAML files
	var yamlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			// Construct the full path to the file
			filePath := path.Join("testdata", "hpsf", entry.Name())
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
			require.NoError(t, errors.Unwrap(err))
		})
	}
}
