package hpsf

import (
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
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
			if err := EnsureHPSF(string(text)); (err != nil) && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("EnsureHPSF() error = %v, should contain '%v'", err, tt.wantErr)
			}
		})
	}
}
