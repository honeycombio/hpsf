package translator

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"sort"
	"strconv"
	"strings"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/honeycombio/hpsf/pkg/validator"
	"golang.org/x/mod/semver"
)

const LatestVersion = "latest"

// extractExpectedPipelines extracts pipeline names referenced in routing connector configuration.
// Returns a map of pipeline names (e.g., "traces/production", "logs/staging") to true.
func extractExpectedPipelines(connectorSection map[string]any) map[string]bool {
	expectedPipelines := make(map[string]bool)

	for key, value := range connectorSection {
		// Check for routing/{signaltype}.default_pipelines
		if strings.HasPrefix(key, "routing/") && strings.HasSuffix(key, ".default_pipelines") {
			if pipelines, ok := value.([]any); ok {
				for _, p := range pipelines {
					if pName, ok := p.(string); ok {
						expectedPipelines[pName] = true
					}
				}
			} else if pipelines, ok := value.([]string); ok {
				for _, pName := range pipelines {
					expectedPipelines[pName] = true
				}
			}
		}

		// Check for routing/{signaltype}.table[N].pipelines
		if strings.HasPrefix(key, "routing/") && strings.Contains(key, ".table[") && strings.HasSuffix(key, ".pipelines") {
			if pipelineList, ok := value.([]any); ok {
				for _, p := range pipelineList {
					if pName, ok := p.(string); ok {
						expectedPipelines[pName] = true
					}
				}
			} else if pipelineList, ok := value.([]string); ok {
				for _, pName := range pipelineList {
					expectedPipelines[pName] = true
				}
			}
		}
	}

	return expectedPipelines
}

// collectPipelineNames extracts all unique pipeline names from the service section.
// Returns a map of pipeline names to true.
func collectPipelineNames(serviceSection map[string]any) map[string]bool {
	pipelineNames := make(map[string]bool)

	for key := range serviceSection {
		if !strings.HasPrefix(key, "pipelines.") {
			continue
		}
		pipelinePath := key[len("pipelines."):]
		parts := strings.SplitN(pipelinePath, ".", 2)
		if len(parts) < 2 {
			continue
		}
		pipelineNames[parts[0]] = true
	}

	return pipelineNames
}

// buildExporterEnvironmentMap creates a mapping from exporter safe names to environment names.
// Reads environment directly from HoneycombExporter components' EnvironmentName property.
func buildExporterEnvironmentMap(h *hpsf.HPSF) map[string]string {
	exporterToEnvironment := make(map[string]string)

	if h == nil {
		return exporterToEnvironment
	}

	for _, hcomp := range h.Components {
		if hcomp.Kind == "HoneycombExporter" {
			safeName := hcomp.GetSafeName()
			// Get EnvironmentName property from exporter
			for _, prop := range hcomp.Properties {
				if prop.Name == "EnvironmentName" {
					if envName, ok := prop.Value.(string); ok && envName != "" {
						exporterToEnvironment[safeName] = envName
					}
					break
				}
			}
		}
	}

	return exporterToEnvironment
}

// getStringListFromAny converts []any or []string to []string, filtering out empty strings.
// Returns nil if the value is not a supported type.
func getStringListFromAny(value any) []string {
	if value == nil {
		return nil
	}

	var result []string
	switch v := value.(type) {
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, s)
			}
		}
	case []string:
		for _, s := range v {
			if s != "" {
				result = append(result, s)
			}
		}
	default:
		return nil
	}

	return result
}

// normalizeRoutingConnector normalizes routing connector names to "routing/{signaltype}" format.
func normalizeRoutingConnector(connectorName, signalType string) string {
	if connectorName == "routing" || strings.HasPrefix(connectorName, "routing_") || strings.HasPrefix(connectorName, "routing/") {
		if signalType != "" {
			return "routing/" + signalType
		}
		return "routing"
	}
	return connectorName
}

// transformIntakePipeline transforms an intake pipeline by moving routing connector from connectors to exporters.
// Returns the new name for the pipeline (e.g., "traces/intake") or empty string if no rename needed.
func transformIntakePipeline(serviceSection map[string]any, pipelineName, signalType string, connectorsList []string) string {
	connectorsKey := fmt.Sprintf("pipelines.%s.connectors", pipelineName)
	exportersKey := fmt.Sprintf("pipelines.%s.exporters", pipelineName)

	// Move connectors to exporters, normalizing routing connector names
	normalizedConnectors := make([]string, len(connectorsList))
	for i, conn := range connectorsList {
		normalizedConnectors[i] = normalizeRoutingConnector(conn, signalType)
	}
	serviceSection[exportersKey] = normalizedConnectors
	delete(serviceSection, connectorsKey)

	// Generate intake pipeline name
	if signalType != "" {
		intakeName := signalType + "/intake"
		if pipelineName != intakeName {
			return intakeName
		}
	}
	return ""
}

// inferEnvironmentFromExporters attempts to infer environment name from exporter components.
func inferEnvironmentFromExporters(exportersList []string, exporterToEnvironment map[string]string, pipelineName string) string {
	// First, check if any exporters have an environment set
	for _, exp := range exportersList {
		// Remove the exporter type prefix (e.g., "otlphttp/") to get the component name
		expParts := strings.SplitN(exp, "/", 2)
		expCompName := exp
		if len(expParts) == 2 {
			expCompName = expParts[1]
		}
		if env, ok := exporterToEnvironment[expCompName]; ok {
			return env
		}
	}

	// Try to extract from pipeline name (e.g., "traces/dev" -> "dev")
	parts := strings.SplitN(pipelineName, "/", 2)
	if len(parts) == 2 && parts[1] != "intake" {
		return parts[1]
	}

	return ""
}

// injectEnvironmentAPIKey injects environment-specific API key for exporters in a pipeline.
func injectEnvironmentAPIKey(exportersSection map[string]any, exportersList []string, envName string) {
	if envName == "" {
		return
	}

	for _, exp := range exportersList {
		headerKey := fmt.Sprintf("%s.headers.x-honeycomb-team", exp)
		existingValue, hasHeader := exportersSection[headerKey]
		// Inject if header doesn't exist or if it's set to the default value
		if !hasHeader || existingValue == "${HTP_EXPORTER_APIKEY}" {
			envVarName := fmt.Sprintf("${HTP_EXPORTER_APIKEY_%s}", strings.ToUpper(envName))
			exportersSection[headerKey] = envVarName
		}
	}
}

// transformOutputPipeline transforms an output pipeline by moving routing connector from connectors to receivers
// and injecting environment-specific API keys.
func transformOutputPipeline(serviceSection map[string]any, cc *tmpl.CollectorConfig, pipelineName, signalType string, connectors any, exportersList []string, exporterToEnvironment map[string]string) {
	receiversKey := fmt.Sprintf("pipelines.%s.receivers", pipelineName)
	connectorsKey := fmt.Sprintf("pipelines.%s.connectors", pipelineName)

	// Move connectors to receivers
	serviceSection[receiversKey] = connectors
	delete(serviceSection, connectorsKey)

	// Infer environment and inject API keys
	if signalType != "" {
		envName := inferEnvironmentFromExporters(exportersList, exporterToEnvironment, pipelineName)
		if envName != "" {
			exportersSection, hasExportersSection := cc.Sections["exporters"]
			if hasExportersSection {
				injectEnvironmentAPIKey(exportersSection, exportersList, envName)
			}
		}
	}
}

// applyPipelineRenames renames pipelines in the service section, merging duplicates if multiple pipelines
// are renamed to the same target name.
func applyPipelineRenames(serviceSection map[string]any, pipelineRenames map[string]string) {
	// Group renames by target name to handle merges
	renamesByTarget := make(map[string][]string) // target name -> list of source names
	for oldName, newName := range pipelineRenames {
		renamesByTarget[newName] = append(renamesByTarget[newName], oldName)
	}

	// Apply pipeline renames, merging duplicates
	for newName, oldNames := range renamesByTarget {
		if len(oldNames) == 1 {
			// Simple rename, no merge needed
			oldName := oldNames[0]
			keysToRename := make([]string, 0)
			for key := range serviceSection {
				if strings.HasPrefix(key, "pipelines."+oldName+".") {
					keysToRename = append(keysToRename, key)
				}
			}

			for _, oldKey := range keysToRename {
				newKey := strings.Replace(oldKey, "pipelines."+oldName+".", "pipelines."+newName+".", 1)
				serviceSection[newKey] = serviceSection[oldKey]
				delete(serviceSection, oldKey)
			}
		} else {
			// Multiple pipelines renaming to same name - merge them
			// Take the first one as the canonical pipeline
			firstOldName := oldNames[0]

			// Rename the first one
			keysToRename := make([]string, 0)
			for key := range serviceSection {
				if strings.HasPrefix(key, "pipelines."+firstOldName+".") {
					keysToRename = append(keysToRename, key)
				}
			}

			for _, oldKey := range keysToRename {
				newKey := strings.Replace(oldKey, "pipelines."+firstOldName+".", "pipelines."+newName+".", 1)
				serviceSection[newKey] = serviceSection[oldKey]
				delete(serviceSection, oldKey)
			}

			// Delete the duplicate pipelines
			for _, oldName := range oldNames[1:] {
				keysToDelete := make([]string, 0)
				for key := range serviceSection {
					if strings.HasPrefix(key, "pipelines."+oldName+".") {
						keysToDelete = append(keysToDelete, key)
					}
				}
				for _, key := range keysToDelete {
					delete(serviceSection, key)
				}
			}
		}
	}
}

