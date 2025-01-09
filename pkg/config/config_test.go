package config

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/validator"
)

func TestDefaultConfigurationIsValidYAML(t *testing.T) {
	_, err := validator.EnsureYAML([]byte(DefaultConfiguration))
	if err != nil {
		t.Errorf("DefaultConfiguration is not valid YAML: %s", err)
	}
}
