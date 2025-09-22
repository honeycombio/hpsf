package translator

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// TestComponentValidationIntegration tests that component validation works end-to-end
func TestComponentValidationIntegration(t *testing.T) {
	// Create a translator with a mock component that has component validations
	components := map[string]config.TemplateComponent{
		"MockAuthExporter": {
			Kind:    "MockAuthExporter",
			Name:    "Mock Authentication Exporter",
			Version: "v1.0.0",
			Properties: []config.TemplateProperty{
				{Name: "APIKey", Type: hpsf.PTYPE_STRING},
				{Name: "BearerToken", Type: hpsf.PTYPE_STRING},
				{Name: "Username", Type: hpsf.PTYPE_STRING},
				{Name: "Password", Type: hpsf.PTYPE_STRING},
				{Name: "EnableTLS", Type: hpsf.PTYPE_BOOL, Default: false},
				{Name: "TLSCertPath", Type: hpsf.PTYPE_STRING},
				{Name: "TLSKeyPath", Type: hpsf.PTYPE_STRING},
			},
			Validations: []string{
				"exactly_one_of(APIKey, BearerToken, Username)",
				"require_together(Username, Password)",
				"conditional_require_together(TLSCertPath, TLSKeyPath | when EnableTLS=true)",
			},
		},
	}

	translator := NewEmptyTranslator()
	translator.InstallComponents(components)

	t.Run("ValidConfiguration", func(t *testing.T) {
		// Create a valid HPSF document
		hpsfDoc := &hpsf.HPSF{
			Components: []*hpsf.Component{
				{
					Name: "test-auth",
					Kind: "MockAuthExporter",
					Properties: []hpsf.Property{
						{Name: "APIKey", Value: "test-key-123"},
						{Name: "EnableTLS", Value: true},
						{Name: "TLSCertPath", Value: "/path/to/cert.pem"},
						{Name: "TLSKeyPath", Value: "/path/to/key.pem"},
					},
				},
			},
		}

		err := translator.ValidateConfig(hpsfDoc)
		if err != nil {
			t.Errorf("Expected valid configuration to pass validation, got: %v", err)
		}
	})

	t.Run("ExactlyOneOfViolation", func(t *testing.T) {
		// Multiple authentication methods specified - should fail
		hpsfDoc := &hpsf.HPSF{
			Components: []*hpsf.Component{
				{
					Name: "test-auth",
					Kind: "MockAuthExporter",
					Properties: []hpsf.Property{
						{Name: "APIKey", Value: "test-key-123"},
						{Name: "BearerToken", Value: "test-token-456"},
					},
				},
			},
		}

		err := translator.ValidateConfig(hpsfDoc)
		if err == nil {
			t.Error("Expected configuration with multiple auth methods to fail validation")
		}
	})

	t.Run("RequireTogetherViolation", func(t *testing.T) {
		// Username without password - should fail
		hpsfDoc := &hpsf.HPSF{
			Components: []*hpsf.Component{
				{
					Name: "test-auth",
					Kind: "MockAuthExporter",
					Properties: []hpsf.Property{
						{Name: "Username", Value: "testuser"},
						// Password missing
					},
				},
			},
		}

		err := translator.ValidateConfig(hpsfDoc)
		if err == nil {
			t.Error("Expected configuration with username but no password to fail validation")
		}
	})

	t.Run("ConditionalRequireTogetherViolation", func(t *testing.T) {
		// TLS enabled but missing cert/key paths - should fail
		hpsfDoc := &hpsf.HPSF{
			Components: []*hpsf.Component{
				{
					Name: "test-auth",
					Kind: "MockAuthExporter",
					Properties: []hpsf.Property{
						{Name: "APIKey", Value: "test-key-123"},
						{Name: "EnableTLS", Value: true},
						{Name: "TLSCertPath", Value: "/path/to/cert.pem"},
						// TLSKeyPath missing
					},
				},
			},
		}

		err := translator.ValidateConfig(hpsfDoc)
		if err == nil {
			t.Error("Expected configuration with TLS enabled but missing key path to fail validation")
		}
	})

	t.Run("ConditionalNotMet", func(t *testing.T) {
		// TLS disabled, missing cert/key paths should be OK
		hpsfDoc := &hpsf.HPSF{
			Components: []*hpsf.Component{
				{
					Name: "test-auth",
					Kind: "MockAuthExporter",
					Properties: []hpsf.Property{
						{Name: "APIKey", Value: "test-key-123"},
						{Name: "EnableTLS", Value: false},
						// TLS cert/key paths missing but that's OK since TLS is disabled
					},
				},
			},
		}

		err := translator.ValidateConfig(hpsfDoc)
		if err != nil {
			t.Errorf("Expected configuration with TLS disabled to pass validation, got: %v", err)
		}
	})
}