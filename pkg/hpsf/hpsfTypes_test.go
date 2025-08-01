package hpsf

import (
	"fmt"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
)

func TestEnsureHPSF(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"0", []string{}, "empty"},
		{"1", []string{"a"}, "unexpected keys"},
		{"2", []string{"connections"}, "unmarshal errors"},
		{"3", []string{"components", "connections"}, "unmarshal errors"},
		{"4", []string{"components", "connections", "something"}, "unexpected keys"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := tmpl.DottedConfig{}
			for i, arg := range tt.args {
				y[arg] = i
			}
			text, _ := y.RenderYAML()
			if err := EnsureHPSFYAML(string(text)); (err != nil) && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("EnsureHPSF() error = %v, should contain '%v'", err, tt.wantErr)
			}
		})
	}
}

func TestHPSF_Validate(t *testing.T) {
	inputData := []byte(`components:
  - name: otlp_in
    kind: OTelReceiver
    properties:
      - name: GRPCPort
        value: 9922
      - name: HTTPPort
        value: 1234
  - name: otlp_out
    kind: OTelGRPCExporter
    properties:
      - name: Host
        value: https://myhost.com
      - name: Port
        value: 1234
      - name: Headers
        value:
          "header1": "1234"
connections:
  - source:
      component: otlp_in
      port: Traces
      type: OTelTraces
    destination:
      component: otlp_out
      port: Traces
      type: OTelTraces`)

	_, err := validator.EnsureYAML(inputData)
	require.NoError(t, err)

	var h HPSF
	err = yaml.Unmarshal(inputData, &h)
	require.NoError(t, err)

	errors := h.Validate()
	require.NoError(t, errors)
}

func TestHPSF_ValidateFailures(t *testing.T) {
	// GRPCPort is missing a value / value type is wrong
	// connection names a non-existent component
	inputData := []byte(`components:
  - name: otlp_in
    kind: OTelReceiver
    properties:
      - name: GRPCPort
      - name: HTTPPort
        value: 1234
  - name: otlp_out
    kind: OTelGRPCExporter
    properties:
      - name: Host
        value: https://myhost.com
      - name: Port
        value: 1234
      - name: Headers
        value:
          "header1": "1234"
connections:
  - source:
      component: otlp_in2
      port: Traces
      type: OTelTraces
    destination:
      component: otlp_out
      port: Traces
      type: OTelTraces`)

	_, err := validator.EnsureYAML(inputData)
	require.NoError(t, err)

	var h HPSF
	err = yaml.Unmarshal(inputData, &h)
	require.NoError(t, err)

	err = h.Validate()
	result, ok := err.(validator.Result)
	assert.True(t, ok)
	assert.Equal(t, 2, result.Len())
	assert.Contains(t, result.Details[0].Error(), "GRPCPort")
	assert.Contains(t, result.Details[1].Error(), "otlp_in2")
}