// findExportersForEnvironment finds all exporters in the exporters section that match the given environment.
func findExportersForEnvironment(exportersSection map[string]any, exporterToEnvironment map[string]string, envName string) []string {
	exportersForEnv := make([]string, 0)

	for exporterSafeName, exporterEnv := range exporterToEnvironment {
		if exporterEnv == envName {
			// Find the exporter key in the exporters section
			for exporterKey := range exportersSection {
				exporterName := exporterKey
				if dotIdx := strings.Index(exporterKey, "."); dotIdx > 0 {
					exporterName = exporterKey[:dotIdx]
				}
				// Exporter names have format: {type}/{componentSafeName}
				if slashIdx := strings.Index(exporterName, "/"); slashIdx >= 0 {
					exporterCompName := exporterName[slashIdx+1:]
					if exporterCompName == exporterSafeName {
						if !sliceContains(exportersForEnv, exporterName) {
							exportersForEnv = append(exportersForEnv, exporterName)
						}
						break
					}
				}
			}
		}
	}

	return exportersForEnv
}

// createMissingOutputPipelines creates output pipelines that are referenced in routing connector config
// but don't exist yet (this happens when output paths weren't generated).
func createMissingOutputPipelines(serviceSection map[string]any, cc *tmpl.CollectorConfig, expectedPipelines map[string]bool, exporterToEnvironment map[string]string) {
	exportersSection, hasExportersSection := cc.Sections["exporters"]
	if !hasExportersSection {
		return
	}

	for expectedPipeline := range expectedPipelines {
		// Check if this pipeline already exists
		pipelineExists := false
		for key := range serviceSection {
			if strings.HasPrefix(key, "pipelines."+expectedPipeline+".") {
				pipelineExists = true
				break
			}
		}

		if !pipelineExists {
			// Create the pipeline with routing connector as receiver
			// Extract signal type from pipeline name (e.g., "traces/dev" -> "traces")
			parts := strings.SplitN(expectedPipeline, "/", 2)
			if len(parts) == 2 {
				signalType := parts[0]
				envName := parts[1]

				// Set receivers to routing connector
				receiversKey := fmt.Sprintf("pipelines.%s.receivers", expectedPipeline)
				serviceSection[receiversKey] = []string{"routing/" + signalType}

				// Find exporters for this environment
				exportersForEnv := findExportersForEnvironment(exportersSection, exporterToEnvironment, envName)

				if len(exportersForEnv) > 0 {
					exportersKey := fmt.Sprintf("pipelines.%s.exporters", expectedPipeline)
					serviceSection[exportersKey] = exportersForEnv
				}
			}
		}
	}
}

// mergeRoutingConnectors finds routing connector entries and merges them per signal type
// Creates separate routing connectors: routing/logs, routing/traces, routing/metrics
func mergeRoutingConnectors(cc *tmpl.CollectorConfig) error {
	connectorSection, ok := cc.Sections["connectors"]
	if !ok {
		return nil // no connectors, nothing to do
	}

	// Track entries by signal type and component: routing/logs, routing/traces, routing/metrics
	// signalType -> component -> entries
	tableEntriesBySignalType := make(map[string]map[string][]map[string]any)
	defaultPipelinesBySignalType := make(map[string][]string) // signalType -> pipelines

	for key, value := range connectorSection {
		// Check for routing/{signaltype}.default_pipelines
		if strings.HasPrefix(key, "routing/") && strings.Contains(key, ".default_pipelines") {
			// Format: routing/logs.default_pipelines
			parts := strings.SplitN(key, "/", 2)
			if len(parts) == 2 {
				signalTypeWithKey := parts[1] // e.g. "logs.default_pipelines"
				signalType := strings.SplitN(signalTypeWithKey, ".", 2)[0]
				if pipelines, ok := value.([]string); ok {
					defaultPipelinesBySignalType[signalType] = append(defaultPipelinesBySignalType[signalType], pipelines...)
				}
			}
		}

		// Check for routing/{signaltype}.table_{component}[N].{field} entries
		// Format: routing/logs.table_router_staging[0].condition
		if strings.HasPrefix(key, "routing/") && strings.Contains(key, ".table_") && strings.Contains(key, "[") {
			// Extract signal type from routing/{signaltype}.table_...
			parts := strings.SplitN(key, "/", 2)
			if len(parts) != 2 {
				continue
			}
			afterSlash := parts[1] // e.g. "logs.table_router_staging[0].condition"

			// Extract signal type (before first dot)
			dotPos := strings.Index(afterSlash, ".")
			if dotPos < 0 {
				continue
			}
			signalType := afterSlash[:dotPos] // e.g. "logs"

			// Extract component name from .table_{component}[N]
			after := strings.SplitN(afterSlash, ".table_", 2)
			if len(after) != 2 {
				continue
			}
			afterTablePrefix := after[1] // e.g. "router_staging[0].condition"

			bracketPos := strings.Index(afterTablePrefix, "[")
			if bracketPos < 0 {
				continue
			}
			componentName := afterTablePrefix[:bracketPos] // e.g. "router_staging"

			// Extract index
			indexStart := bracketPos + 1
			closeBracket := strings.Index(afterTablePrefix[indexStart:], "]")
			if closeBracket < 0 {
				continue
			}
			idx := afterTablePrefix[indexStart : indexStart+closeBracket]
			entryIdx, err := strconv.Atoi(idx)
			if err != nil {
				continue
			}

			// Extract field name (context, condition, pipelines)
			fieldStart := indexStart + closeBracket + 2 // +2 to skip "]."
			if fieldStart >= len(afterTablePrefix) {
				continue
			}
			fieldName := afterTablePrefix[fieldStart:]

			// Initialize nested maps if needed
			if tableEntriesBySignalType[signalType] == nil {
				tableEntriesBySignalType[signalType] = make(map[string][]map[string]any)
			}
			if tableEntriesBySignalType[signalType][componentName] == nil {
				tableEntriesBySignalType[signalType][componentName] = make([]map[string]any, 0)
			}

			// Ensure we have enough entries for this index
			for len(tableEntriesBySignalType[signalType][componentName]) <= entryIdx {
				tableEntriesBySignalType[signalType][componentName] = append(
					tableEntriesBySignalType[signalType][componentName],
					make(map[string]any),
				)
			}

			// Store the field in the entry
			tableEntriesBySignalType[signalType][componentName][entryIdx][fieldName] = value
		}
	}

	// Delete old routing/* keys
	for key := range connectorSection {
		if strings.HasPrefix(key, "routing/") {
			delete(connectorSection, key)
		}
	}

	// For each signal type, merge table entries and create routing/{signaltype} connector
	for signalType, componentEntries := range tableEntriesBySignalType {
		// Sort component names for deterministic output
		componentNames := make([]string, 0, len(componentEntries))
		for componentName := range componentEntries {
			componentNames = append(componentNames, componentName)
		}
		sort.Strings(componentNames)

		// Merge all table entries for this signal type
		allTableEntries := make([]map[string]any, 0)
		for _, componentName := range componentNames {
			entries := componentEntries[componentName]
			allTableEntries = append(allTableEntries, entries...)
		}

		// Create merged table entries for routing/{signaltype}
		for i, entry := range allTableEntries {
			if context, ok := entry["context"].(string); ok {
				key := fmt.Sprintf("routing/%s.table[%d].context", signalType, i)
				connectorSection[key] = context
			}
			if condition, ok := entry["condition"].(string); ok {
				key := fmt.Sprintf("routing/%s.table[%d].condition", signalType, i)
				connectorSection[key] = condition
			}
			if statement, ok := entry["statement"].(string); ok {
				key := fmt.Sprintf("routing/%s.table[%d].statement", signalType, i)
				connectorSection[key] = statement
			}
			if pipelines, ok := entry["pipelines"].([]string); ok {
				key := fmt.Sprintf("routing/%s.table[%d].pipelines", signalType, i)
				connectorSection[key] = pipelines
			}
		}

		// Add default_pipelines if present for this signal type
		if len(defaultPipelinesBySignalType[signalType]) > 0 {
			key := fmt.Sprintf("routing/%s.default_pipelines", signalType)
			connectorSection[key] = defaultPipelinesBySignalType[signalType]
		}
	}

	return nil
}

