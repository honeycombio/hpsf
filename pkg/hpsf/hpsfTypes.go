package hpsf

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/dgryski/go-metro"
	"github.com/honeycombio/hpsf/pkg/validator"
	y "gopkg.in/yaml.v3"
)

type ConnectionType string

const (
	CTYPE_UNKNOWN ConnectionType = "unknown"
	CTYPE_TRACES  ConnectionType = "OTelTraces"
	CTYPE_LOGS    ConnectionType = "OTelLogs"
	CTYPE_METRICS ConnectionType = "OTelMetrics"
	CTYPE_EVENT   ConnectionType = "OTelEvent"
	CTYPE_HONEY   ConnectionType = "Honeycomb"
	CTYPE_NUMBER  ConnectionType = "number"
	CTYPE_STRING  ConnectionType = "string"
	CTYPE_BOOL    ConnectionType = "bool"
)

func (c ConnectionType) AsCollectorType() string {
	switch c {
	case CTYPE_TRACES:
		return "traces"
	case CTYPE_LOGS:
		return "logs"
	case CTYPE_METRICS:
		return "metrics"
	case CTYPE_EVENT:
		return "events"
	default:
		return string(c)
	}
}

type PropType string

const (
	PTYPE_INT    PropType = "int"
	PTYPE_FLOAT  PropType = "float"
	PTYPE_STRING PropType = "string"
	PTYPE_BOOL   PropType = "bool"
	PTYPE_ARRSTR PropType = "stringarray" // []string
	PTYPE_MAPSTR PropType = "map"         // map[string]any
	PTYPE_COND   PropType = "conditions"  // for refinery conditions
)

func (p PropType) Validate() error {
	switch p {
	case PTYPE_INT:
	case PTYPE_FLOAT:
	case PTYPE_STRING:
	case PTYPE_BOOL:
	case PTYPE_ARRSTR:
	case PTYPE_MAPSTR:
	case PTYPE_COND:
	default:
		return errors.New("invalid PropType '" + string(p) + "'")
	}
	return nil
}

// String returns the string representation of the PropType.
func (p PropType) String() string {
	return string(p)
}

// ValueCoerce takes a value and coerces it to the type specified by the
// PropType, if possible, storing the result in target. We try to be as
// forgiving as possible here -- for example, if the PropType is INT and value
// is a string that can be parsed as an int, we will parse it and store the
// result in target. If the value cannot be coerced to the desired type, an
// error is returned. We use this to ensure that all the values in a configuration
// are of the correct type before we try to use them.
func (p PropType) ValueCoerce(a any, target *any) error {
	// empty proptype means anything goes
	if p == "" {
		*target = a
		return nil
	}
	switch p {
	case PTYPE_INT:
		switch v := a.(type) {
		case int:
			*target = v
		case float64:
			if float64(int(v)) != v {
				return errors.New("expected int, got " + fmt.Sprint(a))
			}
			*target = int(v)
		case string:
			i, err := strconv.Atoi(v)
			if err != nil {
				return errors.New("expected int, got " + fmt.Sprint(a))
			}
			*target = i
		default:
			return errors.New("expected int, got " + fmt.Sprint(a))
		}
	case PTYPE_FLOAT:
		switch v := a.(type) {
		case int:
			*target = float64(v)
		case float64:
			*target = v
		case string:
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return errors.New("expected float, got " + fmt.Sprint(a))
			}
			*target = f
		default:
			return errors.New("expected float, got " + fmt.Sprint(a))
		}
	case PTYPE_STRING:
		switch v := a.(type) {
		case int, float64, bool:
			*target = fmt.Sprint(a)
		case string:
			*target = v
		default:
			return errors.New("expected string, got " + fmt.Sprint(a))
		}
	case PTYPE_BOOL:
		switch v := a.(type) {
		case bool:
			*target = v
		case int:
			*target = v != 0
		case float64:
			*target = v != 0
		case string:
			switch v {
			case "true", "True", "TRUE", "YES", "yes", "Yes", "T", "t", "Y", "y":
				*target = true
			case "false", "False", "FALSE", "NO", "no", "No", "F", "f", "N", "n":
				*target = false
			default:
				return errors.New("expected bool, got " + fmt.Sprint(a))
			}
		default:
			return errors.New("expected bool, got " + fmt.Sprint(a))
		}
	case PTYPE_ARRSTR:
		switch v := a.(type) {
		case []string:
			*target = v
		case []any:
			sa := make([]string, len(v))
			for i, a := range v {
				// whatever it was, make it a string
				sa[i] = fmt.Sprint(a)
			}
			*target = sa
		default:
			return errors.New("expected string array, got " + fmt.Sprint(a))
		}
	case PTYPE_MAPSTR:
		switch v := a.(type) {
		case map[string]any:
			*target = v
		default:
			return errors.New("expected dictionary, got " + fmt.Sprint(a))
		}
	case PTYPE_COND:
		switch v := a.(type) {
		case map[string]any:
			*target = v
		default:
			return errors.New("expected dictionary, got " + fmt.Sprint(a))
		}
	default:
		return errors.New("invalid PropType '" + string(p) + "'")
	}
	return nil
}

