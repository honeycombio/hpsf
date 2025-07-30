package hpsf

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/honeycombio/hpsf/pkg/translator"
	collectorConfigProvider "github.com/honeycombio/hpsf/tests/providers/collector"
	refineryConfigProvider "github.com/honeycombio/hpsf/tests/providers/refinery"
	refineryConfig "github.com/honeycombio/refinery/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/otelcol"
)

type ErrorDetails struct {
	Config string
	Error  error
}

type ParserError struct {
	GenerateErrors map[hpsftypes.Type]ErrorDetails

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
	h, err := hpsf.FromYAML(strings.NewReader(hpsfConfig))
	if err != nil {
		log.Fatalf("error unmarshaling HPSF: %v", err)
	}

	hpsfTranslator := translator.NewEmptyTranslator()
	allHpsfComponents, err := data.LoadEmbeddedComponents()
	if err != nil {
		log.Fatalf("error loading embedded components: %v", err)
	}
	hpsfTranslator.InstallComponents(allHpsfComponents)

	errors := make(map[hpsftypes.Type]ErrorDetails)

	refineryRulesTmpl, err := hpsfTranslator.GenerateConfig(&h, hpsftypes.RefineryRules, nil)
	if err != nil {
		errors[hpsftypes.RefineryConfig] = ErrorDetails{Config: hpsfConfig, Error: err}
	} else {
		refineryRules = refineryConfigProvider.GetParsedRulesConfig(t, refineryRulesTmpl.(*tmpl.RulesConfig))
	}

	collectorConfigTmpl, err := hpsfTranslator.GenerateConfig(&h, hpsftypes.CollectorConfig, nil)
	if err != nil {
		errors[hpsftypes.CollectorConfig] = ErrorDetails{Config: hpsfConfig, Error: err}
	} else {
		var parsingErrors collectorConfigProvider.CollectorConfigParseError
		collectorConfig, parsingErrors = collectorConfigProvider.GetParsedConfig(t, collectorConfigTmpl.(*tmpl.CollectorConfig))
		if parsingErrors.HasError {
			errors[hpsftypes.CollectorConfig] = ErrorDetails{Config: parsingErrors.Config, Error: parsingErrors.Error}
		}
	}

	if len(errors) > 0 {
		groupedErrors.GenerateErrors = errors
	}
	groupedErrors.FailIfError(t)
	return

}
