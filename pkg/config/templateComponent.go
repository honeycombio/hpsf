package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	y "gopkg.in/yaml.v3"
)

// This is the Go support for components read as data.
// Right now we read that data from an embedded filesystem, but
// the hope is to replace the filesystem with a database later.

// The premise here is that for each component, we have a block of YAML
// that describes it. This YAML not only contains the component's public
// list of properties, defaults, and documentation, but also contains
// Go templates that can be used to generate the actual configurations.

// A TemplatePort describes a "port" on a component. A port is a place for
// data to flow in or out of the component. A port can be either an input
// or an output, and has a datatype; at least for now, ports of different
// types cannot be connected to each other.

// Index is an optional field that can be used to indicate that this port
// should be treated as an indexed port (e.g. for use in a pipeline). The
// index values for ports should be sequential within a given component, and
// should start at 1. If the index is not specified, it is assumed to be 0.
type TemplatePort struct {
	Name      string              `yaml:"name"`
	Direction string              `yaml:"direction"`
	Type      hpsf.ConnectionType `yaml:"type"`
	Index     int                 `yaml:"index,omitempty"`
	Note      string              `yaml:"note,omitempty"`
}

// A TemplateData describes a template for generating configuration data. It's a
// deliberately simple structure, with a kind (which is the type of
// configuration data it generates), a name (which is used to identify the
// template), a format (which is the format of the data), meta (a map of extra
// component-level info) and the data itself.
type TemplateData struct {
	Kind   hpsftypes.Type
	Name   string
	Format string
	Meta   map[string]any
	Data   []any
}

type ComponentType string

const (
	ComponentTypeBase     ComponentType = "BASE"
	ComponentTypeMeta     ComponentType = "META"
	ComponentTypeTemplate ComponentType = "TEMPLATE"
)

// we need to be able to unmarshal the component style and status from YAML
// and marshal it back to YAML, and the all-caps nature of the DB constants
// is jarring and doesn't fit with the rest of the yaml styling. So we
// can marshal YAML with some case conversions.
// ensure ComponentStyle implements yaml.Marshaler and yaml.Unmarshaler
var _ y.Marshaler = (*ComponentType)(nil)
var _ y.Unmarshaler = (*ComponentType)(nil)

func (c *ComponentType) UnmarshalYAML(value *y.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	cs := ComponentType(strings.ToUpper(s))
	switch cs {
	case ComponentTypeBase, ComponentTypeMeta, ComponentTypeTemplate:
		*c = cs
		return nil
	default:
		return fmt.Errorf("invalid component style: %s", s)
	}
}

func (c ComponentType) MarshalYAML() (any, error) {
	return strings.ToLower(string(c)), nil
}

type ComponentStatus string

const (
	ComponentStatusAlpha       ComponentStatus = "ALPHA"
	ComponentStatusArchived    ComponentStatus = "ARCHIVED"
	ComponentStatusDeprecated  ComponentStatus = "DEPRECATED"
	ComponentStatusDevelopment ComponentStatus = "DEVELOPMENT"
	ComponentStatusStable      ComponentStatus = "STABLE"
)

var _ y.Marshaler = (*ComponentStatus)(nil)
var _ y.Unmarshaler = (*ComponentStatus)(nil)

func (c *ComponentStatus) UnmarshalYAML(value *y.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	cs := ComponentStatus(strings.ToUpper(s))
	switch cs {
	case ComponentStatusAlpha, ComponentStatusArchived, ComponentStatusDeprecated,
		ComponentStatusDevelopment, ComponentStatusStable:
		*c = cs
		return nil
	default:
		return fmt.Errorf("invalid component status: %s", s)
	}
}

func (c ComponentStatus) MarshalYAML() (any, error) {
	return strings.ToLower(string(c)), nil
}