// A Translator is responsible for translating an HPSF document into a
// collection of components, and then further rendering those into configuration
// files.
type Translator struct {
	components map[string]config.TemplateComponent
	templates  map[string]hpsf.HPSF
}

// Deprecated: use NewEmptyTranslator and InstallComponents instead
func NewTranslator() (*Translator, error) {
	tr := &Translator{}
	// autoload the template components because we don't want to break existing code
	err := tr.LoadEmbeddedComponents()
	return tr, err
}

// NewEmptyTranslator creates a translator with no components loaded.
func NewEmptyTranslator() *Translator {
	tr := &Translator{
		components: make(map[string]config.TemplateComponent),
		templates:  make(map[string]hpsf.HPSF),
	}
	return tr
}

// InstallComponents installs the given components into the translator.
func (t *Translator) InstallComponents(components map[string]config.TemplateComponent) {
	maps.Copy(t.components, components)
}

// InstallTemplates installs the given templates into the translator.
func (t *Translator) InstallTemplates(components map[string]hpsf.HPSF) {
	maps.Copy(t.templates, components)
}

// GetComponents returns the components installed in the translator.
func (t *Translator) GetComponents() map[string]config.TemplateComponent {
	return t.components
}

// GetTemplates returns the templates installed in the translator.
func (t *Translator) GetTemplates() map[string]hpsf.HPSF {
	return t.templates
}

// LoadEmbeddedComponents loads the embedded components into the translator.
// Deprecated: use InstallComponents instead
func (t *Translator) LoadEmbeddedComponents() error {
	// load the embedded components
	tcs, err := data.LoadEmbeddedComponents()
	if err != nil {
		return err
	}
	maps.Copy(t.components, tcs)
	return nil
}

// artifactVersionSupported checks if the component supports the artifact version requested
func artifactVersionSupported(component config.TemplateComponent, v string) bool {
	if v == "" || v == LatestVersion {
		return true
	}

	// ensure the version string is prefixed with v otherwise semver.Compare fails
	// to parse the version
	if v[0] != 'v' {
		v = "v" + v
	}

	if component.Minimum != "" && semver.Compare(v, component.Minimum) < 0 {
		return false
	}

	if component.Maximum != "" && semver.Compare(v, component.Maximum) > 0 {
		return false
	}

	return true
}

func (t *Translator) MakeConfigComponent(component *hpsf.Component, artifactVersion string) (config.Component, error) {
	// first look in the template components
	tc, ok := t.components[component.Kind]
	if ok && (len(component.Version) <= 0 || tc.Version == component.Version) && artifactVersionSupported(tc, artifactVersion) {
		// found it, manufacture a new instance of the component
		tc.SetHPSF(component)
		return &tc, nil
	}

	// nothing found so we're done
	return nil, fmt.Errorf("unknown component kind: %s@%s", component.Kind, component.Version)
}

// getMatchingTemplateComponents returns the template components that match the components in the HPSF document.
// It validates components before matching them and returns an error if any components are invalid.
func (t *Translator) getMatchingTemplateComponents(h *hpsf.HPSF) (map[string]config.TemplateComponent, validator.Result) {
	result := validator.NewResult("HPSF component fetch failed")
	templateComps := make(map[string]config.TemplateComponent)
	for _, c := range h.Components {
		err := c.Validate()
		if err != nil {
			result.Add(fmt.Errorf("failed to validate component %s: %w", c.Name, err))
			continue
		}
		if comp, ok := t.components[c.Kind]; ok && (len(c.Version) <= 0 || c.Version == comp.Version) {
			templateComps[c.GetSafeName()] = comp
		} else {
			result.Add(fmt.Errorf("failed to locate corresponding template component for %s@%s: %w", c.Kind, c.Version, err))
		}
	}
	return templateComps, result
}

func (t *Translator) validateProperties(h *hpsf.HPSF, templateComps map[string]config.TemplateComponent) validator.Result {
	result := validator.NewResult("HPSF property validation errors")
	// now we have a map of all the components that were successfully instantiated
	// so we can iterate the properties and validate them according to the validations specified in the template components
	for _, comp := range h.Components {
		tmpl, ok := templateComps[comp.GetSafeName()]
		if !ok {
			// If we don't have a template component for this component, it
			// means we couldn't instantiate it. We caught this earlier, so we
			// should never get here. Just continue.
			continue
		}

		// Get the template properties from the template component.
		templateProperties := tmpl.Props()
		componentProps := make(map[string]hpsf.Property)
		for _, prop := range comp.Properties {
			componentProps[prop.Name] = prop
			_, found := templateProperties[prop.Name]
			if !found {
				// If the property is not found in the template component's
				// properties, something's messed up. This means the property is
				// not defined in the template component.
				err := hpsf.NewError("property not found in template component").
					WithComponent(comp.Name).
					WithProperty(prop.Name)
				result.Add(err)
			}
		}

		for _, prop := range templateProperties {
			// validate each property against the template component's basic validation rules
			suppliedProperty, propertyFound := componentProps[prop.Name]
			if !propertyFound {
				// If the property is not supplied, use the default value from the template component.
				suppliedProperty.Value = prop.Default
			}

			// Now validate the property against the template property's validation rules.
			if validateError := prop.Validate(suppliedProperty); validateError != nil {
				// if the property fails validation, add the error to the result
				// this means the property itself has some issues
				// we want to include the component name and property name in the error message for clarity
				hspfError := hpsf.NewError("failed to validate property").
					WithCause(validateError).
					WithComponent(comp.Name).
					WithProperty(prop.Name)
				result.Add(hspfError)
			}
		}

		// Execute component validations after individual property validations pass
		if validateError := tmpl.Validate(comp); validateError != nil {
			result.Add(validateError)
		}
	}
	return result
}

// this checks that there is exactly one connection on the input and output of each sampler
// and condition component.
func (t *Translator) validateSamplerConnections(h *hpsf.HPSF, templateComps map[string]config.TemplateComponent) validator.Result {
	result := validator.NewResult("HPSF sampler connection validation errors")
	// iterate over the components and check for samplers
	for _, c := range h.Components {
		tmpl, ok := templateComps[c.GetSafeName()]
		if !ok {
			// If we don't have a template component for this component, it
			// means we couldn't instantiate it. We caught this earlier, so we
			// should never get here. Just continue.
			continue
		}

		if tmpl.Style == "sampler" || tmpl.Style == "dropper" || tmpl.Style == "condition" {
			// check the connections for the component
			inputs := 0
			outputs := 0
			for _, conn := range h.Connections {
				if conn.Destination.GetSafeName() == c.GetSafeName() {
					inputs++
				}
				if conn.Source.GetSafeName() == c.GetSafeName() {
					outputs++
				}
			}
			if inputs != 1 {
				err := hpsf.NewError("sampler, dropper, and condition components must have exactly one input connection").
					WithComponent(c.Name)
				result.Add(err)
			}
			if outputs != 1 && tmpl.Style != "dropper" {
				err := hpsf.NewError("sampler and condition components must have exactly one output connection").
					WithComponent(c.Name)
				result.Add(err)
			}
		}
	}
	return result
}

// validateConnectionPorts checks that all connections have valid ports. The name on the connection
// in hpsf must match the port name on the template component.
func (t *Translator) validateConnectionPorts(h *hpsf.HPSF, templateComps map[string]config.TemplateComponent) validator.Result {
	result := validator.NewResult("HPSF connection port validation errors")
	// iterate over the connections and check that the source and destination components have the
	// specified ports. This is a sanity check to ensure that the connections are valid.
	for _, conn := range h.Connections {
		srcComp, ok := templateComps[conn.Source.GetSafeName()]
		if !ok {
			continue
		}

		if srcComp.GetPort(conn.Source.PortName) == nil {
			err := hpsf.NewErrorf("source component does not have a port called %s", conn.Source.PortName).
				WithComponent(conn.Source.Component)
			result.Add(err)
		}

		dstComp, ok := templateComps[conn.Destination.GetSafeName()]
		if !ok {
			continue
		}

		if dstComp.GetPort(conn.Destination.PortName) == nil {
			err := hpsf.NewErrorf("destination component does not have a port called %s", conn.Destination.PortName).
				WithComponent(conn.Destination.Component)
			result.Add(err)
		}
	}
	return result
}

