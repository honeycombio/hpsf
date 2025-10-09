package hpsf

import (
	"os"
	"testing"

	"github.com/honeycombio/hpsf/pkg/layout"
)

func TestHPSF_AutoLayout(t *testing.T) {
	// Create a simple HPSF with three components connected in a chain
	h := &HPSF{
		Components: []*Component{
			{
				Name: "receiver",
				Kind: "OTelReceiver",
				Ports: []Port{
					{Name: "traces_out", Direction: DIR_OUTPUT, Type: CTYPE_TRACES},
				},
			},
			{
				Name: "processor",
				Kind: "Processor",
				Ports: []Port{
					{Name: "traces_in", Direction: DIR_INPUT, Type: CTYPE_TRACES},
					{Name: "traces_out", Direction: DIR_OUTPUT, Type: CTYPE_TRACES},
				},
			},
			{
				Name: "exporter",
				Kind: "OTelExporter",
				Ports: []Port{
					{Name: "traces_in", Direction: DIR_INPUT, Type: CTYPE_TRACES},
				},
			},
		},
		Connections: []*Connection{
			{
				Source:      ConnectionPort{Component: "receiver", PortName: "traces_out", Type: CTYPE_TRACES},
				Destination: ConnectionPort{Component: "processor", PortName: "traces_in", Type: CTYPE_TRACES},
			},
			{
				Source:      ConnectionPort{Component: "processor", PortName: "traces_out", Type: CTYPE_TRACES},
				Destination: ConnectionPort{Component: "exporter", PortName: "traces_in", Type: CTYPE_TRACES},
			},
		},
	}

	// Run AutoLayout
	err := h.AutoLayout(DefaultNodeSize())
	if err != nil {
		t.Fatalf("AutoLayout failed: %v", err)
	}

	// Verify layout was created
	if h.Layout == nil {
		t.Fatal("Layout is nil after AutoLayout")
	}

	// Verify all components have positions
	for _, comp := range h.Components {
		x, y, ok := h.GetComponentPosition(comp.Name)
		if !ok {
			t.Fatalf("Component %s has no position", comp.Name)
		}
		t.Logf("Component %s positioned at (%d, %d)", comp.Name, x, y)
	}

	// Verify components are laid out left to right
	recX, _, _ := h.GetComponentPosition("receiver")
	procX, _, _ := h.GetComponentPosition("processor")
	expX, _, _ := h.GetComponentPosition("exporter")

	if !(recX < procX && procX < expX) {
		t.Errorf("Expected left-to-right layout: receiver(%d) < processor(%d) < exporter(%d)", recX, procX, expX)
	}
}

func TestHPSF_AutoLayout_WithOptions(t *testing.T) {
	h := &HPSF{
		Components: []*Component{
			{Name: "A", Kind: "Source"},
			{Name: "B", Kind: "Sink"},
		},
		Connections: []*Connection{
			{
				Source:      ConnectionPort{Component: "A", PortName: "out", Type: CTYPE_TRACES},
				Destination: ConnectionPort{Component: "B", PortName: "in", Type: CTYPE_TRACES},
			},
		},
	}

	// Run AutoLayout with custom spacing
	err := h.AutoLayout(
		DefaultNodeSize(),
		layout.WithHSeparation(100),
		layout.WithVSeparation(50),
	)
	if err != nil {
		t.Fatalf("AutoLayout failed: %v", err)
	}

	// Verify spacing
	aX, _, _ := h.GetComponentPosition("A")
	bX, _, _ := h.GetComponentPosition("B")

	spacing := bX - aX
	// Spacing should be at least the HSeparation (100) plus node width
	expectedMin := 100 + DefaultNodeSize().Width
	if spacing < expectedMin {
		t.Errorf("Expected spacing >= %d, got %d", expectedMin, spacing)
	}
}

