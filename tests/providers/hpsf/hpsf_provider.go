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
	if e.HasErrors() || len(e.GenerateErrors) > 0 {
		// iterate all the errors in errors.GenerateErrors and log them
		for _, err := range e.GenerateErrors {
			t.Errorf("Failed to parse config: %v \n configFile %s", err.Error, err.Config)
		}
		t.Fatalf("Failed to parse config")
	}
}

func GetParsedConfigsFromFile(t *testing.T, filename string) (refineryRules *refineryConfig.V2SamplerConfig, collectorConfig *otelcol.Config, groupedErrors ParserError) {
	file, err := os.ReadFile("multiple_otlp_exporters.yaml")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

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
	return

}

func unmarshalHPSF(data io.Reader) (*hpsf.HPSF, error) {
	var hpsf hpsf.HPSF
	dec := y.NewDecoder(data)
	err := dec.Decode(&hpsf)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling to yaml: %v", err)
	}
	return &hpsf, nil
}
