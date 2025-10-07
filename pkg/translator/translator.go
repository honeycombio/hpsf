package translator

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"sort"
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

// mergeRoutingConnectors finds all routing/* connectors and merges them into a single routing connector
func mergeRoutingConnectors(cc *tmpl.CollectorConfig) error {
	connectorSection, ok := cc.Sections["connectors"]
	if !ok {
		return nil // no connectors, nothing to do
	}

	// Find all unique routing connector names
	routingConnectorNames := make(map[string]bool)
	for key := range connectorSection {
		if strings.HasPrefix(key, "routing/") {
			// Extract connector name (e.g., "routing/router_production" from "routing/router_production.table.0.statement")
			parts := strings.SplitN(key, ".", 2)
			connectorName := parts[0]
			routingConnectorNames[connectorName] = true
		}
	}

	routingConnectors := make([]string, 0, len(routingConnectorNames))
	for name := range routingConnectorNames {
		routingConnectors = append(routingConnectors, name)
	}

	// Sort for deterministic output
	sort.Strings(routingConnectors)

	if len(routingConnectors) <= 1 {
		return nil // 0 or 1 routing connector, nothing to merge
	}

	// Collect all routing rules
	var defaultPipelines []string
	tableEntries := make([]map[string]any, 0)

	for _, connectorKey := range routingConnectors {
		// Check for default_pipelines
		if defaultPipelinesKey := connectorKey + ".default_pipelines"; connectorSection[defaultPipelinesKey] != nil {
			if pipelines, ok := connectorSection[defaultPipelinesKey].([]string); ok {
				defaultPipelines = pipelines
			}
		}

		// Check for table entries (they have format routing/name.table[0].statement, routing/name.table[0].pipelines)
		// We need to find all table[N] entries for this connector
		tablePrefix := connectorKey + ".table["
		tableIndices := make(map[string]bool)
		for key := range connectorSection {
			if strings.HasPrefix(key, tablePrefix) {
				// Extract the index (e.g., "0" from "routing/name.table[0].statement")
				after := strings.TrimPrefix(key, tablePrefix)
				closeBracket := strings.Index(after, "]")
				if closeBracket > 0 {
					idx := after[:closeBracket]
					tableIndices[idx] = true
				}
			}
		}

		// Collect table entries
		for idx := range tableIndices {
			entry := make(map[string]any)
			statementKey := connectorKey + ".table[" + idx + "].statement"
			pipelinesKey := connectorKey + ".table[" + idx + "].pipelines"

			if statement, ok := connectorSection[statementKey].(string); ok {
				entry["statement"] = statement
			}
			if pipelines, ok := connectorSection[pipelinesKey].([]string); ok {
				entry["pipelines"] = pipelines
			}

			if len(entry) > 0 {
				tableEntries = append(tableEntries, entry)
			}
		}

		// Delete the old routing connector keys
		for key := range connectorSection {
			if strings.HasPrefix(key, connectorKey) {
				delete(connectorSection, key)
			}
		}
	}

	// Create the merged routing connector
	if len(defaultPipelines) > 0 {
		connectorSection["routing.default_pipelines"] = defaultPipelines
	}

	if len(tableEntries) > 0 {
		// Add table entries using indexed format (e.g., routing.table[0], routing.table[1])
		for i, entry := range tableEntries {
			if statement, ok := entry["statement"].(string); ok {
				key := fmt.Sprintf("routing.table[%d].statement", i)
				connectorSection[key] = statement
			}
			if pipelines, ok := entry["pipelines"].([]string); ok {
				key := fmt.Sprintf("routing.table[%d].pipelines", i)
				connectorSection[key] = pipelines
			}
		}
	}

	// Update pipeline connector references from routing/* to routing
	if serviceSection, ok := cc.Sections["service"]; ok {
		for key, value := range serviceSection {
			if strings.Contains(key, ".connectors") {
				if connectors, ok := value.([]string); ok {
					updated := false
					for i, conn := range connectors {
						// Replace any routing/* connector with the merged routing connector
						if strings.HasPrefix(conn, "routing/") {
							connectors[i] = "routing"
							updated = true
						}
					}
					if updated {
						serviceSection[key] = connectors
					}
				}
			}
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
			if startSamplingCount > 1 {
				err := hpsf.NewError("only one StartSampling component is allowed").
					WithComponent(c.Name)
				result.Add(err)
			}
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
func transformRouterPipelines(cc *tmpl.CollectorConfig) error {
	serviceSection, exists := cc.Sections["service"]
	if !exists {
		return nil
	}

	// First, collect all unique pipeline names
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

	// Process each pipeline once
	for pipelineName := range pipelineNames {
		receiversKey := fmt.Sprintf("pipelines.%s.receivers", pipelineName)
		connectorsKey := fmt.Sprintf("pipelines.%s.connectors", pipelineName)
		exportersKey := fmt.Sprintf("pipelines.%s.exporters", pipelineName)

		receivers, hasReceivers := serviceSection[receiversKey]
		connectors, hasConnectors := serviceSection[connectorsKey]
		exporters, hasExporters := serviceSection[exportersKey]

		// Get connectors list if it exists
		var connectorsList []string
		var hasRouting bool
		if hasConnectors {
			switch v := connectors.(type) {
			case []any:
				for _, c := range v {
					if s, ok := c.(string); ok {
						connectorsList = append(connectorsList, s)
					}
				}
			case []string:
				connectorsList = v
			}

			// Check if this has a routing connector
			for _, conn := range connectorsList {
				if strings.HasPrefix(conn, "routing") {
					hasRouting = true
					break
				}
			}
		}

		// Skip if this pipeline has connectors but no routing connector
		if hasConnectors && !hasRouting {
			continue
		}

		// Determine if this is an intake or output pipeline
		// Receivers and exporters can be []any or []string
		var receiversList []string
		if hasReceivers && receivers != nil {
			switch v := receivers.(type) {
			case []any:
				for _, r := range v {
					if s, ok := r.(string); ok {
						receiversList = append(receiversList, s)
					}
				}
			case []string:
				receiversList = v
			}
		}

		var exportersList []string
		if hasExporters && exporters != nil {
			switch v := exporters.(type) {
			case []any:
				for _, e := range v {
					if s, ok := e.(string); ok {
						exportersList = append(exportersList, s)
					}
				}
			case []string:
				exportersList = v
			}
		}

		// Filter out empty strings from lists (sometimes empty arrays contain empty strings)
		receiversFiltered := make([]string, 0)
		for _, r := range receiversList {
			if r != "" {
				receiversFiltered = append(receiversFiltered, r)
			}
		}

		exportersFiltered := make([]string, 0)
		for _, e := range exportersList {
			if e != "" {
				exportersFiltered = append(exportersFiltered, e)
			}
		}

		// Case 1: Intake pipeline - has receivers and routing connector, no exporters → move connector to exporters
		if hasRouting && len(receiversFiltered) > 0 && len(exportersFiltered) == 0 {
			// Move connectors to exporters
			serviceSection[exportersKey] = connectors
			delete(serviceSection, connectorsKey)
		} else if hasRouting && len(receiversFiltered) == 0 && len(exportersFiltered) > 0 {
			// Case 2: Output pipeline - has routing connector, no receivers, has exporters → move connector to receivers
			serviceSection[receiversKey] = connectors
			delete(serviceSection, connectorsKey)
		} else if !hasRouting && len(receiversFiltered) == 0 && len(exportersFiltered) > 0 {
			// Case 3: Output pipeline without routing connector - no receivers, has exporters → add routing to receivers
			serviceSection[receiversKey] = []string{"routing"}
		}
	}

	return nil
}

// generateConfigWithRouters handles special pipeline generation when Router components are present.
// It creates intake pipelines (receiver → router) and environment-specific pipelines (router → exporter).
func (t *Translator) generateConfigWithRouters(h *hpsf.HPSF, comps *OrderedComponentMap, paths []hpsf.PathWithConnections, ct hpsftypes.Type, userdata map[string]any) (tmpl.TemplateConfig, error) {
	dummy := hpsf.Component{Name: "dummy", Kind: "dummy"}

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
					Path:     path.Path[:routerIndex+1],
					ConnType: path.ConnType,
				}
				intakePaths = append(intakePaths, intakePath)
			}

			// Output path: router → ... → exporter (router is first component in output)
			if routerIndex < len(path.Path) {
				outputPath := hpsf.PathWithConnections{
					Path:     path.Path[routerIndex:],
					ConnType: path.ConnType,
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
	for _, path := range intakePaths {
		base := config.GenericBaseComponent{Component: dummy}
		composite, err := base.GenerateConfig(ct, path, userdata)
		if err != nil {
			return nil, err
		}

		mergedSomething := false
		for _, comp := range path.Path {
			c, ok := comps.Get(comp.GetSafeName())
			if !ok {
				return nil, fmt.Errorf("unknown component %s in intake path", comp.GetSafeName())
			}

			compConfig, err := c.GenerateConfig(ct, path, userdata)
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
			intakeComposites = append(intakeComposites, composite)
		}
	}

	// Generate configs for output paths (these will have routing connector as receiver)
	outputComposites := make([]tmpl.TemplateConfig, 0)
	for _, path := range outputPaths {
		base := config.GenericBaseComponent{Component: dummy}
		composite, err := base.GenerateConfig(ct, path, userdata)
		if err != nil {
			return nil, err
		}

		mergedSomething := false
		// Skip the router component itself when generating output pipeline components
		// (we only want processors and exporters after the router)
		for i, comp := range path.Path {
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
			if i == 0 {
				// This is the router itself, skip it
				continue
			}

			compConfig, err := c.GenerateConfig(ct, path, userdata)
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

	// Merge routing connectors and transform pipelines
	if collectorConfig, ok := finalConfig.(*tmpl.CollectorConfig); ok {
		if err := mergeRoutingConnectors(collectorConfig); err != nil {
			return nil, fmt.Errorf("failed to merge routing connectors: %w", err)
		}

		// Transform pipelines to use routing connector properly
		// Intake pipelines (receiver → router) should have routing in exporters
		// Output pipelines (router → exporter) should have routing in receivers
		if err := transformRouterPipelines(collectorConfig); err != nil {
			return nil, fmt.Errorf("failed to transform router pipelines: %w", err)
		}
	}

	return finalConfig, nil
}

func (t *Translator) GenerateConfig(h *hpsf.HPSF, ct hpsftypes.Type, artifactVersion string, userdata map[string]any) (tmpl.TemplateConfig, error) {
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
		// Start with a base component so we always have a valid config
		base := config.GenericBaseComponent{Component: dummy}
		composite, err := base.GenerateConfig(ct, path, userdata)
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

			compConfig, err := c.GenerateConfig(ct, path, userdata)
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

		return config, nil
	}

	// Start with a base component so we always have a valid config
	unconfigured := config.UnconfiguredComponent{Component: dummy}
	return unconfigured.GenerateConfig(ct, hpsf.PathWithConnections{}, nil)
}
