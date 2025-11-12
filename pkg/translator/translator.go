package translator

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"sort"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/honeycombio/hpsf/pkg/validator"
	"golang.org/x/mod/semver"
)

const LatestVersion = "latest"

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
func artifactVersionSupported(component config.TemplateComponent, v string) error {
	if v == "" || v == LatestVersion {
		return nil
	}

	// ensure the version string is prefixed with v otherwise semver.Compare fails
	// to parse the version
	if v[0] != 'v' {
		v = "v" + v
	}

	if component.Minimum != "" && semver.Compare(v, component.Minimum) < 0 {
		return NewVersionError(fmt.Sprintf("agent version %s does not meet component %s requirement minimum version of %s", v, component.Kind, component.Minimum))
	}

	if component.Maximum != "" && semver.Compare(v, component.Maximum) > 0 {
		return NewVersionError(fmt.Sprintf("agent version %s does not meet component %s requirement maximum version of %s", v, component.Kind, component.Maximum))
	}

	return nil
}

// componentVersionSupported checks if the requested version is compatible with the template version using semver.
// It allows patch and minor version upgrades but prevents major version mismatches.
func componentVersionSupported(templateVersion, requestedVersion string) bool {
	// If no version specified, accept any template version
	if requestedVersion == "" {
		return true
	}

	// If template has no version, only accept empty requested version
	if templateVersion == "" {
		return requestedVersion == ""
	}

	// Check if both versions are valid semver
	if !semver.IsValid(templateVersion) || !semver.IsValid(requestedVersion) {
		// Fall back to string equality for invalid semver
		return templateVersion == requestedVersion
	}

	// Check major version compatibility - must match
	if semver.Major(templateVersion) != semver.Major(requestedVersion) {
		return false
	}

	// Template version must be >= requested version (allows upgrades)
	return semver.Compare(templateVersion, requestedVersion) >= 0
}

var _ error = &VersionError{}

type VersionError struct {
	msg string
}

func (e *VersionError) Error() string {
	return e.msg
}

func (e *VersionError) Is(target error) bool {
	_, ok := target.(*VersionError)
	return ok
}

func (e *VersionError) As(target interface{}) bool {
	if t, ok := target.(**VersionError); ok {
		*t = e
		return true
	}
	return false
}

func NewVersionError(msg string) *VersionError {
	return &VersionError{msg: msg}
}

func (t *Translator) MakeConfigComponent(component *hpsf.Component, artifactVersion string) (config.Component, error) {
	// first look in the template components
	tc, ok := t.components[component.Kind]
	if !ok {
		return nil, fmt.Errorf("unknown component kind: %s@%s", component.Kind, component.Version)
	}

	if !componentVersionSupported(tc.Version, component.Version) {
		return nil, NewVersionError(fmt.Sprintf("component %s at version %s is unsupported by agent version %s", component.Kind, tc.Version, artifactVersion))
	}
	if err := artifactVersionSupported(tc, artifactVersion); err != nil {
		return nil, err
	}

	// found it, manufacture a new instance of the component
	tc.SetHPSF(component)
	return &tc, nil
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
		tc, ok := t.components[c.Kind]
		if ok && componentVersionSupported(tc.Version, c.Version) {
			templateComps[c.GetSafeName()] = tc
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
		return finalConfig, nil
	} else if len(composites) == 1 {
		// If we only have one pipeline, we can return it directly.
		return composites[0], nil
	}

	// Start with a base component so we always have a valid config
	unconfigured := config.UnconfiguredComponent{Component: dummy}
	return unconfigured.GenerateConfig(ct, hpsf.PathWithConnections{}, nil)
}

// ComponentInfo represents a component extracted from an HPSF configuration.
// It contains the component's identifying information (name, style, kind) and all
// of its properties, including both explicitly set values and template defaults.
type ComponentInfo struct {
	// Name is the user-defined name of the component instance (e.g., "My S3 Archive")
	Name string
	// Style categorizes the component type: "receiver", "processor", or "exporter"
	Style string
	// Kind identifies the specific component template (e.g., "HoneycombExporter", "OTelReceiver")
	Kind string
	// Version is the component version. If specified in the HPSF document, that version is returned.
	// Otherwise, falls back to the template component's version.
	Version string
	// Properties contains all component properties, merging explicit values with template defaults.
	// Access values directly without type casting: properties["Region"]
	Properties map[string]any
}

// InspectionResult holds all components extracted from an HPSF configuration.
// Access components directly via the Components field, or use the filter methods
// (Exporters, Receivers, Processors) to get components by style.
type InspectionResult struct {
	Components []ComponentInfo
}

// Filter returns a new InspectionResult containing only components that match any of the given predicates.
// This is useful if you need multiple filtering passes on an existing InspectionResult.
// Predicates are ORed together, not ANDed.
// Several common predicates are provided (Exporters, Receivers, Processors, Samplers).
func (r InspectionResult) Filter(predicates ...Predicate) InspectionResult {
	filtered := InspectionResult{
		Components: []ComponentInfo{},
	}

	for _, c := range r.Components {
		matched := false
		for _, p := range predicates {
			if p(c) {
				matched = true
				break
			}
		}
		if matched {
			filtered.Components = append(filtered.Components, c)
		}
	}

	return filtered
}

type Predicate func(ComponentInfo) bool

// Exporters returns true if the component is an exporter.
func Exporters(c ComponentInfo) bool {
	return c.Style == "exporter"
}

// Processors returns true if the component is a processor.
func Processors(c ComponentInfo) bool {
	return c.Style == "processor"
}

// Receivers returns true if the component is a receiver.
func Receivers(c ComponentInfo) bool {
	return c.Style == "receiver"
}

// Samplers returns true if the component is a dropper, condition, sampler, startsampling.
func Samplers(c ComponentInfo) bool {
	switch c.Style {
	case "condition", "dropper", "sampler", "startsampling":
		return true
	default:
		return false
	}
}

// Inspect extracts all components from the HPSF document.
// It returns an InspectionResult containing all components.
// InspectionResult provides filtering methods to get sub sets of components by style.
func (t *Translator) Inspect(h hpsf.HPSF) InspectionResult {
	result := InspectionResult{
		Components: []ComponentInfo{},
	}

	// Iterate through all components
	for _, c := range h.Components {
		// Look up the template for this component
		tc, ok := t.components[c.Kind]
		if !ok {
			continue
		}

		// Determine version: use HPSF component version if specified, otherwise use template version
		version := c.Version
		if version == "" {
			version = tc.Version
		}

		comp := ComponentInfo{
			Name:       c.Name,
			Style:      tc.Style,
			Kind:       c.Kind,
			Version:    version,
			Properties: getProperties(c, tc),
		}

		// Add component to result
		result.Components = append(result.Components, comp)
	}

	return result
}

// getProperties extracts all properties from a component, using template defaults as fallback
func getProperties(c *hpsf.Component, tc config.TemplateComponent) map[string]any {
	properties := make(map[string]any)

	// Iterate through template properties to ensure all defaults are considered
	for _, templateProperty := range tc.Properties {
		// Use the component's property value if set, otherwise use the template default
		var value any
		if componentProp := c.GetProperty(templateProperty.Name); componentProp != nil {
			value = componentProp.Value
		} else {
			value = templateProperty.Default
		}
		properties[templateProperty.Name] = value
	}

	return properties
}
