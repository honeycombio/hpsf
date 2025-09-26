package generator

import (
	"fmt"
	"sort"

	"github.com/honeycombio/hpsf/pkg/hpsf"
	"gopkg.in/yaml.v3"
)

// RefineryRules represents the sampling rules configuration
type RefineryRules struct {
	RulesVersion int                               `yaml:"RulesVersion"`
	Samplers     map[string]map[string]interface{} `yaml:"Samplers"`
}

// RuleCondition represents a condition in a rules-based sampler
type RuleCondition struct {
	Field    string      `yaml:"Field,omitempty"`
	Fields   []string    `yaml:"Fields,omitempty"`
	Operator string      `yaml:"Operator"`
	Value    interface{} `yaml:"Value,omitempty"`
	Datatype string      `yaml:"Datatype,omitempty"`
}

// SamplingRule represents a single sampling rule
type SamplingRule struct {
	Name       string          `yaml:"Name"`
	SampleRate int             `yaml:"SampleRate,omitempty"`
	Drop       bool            `yaml:"Drop,omitempty"`
	Scope      string          `yaml:"Scope,omitempty"`
	Conditions []RuleCondition `yaml:"Conditions,omitempty"`
}

// RulesBasedSampler represents a rules-based sampler configuration
type RulesBasedSampler struct {
	Rules []SamplingRule `yaml:"Rules"`
}

// Generator provides functionality to convert Refinery configurations to HPSF workflows
type Generator struct {
	componentCounter int
}

// NewGenerator creates a new Generator instance
func NewGenerator() *Generator {
	return &Generator{componentCounter: 1}
}

// GenerateWorkflow creates an HPSF workflow from Refinery rules
func (g *Generator) GenerateWorkflow(rulesData []byte) (*hpsf.HPSF, error) {
	// Parse the Refinery rules
	var rules RefineryRules
	if err := yaml.Unmarshal(rulesData, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse Refinery rules: %w", err)
	}

	// Create the HPSF workflow
	workflow := &hpsf.HPSF{
		Kind:        "HPSF",
		Version:     "v1",
		Name:        "Generated_Refinery_Workflow",
		Summary:     "Generated from Refinery sampling rules",
		Description: "HPSF workflow automatically generated from Refinery sampling rules",
		Components:  []*hpsf.Component{},
		Connections: []*hpsf.Connection{},
	}

	// Generate components and connections
	receiverComponent := g.generateOTelReceiver()
	workflow.Components = append(workflow.Components, receiverComponent)

	startSamplingComponent := g.generateStartSampling()
	workflow.Components = append(workflow.Components, startSamplingComponent)

	// Connect OTel Receiver to Start Sampling
	workflow.Connections = append(workflow.Connections, g.createConnection(
		receiverComponent.Name, "Traces", hpsf.CTYPE_TRACES,
		startSamplingComponent.Name, "Traces", hpsf.CTYPE_TRACES,
	))
	workflow.Connections = append(workflow.Connections, g.createConnection(
		receiverComponent.Name, "Logs", hpsf.CTYPE_LOGS,
		startSamplingComponent.Name, "Logs", hpsf.CTYPE_LOGS,
	))

	// Generate sampling components from rules
	ruleComponents, ruleConnections, err := g.generateSamplingComponents(rules, startSamplingComponent.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sampling components: %w", err)
	}

	workflow.Components = append(workflow.Components, ruleComponents...)
	workflow.Connections = append(workflow.Connections, ruleConnections...)

	// Always generate Honeycomb exporter
	exporterComponent := g.generateHoneycombExporter()
	workflow.Components = append(workflow.Components, exporterComponent)

	// Connect all samplers to the Honeycomb exporter
	for _, component := range ruleComponents {
		if g.isSamplerComponent(component.Kind) {
			workflow.Connections = append(workflow.Connections, g.createConnection(
				component.Name, "Events", hpsf.CTYPE_HONEY,
				exporterComponent.Name, "Events", hpsf.CTYPE_HONEY,
			))
		}
	}

	return workflow, nil
}

// generateOTelReceiver creates an OpenTelemetry Collector Receiver component
func (g *Generator) generateOTelReceiver() *hpsf.Component {
	return &hpsf.Component{
		Name: g.getNextComponentName("OTel_Receiver"),
		Kind: "OTelReceiver",
		Properties: []hpsf.Property{
			{Name: "Host", Value: "0.0.0.0"},
			{Name: "GRPCPort", Value: 4317},
			{Name: "HTTPPort", Value: 4318},
		},
	}
}

