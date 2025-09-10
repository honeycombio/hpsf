package config

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// TestValidateAtLeastOneOf tests the at_least_one_of validation type
func TestValidateAtLeastOneOf(t *testing.T) {
	// Create a test template component
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "PropA", Type: hpsf.PTYPE_STRING},
			{Name: "PropB", Type: hpsf.PTYPE_STRING},
			{Name: "PropC", Type: hpsf.PTYPE_STRING},
		},
		Validations: []ComponentValidation{
			{
				Type:       "at_least_one_of",
				Properties: []string{"PropA", "PropB", "PropC"},
				Message:    "At least one of PropA, PropB, or PropC must be set",
			},
		},
	}

	// Test case 1: All properties empty - should fail
	t.Run("AllEmpty", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "PropA", Value: ""},
				{Name: "PropB", Value: ""},
				{Name: "PropC", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err == nil {
			t.Error("Expected validation to fail when all properties are empty")
		}
	})

	// Test case 2: One property set - should pass
	t.Run("OneSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "PropA", Value: "value"},
				{Name: "PropB", Value: ""},
				{Name: "PropC", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when one property is set, got: %v", err)
		}
	})

	// Test case 3: Multiple properties set - should pass
	t.Run("MultipleSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "PropA", Value: "value1"},
				{Name: "PropB", Value: "value2"},
				{Name: "PropC", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when multiple properties are set, got: %v", err)
		}
	})
}

// TestValidateExactlyOneOf tests the exactly_one_of validation type
func TestValidateExactlyOneOf(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "APIKey", Type: hpsf.PTYPE_STRING},
			{Name: "BearerToken", Type: hpsf.PTYPE_STRING},
			{Name: "BasicAuth", Type: hpsf.PTYPE_STRING},
		},
		Validations: []ComponentValidation{
			{
				Type:       "exactly_one_of",
				Properties: []string{"APIKey", "BearerToken", "BasicAuth"},
				Message:    "Exactly one authentication method must be specified",
			},
		},
	}

	// Test case 1: No properties set - should fail
	t.Run("NoneSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "APIKey", Value: ""},
				{Name: "BearerToken", Value: ""},
				{Name: "BasicAuth", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err == nil {
			t.Error("Expected validation to fail when no properties are set")
		}
	})

	// Test case 2: Exactly one property set - should pass
	t.Run("ExactlyOneSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "APIKey", Value: "key123"},
				{Name: "BearerToken", Value: ""},
				{Name: "BasicAuth", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when exactly one property is set, got: %v", err)
		}
	})

	// Test case 3: Multiple properties set - should fail
	t.Run("MultipleSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "APIKey", Value: "key123"},
				{Name: "BearerToken", Value: "token456"},
				{Name: "BasicAuth", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err == nil {
			t.Error("Expected validation to fail when multiple properties are set")
		}
	})
}

// TestValidateMutuallyExclusive tests the mutually_exclusive validation type
func TestValidateMutuallyExclusive(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "GzipCompression", Type: hpsf.PTYPE_BOOL, Default: false},
			{Name: "LZ4Compression", Type: hpsf.PTYPE_BOOL, Default: false},
		},
		Validations: []ComponentValidation{
			{
				Type:       "mutually_exclusive",
				Properties: []string{"GzipCompression", "LZ4Compression"},
				Message:    "GzipCompression and LZ4Compression cannot both be enabled",
			},
		},
	}

	// Test case 1: Both properties false - should pass
	t.Run("BothFalse", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "GzipCompression", Value: false},
				{Name: "LZ4Compression", Value: false},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when both properties are false, got: %v", err)
		}
	})

	// Test case 2: Only one property true - should pass
	t.Run("OnlyOneTrue", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "GzipCompression", Value: true},
				{Name: "LZ4Compression", Value: false},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when only one property is true, got: %v", err)
		}
	})

	// Test case 3: Both properties true - should fail
	t.Run("BothTrue", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "GzipCompression", Value: true},
				{Name: "LZ4Compression", Value: true},
			},
		}
		err := tc.ValidateComponent(component)
		if err == nil {
			t.Error("Expected validation to fail when both properties are true")
		}
	})
}