func (p PropType) ValueConforms(a any) error {
	// null proptype means anything goes
	if p == "" {
		return nil
	}
	switch p {
	case PTYPE_INT:
		if _, ok := a.(int); !ok {
			return errors.New("expected int, got " + fmt.Sprint(a))
		}
	case PTYPE_FLOAT:
		if _, ok := a.(float64); !ok {
			return errors.New("expected float, got " + fmt.Sprint(a))
		}
	case PTYPE_STRING:
		if _, ok := a.(string); !ok {
			return errors.New("expected string, got " + fmt.Sprint(a))
		}
	case PTYPE_BOOL:
		if _, ok := a.(bool); !ok {
			return errors.New("expected bool, got " + fmt.Sprint(a))
		}
	case PTYPE_ARRSTR:
		if _, ok := a.([]string); !ok {
			return errors.New("expected []string, got " + fmt.Sprint(a))
		}
	case PTYPE_MAPSTR:
		if _, ok := a.(map[string]any); !ok {
			return errors.New("expected map[string]any, got " + fmt.Sprint(a))
		}
	default:
		return errors.New("invalid PropType '" + string(p) + "'")
	}
	return nil
}

type Direction string

const (
	DIR_INPUT  Direction = "input"
	DIR_OUTPUT Direction = "output"
)

type Port struct {
	Name      string         `yaml:"name"`
	Direction string         `yaml:"direction"`
	Type      ConnectionType `yaml:"type"`
}

type Property struct {
	Name  string   `yaml:"name"`
	Value any      `yaml:"value"`
	Type  PropType `yaml:"type"`
}

type Component struct {
	Name       string     `yaml:"name"`
	Kind       string     `yaml:"kind"`
	Ports      []Port     `yaml:"ports,omitempty"`
	Properties []Property `yaml:"properties,omitempty"`
	Style      string     `yaml:"style,omitempty"`
}

type ErrorSeverity string

const (
	SEV_ERROR ErrorSeverity = "E"
	SEV_WARN  ErrorSeverity = "W"
)

type HPSFError struct {
	Severity  ErrorSeverity `yaml:"severity"`
	Component string        `yaml:"component,omitempty"`
	Property  string        `yaml:"property,omitempty"`
	Reason    string        `yaml:"reason"`
	Cause     error         `yaml:"cause,omitempty"`
}

func (e *HPSFError) Error() string {
	err := fmt.Sprintf("%s: %s", e.Severity, e.Reason)
	if e.Component != "" {
		err += fmt.Sprintf(" Component: %s", e.Component)
	}
	if e.Property != "" {
		err += fmt.Sprintf(" Property: %s", e.Property)
	}
	if e.Cause != nil {
		err += fmt.Sprintf(" Cause: %s", e.Cause)
	}
	return err
}