func TestComponent_GetSafeName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"a", "a"},
		{"a b", "a_b"},
		{"a b c", "a_b_c"},
		{"a#@#$%^&*()b", "a_b"},
		{"Deterministic Sampler", "Deterministic_Sampler"},
		{"Deterministic_Sampler", "Deterministic_Sampler"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Component{Name: tt.name}
			if got := c.GetSafeName(); got != tt.want {
				t.Errorf("Component.GetSafeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropType_ValueCoerce(t *testing.T) {
	var result any
	tests := []struct {
		p       PropType
		v       any
		target  any
		wantErr bool
	}{
		{PTYPE_STRING, "a", "a", false},
		{PTYPE_STRING, 1, "1", false},
		{PTYPE_STRING, 1.5, "1.5", false},
		{PTYPE_STRING, true, "true", false},
		{PTYPE_STRING, []string{"a"}, nil, true},
		{PTYPE_STRING, map[string]any{"a": 3}, nil, true},
		{PTYPE_INT, "a", nil, true},
		{PTYPE_INT, "17", 17, false},
		{PTYPE_INT, "17.5", nil, true},
		{PTYPE_INT, 1, 1, false},
		{PTYPE_INT, 1.0, 1, false},
		{PTYPE_INT, 1.5, nil, true},
		{PTYPE_INT, true, nil, true},
		{PTYPE_INT, []string{"a"}, nil, true},
		{PTYPE_INT, map[string]any{"a": 3}, nil, true},
		{PTYPE_FLOAT, "a", nil, true},
		{PTYPE_FLOAT, "17", 17.0, false},
		{PTYPE_FLOAT, "17.5", 17.5, false},
		{PTYPE_FLOAT, 1, 1.0, false},
		{PTYPE_FLOAT, 1.0, 1.0, false},
		{PTYPE_FLOAT, 1.5, 1.5, false},
		{PTYPE_FLOAT, true, nil, true},
		{PTYPE_FLOAT, []string{"a"}, nil, true},
		{PTYPE_FLOAT, map[string]any{"a": 3}, nil, true},
		{PTYPE_BOOL, "a", nil, true},
		{PTYPE_BOOL, "true", true, false},
		{PTYPE_BOOL, "True", true, false},
		{PTYPE_BOOL, "TRUE", true, false},
		{PTYPE_BOOL, "T", true, false},
		{PTYPE_BOOL, "t", true, false},
		{PTYPE_BOOL, "YES", true, false},
		{PTYPE_BOOL, "yes", true, false},
		{PTYPE_BOOL, "Yes", true, false},
		{PTYPE_BOOL, "Y", true, false},
		{PTYPE_BOOL, "y", true, false},
		{PTYPE_BOOL, "1", nil, true},
		{PTYPE_BOOL, "0", nil, true},
		{PTYPE_BOOL, 1, true, false},
		{PTYPE_BOOL, 0, false, false},
		{PTYPE_BOOL, 1.0, true, false},
		{PTYPE_BOOL, 0.0, false, false},
		{PTYPE_BOOL, 1.5, true, false},
		{PTYPE_BOOL, true, true, false},
		{PTYPE_BOOL, false, false, false},
		{PTYPE_BOOL, []string{"true"}, nil, true},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%s_%#v", tt.p, tt.v)
		t.Run(name, func(t *testing.T) {
			result = nil
			if err := tt.p.ValueCoerce(tt.v, &result); (err != nil) != tt.wantErr {
				t.Errorf("PropType.ValueCoerce() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.target, result)
		})
	}
}

func TestHPSF_VisitComponents(t *testing.T) {
	tests := []struct {
		name          string
		components    []*Component
		connections   []*Connection
		expectedOrder []string
		errorOn       string // force error when visiting this component (empty string means no error)
		expectedError string // expect an error message containing this string
	}{
		{
			name: "no connections",
			components: []*Component{
				{Name: "A", Kind: "test"},
				{Name: "B", Kind: "test"},
				{Name: "C", Kind: "test"},
			},
			connections:   []*Connection{},
			expectedOrder: []string{"A", "B", "C"}, // Alphabetical because of sorting
		},
		{
			name: "linear path",
			components: []*Component{
				{Name: "A", Kind: "test"},
				{Name: "B", Kind: "test"},
				{Name: "C", Kind: "test"},
			},
			connections: []*Connection{
				{
					Source:      ConnectionPort{Component: "A", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "B", PortName: "in", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "B", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "C", PortName: "in", Type: "test"},
				},
			},
			expectedOrder: []string{"A", "B", "C"},
		},
		{
			name: "fork",
			components: []*Component{
				{Name: "A", Kind: "test"},
				{Name: "B", Kind: "test"},
				{Name: "C", Kind: "test"},
			},
			connections: []*Connection{
				{
					Source:      ConnectionPort{Component: "A", PortName: "out1", Type: "test"},
					Destination: ConnectionPort{Component: "B", PortName: "in", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "A", PortName: "out2", Type: "test"},
					Destination: ConnectionPort{Component: "C", PortName: "in", Type: "test"},
				},
			},
			expectedOrder: []string{"A", "B", "C"},
		},
		{
			name: "join",
			components: []*Component{
				{Name: "A", Kind: "test"},
				{Name: "B", Kind: "test"},
				{Name: "C", Kind: "test"},
			},
			connections: []*Connection{
				{
					Source:      ConnectionPort{Component: "A", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "C", PortName: "in1", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "B", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "C", PortName: "in2", Type: "test"},
				},
			},
			expectedOrder: []string{"A", "C", "B"}, // A and B are start nodes (alphabetical), and C is visited from A
		},
		{
			name: "cycle",
			components: []*Component{
				{Name: "A", Kind: "test"},
				{Name: "B", Kind: "test"},
				{Name: "C", Kind: "test"},
			},
			connections: []*Connection{
				{
					Source:      ConnectionPort{Component: "A", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "B", PortName: "in", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "B", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "C", PortName: "in", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "C", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "A", PortName: "in", Type: "test"},
				},
			},
			expectedOrder: nil,
			expectedError: "cycle detected",
		},
		{
			name: "error during visit",
			components: []*Component{
				{Name: "A", Kind: "test"},
				{Name: "B", Kind: "test"},
				{Name: "C", Kind: "test"},
			},
			connections: []*Connection{
				{
					Source:      ConnectionPort{Component: "A", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "B", PortName: "in", Type: "test"},
				},
			},
			expectedOrder: []string{"A"},
			errorOn:       "B",
		},
		{
			name:          "empty HPSF",
			components:    []*Component{},
			connections:   []*Connection{},
			expectedOrder: nil,
		},
		{
			name: "complex graph",
			components: []*Component{
				{Name: "A", Kind: "test"},
				{Name: "B", Kind: "test"},
				{Name: "C", Kind: "test"},
				{Name: "D", Kind: "test"},
				{Name: "E", Kind: "test"},
			},
			connections: []*Connection{
				{
					Source:      ConnectionPort{Component: "A", PortName: "out1", Type: "test"},
					Destination: ConnectionPort{Component: "B", PortName: "in", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "A", PortName: "out2", Type: "test"},
					Destination: ConnectionPort{Component: "C", PortName: "in", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "B", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "D", PortName: "in1", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "C", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "D", PortName: "in2", Type: "test"},
				},
				{
					Source:      ConnectionPort{Component: "D", PortName: "out", Type: "test"},
					Destination: ConnectionPort{Component: "E", PortName: "in", Type: "test"},
				},
			},
			expectedOrder: []string{"A", "B", "D", "E", "C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HPSF{
				Components:  tt.components,
				Connections: tt.connections,
			}

			var visited []string
			err := h.VisitComponents(func(c *Component) error {
				visited = append(visited, c.Name)
				if tt.errorOn == c.Name {
					return fmt.Errorf("test error expected %s, got %s", tt.errorOn, c.Name)
				}
				return nil
			})

			if tt.errorOn != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "test error")
			} else {
				if tt.expectedError != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectedOrder, visited)
				}
			}
		})
	}
}