// A TemplateComponent is a component that can be described with a template.
// We're hoping that most components will be described this way, so that we
// can store most templates in a database and not have to change the code when
// we add new components.
//
// A few notes on the fields:
//   - Kind is the kind of component, e.g. "TraceGRPC", "LogHTTP", etc. The combination of
//     Kind and Version is used to uniquely identify a component.
//   - Version is the version of the component as a semver string.
//   - Minimum is the minimum version of the artifact that the component supports as a semver string.
//   - Maximum is the maximum version of the artifact that the component supports as a semver string.
//   - Name is the name of the component. In a templateComponent, it is used to suggest a name that the
//     end user might want to call the component. It is not used to identify the component in a template,
//     but is used to identify the component in the UI.
//   - Style is currently a string, will be used to help the frontend figure out how to display the component.
//     It will likely become some sort of enum, but for now we don't know what the values will be.
//   - Type is the generalized type of component for broad classification - Base, Meta, or Template.
//   - Status is the development status of the component.
//   - User is only used for templating, but it needs to be exported, so its yaml tag is set to "-"
//   - collName is the name of the OTel collector component that this component is associated with; it may
//     be empty if the component is not associated with a collector. We need to store it in this data type
//     because it's used in the template rendering, but it's not part of the component itself (it's specified
//     in the template metadata).
type TemplateComponent struct {
	Kind        string             `yaml:"kind"`
	Version     string             `yaml:"version"`
	Minimum     string             `yaml:"minimum"`
	Maximum     string             `yaml:"maximum"`
	Name        string             `yaml:"name"`
	Logo        string             `yaml:"logo,omitempty"`
	Summary     string             `yaml:"summary,omitempty"`
	Description string             `yaml:"description,omitempty"`
	Comment     string             `yaml:"comment,omitempty"`
	Tags        []string           `yaml:"tags,omitempty"`
	Type        ComponentType      `yaml:"type,omitempty"`
	Style       string             `yaml:"style,omitempty"`
	Status      ComponentStatus    `yaml:"status,omitempty"`
	Metadata    map[string]string  `yaml:"metadata,omitempty"`
	Ports       []TemplatePort     `yaml:"ports,omitempty"`
	Properties  []TemplateProperty `yaml:"properties,omitempty"`
	Validations []string           `yaml:"validations,omitempty"`
	Templates   []TemplateData     `yaml:"templates,omitempty"`
	User        map[string]any     `yaml:"-"`
	hpsf        *hpsf.Component    // the component from the hpsf document
	connections []*hpsf.Connection
	collName    string
}

// SetHPSF stores the original component's details and may modify their contents. To
// prevent the original being modified, the argument here should never be changed to a pointer.
func (t *TemplateComponent) SetHPSF(c *hpsf.Component) {
	t.hpsf = c
}

// HProps is a template helper that gets a map of all properties specified in the hpsf document.
func (t *TemplateComponent) HProps() map[string]any {
	props := make(map[string]any)
	if t.hpsf != nil {
		for _, name := range t.hpsf.GetPropertyNames() {
			p := t.hpsf.GetProperty(name)
			props[p.Name] = p.Value
		}
	}
	return props
}

// Props is a helper for templates that gets all properties in a template
// component as a map. It's mainly used to look up property defaults.
func (t *TemplateComponent) Props() map[string]TemplateProperty {
	props := make(map[string]TemplateProperty)
	for _, prop := range t.Properties {
		props[prop.Name] = prop
	}
	return props
}

// Values is a helper for templates to get the data values that are available in
// the component including those specified as defaults and the environment. This
// composes HProps, User, and Properties into a single map that can be used in
// templates. You can still use them individually for special cases.
func (t *TemplateComponent) Values() map[string]any {
	result := make(map[string]any)
	for k, v := range t.HProps() {
		if !_isZeroValue(v) {
			// we only want to include non-zero values in the result
			result[k] = v
		}
	}
	for k, v := range t.User {
		if !_isZeroValue(v) {
			// don't overwrite existing values
			if _, exists := result[k]; exists {
				continue
			}
			result[k] = v
		}
	}
	for _, prop := range t.Properties {
		// don't overwrite existing values
		if _, exists := result[prop.Name]; exists {
			continue
		}
		result[prop.Name] = prop.Default
	}
	return result
}

func (t *TemplateComponent) ComponentName() string {
	if t.collName != "" {
		return t.collName + "/" + t.hpsf.GetSafeName()
	}
	return t.Name
}

// ConnectsUsingAppropriateType tests if this component has a connection with this signal type
func (t *TemplateComponent) ConnectsUsingAppropriateType(connType hpsf.ConnectionType) bool {
	for _, conn := range t.connections {
		if conn.Source.Type == connType || conn.Destination.Type == connType {
			return true
		}
	}
	return false
}

// GetPort returns the port with the given name, or nil if it doesn't exist
func (t *TemplateComponent) GetPort(name string) *TemplatePort {
	for _, port := range t.Ports {
		if port.Name == name {
			return &port
		}
	}
	return nil
}