func TestHPSF_AutoLayout_Diamond(t *testing.T) {
	// Create a diamond-shaped graph: A -> B, A -> C, B -> D, C -> D
	h := &HPSF{
		Components: []*Component{
			{Name: "A", Kind: "Source"},
			{Name: "B", Kind: "Process1"},
			{Name: "C", Kind: "Process2"},
			{Name: "D", Kind: "Sink"},
		},
		Connections: []*Connection{
			{
				Source:      ConnectionPort{Component: "A", PortName: "out", Type: CTYPE_TRACES},
				Destination: ConnectionPort{Component: "B", PortName: "in", Type: CTYPE_TRACES},
			},
			{
				Source:      ConnectionPort{Component: "A", PortName: "out", Type: CTYPE_TRACES},
				Destination: ConnectionPort{Component: "C", PortName: "in", Type: CTYPE_TRACES},
			},
			{
				Source:      ConnectionPort{Component: "B", PortName: "out", Type: CTYPE_TRACES},
				Destination: ConnectionPort{Component: "D", PortName: "in", Type: CTYPE_TRACES},
			},
			{
				Source:      ConnectionPort{Component: "C", PortName: "out", Type: CTYPE_TRACES},
				Destination: ConnectionPort{Component: "D", PortName: "in", Type: CTYPE_TRACES},
			},
		},
	}

	err := h.AutoLayout(DefaultNodeSize())
	if err != nil {
		t.Fatalf("AutoLayout failed: %v", err)
	}

	// Verify diamond layout structure
	aX, _, _ := h.GetComponentPosition("A")
	bX, _, _ := h.GetComponentPosition("B")
	cX, _, _ := h.GetComponentPosition("C")
	dX, _, _ := h.GetComponentPosition("D")

	// A should be leftmost
	if aX >= bX || aX >= cX {
		t.Error("A should be leftmost in diamond layout")
	}

	// B and C should be in the middle (same column)
	if bX != cX {
		t.Error("B and C should be in the same column")
	}

	// D should be rightmost
	if dX <= bX || dX <= cX {
		t.Error("D should be rightmost in diamond layout")
	}

	t.Logf("Diamond layout: A(%d) -> B(%d), C(%d) -> D(%d)", aX, bX, cX, dX)
}

func TestHPSF_AutoLayout_EmptyGraph(t *testing.T) {
	h := &HPSF{
		Components:  []*Component{},
		Connections: []*Connection{},
	}

	err := h.AutoLayout(DefaultNodeSize())
	if err != nil {
		t.Fatalf("AutoLayout should handle empty graph: %v", err)
	}
}

func TestHPSF_GetComponentPosition_NoLayout(t *testing.T) {
	h := &HPSF{
		Components: []*Component{
			{Name: "test", Kind: "Test"},
		},
	}

	x, y, ok := h.GetComponentPosition("test")
	if ok {
		t.Errorf("Expected no position, got (%d, %d)", x, y)
	}
}

func TestHPSF_ReadLayoutFromYAML(t *testing.T) {
	// Read the layoutTester.yaml file
	data, err := os.ReadFile("../../examples/layoutTester.yaml")
	if err != nil {
		t.Skipf("Skipping test, layoutTester.yaml not found: %v", err)
		return
	}

	h, err := FromYAML(string(data))
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// Verify we can read positions from the layout
	testCases := []struct {
		name string
		x    int
		y    int
	}{
		{"Check for Errors_1", 260, 160},
		{"Check Duration_1", 240, 320},
		{"Start Sampling_1", 0, 120},
		{"Keep All_1", 520, 0},
		{"Sample at a Fixed Rate_1", 500, 160},
		{"Keep All_2", 520, 320},
		{"Send to Honeycomb_1", 780, 120},
	}

	for _, tc := range testCases {
		x, y, ok := h.GetComponentPosition(tc.name)
		if !ok {
			t.Errorf("Component %s has no position", tc.name)
			continue
		}
		if x != tc.x || y != tc.y {
			t.Errorf("Component %s: expected (%d, %d), got (%d, %d)", tc.name, tc.x, tc.y, x, y)
		}
	}
}