// findPathComponents finds all the components in paths starting from the given component.
// It returns a slice of component names that represent all components in paths
// that start from the given component.
func (t *Translator) findPathComponents(h *hpsf.HPSF, startComp string) []string {
	visited := make(map[string]bool)
	components := make([]string, 0)

	var dfs func(comp string)
	dfs = func(comp string) {
		if visited[comp] {
			return
		}
		visited[comp] = true
		components = append(components, comp)

		// Continue traversing to find all connected components
		for _, conn := range h.Connections {
			if conn.Source.GetSafeName() == comp {
				dfs(conn.Destination.GetSafeName())
			}
		}
	}

	dfs(startComp)
	return components
}

// The rules for sampling in HPSF are as follows:
// - If there are any sampling components, there must be at least one component with style "startsampling".
// - Each path connected to a "startsampling" component's output must lead to exactly one "sampler" or "dropper".
// - There may be multiple "condition" components between startsampling and the sampler or dropper.
// - Every path on a startsampler except the one with the highest index must connect to a condition.
// - Droppers can terminate a path (since they do not have an output port).
// - The output of samplers must be connected to an "exporter" component.
func (t *Translator) validateStartSampling(h *hpsf.HPSF, templateComps map[string]config.TemplateComponent) validator.Result {
	result := validator.NewResult("HPSF start sampling validation errors")
	startSamplingCount := 0
	var startSamplingComp string
	for _, c := range h.Components {
		tmpl, ok := templateComps[c.GetSafeName()]
		if !ok {
			continue
		}

		if tmpl.Style == "startsampling" {
			startSamplingCount++
			startSamplingComp = c.GetSafeName()
		}
	}
	if startSamplingCount == 0 {
		// if there is no StartSampling component, we cannot have any samplers in the configuration
		for _, c := range h.Components {
			tmpl, ok := templateComps[c.GetSafeName()]
			if !ok {
				continue
			}

			if tmpl.Style == "sampler" {
				err := hpsf.NewError("if there is no StartSampling component, no samplers are allowed").
					WithComponent(c.Name)
				result.Add(err)
			}
		}
	} else {
		// if there is a StartSampling component, we must have at least one sampler or dropper in the configuration
		hasSamplerOrDropper := false
		for _, c := range h.Components {
			tmpl, ok := templateComps[c.GetSafeName()]
			if !ok {
				continue
			}

			if tmpl.Style == "sampler" || tmpl.Style == "dropper" {
				hasSamplerOrDropper = true
				break
			}
		}
		if !hasSamplerOrDropper {
			err := hpsf.NewError("if there is a StartSampling component, at least one sampler or dropper is required").
				WithComponent(startSamplingComp)
			result.Add(err)
		}
	}
	// now we need to check that each path from the StartSampling component leads to exactly one sampler or dropper
	if startSamplingCount == 1 {
		// Find all connections from StartSampling
		startSamplingConnections := make([]*hpsf.Connection, 0)
		for _, conn := range h.Connections {
			if conn.Source.GetSafeName() == startSamplingComp {
				startSamplingConnections = append(startSamplingConnections, conn)
			}
		}

		// For each connection from StartSampling, trace the path to find if it leads to exactly one sampler or dropper
		for _, startConn := range startSamplingConnections {
			pathComponents := t.findPathComponents(h, startConn.Destination.GetSafeName())
			samplerOrDropperCount := 0
			for _, comp := range pathComponents {
				tmpl, ok := templateComps[comp]
				if ok && (tmpl.Style == "sampler" || tmpl.Style == "dropper") {
					samplerOrDropperCount++
				}
			}
			if samplerOrDropperCount != 1 {
				err := hpsf.NewError("Each path from StartSampling must lead to exactly one sampler or dropper").
					WithComponent(startSamplingComp)
				result.Add(err)
			}
		}

		// Validate that every path except the one with the highest index connects to a condition
		// Find the highest index among all StartSampling connections
		highestIndex := -1
		startSamplingTemplate := templateComps[startSamplingComp]
		for _, startConn := range startSamplingConnections {
			// Get the port index from the connection's source port
			portIndex := startSamplingTemplate.GetPortIndex(startConn.Source.PortName)
			if portIndex > highestIndex {
				highestIndex = portIndex
			}
		}

		// Check each path except the one with the highest index
		for _, startConn := range startSamplingConnections {
			portIndex := startSamplingTemplate.GetPortIndex(startConn.Source.PortName)
			if portIndex != highestIndex {
				// This path must connect to a condition
				pathComponents := t.findPathComponents(h, startConn.Destination.GetSafeName())
				hasCondition := false
				for _, comp := range pathComponents {
					tmpl, ok := templateComps[comp]
					if ok && tmpl.Style == "condition" {
						hasCondition = true
						break
					}
				}
				if !hasCondition {
					err := hpsf.NewError("Every path on a startsampler except the one with the highest index must connect to a condition").
						WithComponent(startSamplingComp)
					result.Add(err)
				}
			}
		}
	}

	return result
}

// ValidateConfig validates the configuration of the HPSF document as it stands with respect to the
// components and templates installed in the translator.
// Note that it returns a validation.Result so that the errors can be collected and reported in a
// structured way. This allows for multiple validation errors to be returned at once, rather than
// stopping at the first error. This is useful for providing feedback to users on multiple issues
// in their configuration.
func (t *Translator) ValidateConfig(h *hpsf.HPSF) error {
	if h == nil {
		return errors.New("nil HPSF document provided for validation")
	}

	// if we don't pass basic validation, we can't continue
	if err := h.Validate(); err != nil {
		return err
	}

	// We assume that the HPSF document has already been validated for syntax and structure since
	// it's already in hpsf format. Our goal here is to make sure that the components and templates
	// can be used to generate a valid configuration. This means checking that all components referenced
	// in the HPSF document are available in the translator's component map and that they can be instantiated
	// correctly, and that all the properties are of the correct type.
	templateComps, result := t.getMatchingTemplateComponents(h)
	if !result.IsEmpty() {
		// if we have errors at this point, return early
		// this means we couldn't even instantiate the components
		// so there's no point in continuing to validate the connections
		return result
	}

	result.Add(t.validateProperties(h, templateComps))
	result.Add(t.validateConnectionPorts(h, templateComps))
	result.Add(t.validateStartSampling(h, templateComps))
	result.Add(t.validateSamplerConnections(h, templateComps))

	return result.ErrOrNil()
}

// OrderedComponentMap is a generic map that maintains the order of insertion.
// It is used to ensure that the order of components and properties is preserved
// when generating the configuration.
type OrderedComponentMap struct {
	// Keys is the list of keys in the order they were added.
	Keys []string
	// Values is the map of keys to values.
	Values map[string]config.Component
}

func NewOrderedComponentMap() *OrderedComponentMap {
	return &OrderedComponentMap{
		Keys:   make([]string, 0),
		Values: make(map[string]config.Component),
	}
}

// Set adds a key-value pair to the ordered map.
func (om *OrderedComponentMap) Set(key string, value config.Component) {
	if _, exists := om.Values[key]; !exists {
		// Only add the key to the Keys slice if it doesn't already exist
		om.Keys = append(om.Keys, key)
	}
	om.Values[key] = value
}

// Get retrieves a value from the ordered map by key.
func (om *OrderedComponentMap) Get(key string) (config.Component, bool) {
	value, exists := om.Values[key]
	return value, exists
}

// Items returns a Go iterable
func (om *OrderedComponentMap) Items() iter.Seq[config.Component] {
	return func(yield func(config.Component) bool) {
		for _, key := range om.Keys {
			if value, exists := om.Values[key]; exists {
				if !yield(value) {
					return
				}
			}
		}
	}
}

// getFirstConnectionPortIndex attempts to read the index of the source port on the first
// connection of a path. It returns (index, true) if a positive (non-zero) index is found.
// Index 0 or any failure to determine the index returns (0, false) to indicate "unspecified".
func getFirstConnectionPortIndex(path hpsf.PathWithConnections, comps *OrderedComponentMap) (int, bool) {
	if len(path.Connections) == 0 {
		return 0, false
	}
	first := path.Connections[0]
	comp, ok := comps.Get(first.Source.GetSafeName())
	if !ok {
		return 0, false
	}
	tc, ok := comp.(*config.TemplateComponent)
	if !ok {
		return 0, false
	}
	idx := tc.GetPortIndex(first.Source.PortName)
	if idx > 0 {
		return idx, true
	}
	return 0, false
}