func (t *TemplateComponent) GetPortIndex(name string) int {
	// Returns the index for a given port name, or 0 if it's unspecified.
	// This implies that indices start at 1.
	for _, port := range t.Ports {
		if port.Name == name {
			return port.Index
		}
	}
	return 0
}

// // ensure that TemplateComponent implements Component
var _ Component = (*TemplateComponent)(nil)

func (t *TemplateComponent) GenerateConfig(cfgType hpsftypes.Type, pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// we have to find a template with the kind of the config; if it
	// doesn't exist, we return an error

	// we might have more than one template for the same config type,
	// so we need to generate all of them and return the merged result.
	generatedTemplates := make([]tmpl.TemplateConfig, 0)

	for _, template := range t.Templates {
		if template.Kind == cfgType {
			switch template.Format {
			case "dotted":
				dct, err := buildDottedConfigTemplate(template.Data)
				if err != nil {
					return nil, fmt.Errorf("error %w building dotted config template for %s",
						err, t.Kind)
				}
				tmpl, err := t.generateDottedConfig(dct, userdata)
				if err != nil {
					return nil, err
				}
				generatedTemplates = append(generatedTemplates, tmpl)
			case "collector":
				ct, err := buildCollectorTemplate(template)
				if err != nil {
					return nil, fmt.Errorf("error %w building collector template for %s",
						err, t.Kind)
				}
				tmpl, err := t.generateCollectorConfig(ct, pipeline, userdata)
				if err != nil {
					return nil, err
				}
				generatedTemplates = append(generatedTemplates, tmpl)
			case "rules":
				// a rules template expects the metadata to include environment
				// information.
				if pipeline.ConnType != hpsf.CTYPE_SAMPLE {
					continue // rules templates are only for sampling pipelines
				}
				rt, err := buildRulesTemplate(template)
				if err != nil {
					return nil, fmt.Errorf("error %w building rules template for %s",
						err, t.Kind)
				}

				// Determine the pipeline index based on the SamplingSequencer's port
				// This index will be propagated through the merge chain
				index := 0
				if t.Style == "startsampling" {
					// For SamplingSequencer, determine index from the connection leading to it
					conn := pipeline.GetConnectionLeadingTo(t.hpsf.GetSafeName())
					if conn != nil {
						index = t.GetPortIndex(conn.Source.PortName)
					}
				} else if len(pipeline.Connections) > 0 {
					// For downstream components, find the SamplingSequencer in
					// the path (it's always the first component in the
					// pipeline) and determine the index from its output port
					firstConn := pipeline.Connections[0]
					// Use GetPortIndex to determine the index from the port name
					index = t.GetPortIndex(firstConn.Source.PortName)
				}

				rmt, err := tmpl.RMTFromStyle(t.Style)
				if err != nil {
					return nil, fmt.Errorf("error %w getting RulesComponentType from style %s for %s",
						err, t.Style, t.Kind)
				}
				tmpl, err := t.generateRulesConfig(rt, rmt, index, userdata)
				if err != nil {
					return nil, err
				}
				generatedTemplates = append(generatedTemplates, tmpl)
			default:
				return nil, fmt.Errorf("unknown template format %s", template.Format)
			}
		}
	}

	if len(generatedTemplates) == 0 {
		return nil, nil // no templates found for this config type
	}

	ret := generatedTemplates[0]
	for _, tmpl := range generatedTemplates[1:] {
		if err := ret.Merge(tmpl); err != nil {
			return nil, fmt.Errorf("failed to merge template: %w", err)
		}
	}

	return ret, nil
}

func (t *TemplateComponent) AddConnection(conn *hpsf.Connection) {
	t.connections = append(t.connections, conn)
}

func (t *TemplateComponent) expandTemplateVariable(tmplText string, userdata map[string]any) (string, error) {
	if tmplText == "" || !strings.Contains(tmplText, "{{") {
		return tmplText, nil
	}
	tmpl, err := template.New("template").Funcs(helpers()).Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("error %w parsing template", err)
	}

	t.User = userdata
	var b bytes.Buffer
	err = tmpl.Execute(&b, t)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}
	return b.String(), nil
}