// generateStartSampling creates a Start Sampling component
func (g *Generator) generateStartSampling() *hpsf.Component {
	return &hpsf.Component{
		Name: g.getNextComponentName("Start_Sampling"),
		Kind: "SamplingSequencer",
		Properties: []hpsf.Property{
			{Name: "Host", Value: "refinery"},
			{Name: "Port", Value: 8080},
		},
	}
}

// generateSamplingComponents creates condition and sampler components based on rules
func (g *Generator) generateSamplingComponents(rules RefineryRules, startSamplingName string) ([]*hpsf.Component, []*hpsf.Connection, error) {
	var components []*hpsf.Component
	var connections []*hpsf.Connection

	ruleIndex := 0

	// Process each environment's samplers
	for env, samplers := range rules.Samplers {
		if env != "__default__" {
			// Skip non-default environments for now
			continue
		}

		// Sort sampler types for deterministic ordering
		samplerTypes := make([]string, 0, len(samplers))
		for samplerType := range samplers {
			samplerTypes = append(samplerTypes, samplerType)
		}
		sort.Strings(samplerTypes)

		for _, samplerType := range samplerTypes {
			samplerConfig := samplers[samplerType]
			switch samplerType {
			case "RulesBasedSampler":
				ruleComponents, ruleConnections := g.generateRulesBasedSampler(samplerConfig, startSamplingName, &ruleIndex)
				components = append(components, ruleComponents...)
				connections = append(connections, ruleConnections...)

			case "DeterministicSampler":
				samplerComponent := g.generateDeterministicSampler(samplerConfig)
				components = append(components, samplerComponent)

				// Connect from Start Sampling to this sampler
				connections = append(connections, g.createConnection(
					startSamplingName, fmt.Sprintf("Rule %d", ruleIndex+1), hpsf.CTYPE_SAMPLE,
					samplerComponent.Name, "Sample", hpsf.CTYPE_SAMPLE,
				))
				ruleIndex++

			case "EMAThroughputSampler":
				samplerComponent := g.generateEMAThroughputSampler(samplerConfig)
				components = append(components, samplerComponent)

				// Connect from Start Sampling to this sampler
				connections = append(connections, g.createConnection(
					startSamplingName, fmt.Sprintf("Rule %d", ruleIndex+1), hpsf.CTYPE_SAMPLE,
					samplerComponent.Name, "Sample", hpsf.CTYPE_SAMPLE,
				))
				ruleIndex++

			case "EMADynamicSampler":
				samplerComponent := g.generateEMADynamicSampler(samplerConfig)
				components = append(components, samplerComponent)

				// Connect from Start Sampling to this sampler
				connections = append(connections, g.createConnection(
					startSamplingName, fmt.Sprintf("Rule %d", ruleIndex+1), hpsf.CTYPE_SAMPLE,
					samplerComponent.Name, "Sample", hpsf.CTYPE_SAMPLE,
				))
				ruleIndex++
			}
		}
	}

	return components, connections, nil
}

// generateRulesBasedSampler creates components for rules-based sampling
func (g *Generator) generateRulesBasedSampler(samplerConfig interface{}, startSamplingName string, ruleIndex *int) ([]*hpsf.Component, []*hpsf.Connection) {
	var components []*hpsf.Component
	var connections []*hpsf.Connection

	// Convert the sampler config to RulesBasedSampler struct
	configBytes, _ := yaml.Marshal(samplerConfig)
	var rulesBasedSampler RulesBasedSampler
	yaml.Unmarshal(configBytes, &rulesBasedSampler)

	var previousComponent string = startSamplingName
	var previousPort string

	for _, rule := range rulesBasedSampler.Rules {
		*ruleIndex++
		previousPort = fmt.Sprintf("Rule %d", *ruleIndex)

		// Generate condition components for this rule
		if len(rule.Conditions) > 0 {
			conditionComponent := g.generateConditionComponent(rule.Conditions, rule.Name)
			components = append(components, conditionComponent)

			// Connect from previous component to condition
			connections = append(connections, g.createConnection(
				previousComponent, previousPort, hpsf.CTYPE_SAMPLE,
				conditionComponent.Name, "Sample", hpsf.CTYPE_SAMPLE,
			))

			previousComponent = conditionComponent.Name
			previousPort = "Sample"
		}

		// Generate sampler component for this rule
		if rule.Drop {
			// Generate Dropper component
			dropperComponent := g.generateDropperComponent(rule.Name)
			components = append(components, dropperComponent)

			connections = append(connections, g.createConnection(
				previousComponent, previousPort, hpsf.CTYPE_SAMPLE,
				dropperComponent.Name, "Sample", hpsf.CTYPE_SAMPLE,
			))
		} else {
			// Generate appropriate sampler component
			samplerComponent := g.generateRuleSampler(rule)
			components = append(components, samplerComponent)

			connections = append(connections, g.createConnection(
				previousComponent, previousPort, hpsf.CTYPE_SAMPLE,
				samplerComponent.Name, "Sample", hpsf.CTYPE_SAMPLE,
			))
		}
	}

	return components, connections
}