func (e *HPSFError) Unwrap() error {
	return e.Cause
}

func (e *HPSFError) WithComponent(c string) *HPSFError {
	e.Component = c
	return e
}

func (e *HPSFError) WithProperty(p string) *HPSFError {
	e.Property = p
	return e
}

// WithCause accepts an error that will be used to populate the Cause field of the HPSFError struct.
// This allows you to wrap another error inside an HPSFError, which can be useful for debugging.
func (e *HPSFError) WithCause(c error) *HPSFError {
	e.Cause = c
	return e
}

func NewError(reason string) *HPSFError {
	return &HPSFError{
		Severity: SEV_ERROR,
		Reason:   reason,
	}
}

func NewWarning(reason string) *HPSFError {
	return &HPSFError{
		Severity: SEV_WARN,
		Reason:   reason,
	}
}

func (c *Component) Validate() error {
	result := validator.NewResult("component validation errors")
	if c.Name == "" {
		result.Add(NewError("Name must be set"))
	}
	if c.Kind == "" {
		result.Add(NewError("Kind must be set").WithComponent(c.Name))
	}
	// base components mentioned in typical configurations don't need to set up
	// ports, because those come from the templatecomponents, but composite
	// components might have ports, so we do want to check them if they exist
	for _, p := range c.Ports {
		if p.Direction != string(DIR_INPUT) && p.Direction != string(DIR_OUTPUT) {
			result.Add(NewError("Port " + p.Name + " Direction must be 'Input' or 'Output'").WithComponent(c.Name))
		}
	}
	// any properties specified need to have a value
	for _, p := range c.Properties {
		if p.Name == "" {
			result.Add(NewError("Property Name must be set").WithComponent(c.Name))
		}
		if p.Type != "" {
			if err := p.Type.Validate(); err != nil {
				result.Add(NewError("Type is invalid").WithComponent(c.Name).WithProperty(p.Name).WithCause(err))
			}
		}
		if p.Value == nil {
			result.Add(NewError("Value must be set").WithComponent(c.Name).WithProperty(p.Name))
			// can't check values after this
			continue
		}

		// we can only support specific types for the values we get from the YAML, so we coerce the values
		// we have to the types we expect
		switch p.Value.(type) {
		case string, int, float64, bool, []any, []string, map[string]any:
			err := p.Type.ValueCoerce(p.Value, &p.Value)
			if err != nil {
				result.Add(NewError("Value error").WithComponent(c.Name).WithProperty(p.Name).WithCause(err))
			}
		default:
			result.Add(NewError("Value must be a string, number, bool, array, or dictionary").WithComponent(c.Name).WithProperty(p.Name))
		}

		// This is a sanity check; belt and suspenders since the above should have done it right.
		// This was the first implementation, and we should be able to delete it once we're comfortable.
		err := p.Type.ValueConforms(p.Value)
		if err != nil {
			result.Add(NewError("Value does not conform").WithComponent(c.Name).WithProperty(p.Name).WithCause(err))
		}
	}
	return result.ErrOrNil()
}

func safeName(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return re.ReplaceAllString(s, "_")
}

// Returns the safe name of the component (no spaces or special characters)
// This has potential to cause a problem if the resulting name is not unique -- so uniqueness
// should be tested with this name, not the original name.
// we replace any runs of characters not in [a-zA-Z0-9] with an underscore
func (c *Component) GetSafeName() string {
	return safeName(c.Name)
}