// orderPaths sorts paths deterministically. Precedence within a connection type:
//  1. Has connections (paths with zero connections come last)
//  2. Presence of a positive port index (indexed paths before non-indexed)
//  3. Ascending numeric port index (if both indexed)
//  4. Source component name (lexicographically)
//  5. Source port name (lexicographically)
//  6. Path ID (stable final tie breaker)
//
// This matches updated requirement: index ordering takes priority over component name.
func orderPaths(paths []hpsf.PathWithConnections, comps *OrderedComponentMap) {
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].ConnType != paths[j].ConnType {
			return paths[i].ConnType < paths[j].ConnType
		}

		// Handle zero-connection paths: they go last within the connection type group
		li := len(paths[i].Connections)
		lj := len(paths[j].Connections)
		if li == 0 || lj == 0 {
			if li == 0 && lj == 0 {
				// stable ordering: compare IDs to keep determinate
				return paths[i].GetID() < paths[j].GetID()
			}
			return lj != 0 // true if i has connections and j does not
		}

		// Both have at least one connection
		// Primary: port index presence / value
		idxI, hasIdxI := getFirstConnectionPortIndex(paths[i], comps)
		idxJ, hasIdxJ := getFirstConnectionPortIndex(paths[j], comps)
		if hasIdxI != hasIdxJ { // indexed before non-indexed
			return hasIdxI
		}
		if hasIdxI && idxI != idxJ { // both indexed, numeric order
			return idxI < idxJ
		}

		// Next: component name
		srcCompI := paths[i].Connections[0].Source.Component
		srcCompJ := paths[j].Connections[0].Source.Component
		if srcCompI != srcCompJ {
			return srcCompI < srcCompJ
		}

		// Next: source port name
		portI := paths[i].Connections[0].Source.PortName
		portJ := paths[j].Connections[0].Source.PortName
		if portI != portJ {
			return portI < portJ
		}

		// Last resort: deterministic by path ID
		return paths[i].GetID() < paths[j].GetID()
	})
}

// transformRouterPipelines transforms pipelines that use routing connectors to follow OTel conventions:
// - Intake pipelines (with receivers and connectors but no exporters) move connector to exporters
// - Output pipelines (with exporters but empty receivers and connectors) move connector to receivers
// - Renames output pipelines to match environment names in routing connector config
// - Extracts environment information from SetEnvironment components for API key injection
func transformRouterPipelines(cc *tmpl.CollectorConfig, h *hpsf.HPSF, comps *OrderedComponentMap) error {
	serviceSection, exists := cc.Sections["service"]
	if !exists {
		return nil
	}

	connectorSection, hasConnectors := cc.Sections["connectors"]
	if !hasConnectors {
		return nil
	}

	// Extract configuration from routing connectors and existing pipelines
	expectedPipelines := extractExpectedPipelines(connectorSection)
	pipelineNames := collectPipelineNames(serviceSection)
	exporterToEnvironment := buildExporterEnvironmentMap(h)

	// Track pipeline renames for output pipelines
	pipelineRenames := make(map[string]string) // old name -> new name

	// Process each pipeline once
	for pipelineName := range pipelineNames {
		receiversKey := fmt.Sprintf("pipelines.%s.receivers", pipelineName)
		connectorsKey := fmt.Sprintf("pipelines.%s.connectors", pipelineName)
		exportersKey := fmt.Sprintf("pipelines.%s.exporters", pipelineName)

		receivers := serviceSection[receiversKey]
		connectors := serviceSection[connectorsKey]
		exporters := serviceSection[exportersKey]

		// Convert pipeline components to string lists
		connectorsList := getStringListFromAny(connectors)
		receiversFiltered := getStringListFromAny(receivers)
		exportersFiltered := getStringListFromAny(exporters)

		// Check if this pipeline has a routing connector
		hasRouting := false
		for _, conn := range connectorsList {
			if conn == "routing" || strings.HasPrefix(conn, "routing_") || strings.HasPrefix(conn, "routing/") {
				hasRouting = true
				break
			}
		}

		// Skip if this pipeline has connectors but no routing connector
		if len(connectorsList) > 0 && !hasRouting {
			continue
		}

		// Determine signal type from pipeline name
		parts := strings.SplitN(pipelineName, "/", 2)
		var signalType string
		if len(parts) == 2 {
			signalType = parts[0] // e.g., "traces", "logs", "metrics"
		}

		// Check if routing connector is in exporters (and normalize the name to routing/{signaltype})
		var hasRoutingInExporters bool
		for i, exp := range exportersFiltered {
			if exp == "routing" || strings.HasPrefix(exp, "routing_") || strings.HasPrefix(exp, "routing/") {
				// Normalize to "routing/{signaltype}" if we know the signal type
				if signalType != "" {
					exportersFiltered[i] = "routing/" + signalType
				} else {
					exportersFiltered[i] = "routing"
				}
				hasRoutingInExporters = true
			}
		}
		if hasRoutingInExporters {
			serviceSection[exportersKey] = exportersFiltered
		}

		// Case 1: Intake pipeline - has receivers and routing connector, no other exporters
		if hasRouting && len(receiversFiltered) > 0 && len(exportersFiltered) == 0 {
			if newName := transformIntakePipeline(serviceSection, pipelineName, signalType, connectorsList); newName != "" {
				pipelineRenames[pipelineName] = newName
			}
		} else if hasRoutingInExporters && len(receiversFiltered) > 0 {
			// Case 1b: Intake pipeline - has receivers and routing connector already in exporters
			// Just rename the pipeline
			if signalType != "" {
				intakeName := signalType + "/intake"
				if pipelineName != intakeName {
					pipelineRenames[pipelineName] = intakeName
				}
			}
		} else if hasRouting && len(receiversFiltered) == 0 && len(exportersFiltered) > 0 {
			// Case 2: Output pipeline - has routing connector, no receivers, has exporters
			transformOutputPipeline(serviceSection, cc, pipelineName, signalType, connectors, exportersFiltered, exporterToEnvironment)
		} else if !hasRouting && len(receiversFiltered) == 0 && len(exportersFiltered) > 0 {
			// Case 3: Output pipeline without routing connector - no receivers, has exporters → add routing to receivers
			// Use signal-type-specific routing connector name
			if signalType != "" {
				routingConnectorName := "routing/" + signalType
				serviceSection[receiversKey] = []string{routingConnectorName}

				// Try to infer environment from pipeline name first, then exporter name
				var envName string

				// For Case 3, the pipeline name might not be renamed yet, so we need to find which
				// environment this pipeline should belong to by matching exporter names first
				for _, exp := range exportersFiltered {
					// Extract potential environment name from exporter
					// Look for common patterns like "_staging", "_production", "Staging", "Production"
					expLower := strings.ToLower(exp)
					for expectedName := range expectedPipelines {
						if strings.HasPrefix(expectedName, signalType+"/") {
							envPart := strings.TrimPrefix(expectedName, signalType+"/")
							if strings.Contains(expLower, strings.ToLower(envPart)) {
								envName = envPart
								break
							}
						}
					}
					if envName != "" {
						break
					}
				}

				// If we found an environment name, use it
				if envName != "" {
					targetName := signalType + "/" + envName
					if expectedPipelines[targetName] {
						pipelineRenames[pipelineName] = targetName

						// Inject environment-specific API key header into otlphttp exporters
						exportersSection, hasExportersSection := cc.Sections["exporters"]
						if hasExportersSection {
							for _, exp := range exportersFiltered {
								headerKey := fmt.Sprintf("%s.headers.x-honeycomb-team", exp)
								existingValue, hasHeader := exportersSection[headerKey]
								// Inject if header doesn't exist or if it's set to the default value
								if !hasHeader || existingValue == "${HTP_EXPORTER_APIKEY}" {
									envVarName := fmt.Sprintf("${HTP_EXPORTER_APIKEY_%s}", strings.ToUpper(envName))
									exportersSection[headerKey] = envVarName
								}
							}
						}
					}
				} else {
					// Fall back to finding any unused expected pipeline name for this signal type
					for expectedName := range expectedPipelines {
						if strings.HasPrefix(expectedName, signalType+"/") {
							// Check if this expected pipeline name is already assigned to another pipeline
							alreadyAssigned := false
							for _, assignedName := range pipelineRenames {
								if assignedName == expectedName {
									alreadyAssigned = true
									break
								}
							}
							if !alreadyAssigned {
								// Use this environment name for renaming
								pipelineRenames[pipelineName] = expectedName
								break
							}
						}
					}
				}
			}
		}
	}

	// Apply pipeline renames and create missing pipelines
	applyPipelineRenames(serviceSection, pipelineRenames)
	createMissingOutputPipelines(serviceSection, cc, expectedPipelines, exporterToEnvironment)

	return nil
}