// undecorate removes type decorations from strings and returns the desired type.
// These decorations were placed there by the functions in helpers.go.
// Since everything that comes out of a Go template is a string, for things that
// needed to not be strings, we flagged them with a decoration indicating the
// desired type. Now we need to do some extra work to make sure that we return
// the indicated type. If it can't be converted to the desired type, we return
// the string as is.
func undecorate(s string) any {
	switch {
	case strings.HasPrefix(s, IntPrefix):
		s = strings.TrimPrefix(s, IntPrefix)
		i, err := strconv.Atoi(s)
		if err == nil {
			return i
		}
	case strings.HasPrefix(s, BoolPrefix):
		s = strings.TrimPrefix(s, BoolPrefix)
		b, err := strconv.ParseBool(s)
		if err == nil {
			return b
		}
	case strings.HasPrefix(s, FloatPrefix):
		s = strings.TrimPrefix(s, FloatPrefix)
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f
		}
	case strings.HasPrefix(s, ArrPrefix):
		s = strings.TrimPrefix(s, ArrPrefix)
		items := strings.Split(s, FieldSeparator)
		// we need to trim the spaces from the items and we don't want blanks
		// in the array
		var arr []string
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item != "" {
				arr = append(arr, item)
			}
		}
		return arr
	case strings.HasPrefix(s, MapPrefix):
		s = strings.TrimPrefix(s, MapPrefix)
		// s is encoded as a JSON map, so we need to decode it
		var m map[string]any
		// we ignore the error here because the input string
		// was marshaled by us and we know it's valid JSON,
		// and there's nothing we can do with it anyway.
		json.Unmarshal([]byte(s), &m)
		return m
	}
	return s
}

func (t *TemplateComponent) applyTemplate(tmplVal any, userdata map[string]any) (any, error) {
	switch k := tmplVal.(type) {
	case string:
		if !strings.Contains(k, "{{") {
			return k, nil
		}

		// expand the template, but before we return it we need to check if
		// it needs extra handling
		value, err := t.expandTemplateVariable(k, userdata)
		if err != nil {
			return nil, err
		}

		result := undecorate(value)
		return result, nil
	// right now this is dealing with nop receiver/exporter case
	case map[string]string:
		return k, nil
	case []string:
		for i, v := range k {
			// we need to expand the template for each value in the array
			expanded, err := t.expandTemplateVariable(v, userdata)
			if err != nil {
				return nil, err
			}
			k[i] = expanded
		}
		return k, nil
	case int:
		return k, nil
	case float64:
		return k, nil
	case bool:
		return k, nil
	default:
		return "", fmt.Errorf("invalid templated variable type %T", k)
	}
}

// this is where we do the actual work of generating the config; this thing knows about
// the structure of the collector config and how to fill it in
func (t *TemplateComponent) generateCollectorConfig(ct collectorTemplate, pipeline hpsf.PathWithConnections, userdata map[string]any) (*tmpl.CollectorConfig, error) {
	// we have to fill in the template with the default values
	// and the values from the properties
	t.collName = ct.collectorComponentName
	config := tmpl.NewCollectorConfig()
	sectionOrder := []string{"receivers", "processors", "exporters", "extensions"}
	for _, section := range sectionOrder {
		for _, signalType := range hpsf.CollectorSignalTypes {
			if pipeline.ConnType != signalType {
				continue // skip this signal type if it doesn't match the pipeline
			}
			// if this template doesn't have a connection for this signal type, skip it
			if !t.ConnectsUsingAppropriateType(signalType) {
				continue
			}
			svcKey := fmt.Sprintf("pipelines.%s/%s.%s", signalType.AsCollectorSignalType(), pipeline.GetID(), section)
			for _, kv := range ct.kvs[section] {
				if kv.suppressIf != "" {
					// if the suppress_if condition is met, we skip this key
					condition, err := t.applyTemplate(kv.suppressIf, userdata)
					if err != nil {
						return nil, err
					}
					if condition == "true" {
						continue
					}
				}
				config.Set("service", svcKey, []string{t.ComponentName()})
				key, err := t.expandTemplateVariable(kv.key, userdata)
				if err != nil {
					return nil, err
				}
				value, err := t.applyTemplate(kv.value, userdata)
				if err != nil {
					return nil, err
				}
				config.Set(section, key, value)
			}
		}
	}
	return config, nil
}

func (t *TemplateComponent) AsYAML() (string, error) {
	// this is a mechanism to marshal the template component to YAML
	data, err := y.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("error marshaling template component to YAML: %w", err)
	}
	return string(data), nil
}

