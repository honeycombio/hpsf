package hpsf

import (
	"fmt"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/stretchr/testify/assert"
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

func TestHPSF_ValidateFailures(t *testing.T) {
	// GRPCPort is missing a value / value type is wrong
	// connection names a non-existent component
	inputData := []byte(`components:
  - name: otlp_in
    kind: OTelReceiver
    properties:
      - name: GRPCPort
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
      component: otlp_in2
      port: Traces
      type: OTelTrace
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
	errs, ok := errors.(validator.Result)
	assert.True(t, ok)
	assert.Equal(t, 2, errs.Len())
	unwrapped := errs.Unwrap()
	assert.Equal(t, 2, len(unwrapped))
	assert.Contains(t, unwrapped[0].Error(), "GRPCPort")
	assert.Contains(t, unwrapped[1].Error(), "otlp_in2")
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

func TestPropType_ValueCoerce(t *testing.T) {
	var result any
	tests := []struct {
		p       PropType
		v       any
		target  any
		wantErr bool
	}{
		{PTYPE_STRING, "a", "a", false},
		{PTYPE_STRING, 1, "1", false},
		{PTYPE_STRING, 1.5, "1.5", false},
		{PTYPE_STRING, true, "true", false},
		{PTYPE_STRING, []string{"a"}, nil, true},
		{PTYPE_STRING, map[string]any{"a": 3}, nil, true},
		{PTYPE_INT, "a", nil, true},
		{PTYPE_INT, "17", 17, false},
		{PTYPE_INT, "17.5", nil, true},
		{PTYPE_INT, 1, 1, false},
		{PTYPE_INT, 1.0, 1, false},
		{PTYPE_INT, 1.5, nil, true},
		{PTYPE_INT, true, nil, true},
		{PTYPE_INT, []string{"a"}, nil, true},
		{PTYPE_INT, map[string]any{"a": 3}, nil, true},
		{PTYPE_FLOAT, "a", nil, true},
		{PTYPE_FLOAT, "17", 17.0, false},
		{PTYPE_FLOAT, "17.5", 17.5, false},
		{PTYPE_FLOAT, 1, 1.0, false},
		{PTYPE_FLOAT, 1.0, 1.0, false},
		{PTYPE_FLOAT, 1.5, 1.5, false},
		{PTYPE_FLOAT, true, nil, true},
		{PTYPE_FLOAT, []string{"a"}, nil, true},
		{PTYPE_FLOAT, map[string]any{"a": 3}, nil, true},
		{PTYPE_BOOL, "a", nil, true},
		{PTYPE_BOOL, "true", true, false},
		{PTYPE_BOOL, "True", true, false},
		{PTYPE_BOOL, "TRUE", true, false},
		{PTYPE_BOOL, "T", true, false},
		{PTYPE_BOOL, "t", true, false},
		{PTYPE_BOOL, "YES", true, false},
		{PTYPE_BOOL, "yes", true, false},
		{PTYPE_BOOL, "Yes", true, false},
		{PTYPE_BOOL, "Y", true, false},
		{PTYPE_BOOL, "y", true, false},
		{PTYPE_BOOL, "1", nil, true},
		{PTYPE_BOOL, "0", nil, true},
		{PTYPE_BOOL, 1, true, false},
		{PTYPE_BOOL, 0, false, false},
		{PTYPE_BOOL, 1.0, true, false},
		{PTYPE_BOOL, 0.0, false, false},
		{PTYPE_BOOL, 1.5, true, false},
		{PTYPE_BOOL, true, true, false},
		{PTYPE_BOOL, false, false, false},
		{PTYPE_BOOL, []string{"true"}, nil, true},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%s_%#v", tt.p, tt.v)
		t.Run(name, func(t *testing.T) {
			result = nil
			if err := tt.p.ValueCoerce(tt.v, &result); (err != nil) != tt.wantErr {
				t.Errorf("PropType.ValueCoerce() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.target, result)
		})
	}
}