// sliceContains checks if a string slice contains a specific string
func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// injectEnvironmentAPIKeys injects environment-specific API keys into exporters
// based on their EnvironmentName property. This ensures ${HTP_EXPORTER_APIKEY_ENV}
// variables are set correctly for each environment.
func injectEnvironmentAPIKeys(cc *tmpl.CollectorConfig, exporterToEnvironment map[string]string) error {
	if cc == nil {
		return nil
	}

	exportersSection, hasExportersSection := cc.Sections["exporters"]
	if !hasExportersSection {
		return nil
	}

	// For each exporter, inject environment-specific API key based on its environment property
	for exporterSafeName, envName := range exporterToEnvironment {
		// Find the exporter key in the exporters section
		// Exporter keys have format: {type}/{componentSafeName}.headers.x-honeycomb-team
		for exporterKey := range exportersSection {
			exporterName := exporterKey
			if dotIdx := strings.Index(exporterKey, "."); dotIdx > 0 {
				exporterName = exporterKey[:dotIdx]
			}

			// Exporter names have format: {type}/{componentSafeName}
			if slashIdx := strings.Index(exporterName, "/"); slashIdx >= 0 {
				exporterCompName := exporterName[slashIdx+1:]
				if exporterCompName == exporterSafeName {
					// Check if this is the API key header
					headerKey := fmt.Sprintf("%s.headers.x-honeycomb-team", exporterName)
					existingValue, hasHeader := exportersSection[headerKey]

					// Inject if header doesn't exist or if it's set to the default value
					if !hasHeader || existingValue == "${HTP_EXPORTER_APIKEY}" {
						envVarName := fmt.Sprintf("${HTP_EXPORTER_APIKEY_%s}", strings.ToUpper(envName))
						exportersSection[headerKey] = envVarName
					}
					break
				}
			}
		}
	}

	return nil
}

// generateRefineryRulesWithRouter handles refinery rules generation for multi-environment routing.
// It processes sampling paths and sets the environment context from HoneycombExporter components.
func (t *Translator) generateRefineryRulesWithRouter(h *hpsf.HPSF, comps *OrderedComponentMap, paths []hpsf.PathWithConnections, ct hpsftypes.Type, userdata map[string]any) (tmpl.TemplateConfig, error) {
	dummy := hpsf.Component{Name: "dummy", Kind: "dummy"}

	// Build exporter to environment mapping
	exporterToEnvironment := make(map[string]string)
	for _, hcomp := range h.Components {
		if hcomp.Kind == "HoneycombExporter" {
			for _, prop := range hcomp.Properties {
				if prop.Name == "EnvironmentName" {
					if envName, ok := prop.Value.(string); ok && envName != "" {
						exporterToEnvironment[hcomp.Name] = envName
					}
					break
				}
			}
		}
	}

	// Generate configs for each sampling path with environment context
	composites := make([]tmpl.TemplateConfig, 0)
	for _, path := range paths {
		// Only process sampling paths
		if path.ConnType != hpsf.CTYPE_SAMPLE {
			continue
		}

		// Create a copy of userdata for this path
		pathUserdata := make(map[string]any)
		for k, v := range userdata {
			pathUserdata[k] = v
		}

		// Find environment by looking for HoneycombExporter downstream from the SamplingSequencer
		// Sampling paths follow SampleData connections, so we need to traverse OTel signal connections
		// forward to find exporters
		var samplingSequencer *hpsf.Component
		for _, comp := range path.Path {
			if comp.Kind == "SamplingSequencer" {
				samplingSequencer = comp
				break
			}
		}

		if samplingSequencer != nil {
			// Traverse forward from SamplingSequencer to find HoneycombExporter
			visited := make(map[string]bool)
			var findExporterEnv func(compName string) string
			findExporterEnv = func(compName string) string {
				if visited[compName] {
					return ""
				}
				visited[compName] = true

				// Check if this is an exporter
				if envName, ok := exporterToEnvironment[compName]; ok {
					return envName
				}

				// Follow OTel signal connections downstream
				for _, conn := range h.Connections {
					if conn.Source.Component == compName &&
						(conn.Source.Type == hpsf.CTYPE_TRACES || conn.Source.Type == hpsf.CTYPE_LOGS || conn.Source.Type == hpsf.CTYPE_METRICS) {
						if env := findExporterEnv(conn.Destination.Component); env != "" {
							return env
						}
					}
				}
				return ""
			}

			if env := findExporterEnv(samplingSequencer.Name); env != "" {
				pathUserdata["environment"] = env
			}
		}

		// Generate config for this path
		base := config.GenericBaseComponent{Component: dummy}
		composite, err := base.GenerateConfig(ct, path, pathUserdata)
		if err != nil {
			return nil, err
		}

		mergedSomething := false
		for _, comp := range path.Path {
			c, ok := comps.Get(comp.GetSafeName())
			if !ok {
				return nil, fmt.Errorf("unknown component %s in sampling path", comp.GetSafeName())
			}

			compConfig, err := c.GenerateConfig(ct, path, pathUserdata)
			if err != nil {
				return nil, err
			}
			if compConfig != nil {
				if err := composite.Merge(compConfig); err != nil {
					return nil, fmt.Errorf("failed to merge component config: %w", err)
				}
				mergedSomething = true
			}
		}

		if mergedSomething {
			composites = append(composites, composite)
		}
	}

	if len(composites) == 0 {
		unconfigured := config.UnconfiguredComponent{Component: dummy}
		return unconfigured.GenerateConfig(ct, hpsf.PathWithConnections{}, nil)
	}

	// Merge all composites
	finalConfig := composites[0]
	for _, comp := range composites[1:] {
		if err := finalConfig.Merge(comp); err != nil {
			return nil, fmt.Errorf("failed to merge refinery rules configs: %w", err)
		}
	}

	// For refinery rules, set the default environment from Router if present
	if rulesConfig, ok := finalConfig.(*tmpl.RulesConfig); ok {
		if defaultEnv, exists := userdata["router_default_env"]; exists {
			if defaultEnvStr, ok := defaultEnv.(string); ok {
				rulesConfig.SetDefaultEnv(defaultEnvStr)
			}
		}
	}

	return finalConfig, nil
}