func TestHPSF_FindAllPaths(t *testing.T) {
	// Create a simple HPSF with 3 components connected in a line: A -> B -> C
	hpsf := &HPSF{
		Kind:    "test",
		Version: "1.0",
		Name:    "test-pipeline",
		Components: []*Component{
			{Name: "component_a", Kind: "receiver"},
			{Name: "component_b", Kind: "processor"},
			{Name: "component_c", Kind: "exporter"},
		},
		Connections: []*Connection{
			{
				Source: ConnectionPort{
					Component: "component_a",
					PortName:  "out",
					Type:      CTYPE_TRACES,
				},
				Destination: ConnectionPort{
					Component: "component_b",
					PortName:  "in",
					Type:      CTYPE_TRACES,
				},
			},
			{
				Source: ConnectionPort{
					Component: "component_b",
					PortName:  "out",
					Type:      CTYPE_TRACES,
				},
				Destination: ConnectionPort{
					Component: "component_c",
					PortName:  "in",
					Type:      CTYPE_TRACES,
				},
			},
		},
	}

	// Find all paths
	paths := hpsf.FindAllPaths(nil)

	// Should find exactly one path for CTYPE_TRACES
	assert.Len(t, paths, 1, "Should find exactly one path")

	path := paths[0]
	assert.Equal(t, CTYPE_TRACES, path.ConnType, "Path should be for traces")

	// Should have 3 components in the path
	assert.Len(t, path.Path, 3, "Path should have 3 components")
	assert.Equal(t, "component_a", path.Path[0].Name, "First component should be component_a")
	assert.Equal(t, "component_b", path.Path[1].Name, "Second component should be component_b")
	assert.Equal(t, "component_c", path.Path[2].Name, "Third component should be component_c")

	// Should have exactly 2 connections (A->B and B->C)
	assert.Len(t, path.Connections, 2, "Pipeline should have exactly 2 connections")

	// Verify the first connection (A->B)
	assert.Equal(t, "component_a", path.Connections[0].Source.Component, "First connection source should be component_a")
	assert.Equal(t, "component_b", path.Connections[0].Destination.Component, "First connection destination should be component_b")

	// Verify the second connection (B->C)
	assert.Equal(t, "component_b", path.Connections[1].Source.Component, "Second connection source should be component_b")
	assert.Equal(t, "component_c", path.Connections[1].Destination.Component, "Second connection destination should be component_c")
}