// Validate executes component validation rules for a TemplateComponent
// using the provided hpsf.Component data. It returns an error if any validation fails.
func (t *TemplateComponent) Validate(component *hpsf.Component) error {
	if len(t.Validations) == 0 {
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
	for _, validationStr := range t.Validations {
		if err := t.executeComponentValidation(validationStr, propertyValues, component.Name); err != nil {
			return err
		}
	}

	return nil
}

// parseComponentValidation parses a component validation string like "at_least_one_of(PropA, PropB, PropC)"
func parseComponentValidation(validationStr string) (validationType string, properties []string, conditionProperty string, conditionValue any, err error) {
	// Parse validation strings in formats like:
	// "at_least_one_of(PropA, PropB, PropC)"
	// "conditional_require_together(PropA, PropB | when EnableTLS=true)"

	valpat := regexp.MustCompile(`^(\w+)\(([^)]+)\)$`)
	argpat := regexp.MustCompile(`[\t ]*,[\t ]*`)

	matches := valpat.FindStringSubmatch(strings.TrimSpace(validationStr))
	if len(matches) != 3 {
		return "", nil, "", nil, fmt.Errorf("invalid component validation format: %s", validationStr)
	}

	validationType = matches[1]
	argsStr := matches[2]

	// Check for conditional validation format: "PropA, PropB | when ConditionProp=value"
	if strings.Contains(argsStr, " | when ") {
		parts := strings.Split(argsStr, " | when ")
		if len(parts) != 2 {
			return "", nil, "", nil, fmt.Errorf("invalid conditional validation format: %s", validationStr)
		}

		// Parse properties
		propsPart := strings.TrimSpace(parts[0])
		if propsPart != "" {
			properties = argpat.Split(propsPart, -1)
			for i, prop := range properties {
				properties[i] = strings.TrimSpace(prop)
			}
		}

		// Parse condition
		conditionPart := strings.TrimSpace(parts[1])
		conditionPat := regexp.MustCompile(`^(\w+)\s*=\s*(.+)$`)
		condMatches := conditionPat.FindStringSubmatch(conditionPart)
		if len(condMatches) != 3 {
			return "", nil, "", nil, fmt.Errorf("invalid condition format: %s", conditionPart)
		}

		conditionProperty = strings.TrimSpace(condMatches[1])
		conditionValueStr := strings.TrimSpace(condMatches[2])

		// Parse condition value (try bool, then string)
		if conditionValueStr == "true" {
			conditionValue = true
		} else if conditionValueStr == "false" {
			conditionValue = false
		} else {
			conditionValue = conditionValueStr
		}
	} else {
		// Simple validation - just parse properties
		properties = argpat.Split(argsStr, -1)
		for i, prop := range properties {
			properties[i] = strings.TrimSpace(prop)
		}
	}

	return validationType, properties, conditionProperty, conditionValue, nil
}

// executeComponentValidation runs a single component validation rule
func (t *TemplateComponent) executeComponentValidation(validationStr string, propertyValues map[string]any, componentName string) error {
	validationType, properties, conditionProperty, conditionValue, err := parseComponentValidation(validationStr)
	if err != nil {
		return hpsf.NewError("failed to parse component validation: " + err.Error()).WithComponent(componentName)
	}

	// Validate that all referenced properties exist
	for _, propName := range properties {
		if !t.propertyExists(propName) {
			return hpsf.NewError("component validation references unknown property: " + propName).
				WithComponent(componentName)
		}
	}

	// Validate condition property if specified
	if conditionProperty != "" && !t.propertyExists(conditionProperty) {
		return hpsf.NewError("component validation references unknown condition property: " + conditionProperty).
			WithComponent(componentName)
	}

	switch validationType {
	case "at_least_one_of":
		return t.validateAtLeastOneOf(properties, propertyValues, componentName)
	case "exactly_one_of":
		return t.validateExactlyOneOf(properties, propertyValues, componentName)
	case "mutually_exclusive":
		return t.validateMutuallyExclusive(properties, propertyValues, componentName)
	case "require_together":
		return t.validateRequireTogether(properties, propertyValues, componentName)
	case "conditional_require_together":
		return t.validateConditionalRequireTogether(properties, conditionProperty, conditionValue, propertyValues, componentName)
	default:
		return hpsf.NewError("unknown component validation type: " + validationType).
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

// GenerateValidationErrorMessage creates an error message based on validation type and properties
func GenerateValidationErrorMessage(validationType string, properties []string, conditionProperty string, conditionValue any) string {
	propsStr := strings.Join(properties, ", ")

	switch validationType {
	case "at_least_one_of":
		return fmt.Sprintf("At least one of [%s] must be provided", propsStr)
	case "exactly_one_of":
		return fmt.Sprintf("Exactly one of [%s] must be provided", propsStr)
	case "mutually_exclusive":
		return fmt.Sprintf("Only one of [%s] can be provided", propsStr)
	case "require_together":
		return fmt.Sprintf("Either all or none of [%s] must be provided", propsStr)
	case "conditional_require_together":
		return fmt.Sprintf("When %s is %v, all of [%s] must be provided", conditionProperty, conditionValue, propsStr)
	default:
		return fmt.Sprintf("Validation failed for properties [%s]", propsStr)
	}
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
func (t *TemplateComponent) validateAtLeastOneOf(properties []string, propertyValues map[string]any, componentName string) error {
	for _, propName := range properties {
		if value, exists := propertyValues[propName]; exists && !isPropertyEmpty(value) {
			return nil // At least one property has a non-empty value
		}
	}
	return hpsf.NewError(GenerateValidationErrorMessage("at_least_one_of", properties, "", nil)).WithComponent(componentName)
}

// validateExactlyOneOf ensures exactly one of the specified properties is non-empty
func (t *TemplateComponent) validateExactlyOneOf(properties []string, propertyValues map[string]any, componentName string) error {
	nonEmptyCount := 0
	for _, propName := range properties {
		if value, exists := propertyValues[propName]; exists && !isPropertyEmpty(value) {
			nonEmptyCount++
		}
	}
	if nonEmptyCount != 1 {
		return hpsf.NewError(GenerateValidationErrorMessage("exactly_one_of", properties, "", nil)).WithComponent(componentName)
	}
	return nil
}

// validateMutuallyExclusive ensures at most one of the specified properties is non-empty
func (t *TemplateComponent) validateMutuallyExclusive(properties []string, propertyValues map[string]any, componentName string) error {
	nonEmptyCount := 0
	for _, propName := range properties {
		if value, exists := propertyValues[propName]; exists && !isPropertyEmpty(value) {
			nonEmptyCount++
			if nonEmptyCount > 1 {
				return hpsf.NewError(GenerateValidationErrorMessage("mutually_exclusive", properties, "", nil)).WithComponent(componentName)
			}
		}
	}
	return nil
}

// validateRequireTogether ensures all properties are either all empty or all non-empty
func (t *TemplateComponent) validateRequireTogether(properties []string, propertyValues map[string]any, componentName string) error {
	hasNonEmpty := false
	hasEmpty := false

	for _, propName := range properties {
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
		return hpsf.NewError(GenerateValidationErrorMessage("require_together", properties, "", nil)).WithComponent(componentName)
	}

	return nil
}

// validateConditionalRequireTogether ensures all properties are non-empty when condition is met
func (t *TemplateComponent) validateConditionalRequireTogether(properties []string, conditionProperty string, conditionValue any, propertyValues map[string]any, componentName string) error {
	// Check if condition is met
	actualValue, exists := propertyValues[conditionProperty]
	if !exists || isPropertyEmpty(actualValue) {
		return nil // Condition not met, validation doesn't apply
	}

	// Compare condition value with expected value
	conditionMet := false
	switch expectedVal := conditionValue.(type) {
	case bool:
		if boolVal, ok := actualValue.(bool); ok && boolVal == expectedVal {
			conditionMet = true
		}
	case string:
		if strVal, ok := actualValue.(string); ok && strVal == expectedVal {
			conditionMet = true
		}
	case int:
		if intVal, ok := actualValue.(int); ok && intVal == expectedVal {
			conditionMet = true
		}
	case float64:
		if floatVal, ok := actualValue.(float64); ok && floatVal == expectedVal {
			conditionMet = true
		}
	default:
		// For other types, use string comparison
		conditionMet = fmt.Sprint(actualValue) == fmt.Sprint(expectedVal)
	}

	if !conditionMet {
		return nil // Condition not met, validation doesn't apply
	}

	// Condition is met, ensure all specified properties are non-empty
	for _, propName := range properties {
		value, exists := propertyValues[propName]
		if !exists || isPropertyEmpty(value) {
			return hpsf.NewError(GenerateValidationErrorMessage("conditional_require_together", properties, conditionProperty, conditionValue)).WithComponent(componentName)
		}
	}

	return nil
}
