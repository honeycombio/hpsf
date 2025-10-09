package hpsf

import (
	"github.com/honeycombio/hpsf/pkg/layout"
)

// NodeSize contains default width and height for nodes in the layout
type NodeSize struct {
	Width  int
	Height int
}

// DefaultNodeSize returns the default node size for layout
func DefaultNodeSize() NodeSize {
	return NodeSize{Width: 100, Height: 50}
}

// AutoLayout computes an automatic layout for the HPSF components and stores
// the positions in the Layout field. It uses the layout package to compute
// optimal positions for components based on their connections.
//
// Options can be passed to customize the layout behavior (e.g., WithHSeparation,
// WithVSeparation, WithSnapGridSize, etc.).
//
// The resulting layout is stored in h.Layout as a map with component names as keys,
// and each value is a map containing "x" and "y" integer coordinates, and optionally
// "w" and "h" size values. If sizes are present in the existing layout, they will be
// preserved.
func (h *HPSF) AutoLayout(nodeSize NodeSize, opts ...layout.LayoutOption) error {
	// Build the layout graph from HPSF components and connections
	g := layout.Graph{}

	// Create nodes for each component
	nodeMap := make(map[string]*layout.Node)
	for _, comp := range h.Components {
		// Check if there's an existing size in the layout
		width, height := nodeSize.Width, nodeSize.Height
		if existingW, existingH, ok := h.getComponentSize(comp.Name); ok {
			width, height = existingW, existingH
		}

		node := &layout.Node{
			Id: comp.Name,
			Rect: layout.Rect{
				Position: layout.Position{X: 0, Y: 0},
				Size:     layout.Size{W: width, H: height},
			},
		}

		// Create input ports based on component ports
		inputIdx := 1
		outputIdx := 1
		for _, port := range comp.Ports {
			if port.Direction == DIR_INPUT {
				p := &layout.Port{Node: node, Index: inputIdx}
				node.Inputs = append(node.Inputs, p)
				inputIdx++
			} else if port.Direction == DIR_OUTPUT {
				p := &layout.Port{Node: node, Index: outputIdx}
				node.Outputs = append(node.Outputs, p)
				outputIdx++
			}
		}

		// If no ports are defined, create default ports
		// (most components will have ports defined by template expansion)
		if len(node.Inputs) == 0 {
			node.Inputs = append(node.Inputs, &layout.Port{Node: node, Index: 1})
		}
		if len(node.Outputs) == 0 {
			node.Outputs = append(node.Outputs, &layout.Port{Node: node, Index: 1})
		}

		g.AddNode(node)
		nodeMap[comp.Name] = node
	}

	// Create edges for each connection
	// Group connections by source component to assign port indices
	connectionsBySource := make(map[string][]*Connection)
	for _, conn := range h.Connections {
		connectionsBySource[conn.Source.Component] = append(connectionsBySource[conn.Source.Component], conn)
	}

	// For each source component, create edges with sequential port indices
	for _, conn := range h.Connections {
		sourceNode := nodeMap[conn.Source.Component]
		destNode := nodeMap[conn.Destination.Component]

		if sourceNode == nil || destNode == nil {
			// Skip connections to/from non-existent components
			continue
		}

		// Find or create the appropriate ports
		// Use port index 1 as default (most connections will use the default port)
		sourcePortIdx := 1
		destPortIdx := 1

		// Find the output port for the source
		var sourcePort *layout.Port
		if len(sourceNode.Outputs) > 0 {
			sourcePort = sourceNode.Outputs[sourcePortIdx-1]
		}

		// Find the input port for the destination
		var destPort *layout.Port
		if len(destNode.Inputs) > 0 {
			destPort = destNode.Inputs[destPortIdx-1]
		}

		if sourcePort != nil && destPort != nil {
			edge := &layout.Edge{
				From: sourcePort,
				To:   destPort,
			}
			g.AddEdge(edge)
		}
	}

	// Run the auto-layout algorithm
	if err := g.AutoLayout(opts...); err != nil {
		return err
	}

	// Extract positions and store them in the Layout field
	// Format matches layoutTester.yaml structure
	if h.Layout == nil {
		h.Layout = make(Layout)
	}

	components := make([]any, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		compLayout := Layout{
			"name": node.Id,
			"position": Layout{
				"x": node.Rect.Position.X,
				"y": node.Rect.Position.Y,
			},
		}
		// Preserve size if it was set (non-default)
		if node.Rect.Size.W != 0 || node.Rect.Size.H != 0 {
			compLayout["size"] = Layout{
				"w": node.Rect.Size.W,
				"h": node.Rect.Size.H,
			}
		}
		components = append(components, compLayout)
	}
	h.Layout["components"] = components

	return nil
}

