package hpsf

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/translator"
	collectorConfigProvider "github.com/honeycombio/hpsf/tests/providers/collector"
	refineryConfigProvider "github.com/honeycombio/hpsf/tests/providers/refinery"
	refineryConfig "github.com/honeycombio/refinery/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/otelcol"
	y "gopkg.in/yaml.v3"
)

type ErrorDetails struct {
	Config string
	Error  error
}

type ParserError struct {
	GenerateErrors map[config.Type]ErrorDetails

	error
}

func (e ParserError) HasErrors() bool {
	return len(e.GenerateErrors) > 0
}

func (e ParserError) FailIfError(t testing.TB) {
	require.False(t, e.HasErrors() || len(e.GenerateErrors) > 0, "Failed to parse config with errors: %+v", e.GenerateErrors)
}

func GetParsedConfigsFromFile(t *testing.T, filename string) (refineryRules *refineryConfig.V2SamplerConfig, collectorConfig *otelcol.Config, groupedErrors ParserError) {
	file, err := os.ReadFile(filename)
	require.NoError(t, err, "Failed to read file")

	return GetParsedConfigs(t, string(file))
}

func GetParsedConfigs(t *testing.T, hpsfConfig string) (refineryRules *refineryConfig.V2SamplerConfig, collectorConfig *otelcol.Config, groupedErrors ParserError) {
	hpsf, err := unmarshalHPSF(strings.NewReader(hpsfConfig))
	if err != nil {
		log.Fatalf("error unmarshaling HPSF: %v", err)
	}

	hpsfTranslator := translator.NewEmptyTranslator()
	allHpsfComponents, err := data.LoadEmbeddedComponents()
	if err != nil {
		log.Fatalf("error loading embedded components: %v", err)
	}
	hpsfTranslator.InstallComponents(allHpsfComponents)

	errors := make(map[config.Type]ErrorDetails)

	refineryRulesTmpl, err := hpsfTranslator.GenerateConfig(hpsf, config.RefineryRulesType, nil)
	if err != nil {
		errors[config.RefineryConfigType] = ErrorDetails{Config: hpsfConfig, Error: err}
	} else {
		refineryRules = refineryConfigProvider.GetParsedRulesConfig(t, refineryRulesTmpl.(*tmpl.RulesConfig))
	}

	collectorConfigTmpl, err := hpsfTranslator.GenerateConfig(hpsf, config.CollectorConfigType, nil)
	if err != nil {
		errors[config.CollectorConfigType] = ErrorDetails{Config: hpsfConfig, Error: err}
	} else {
		var parsingErrors collectorConfigProvider.CollectorConfigParseError
		collectorConfig, parsingErrors = collectorConfigProvider.GetParsedConfig(t, collectorConfigTmpl.(*tmpl.CollectorConfig))
		if parsingErrors.HasError {
			errors[config.CollectorConfigType] = ErrorDetails{Config: parsingErrors.Config, Error: parsingErrors.Error}
		}
	}

	if len(errors) > 0 {
		groupedErrors.GenerateErrors = errors
	}
	groupedErrors.FailIfError(t)
	return

}

func unmarshalHPSF(data io.Reader) (*hpsf.HPSF, error) {
	var h hpsf.HPSF
	dec := y.NewDecoder(data)
	err := dec.Decode(&h)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling to yaml: %v", err)
	}
	return &h, nil
}