// generateConfigWithRouters handles special pipeline generation when Router components are present.
// It creates intake pipelines (receiver → router) and environment-specific pipelines (router → exporter).
func (t *Translator) generateConfigWithRouters(h *hpsf.HPSF, comps *OrderedComponentMap, paths []hpsf.PathWithConnections, ct hpsftypes.Type, userdata map[string]any) (tmpl.TemplateConfig, error) {
	dummy := hpsf.Component{Name: "dummy", Kind: "dummy"}

	// Add the HPSF document to userdata so components can access it during generation
	if userdata == nil {
		userdata = make(map[string]any)
	}
	userdata["hpsf"] = h

	// For refinery rules, we need to handle sampling paths with environment context
	if ct == hpsftypes.RefineryRules {
		return t.generateRefineryRulesWithRouter(h, comps, paths, ct, userdata)
	}

	// Separate paths into those before and after routers
	// We'll build intake pipelines (receiver → router) and output pipelines (router → exporter)
	intakePaths := make([]hpsf.PathWithConnections, 0)
	outputPaths := make([]hpsf.PathWithConnections, 0)

	for _, path := range paths {
		// Find if this path contains a router
		routerIndex := -1
		for i, comp := range path.Path {
			if c, ok := comps.Get(comp.GetSafeName()); ok {
				if tc, ok := c.(*config.TemplateComponent); ok {
					if tc.Style == "router" {
						routerIndex = i
						break
					}
				}
			}
		}

		if routerIndex >= 0 {
			// Split the path at the router
			// Intake path: receiver → ... → router (router is last component in intake)
			if routerIndex >= 0 {
				intakePath := hpsf.PathWithConnections{
					Path:        path.Path[:routerIndex+1],
					Connections: path.Connections[:routerIndex], // Connections up to (but not including) the one leading to router
					ConnType:    path.ConnType,
				}
				intakePaths = append(intakePaths, intakePath)
			}

			// Output path: router → ... → exporter (router is first component in output)
			if routerIndex < len(path.Path)-1 {
				outputPath := hpsf.PathWithConnections{
					Path:        path.Path[routerIndex:],
					Connections: path.Connections[routerIndex:], // Connections from router onwards
					ConnType:    path.ConnType,
				}
				outputPaths = append(outputPaths, outputPath)
			}
		} else {
			// No router in this path, treat as regular path
			intakePaths = append(intakePaths, path)
		}
	}

	// Generate configs for intake paths (these will have routing connector as exporter)
	intakeComposites := make([]tmpl.TemplateConfig, 0)
	for i := range intakePaths {
		// Set custom pipeline name for intake paths: intake (signal type is prepended by template)
		intakePaths[i].PipelineName = "intake"

		base := config.GenericBaseComponent{Component: dummy}
		composite, err := base.GenerateConfig(ct, intakePaths[i], userdata)
		if err != nil {
			return nil, err
		}

		mergedSomething := false
		for _, comp := range intakePaths[i].Path {
			c, ok := comps.Get(comp.GetSafeName())
			if !ok {
				return nil, fmt.Errorf("unknown component %s in intake path", comp.GetSafeName())
			}

			compConfig, err := c.GenerateConfig(ct, intakePaths[i], userdata)
			if err != nil {
				return nil, err
			}
			if compConfig != nil {
				if err := composite.Merge(compConfig); err != nil {
					return nil, fmt.Errorf("failed to merge component config: %w", err)
				}
				mergedSomething = true
			}
		}
		if mergedSomething {
			// Add routing connector to the intake pipeline's connectors field
			// This is needed because the Router component no longer has a template that generates this
			if collectorConfig, ok := composite.(*tmpl.CollectorConfig); ok {
				signalType := intakePaths[i].ConnType.AsCollectorSignalType()
				pipelineName := intakePaths[i].GetID() // This will return "intake" because we set PipelineName
				connectorKey := fmt.Sprintf("pipelines.%s/%s.connectors", signalType, pipelineName)
				routingConnector := "routing/" + signalType

				if serviceSection, exists := collectorConfig.Sections["service"]; exists {
					// Check if connectors already exist and append, otherwise create new list
					if existing, ok := serviceSection[connectorKey]; ok {
						if existingList, ok := existing.([]string); ok {
							serviceSection[connectorKey] = append(existingList, routingConnector)
						}
					} else {
						serviceSection[connectorKey] = []string{routingConnector}
					}
				}
			}
			intakeComposites = append(intakeComposites, composite)
		}
	}

	// Build exporter to environment mapping for this generation
	// Get environment names from HoneycombExporter components' EnvironmentName property
	exporterToEnvironment := make(map[string]string)
	for _, hcomp := range h.Components {
		if hcomp.Kind == "HoneycombExporter" {
			safeName := hcomp.GetSafeName()
			// Get EnvironmentName property from exporter
			for _, prop := range hcomp.Properties {
				if prop.Name == "EnvironmentName" {
					if envName, ok := prop.Value.(string); ok && envName != "" {
						exporterToEnvironment[safeName] = envName
					}
					break
				}
			}
		}
	}

	// Generate configs for output paths (these will have routing connector as receiver)
	outputComposites := make([]tmpl.TemplateConfig, 0)
	for i, path := range outputPaths {
		// Check if this path contains a HoneycombExporter component and get its environment
		pathEnv := ""
		for _, comp := range path.Path {
			safeName := comp.GetSafeName()
			if env, ok := exporterToEnvironment[safeName]; ok {
				pathEnv = env
				break
			}
		}

		// Set custom pipeline name for output paths: env-name (signal type is prepended by template)
		if pathEnv != "" {
			outputPaths[i].PipelineName = pathEnv
		}

		// Create a copy of userdata for this path with environment context
		pathUserdata := make(map[string]any)
		for k, v := range userdata {
			pathUserdata[k] = v
		}
		if pathEnv != "" {
			pathUserdata["environment"] = pathEnv
		}

		base := config.GenericBaseComponent{Component: dummy}
		composite, err := base.GenerateConfig(ct, outputPaths[i], pathUserdata)
		if err != nil {
			return nil, err
		}

		mergedSomething := false
		// Skip the router component itself when generating output pipeline components
		// (we only want processors and exporters after the router)
		for j, comp := range outputPaths[i].Path {
			c, ok := comps.Get(comp.GetSafeName())
			if !ok {
				return nil, fmt.Errorf("unknown component %s in output path", comp.GetSafeName())
			}

			// Skip router component config generation for output paths
			if tc, ok := c.(*config.TemplateComponent); ok && tc.Style == "router" {
				continue
			}

			// For the first non-router component, we need to note that it comes from routing connector
			// This is handled by the path connection type
			if j == 0 {
				// This is the router itself, skip it
				continue
			}

			compConfig, err := c.GenerateConfig(ct, outputPaths[i], pathUserdata)
			if err != nil {
				return nil, err
			}
			if compConfig != nil {
				if err := composite.Merge(compConfig); err != nil {
					return nil, fmt.Errorf("failed to merge component config: %w", err)
				}
				mergedSomething = true
			}
		}
		if mergedSomething {
			outputComposites = append(outputComposites, composite)
		}
	}

	// Merge all composites
	allComposites := append(intakeComposites, outputComposites...)

	if len(allComposites) == 0 {
		unconfigured := config.UnconfiguredComponent{Component: dummy}
		return unconfigured.GenerateConfig(ct, hpsf.PathWithConnections{}, nil)
	}

	finalConfig := allComposites[0]
	for _, comp := range allComposites[1:] {
		if err := finalConfig.Merge(comp); err != nil {
			return nil, fmt.Errorf("failed to merge pipeline configs: %w", err)
		}
	}

	// Dynamically generate routing connector configuration
	// This replaces the static template that was in Router.yaml
	if collectorConfig, ok := finalConfig.(*tmpl.CollectorConfig); ok && len(exporterToEnvironment) > 0 {
		// Get unique environment names
		environments := make([]string, 0, len(exporterToEnvironment))
		envSet := make(map[string]bool)
		for _, envName := range exporterToEnvironment {
			if !envSet[envName] {
				environments = append(environments, envName)
				envSet[envName] = true
			}
		}

		// Determine which signal types are actually used in the paths
		usedSignalTypes := make(map[hpsf.ConnectionType]bool)
		for _, path := range paths {
			usedSignalTypes[path.ConnType] = true
		}

		// Get Router component properties
		var routingAttribute string
		var defaultEnvironment string
		for _, comp := range h.Components {
			if comp.Kind == "Router" {
				for _, prop := range comp.Properties {
					if prop.Name == "RoutingAttribute" {
						if val, ok := prop.Value.(string); ok {
							routingAttribute = val
						}
					} else if prop.Name == "DefaultEnvironment" {
						if val, ok := prop.Value.(string); ok {
							defaultEnvironment = val
						}
					}
				}
				// If RoutingAttribute wasn't specified, use the default from the Router component definition
				if routingAttribute == "" {
					// Get the Router component definition to check for default value
					if routerComp, ok := comps.Get(comp.GetSafeName()); ok {
						if tc, ok := routerComp.(*config.TemplateComponent); ok {
							for _, prop := range tc.Properties {
								if prop.Name == "RoutingAttribute" && prop.Default != nil {
									if defaultVal, ok := prop.Default.(string); ok {
										routingAttribute = defaultVal
									}
								}
							}
						}
					}
				}
				break
			}
		}

		// Create or get connectors section
		connectorsSection := collectorConfig.Sections["connectors"]
		if connectorsSection == nil {
			connectorsSection = make(map[string]any)
			collectorConfig.Sections["connectors"] = connectorsSection
		}

		// Generate routing connector config only for signal types that are actually used
		for _, connType := range hpsf.CollectorSignalTypes {
			// Skip signal types that aren't used in any paths
			if !usedSignalTypes[connType] {
				continue
			}

			signalType := connType.AsCollectorSignalType()
			routingKey := "routing/" + signalType

			// Set default_pipelines if DefaultEnvironment is specified
			if defaultEnvironment != "" {
				defaultPipelines := []string{signalType + "/" + defaultEnvironment}
				connectorsSection[routingKey+".default_pipelines"] = defaultPipelines
			}

			// Generate table entries for non-default environments
			tableIdx := 0
			for _, envName := range environments {
				if envName == defaultEnvironment {
					continue // Skip default environment
				}

				tableKey := fmt.Sprintf("%s.table[%d]", routingKey, tableIdx)
				connectorsSection[tableKey+".context"] = "resource"
				connectorsSection[tableKey+".condition"] = fmt.Sprintf("attributes[\"%s\"] == \"%s\"", routingAttribute, envName)
				connectorsSection[tableKey+".pipelines"] = []string{signalType + "/" + envName}
				tableIdx++
			}
		}
	}

	// Merge routing connectors and transform pipelines
	if collectorConfig, ok := finalConfig.(*tmpl.CollectorConfig); ok {
		// Only merge routing connectors if we didn't dynamically generate them
		// (dynamically generated connectors are already in the final merged format)
		if len(exporterToEnvironment) == 0 {
			if err := mergeRoutingConnectors(collectorConfig); err != nil {
				return nil, fmt.Errorf("failed to merge routing connectors: %w", err)
			}
		}

		// Transform pipelines to use routing connector properly
		// Intake pipelines (receiver → router) should have routing in exporters
		// Output pipelines (router → exporter) should have routing in receivers
		if err := transformRouterPipelines(collectorConfig, h, comps); err != nil {
			return nil, fmt.Errorf("failed to transform router pipelines: %w", err)
		}

		// Inject environment-specific API keys for all environment pipelines
		// This ensures exporters get the correct ${HTP_EXPORTER_APIKEY_ENV} variables
		if err := injectEnvironmentAPIKeys(collectorConfig, exporterToEnvironment); err != nil {
			return nil, fmt.Errorf("failed to inject environment API keys: %w", err)
		}
	}

	// For refinery rules, set the default environment from Router if present
	if rulesConfig, ok := finalConfig.(*tmpl.RulesConfig); ok {
		if defaultEnv, exists := userdata["router_default_env"]; exists {
			if defaultEnvStr, ok := defaultEnv.(string); ok {
				rulesConfig.SetDefaultEnv(defaultEnvStr)
			}
		}
	}

	return finalConfig, nil
}

