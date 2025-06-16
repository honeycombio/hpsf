package translator

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropertyValidation(t *testing.T) {

	t.Run("Validate should error if a noblanks string is not supplied and there is no default", func(t *testing.T) {

		testComponent := config.TemplateComponent{
			Kind: "NoBlanksComponent",
			Properties: []config.TemplateProperty{
				{Name: "Mandatory", Type: hpsf.PTYPE_STRING, Validations: []string{"noblanks"}},
			},
		}

		translator := NewEmptyTranslator()

		translator.InstallComponents(map[string]config.TemplateComponent{
			testComponent.Kind: testComponent,
		})

		hpsfDocument := hpsf.HPSF{
			Components: []*hpsf.Component{
				{
					Name: "TestComponent",
					Kind: testComponent.Kind,
				},
			},
		}

		err := translator.ValidateConfig(&hpsfDocument)
		require.Error(t, err)
		require.IsType(t, validator.Result{}, err)
		validationError := err.(validator.Result)
		fieldValidationError := validationError.Details[0].(*hpsf.HPSFError)
		require.Equal(t, testComponent.Properties[0].Name, fieldValidationError.Property)
		assert.Equal(t, "failed to validate property", fieldValidationError.Reason)
		assert.Equal(t, "TestComponent", fieldValidationError.Component)
		assert.Equal(t, hpsf.ErrorSeverity("E"), fieldValidationError.Severity)
	})

	t.Run("Validate should error if a noblanks string is not supplied and there is a default", func(t *testing.T) {
		translator := NewEmptyTranslator()

		testComponent := config.TemplateComponent{
			Kind: "NoBlanksComponent",
			Properties: []config.TemplateProperty{
				{Name: "Mandatory", Type: hpsf.PTYPE_STRING, Validations: []string{"noblanks"}, Default: "default"},
			},
		}

		translator.InstallComponents(map[string]config.TemplateComponent{
			testComponent.Kind: testComponent,
		})

		hpsfDocument := hpsf.HPSF{
			Components: []*hpsf.Component{
				{
					Name: "TestComponent",
					Kind: testComponent.Kind,
				},
			},
		}

		err := translator.ValidateConfig(&hpsfDocument)
		assert.NoError(t, err)
	})

	t.Run("Validate should fail when one property passes and another does not", func(t *testing.T) {

		translator := NewEmptyTranslator()

		testComponent := config.TemplateComponent{
			Kind: "NoBlanksComponent",
			Properties: []config.TemplateProperty{
				{Name: "Mandatory", Type: hpsf.PTYPE_STRING, Validations: []string{"noblanks"}},
				{Name: "AlsoMandatory", Type: hpsf.PTYPE_STRING, Validations: []string{"noblanks"}},
			},
		}

		translator.InstallComponents(map[string]config.TemplateComponent{
			testComponent.Kind: testComponent,
		})

		hpsfDocument := hpsf.HPSF{
			Components: []*hpsf.Component{
				{
					Name: "TestComponent",
					Kind: testComponent.Kind,
					Properties: []hpsf.Property{
						{Name: "Mandatory", Value: "value"},
					},
				},
			},
		}

		err := translator.ValidateConfig(&hpsfDocument)
		require.Error(t, err)
		require.IsType(t, validator.Result{}, err)
		validationError := err.(validator.Result)
		assert.Equal(t, 1, validationError.Len())

	})
}
