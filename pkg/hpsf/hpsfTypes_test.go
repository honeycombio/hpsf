package hpsf

import (
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
)

func TestEnsureHPSF(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"0", []string{}, "empty"},
		{"1", []string{"a"}, "unexpected keys"},
		{"2", []string{"connections"}, "unmarshal errors"},
		{"3", []string{"components", "connections"}, "unmarshal errors"},
		{"4", []string{"components", "connections", "something"}, "unexpected keys"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := tmpl.DottedConfig{}
			for i, arg := range tt.args {
				y[arg] = i
			}
			text, _ := y.RenderYAML()
			if err := EnsureHPSFYAML(string(text)); (err != nil) && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("EnsureHPSF() error = %v, should contain '%v'", err, tt.wantErr)
			}
		})
	}
}

func TestHPSF_Validate(t *testing.T) {
	inputData := []byte(`components:
  - name: otlp_in
    kind: OTelReceiver
    properties:
      - name: GRPCPort
        value: 9922
      - name: HTTPPort
        value: 1234
  - name: otlp_out
    kind: OTelGRPCExporter
    properties:
      - name: Host
        value: myhost.com
      - name: Port
        value: 1234
      - name: Headers
        value:
          "header1": "1234"
connections:
  - source:
      component: otlp_in
      port: Traces
      type: OTelTraces
    destination:
      component: otlp_out
      port: Traces
      type: OTelTraces`)

	_, err := validator.EnsureYAML(inputData)
	assert.NoError(t, err)

	var hpsf HPSF
	err = yaml.Unmarshal(inputData, &hpsf)
	assert.NoError(t, err)

	errors := hpsf.Validate()
	assert.Empty(t, errors)
}

func TestDefaultConfigurationIsValidYAML(t *testing.T) {
	err := EnsureHPSFYAML(DefaultConfiguration)
	require.NoError(t, err)
}

func TestComponent_GetSafeName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"a", "a"},
		{"a b", "a_b"},
		{"a b c", "a_b_c"},
		{"a#@#$%^&*()b", "a_b"},
		{"Deterministic Sampler", "Deterministic_Sampler"},
		{"Deterministic_Sampler", "Deterministic_Sampler"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Component{Name: tt.name}
			if got := c.GetSafeName(); got != tt.want {
				t.Errorf("Component.GetSafeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