func (t *Translator) GenerateConfig(h *hpsf.HPSF, ct hpsftypes.Type, artifactVersion string, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// Add the HPSF document to userdata so components can access it during generation
	if userdata == nil {
		userdata = make(map[string]any)
	}
	userdata["hpsf"] = h

	comps := NewOrderedComponentMap()
	receiverNames := make(map[string]bool)
	// make all the components
	visitFunc := func(c *hpsf.Component) error {
		comp, err := t.MakeConfigComponent(c, artifactVersion)
		if err != nil {
			return err
		}
		comps.Set(c.GetSafeName(), comp)
		if tc, ok := comp.(*config.TemplateComponent); ok {
			if tc.Style == "receiver" {
				receiverNames[c.GetSafeName()] = true
			}
		}
		return nil
	}

	if err := h.VisitComponents(visitFunc); err != nil {
		return nil, fmt.Errorf("failed to create components: %w", err)
	}

	// now add the connections
	for _, conn := range h.Connections {
		comp, ok := comps.Get(conn.Source.GetSafeName())
		if !ok {
			return nil, fmt.Errorf("unknown source component %s in connection", conn.Source.Component)
		}
		comp.AddConnection(conn)

		comp, ok = comps.Get(conn.Destination.GetSafeName())
		if !ok {
			return nil, fmt.Errorf("unknown target component %s in connection", conn.Destination.Component)
		}
		comp.AddConnection(conn)
	}

	// We need to generate our collection of unique paths. A pipeline in
	// this context is the shortest path from a source component to a
	// destination component. We iterate over all starting components (those
	// with no incoming connections) and all ending components (those with no
	// outgoing connections).
	paths := h.FindAllPaths(receiverNames)
	if len(paths) == 0 {
		// there were no complete paths found, so we construct dummy paths with all the components
		// so that all the unconnected components can play
		paths = []hpsf.PathWithConnections{
			{Path: h.Components, ConnType: hpsf.CTYPE_LOGS},
			{Path: h.Components, ConnType: hpsf.CTYPE_METRICS},
			{Path: h.Components, ConnType: hpsf.CTYPE_TRACES},
			{Path: h.Components, ConnType: hpsf.CTYPE_HONEY},
			{Path: h.Components, ConnType: hpsf.CTYPE_SAMPLE},
		}
	}

	// Order the paths using port index (if specified) as a secondary key.
	orderPaths(paths, comps)

	// we need a dummy component to start with so that we can always have a valid config
	dummy := hpsf.Component{Name: "dummy", Kind: "dummy"}
	composites := make([]tmpl.TemplateConfig, 0, len(paths))

	// Check if any paths contain Router components - if so, we need special handling
	hasRouter := false
	for comp := range comps.Items() {
		if tc, ok := comp.(*config.TemplateComponent); ok {
			if tc.Style == "router" {
				hasRouter = true
				break
			}
		}
	}

	// If we have routers, use special pipeline generation
	if hasRouter {
		return t.generateConfigWithRouters(h, comps, paths, ct, userdata)
	}

	// now we can iterate over the paths and generate a configuration for each
	for _, path := range paths {
		// Create a copy of userdata for this path
		pathUserdata := make(map[string]any)
		for k, v := range userdata {
			pathUserdata[k] = v
		}

		// For sampling paths, check if there's a SetEnvironment component and extract environment
		if path.ConnType == hpsf.CTYPE_SAMPLE {
			for _, comp := range path.Path {
				if comp.Kind == "SetEnvironment" {
					envNameProp := comp.GetProperty("EnvironmentName")
					if envNameProp != nil && envNameProp.Value != nil {
						if envName, ok := envNameProp.Value.(string); ok && envName != "" {
							pathUserdata["environment"] = envName
							break
						}
					}
				}
			}
		}

		// Start with a base component so we always have a valid config
		base := config.GenericBaseComponent{Component: dummy}
		composite, err := base.GenerateConfig(ct, path, pathUserdata)
		if err != nil {
			return nil, err
		}

		mergedSomething := false
		for _, comp := range path.Path {
			// look up the component in the ordered map
			c, ok := comps.Get(comp.GetSafeName())
			if !ok {
				return nil, fmt.Errorf("unknown component %s in path", comp.GetSafeName())
			}

			compConfig, err := c.GenerateConfig(ct, path, pathUserdata)
			if err != nil {
				return nil, err
			}
			if compConfig != nil {
				if err := composite.Merge(compConfig); err != nil {
					return nil, fmt.Errorf("failed to merge component config: %w", err)
				}
				mergedSomething = true
			}
		}
		if mergedSomething {
			composites = append(composites, composite)
		}
	}
	// If we have multiple pipelines, we need to merge them into a single config.
	if len(composites) > 1 {
		// We can use the Merge method to combine all the configurations into one.
		finalConfig := composites[0]
		for _, comp := range composites[1:] {
			if err := finalConfig.Merge(comp); err != nil {
				return nil, fmt.Errorf("failed to merge pipeline configs: %w", err)
			}
		}

		// Post-process: merge multiple routing connectors into a single one
		if collectorConfig, ok := finalConfig.(*tmpl.CollectorConfig); ok {
			if err := mergeRoutingConnectors(collectorConfig); err != nil {
				return nil, fmt.Errorf("failed to merge routing connectors: %w", err)
			}
		}

		// For refinery rules, set the default environment from Router if present
		if rulesConfig, ok := finalConfig.(*tmpl.RulesConfig); ok {
			if defaultEnv, exists := userdata["router_default_env"]; exists {
				if defaultEnvStr, ok := defaultEnv.(string); ok {
					rulesConfig.SetDefaultEnv(defaultEnvStr)
				}
			}
		}

		return finalConfig, nil
	} else if len(composites) == 1 {
		// If we only have one pipeline, we can return it directly.
		config := composites[0]

		// Post-process: merge multiple routing connectors into a single one
		if collectorConfig, ok := config.(*tmpl.CollectorConfig); ok {
			if err := mergeRoutingConnectors(collectorConfig); err != nil {
				return nil, fmt.Errorf("failed to merge routing connectors: %w", err)
			}
		}

		// For refinery rules, set the default environment from Router if present
		if rulesConfig, ok := config.(*tmpl.RulesConfig); ok {
			if defaultEnv, exists := userdata["router_default_env"]; exists {
				if defaultEnvStr, ok := defaultEnv.(string); ok {
					rulesConfig.SetDefaultEnv(defaultEnvStr)
				}
			}
		}

		return config, nil
	}

	// Start with a base component so we always have a valid config
	unconfigured := config.UnconfiguredComponent{Component: dummy}
	return unconfigured.GenerateConfig(ct, hpsf.PathWithConnections{}, nil)
}