// TestValidateRequireTogether tests the require_together validation type
func TestValidateRequireTogether(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "Username", Type: hpsf.PTYPE_STRING},
			{Name: "Password", Type: hpsf.PTYPE_STRING},
		},
		Validations: []ComponentValidation{
			{
				Type:       "require_together",
				Properties: []string{"Username", "Password"},
				Message:    "Username and Password must both be provided when using authentication",
			},
		},
	}

	// Test case 1: Both properties empty - should pass
	t.Run("BothEmpty", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "Username", Value: ""},
				{Name: "Password", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when both properties are empty, got: %v", err)
		}
	})

	// Test case 2: Both properties set - should pass
	t.Run("BothSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "Username", Value: "user123"},
				{Name: "Password", Value: "pass456"},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when both properties are set, got: %v", err)
		}
	})

	// Test case 3: Only one property set - should fail
	t.Run("OnlyOneSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "Username", Value: "user123"},
				{Name: "Password", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err == nil {
			t.Error("Expected validation to fail when only one property is set")
		}
	})
}

// TestValidateConditionalRequireTogether tests the conditional_require_together validation type
func TestValidateConditionalRequireTogether(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "EnableTLS", Type: hpsf.PTYPE_BOOL, Default: false},
			{Name: "TLSCertPath", Type: hpsf.PTYPE_STRING},
			{Name: "TLSKeyPath", Type: hpsf.PTYPE_STRING},
		},
		Validations: []ComponentValidation{
			{
				Type:              "conditional_require_together",
				ConditionProperty: "EnableTLS",
				ConditionValue:    true,
				Properties:        []string{"TLSCertPath", "TLSKeyPath"},
				Message:           "When EnableTLS is true, both TLSCertPath and TLSKeyPath must be provided",
			},
		},
	}

	// Test case 1: Condition false - should pass regardless of other properties
	t.Run("ConditionFalse", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "EnableTLS", Value: false},
				{Name: "TLSCertPath", Value: ""},
				{Name: "TLSKeyPath", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when condition is false, got: %v", err)
		}
	})

	// Test case 2: Condition true and all required properties set - should pass
	t.Run("ConditionTrueAllSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "EnableTLS", Value: true},
				{Name: "TLSCertPath", Value: "/path/to/cert"},
				{Name: "TLSKeyPath", Value: "/path/to/key"},
			},
		}
		err := tc.ValidateComponent(component)
		if err != nil {
			t.Errorf("Expected validation to pass when condition is true and all properties are set, got: %v", err)
		}
	})

	// Test case 3: Condition true but required properties missing - should fail
	t.Run("ConditionTrueMissingProperties", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "EnableTLS", Value: true},
				{Name: "TLSCertPath", Value: "/path/to/cert"},
				{Name: "TLSKeyPath", Value: ""},
			},
		}
		err := tc.ValidateComponent(component)
		if err == nil {
			t.Error("Expected validation to fail when condition is true but required properties are missing")
		}
	})
}

// TestUnknownValidationType tests that unknown validation types are handled correctly
func TestUnknownValidationType(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "PropA", Type: hpsf.PTYPE_STRING},
		},
		Validations: []ComponentValidation{
			{
				Type:       "unknown_validation_type",
				Properties: []string{"PropA"},
				Message:    "Test message",
			},
		},
	}

	component := &hpsf.Component{
		Name: "TestComponent",
		Properties: []hpsf.Property{
			{Name: "PropA", Value: "value"},
		},
	}

	err := tc.ValidateComponent(component)
	if err == nil {
		t.Error("Expected validation to fail for unknown validation type")
	}
}

// TestNonExistentProperty tests that referencing non-existent properties is handled correctly
func TestNonExistentProperty(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "PropA", Type: hpsf.PTYPE_STRING},
		},
		Validations: []ComponentValidation{
			{
				Type:       "at_least_one_of",
				Properties: []string{"PropA", "NonExistentProp"},
				Message:    "Test message",
			},
		},
	}

	component := &hpsf.Component{
		Name: "TestComponent",
		Properties: []hpsf.Property{
			{Name: "PropA", Value: "value"},
		},
	}

	err := tc.ValidateComponent(component)
	if err == nil {
		t.Error("Expected validation to fail when referencing non-existent property")
	}
}