// returns the port with the given name, or nil if not found
func (c *Component) GetPort(name string) *Port {
	for _, p := range c.Ports {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// returns the property with the given name, or nil if not found
func (c *Component) GetProperty(name string) *Property {
	for _, p := range c.Properties {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// returns all specified property names as a slice of strings
func (c *Component) GetPropertyNames() []string {
	props := make([]string, len(c.Properties))
	for i, p := range c.Properties {
		props[i] = p.Name
	}
	return props
}

type ConnectionPort struct {
	Component string         `yaml:"component"`
	PortName  string         `yaml:"port"`
	Type      ConnectionType `yaml:"type"`
}

func (cp *ConnectionPort) GetSafeName() string {
	return safeName(cp.Component)
}

func (cp *ConnectionPort) Validate() error {
	result := validator.NewResult("connection port validation errors")
	if cp.Component == "" {
		result.Add(validator.NewResult("ConnectionPort Component must be set"))
	}
	if cp.PortName == "" {
		result.Add(validator.NewResult("ConnectionPort PortName must be set"))
	}
	if cp.Type == "" {
		result.Add(validator.NewResult("ConnectionPort Type must be set"))
	}
	return result
}

type Connection struct {
	Source      ConnectionPort `yaml:"source"`
	Destination ConnectionPort `yaml:"destination"`
}

func (c *Connection) Validate() error {
	result := validator.NewResult("connection validation errors")
	e := c.Source.Validate()
	result.Add(e)
	e = c.Destination.Validate()
	result.Add(e)
	return result
}

type PublicPort struct {
	Name      string `yaml:"name"`
	Component string `yaml:"component"`
	Port      string `yaml:"port"`
}

func (pp *PublicPort) Validate() error {
	result := validator.NewResult("port validation errors")
	if pp.Name == "" {
		result.Add(validator.NewResult("PublicPort Name must be set"))
	}
	if pp.Component == "" {
		result.Add(validator.NewResult("PublicPort Component must be set"))
	}
	if pp.Port == "" {
		result.Add(validator.NewResult("PublicPort Port must be set"))
	}
	return result
}

type PublicProp struct {
	Name      string `yaml:"name"`
	Component string `yaml:"component"`
	Property  string `yaml:"property"`
}

func (pp *PublicProp) Validate() error {
	result := validator.NewResult("prop validation errors")
	if pp.Name == "" {
		result.Add(validator.NewResult("PublicProp Name must be set"))
	}
	if pp.Component == "" {
		result.Add(validator.NewResult("PublicProp Component must be set"))
	}
	if pp.Property == "" {
		result.Add(validator.NewResult("PublicProp Property must be set"))
	}
	return result
}

type Container struct {
	Name       string       `yaml:"name"`
	Components []Component  `yaml:"components,omitempty"`
	Ports      []PublicPort `yaml:"ports,omitempty"`
	Props      []PublicProp `yaml:"props,omitempty"`
}

func (c *Container) Validate() error {
	result := validator.NewResult("container validation errors")
	if c.Name == "" {
		result.Add(validator.NewResult("Container Name must be set"))
	}
	for _, p := range c.Ports {
		e := p.Validate()
		result.Add(e)
	}
	for _, p := range c.Props {
		e := p.Validate()
		result.Add(e)
	}
	return result
}

// placeholder for where we'll store layout information later
type Layout map[string]any

type HPSF struct {
	Kind        string        `yaml:"kind"`
	Version     string        `yaml:"version"`
	Name        string        `yaml:"name"`
	Summary     string        `yaml:"summary"`
	Description string        `yaml:"description"`
	Components  []*Component  `yaml:"components,omitempty"`
	Connections []*Connection `yaml:"connections,omitempty"`
	Containers  []Container   `yaml:"containers,omitempty"`
	Layout      Layout        `yaml:"layout,omitempty"`
}

// generate a list of components that are not named as the destination of a connection
func (h *HPSF) GetStartComponents() []*Component {
	startComps := make([]*Component, 0)
	// make a map of all components that are destinations of connections
	destinations := make(map[string]bool)
	for _, conn := range h.Connections {
		destinations[conn.Destination.Component] = true
	}
	for _, c := range h.Components {
		// if the component is not a destination of a connection, add it to the list
		if !destinations[c.Name] {
			startComps = append(startComps, c)
		}
	}

	return startComps
}

func (h *HPSF) getComponent(name string) *Component {
	// find the component with the given name
	for _, c := range h.Components {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (h *HPSF) isSourceComponent(c *Component, signalType ConnectionType) bool {
	// check if the component is a source of a connection of this signal type
	for _, conn := range h.Connections {
		if conn.Source.Component == c.Name && conn.Source.Type == signalType {
			return true
		}
	}
	return false
}

type PipelineWithSignalType struct {
	SignalType ConnectionType
	Pipeline   []*Component
}

func (p PipelineWithSignalType) GetID() string {
	// return the ID of the pipeline, which is a hash of the names of components in its pipeline
	// and the signal type, truncated to 4 characters.
	buf := bytes.Buffer{}
	for _, comp := range p.Pipeline {
		buf.WriteString(comp.GetSafeName())
	}
	buf.WriteString(string(p.SignalType))
	hash := metro.Hash64(buf.Bytes(), 0x234da488) // use a fixed seed for reproducibility
	shash := strconv.FormatUint(hash, 16)
	return shash[len(shash)-4:] // return the last 4 characters of the hash
}

// FindAllPipelines generates all paths from the start components to the end
// components where end components are those that are not sources of any
// connections. It returns a slice of slices of components, where each inner slice
// is a path from a start component to an end component. If there are no start
// components, it returns nil.
func (h *HPSF) FindAllPipelines(receivers map[string]bool) []PipelineWithSignalType {
	startComps := h.GetStartComponents()
	if len(startComps) == 0 {
		return nil // no start components, no paths
	}
	// copy the startComps list, but skip components whose names are not in the receivers map
	receiverComps := make([]*Component, 0)
	for _, c := range startComps {
		if _, ok := receivers[c.GetSafeName()]; ok {
			receiverComps = append(receiverComps, c)
		}
	}

	var pipelines []PipelineWithSignalType
	var path []*Component

	var findPaths func(ConnectionType, *Component)
	findPaths = func(signalType ConnectionType, c *Component) {
		path = append(path, c)
		if !h.isSourceComponent(c, signalType) {
			// we reached an end component, create a pipeline
			pipeline := PipelineWithSignalType{
				SignalType: signalType,
				Pipeline:   slices.Clone(path),
			}
			pipelines = append(pipelines, pipeline)
		} else {
			// for each of these sources, we don't want to visit the same component again,
			visited := make(map[string]bool)
			for _, conn := range h.Connections {
				if conn.Source.Component == c.Name && conn.Source.Type == signalType && !visited[conn.Destination.Component] {
					destComp := h.getComponent(conn.Destination.Component)
					visited[conn.Destination.Component] = true // mark as visited
					if destComp != nil {
						findPaths(signalType, destComp) // look deeper
					}
				}
			}
		}
		path = path[:len(path)-1] // backtrack
	}

	// start the search from each start component
	for _, c := range receiverComps {
		for _, signalType := range []ConnectionType{CTYPE_LOGS, CTYPE_METRICS, CTYPE_TRACES, CTYPE_HONEY} {
			findPaths(signalType, c)
		}
	}

	return pipelines
}

// visit all components in the HPSF in order of connections, starting from the components
// that are not destinations of any connections. This is a depth-first search
// that will visit all components that are reachable from the start components.
func (h *HPSF) VisitComponents(fn func(*Component) error) error {
	if len(h.Components) == 0 {
		// nothing to do, no components to visit
		return nil
	}

	startComps := h.GetStartComponents()
	if len(startComps) == 0 {
		return errors.New("cycle detected: component loops are not supported")
	}
	// let's sort this so that we always visit the same components in the same order
	slices.SortFunc(startComps, func(a, b *Component) int {
		return strings.Compare(a.Name, b.Name)
	})
	visited := make(map[string]bool)
	// we need the visit function to be recursive, so we define it first
	var visit func(*Component) error
	visit = func(c *Component) error {
		if c == nil {
			return nil
		}
		if visited[c.Name] {
			// already visited this component, skip it
			return nil
		}
		visited[c.Name] = true
		// call the function on the component
		err := fn(c)
		if err != nil {
			return fmt.Errorf("error visiting component %s: %w", c.Name, err)
		}
		// now visit all connections that have this component as a source
		for _, conn := range h.Connections {
			if conn.Source.Component == c.Name {
				// find the destination component
				for _, destComp := range h.Components {
					if destComp.Name == conn.Destination.Component {
						// visit the destination component
						err := visit(destComp)
						if err != nil {
							return fmt.Errorf("error visiting destination component %s from source %s: %w", destComp.Name, c.Name, err)
						}
					}
				}
			}
		}
		return nil
	}
	for _, c := range startComps {
		// visit each start component
		err := visit(c)
		if err != nil {
			return fmt.Errorf("error visiting start component %s: %w", c.Name, err)
		}
	}
	return nil
}

// use reflect to generate a list of valid yaml tags in a pointer to
// a struct
func getValidKeys(p any) []string {
	keys := []string{}
	v := reflect.ValueOf(p).Elem()
	for i := range v.NumField() {
		f := v.Type().Field(i)
		yamltag := f.Tag.Get("yaml")
		if yamltag != "" {
			// ignore any options like "omitempty"
			if strings.Contains(yamltag, ",") {
				yamltag = strings.Split(yamltag, ",")[0]
			}
			keys = append(keys, yamltag)
		}
	}
	return keys
}

// Validate checks that the HPSF is valid, returning a list of errors if it is not.
// If it detects minor issues that can be corrected, it will fix them and return.
// For example, if a property specifies that it requires an integer but the value
// is a string that can be parsed as an integer, it will parse it and store the
// result as an integer in the value.
func (h *HPSF) Validate() error {
	result := validator.NewResult("hpsf validation errors")

	// if the HPSF is empty, it's invalid
	if len(h.Components) == 0 && len(h.Containers) == 0 {
		result.Add(errors.New("empty HPSF is not valid"))
	}

	for _, c := range h.Components {
		e := c.Validate()
		result.Add(e)
	}
	for _, c := range h.Connections {
		e := c.Validate()
		result.Add(e)
	}
	for _, c := range h.Containers {
		e := c.Validate()
		result.Add(e)
	}

	// crosscheck the components and connections to make sure that all connections
	// have valid source and destination components
	components := make(map[string]bool)
	for _, c := range h.Components {
		components[c.Name] = true
	}
	for _, c := range h.Connections {
		if _, ok := components[c.Source.Component]; !ok {
			result.Add(NewError("Connection source component not found").WithComponent(c.Source.Component))
		}
		if _, ok := components[c.Destination.Component]; !ok {
			result.Add(NewError("Connection destination component not found").WithComponent(c.Destination.Component))
		}
	}

	return result.ErrOrNil()
}

// EnsureHPSFYAML returns an error if the input is not HPSF yaml or invalid HPSF
func EnsureHPSFYAML(input string) error {
	m, err := validator.EnsureYAML([]byte(input))
	if err != nil {
		return err
	}
	// it has to have at least one key
	if len(m) == 0 {
		return errors.New("HPSF yaml is empty")
	}

	// check to see if it has only expected top-level keys
	// (it would be interesting to do this recursively someday, but it's a lot)
	keys := getValidKeys(&HPSF{})
	badkeys := make([]string, 0)
	for k := range m {
		if !slices.Contains(keys, k) {
			badkeys = append(badkeys, k)
		}
	}
	if len(badkeys) > 0 {
		return errors.New("HPSF yaml contains unexpected keys: " + strings.Join(badkeys, ", "))
	}

	var hpsf HPSF
	dec := y.NewDecoder(strings.NewReader(input))
	err = dec.Decode(&hpsf)
	if err != nil {
		return err
	}
	return hpsf.Validate()
}

func (h *HPSF) AsYAML() (string, error) {
	// this is a mechanism to marshal the template to YAML
	data, err := y.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("error marshalling hpsf to YAML: %w", err)
	}
	return string(data), nil
}