// Helper methods for component generation
func (g *Generator) generateDeterministicSampler(config interface{}) *hpsf.Component {
	sampleRate := 100 // default

	// Handle both map types that can come from YAML unmarshaling
	switch configMap := config.(type) {
	case map[interface{}]interface{}:
		if rate, ok := configMap["SampleRate"]; ok {
			if rateInt, ok := rate.(int); ok {
				sampleRate = rateInt
			}
		}
	case map[string]interface{}:
		if rate, ok := configMap["SampleRate"]; ok {
			if rateInt, ok := rate.(int); ok {
				sampleRate = rateInt
			}
		}
	}

	return &hpsf.Component{
		Name: g.getNextComponentName("Deterministic_Sampler"),
		Kind: "DeterministicSampler",
		Properties: []hpsf.Property{
			{Name: "SampleRate", Value: sampleRate},
		},
	}
}

// generateEMAThroughputSampler creates an EMA Throughput sampler component
func (g *Generator) generateEMAThroughputSampler(config interface{}) *hpsf.Component {
	component := &hpsf.Component{
		Name:       g.getNextComponentName("EMA_Throughput_Sampler"),
		Kind:       "EMAThroughputSampler",
		Properties: []hpsf.Property{},
	}

	// Handle both map types that can come from YAML unmarshaling
	switch configMap := config.(type) {
	case map[interface{}]interface{}:
		if goalThroughput, ok := configMap["GoalThroughputPerSec"]; ok {
			component.Properties = append(component.Properties, hpsf.Property{
				Name: "GoalThroughputPerSec", Value: goalThroughput,
			})
		}
		if fieldList, ok := configMap["FieldList"]; ok {
			if fields, ok := fieldList.([]interface{}); ok {
				var stringFields []string
				for _, field := range fields {
					if str, ok := field.(string); ok {
						stringFields = append(stringFields, str)
					}
				}
				component.Properties = append(component.Properties, hpsf.Property{
					Name: "FieldList", Value: stringFields,
				})
			}
		}
	case map[string]interface{}:
		if goalThroughput, ok := configMap["GoalThroughputPerSec"]; ok {
			component.Properties = append(component.Properties, hpsf.Property{
				Name: "GoalThroughputPerSec", Value: goalThroughput,
			})
		}
		if fieldList, ok := configMap["FieldList"]; ok {
			if fields, ok := fieldList.([]interface{}); ok {
				var stringFields []string
				for _, field := range fields {
					if str, ok := field.(string); ok {
						stringFields = append(stringFields, str)
					}
				}
				component.Properties = append(component.Properties, hpsf.Property{
					Name: "FieldList", Value: stringFields,
				})
			}
		}
	}

	return component
}

// generateEMADynamicSampler creates an EMA Dynamic sampler component
func (g *Generator) generateEMADynamicSampler(config interface{}) *hpsf.Component {
	component := &hpsf.Component{
		Name:       g.getNextComponentName("EMA_Dynamic_Sampler"),
		Kind:       "EMADynamicSampler",
		Properties: []hpsf.Property{},
	}

	// Handle both map types that can come from YAML unmarshaling
	switch configMap := config.(type) {
	case map[interface{}]interface{}:
		if goalSampleRate, ok := configMap["GoalSampleRate"]; ok {
			component.Properties = append(component.Properties, hpsf.Property{
				Name: "GoalSampleRate", Value: goalSampleRate,
			})
		}
		if fieldList, ok := configMap["FieldList"]; ok {
			if fields, ok := fieldList.([]interface{}); ok {
				var stringFields []string
				for _, field := range fields {
					if str, ok := field.(string); ok {
						stringFields = append(stringFields, str)
					}
				}
				component.Properties = append(component.Properties, hpsf.Property{
					Name: "FieldList", Value: stringFields,
				})
			}
		}
	case map[string]interface{}:
		if goalSampleRate, ok := configMap["GoalSampleRate"]; ok {
			component.Properties = append(component.Properties, hpsf.Property{
				Name: "GoalSampleRate", Value: goalSampleRate,
			})
		}
		if fieldList, ok := configMap["FieldList"]; ok {
			if fields, ok := fieldList.([]interface{}); ok {
				var stringFields []string
				for _, field := range fields {
					if str, ok := field.(string); ok {
						stringFields = append(stringFields, str)
					}
				}
				component.Properties = append(component.Properties, hpsf.Property{
					Name: "FieldList", Value: stringFields,
				})
			}
		}
	}

	return component
}

