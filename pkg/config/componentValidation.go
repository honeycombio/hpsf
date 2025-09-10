package config

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// ValidateComponent executes component validation rules for a TemplateComponent
// using the provided hpsf.Component data. It returns an error if any validation fails.
func (t *TemplateComponent) ValidateComponent(component *hpsf.Component) error {
	if len(t.ComponentValidations) == 0 {
		return nil
	}

	// Build a map of property values for easy lookup
	propertyValues := make(map[string]any)
	for _, prop := range component.Properties {
		propertyValues[prop.Name] = prop.Value
	}

	// Add defaults for properties not explicitly set
	for _, templateProp := range t.Properties {
		if _, exists := propertyValues[templateProp.Name]; !exists && templateProp.Default != nil {
			propertyValues[templateProp.Name] = templateProp.Default
		}
	}

	// Execute each validation rule
	for _, validation := range t.ComponentValidations {
		if err := t.executeComponentValidation(validation, propertyValues, component.Name); err != nil {
			return err
		}
	}

	return nil
}

// executeComponentValidation runs a single component validation rule
func (t *TemplateComponent) executeComponentValidation(validation ComponentValidation, propertyValues map[string]any, componentName string) error {
	// Validate that all referenced properties exist
	for _, propName := range validation.Properties {
		if !t.propertyExists(propName) {
			return hpsf.NewError("component validation references unknown property: " + propName).
				WithComponent(componentName)
		}
	}

	// Validate condition property if specified
	if validation.ConditionProperty != "" && !t.propertyExists(validation.ConditionProperty) {
		return hpsf.NewError("component validation references unknown condition property: " + validation.ConditionProperty).
			WithComponent(componentName)
	}

	switch validation.Type {
	case "at_least_one_of":
		return t.validateAtLeastOneOf(validation, propertyValues, componentName)
	case "exactly_one_of":
		return t.validateExactlyOneOf(validation, propertyValues, componentName)
	case "mutually_exclusive":
		return t.validateMutuallyExclusive(validation, propertyValues, componentName)
	case "require_together":
		return t.validateRequireTogether(validation, propertyValues, componentName)
	case "conditional_require_together":
		return t.validateConditionalRequireTogether(validation, propertyValues, componentName)
	default:
		return hpsf.NewError("unknown component validation type: " + validation.Type).
			WithComponent(componentName)
	}
}

// propertyExists checks if a property with the given name exists in the template component
func (t *TemplateComponent) propertyExists(propName string) bool {
	for _, prop := range t.Properties {
		if prop.Name == propName {
			return true
		}
	}
	return false
}

// isPropertyEmpty determines if a property value is considered "empty" according to the spec
func isPropertyEmpty(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case bool:
		return !v
	case int:
		return v == 0
	case float64:
		return v == 0.0
	case []string:
		return len(v) == 0
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		// For other types, consider nil/zero values as empty
		return value == nil
	}
}

// validateAtLeastOneOf ensures at least one of the specified properties is non-empty
func (t *TemplateComponent) validateAtLeastOneOf(validation ComponentValidation, propertyValues map[string]any, componentName string) error {
	for _, propName := range validation.Properties {
		if value, exists := propertyValues[propName]; exists && !isPropertyEmpty(value) {
			return nil // At least one property has a non-empty value
		}
	}
	return hpsf.NewError(validation.Message).WithComponent(componentName)
}

// validateExactlyOneOf ensures exactly one of the specified properties is non-empty
func (t *TemplateComponent) validateExactlyOneOf(validation ComponentValidation, propertyValues map[string]any, componentName string) error {
	nonEmptyCount := 0
	for _, propName := range validation.Properties {
		if value, exists := propertyValues[propName]; exists && !isPropertyEmpty(value) {
			nonEmptyCount++
		}
	}
	if nonEmptyCount != 1 {
		return hpsf.NewError(validation.Message).WithComponent(componentName)
	}
	return nil
}

// validateMutuallyExclusive ensures at most one of the specified properties is non-empty
func (t *TemplateComponent) validateMutuallyExclusive(validation ComponentValidation, propertyValues map[string]any, componentName string) error {
	nonEmptyCount := 0
	for _, propName := range validation.Properties {
		if value, exists := propertyValues[propName]; exists && !isPropertyEmpty(value) {
			nonEmptyCount++
			if nonEmptyCount > 1 {
				return hpsf.NewError(validation.Message).WithComponent(componentName)
			}
		}
	}
	return nil
}

// validateRequireTogether ensures all properties are either all empty or all non-empty
func (t *TemplateComponent) validateRequireTogether(validation ComponentValidation, propertyValues map[string]any, componentName string) error {
	hasNonEmpty := false
	hasEmpty := false

	for _, propName := range validation.Properties {
		value, exists := propertyValues[propName]
		isEmpty := !exists || isPropertyEmpty(value)
		
		if isEmpty {
			hasEmpty = true
		} else {
			hasNonEmpty = true
		}
	}

	// If we have both empty and non-empty properties, that's an error
	if hasEmpty && hasNonEmpty {
		return hpsf.NewError(validation.Message).WithComponent(componentName)
	}
	
	return nil
}

// validateConditionalRequireTogether ensures all properties are non-empty when condition is met
func (t *TemplateComponent) validateConditionalRequireTogether(validation ComponentValidation, propertyValues map[string]any, componentName string) error {
	// Check if condition is met
	conditionValue, exists := propertyValues[validation.ConditionProperty]
	if !exists || isPropertyEmpty(conditionValue) {
		return nil // Condition not met, validation doesn't apply
	}

	// Compare condition value with expected value
	conditionMet := false
	switch expectedVal := validation.ConditionValue.(type) {
	case bool:
		if boolVal, ok := conditionValue.(bool); ok && boolVal == expectedVal {
			conditionMet = true
		}
	case string:
		if strVal, ok := conditionValue.(string); ok && strVal == expectedVal {
			conditionMet = true
		}
	case int:
		if intVal, ok := conditionValue.(int); ok && intVal == expectedVal {
			conditionMet = true
		}
	case float64:
		if floatVal, ok := conditionValue.(float64); ok && floatVal == expectedVal {
			conditionMet = true
		}
	default:
		// For other types, use string comparison
		conditionMet = fmt.Sprint(conditionValue) == fmt.Sprint(expectedVal)
	}

	if !conditionMet {
		return nil // Condition not met, validation doesn't apply
	}

	// Condition is met, ensure all specified properties are non-empty
	for _, propName := range validation.Properties {
		value, exists := propertyValues[propName]
		if !exists || isPropertyEmpty(value) {
			return hpsf.NewError(validation.Message).WithComponent(componentName)
		}
	}

	return nil
}