func TestHPSF_FindAllPipelines_MultiplePaths(t *testing.T) {
	// Create an HPSF with multiple paths: A -> B -> C and A -> D -> C
	hpsf := &HPSF{
		Kind:    "test",
		Version: "1.0",
		Name:    "test-multiple-pipelines",
		Components: []*Component{
			{Name: "component_a", Kind: "receiver"},
			{Name: "component_b", Kind: "processor"},
			{Name: "component_c", Kind: "exporter"},
			{Name: "component_d", Kind: "processor"},
		},
		Connections: []*Connection{
			{
				Source: ConnectionPort{
					Component: "component_a",
					PortName:  "out",
					Type:      CTYPE_TRACES,
				},
				Destination: ConnectionPort{
					Component: "component_b",
					PortName:  "in",
					Type:      CTYPE_TRACES,
				},
			},
			{
				Source: ConnectionPort{
					Component: "component_b",
					PortName:  "out",
					Type:      CTYPE_TRACES,
				},
				Destination: ConnectionPort{
					Component: "component_c",
					PortName:  "in",
					Type:      CTYPE_TRACES,
				},
			},
			{
				Source: ConnectionPort{
					Component: "component_a",
					PortName:  "out2",
					Type:      CTYPE_TRACES,
				},
				Destination: ConnectionPort{
					Component: "component_d",
					PortName:  "in",
					Type:      CTYPE_TRACES,
				},
			},
			{
				Source: ConnectionPort{
					Component: "component_d",
					PortName:  "out",
					Type:      CTYPE_TRACES,
				},
				Destination: ConnectionPort{
					Component: "component_c",
					PortName:  "in2",
					Type:      CTYPE_TRACES,
				},
			},
		},
	}

	// Find all paths
	paths := hpsf.FindAllPaths(nil)

	// Should find exactly two paths for CTYPE_TRACES
	assert.Len(t, paths, 2, "Should find exactly two paths")

	// Both paths should have the same connection type
	for _, path := range paths {
		assert.Equal(t, CTYPE_TRACES, path.ConnType, "All paths should be for traces")
		assert.Len(t, path.Path, 3, "Each path should have 3 components")
		assert.Equal(t, "component_a", path.Path[0].Name, "First component should be component_a")
		assert.Equal(t, "component_c", path.Path[2].Name, "Last component should be component_c")
		assert.Len(t, path.Connections, 2, "Each path should have exactly 2 connections")
	}

	// Verify that we have both paths: A->B->C and A->D->C
	foundPath1 := false
	foundPath2 := false

	for _, path := range paths {
		switch path.Path[1].Name {
		case "component_b":
			foundPath1 = true
			assert.Equal(t, "component_b", path.Connections[0].Destination.Component, "Path 1 should go through component_b")
		case "component_d":
			foundPath2 = true
			assert.Equal(t, "component_d", path.Connections[0].Destination.Component, "Path 2 should go through component_d")
		}
	}

	assert.True(t, foundPath1, "Should find path A->B->C")
	assert.True(t, foundPath2, "Should find path A->D->C")
}