func (g *Generator) generateConditionComponent(conditions []RuleCondition, ruleName string) *hpsf.Component {
	// For simplicity, we'll generate a single condition component for the first condition
	// In a more sophisticated implementation, we might handle multiple conditions
	if len(conditions) == 0 {
		return nil
	}

	condition := conditions[0]
	var kind string
	var properties []hpsf.Property

	// Determine the appropriate condition component kind based on the condition
	switch condition.Operator {
	case "exists", "does-not-exist":
		kind = "FieldExistsCondition"
		field := condition.Field
		if len(condition.Fields) > 0 {
			field = condition.Fields[0]
		}
		properties = []hpsf.Property{
			{Name: "Field", Value: field},
			{Name: "ShouldExist", Value: condition.Operator == "exists"},
		}

	case ">=", ">", "<=", "<", "=", "!=":
		switch condition.Datatype {
		case "int":
			kind = "CompareIntegerFieldCondition"
		case "float":
			kind = "CompareDecimalFieldCondition"
		case "bool":
			kind = "CompareBoolFieldCondition"
		default:
			kind = "CompareStringFieldCondition"
		}

		field := condition.Field
		if len(condition.Fields) > 0 {
			field = condition.Fields[0]
		}

		properties = []hpsf.Property{
			{Name: "Field", Value: field},
			{Name: "Operator", Value: condition.Operator},
			{Name: "Value", Value: condition.Value},
		}

	case "contains":
		kind = "FieldContainsCondition"
		field := condition.Field
		if len(condition.Fields) > 0 {
			field = condition.Fields[0]
		}
		properties = []hpsf.Property{
			{Name: "Field", Value: field},
			{Name: "Value", Value: condition.Value},
		}

	case "starts-with":
		kind = "FieldStartsWithCondition"
		field := condition.Field
		if len(condition.Fields) > 0 {
			field = condition.Fields[0]
		}
		properties = []hpsf.Property{
			{Name: "Field", Value: field},
			{Name: "Value", Value: condition.Value},
		}

	default:
		// Fallback to a generic condition
		kind = "FieldExistsCondition"
		field := condition.Field
		if len(condition.Fields) > 0 {
			field = condition.Fields[0]
		}
		properties = []hpsf.Property{
			{Name: "Field", Value: field},
			{Name: "ShouldExist", Value: true},
		}
	}

	return &hpsf.Component{
		Name:       g.getNextComponentName(fmt.Sprintf("Condition_%s", ruleName)),
		Kind:       kind,
		Properties: properties,
	}
}

func (g *Generator) generateDropperComponent(ruleName string) *hpsf.Component {
	return &hpsf.Component{
		Name: g.getNextComponentName(fmt.Sprintf("Drop_%s", ruleName)),
		Kind: "Dropper",
	}
}

func (g *Generator) generateRuleSampler(rule SamplingRule) *hpsf.Component {
	return &hpsf.Component{
		Name: g.getNextComponentName(fmt.Sprintf("Sample_%s", rule.Name)),
		Kind: "DeterministicSampler",
		Properties: []hpsf.Property{
			{Name: "SampleRate", Value: rule.SampleRate},
		},
	}
}

func (g *Generator) generateHoneycombExporter() *hpsf.Component {
	return &hpsf.Component{
		Name: g.getNextComponentName("Send_to_Honeycomb"),
		Kind: "HoneycombExporter",
		Properties: []hpsf.Property{
			{Name: "APIEndpoint", Value: "api.honeycomb.io"},
		},
	}
}

// Helper methods
func (g *Generator) getNextComponentName(baseName string) string {
	name := fmt.Sprintf("%s_%d", baseName, g.componentCounter)
	g.componentCounter++
	return name
}

func (g *Generator) createConnection(srcComp, srcPort string, srcType hpsf.ConnectionType, destComp, destPort string, destType hpsf.ConnectionType) *hpsf.Connection {
	return &hpsf.Connection{
		Source: hpsf.ConnectionPort{
			Component: srcComp,
			PortName:  srcPort,
			Type:      srcType,
		},
		Destination: hpsf.ConnectionPort{
			Component: destComp,
			PortName:  destPort,
			Type:      destType,
		},
	}
}

func (g *Generator) isSamplerComponent(kind string) bool {
	samplerKinds := []string{
		"DeterministicSampler",
		"EMAThroughputSampler",
		"EMADynamicSampler",
		"KeepAllSampler",
	}

	for _, samplerKind := range samplerKinds {
		if kind == samplerKind {
			return true
		}
	}
	return false
}