// GetComponentPosition retrieves the layout position for a component.
// Returns (0, 0, false) if the component has no layout information.
func (h *HPSF) GetComponentPosition(componentName string) (x, y int, ok bool) {
	if h.Layout == nil {
		return 0, 0, false
	}

	componentsVal, exists := h.Layout["components"]
	if !exists {
		return 0, 0, false
	}

	// Handle []interface{} (from YAML unmarshaling or our code)
	components, ok := componentsVal.([]any)
	if !ok {
		return 0, 0, false
	}

	for _, compVal := range components {
		comp, ok := compVal.(Layout)
		if !ok {
			continue
		}

		name, nameOk := comp["name"].(string)
		if !nameOk || name != componentName {
			continue
		}

		return extractPosition(comp)
	}

	return 0, 0, false
}

// getComponentSize retrieves the layout size for a component.
// Returns (0, 0, false) if the component has no size information.
func (h *HPSF) getComponentSize(componentName string) (width, height int, ok bool) {
	if h.Layout == nil {
		return 0, 0, false
	}

	componentsVal, exists := h.Layout["components"]
	if !exists {
		return 0, 0, false
	}

	// Handle []interface{} (from YAML unmarshaling or our code)
	components, ok := componentsVal.([]any)
	if !ok {
		return 0, 0, false
	}

	for _, compVal := range components {
		comp, ok := compVal.(Layout)
		if !ok {
			continue
		}

		name, nameOk := comp["name"].(string)
		if !nameOk || name != componentName {
			continue
		}

		return extractSize(comp)
	}

	return 0, 0, false
}

func extractPosition(comp Layout) (x, y int, ok bool) {
	posVal, posOk := comp["position"]
	if !posOk {
		return 0, 0, false
	}

	posMap, ok := posVal.(Layout)
	if !ok {
		return 0, 0, false
	}

	xVal, xOk := posMap["x"]
	yVal, yOk := posMap["y"]
	if !xOk || !yOk {
		return 0, 0, false
	}

	// Handle both int and float64 (from YAML unmarshaling)
	var xInt, yInt int
	switch v := xVal.(type) {
	case int:
		xInt = v
	case float64:
		xInt = int(v)
	default:
		return 0, 0, false
	}

	switch v := yVal.(type) {
	case int:
		yInt = v
	case float64:
		yInt = int(v)
	default:
		return 0, 0, false
	}

	return xInt, yInt, true
}

func extractSize(comp Layout) (w, h int, ok bool) {
	sizeVal, sizeOk := comp["size"]
	if !sizeOk {
		return 0, 0, false
	}

	sizeMap, ok := sizeVal.(Layout)
	if !ok {
		return 0, 0, false
	}

	wVal, wOk := sizeMap["w"]
	hVal, hOk := sizeMap["h"]
	if !wOk || !hOk {
		return 0, 0, false
	}

	// Handle both int and float64 (from YAML unmarshaling)
	var wInt, hInt int
	switch v := wVal.(type) {
	case int:
		wInt = v
	case float64:
		wInt = int(v)
	default:
		return 0, 0, false
	}

	switch v := hVal.(type) {
	case int:
		hInt = v
	case float64:
		hInt = int(v)
	default:
		return 0, 0, false
	}

	return wInt, hInt, true
}
