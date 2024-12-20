package config

import (
	"testing"
)

func TestLoadTemplateComponents(t *testing.T) {
	got, err := LoadTemplateComponents()
	if err != nil {
		t.Errorf("LoadTemplateComponents() error = '%v', want nil", err)
	}
	if len(got) == 0 {
		t.Errorf("LoadTemplateComponents() = %v, want non-empty", got)
	}
}
