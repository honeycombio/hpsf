package layout

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// Layouter handles automatic layout of HPSF configurations.
// It loads component templates to understand port structures for accurate edge routing.
type Layouter struct {
	templates map[string]config.TemplateComponent // kind -> template
}

// NewLayouter creates a new Layouter with embedded component templates loaded.
func NewLayouter() (*Layouter, error) {
	templates, err := data.LoadEmbeddedComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded components: %w", err)
	}

	// templates is already a map[string]config.TemplateComponent keyed by kind

	return &Layouter{
		templates: templates,
	}, nil
}

// buildLayoutGraph constructs a Graph from the HPSF components and connections.
// If the HPSF has existing layout information, node sizes are preserved.
// Returns an error if a component template is not found.
func (l *Layouter) buildLayoutGraph(h *hpsf.HPSF, nodeSize hpsf.NodeSize) (*Graph, error) {
	g := &Graph{}

	// Create nodes for each component
	nodeMap := make(map[string]*Node)
	// Also create port name -> index maps for each component
	type portMaps struct {
		inputPorts  map[string]int // port name -> port index (1-based)
		outputPorts map[string]int
	}
	componentPortMaps := make(map[string]*portMaps)

	for _, comp := range h.Components {
		// Look up the template for this component (keyed by Kind only)
		template, ok := l.templates[comp.Kind]
		if !ok {
			return nil, fmt.Errorf("component template not found: %s (kind=%s, version=%s)", comp.Name, comp.Kind, comp.Version)
		}

		// Check if there's an existing size in the layout
		width, height := nodeSize.Width, nodeSize.Height
		if existingW, existingH, ok := h.GetComponentSize(comp.Name); ok {
			width, height = existingW, existingH
		}

		node := &Node{
			Id: comp.Name,
			Rect: Rect{
				Position: Position{X: 0, Y: 0},
				Size:     Size{W: width, H: height},
			},
		}

		// Create ports based on template ports and build port name -> index maps
		inputIdx := 1
		outputIdx := 1
		portMaps := &portMaps{
			inputPorts:  make(map[string]int),
			outputPorts: make(map[string]int),
		}

		for _, port := range template.Ports {
			if port.Direction == "input" {
				p := &Port{Node: node, Index: inputIdx}
				node.Inputs = append(node.Inputs, p)
				portMaps.inputPorts[port.Name] = inputIdx
				inputIdx++
			} else if port.Direction == "output" {
				p := &Port{Node: node, Index: outputIdx}
				node.Outputs = append(node.Outputs, p)
				portMaps.outputPorts[port.Name] = outputIdx
				outputIdx++
			}
		}

		// If no ports are defined in template, create default ports
		if len(node.Inputs) == 0 {
			node.Inputs = append(node.Inputs, &Port{Node: node, Index: 1})
			portMaps.inputPorts["default"] = 1
		}
		if len(node.Outputs) == 0 {
			node.Outputs = append(node.Outputs, &Port{Node: node, Index: 1})
			portMaps.outputPorts["default"] = 1
		}

		g.AddNode(node)
		nodeMap[comp.Name] = node
		componentPortMaps[comp.Name] = portMaps
	}

	// Create edges for each connection using the port name -> index maps
	for _, conn := range h.Connections {
		sourceNode := nodeMap[conn.Source.Component]
		destNode := nodeMap[conn.Destination.Component]

		if sourceNode == nil || destNode == nil {
			return nil, fmt.Errorf("connection references unknown component: %s -> %s",
				conn.Source.Component, conn.Destination.Component)
		}

		sourcePorts := componentPortMaps[conn.Source.Component]
		destPorts := componentPortMaps[conn.Destination.Component]

		// Look up source port by name
		sourcePortIdx, ok := sourcePorts.outputPorts[conn.Source.PortName]
		if !ok {
			return nil, fmt.Errorf("output port not found: %s.%s", conn.Source.Component, conn.Source.PortName)
		}

		// Look up destination port by name
		destPortIdx, ok := destPorts.inputPorts[conn.Destination.PortName]
		if !ok {
			return nil, fmt.Errorf("input port not found: %s.%s", conn.Destination.Component, conn.Destination.PortName)
		}

		// Get the actual port objects (indices are 1-based, arrays are 0-based)
		if sourcePortIdx < 1 || sourcePortIdx > len(sourceNode.Outputs) {
			return nil, fmt.Errorf("source port index out of range: %s.%s (index=%d, max=%d)",
				conn.Source.Component, conn.Source.PortName, sourcePortIdx, len(sourceNode.Outputs))
		}
		if destPortIdx < 1 || destPortIdx > len(destNode.Inputs) {
			return nil, fmt.Errorf("dest port index out of range: %s.%s (index=%d, max=%d)",
				conn.Destination.Component, conn.Destination.PortName, destPortIdx, len(destNode.Inputs))
		}

		sourcePort := sourceNode.Outputs[sourcePortIdx-1]
		destPort := destNode.Inputs[destPortIdx-1]

		edge := &Edge{
			From: sourcePort,
			To:   destPort,
		}
		g.AddEdge(edge)
	}

	return g, nil
}

// AutoLayout computes an automatic layout for the HPSF components and stores
// the positions in h.Layout. It uses the layout package to compute optimal
// positions for components based on their connections.
//
// Options can be passed to customize the layout behavior (e.g., WithHSeparation,
// WithVSeparation, WithSnapGridSize, etc.).
func (l *Layouter) AutoLayout(h *hpsf.HPSF, nodeSize hpsf.NodeSize, opts ...LayoutOption) error {
	g, err := l.buildLayoutGraph(h, nodeSize)
	if err != nil {
		return err
	}

	// Run the auto-layout algorithm
	if err := g.AutoLayout(opts...); err != nil {
		return err
	}

	// Apply the layout back to HPSF
	l.applyLayoutFromGraph(h, g)

	return nil
}

// CountCrossings counts the number of edge crossings in the current layout
// using the graph-based intersection detection. Returns 0 if there is no layout
// or if an error occurs building the graph.
func (l *Layouter) CountCrossings(h *hpsf.HPSF) int {
	if h.Layout == nil {
		return 0
	}

	// Build the graph with current layout positions
	g, err := l.buildLayoutGraph(h, hpsf.DefaultNodeSize())
	if err != nil {
		// Return 0 on error (could also log the error)
		return 0
	}

	// Apply current positions to the graph nodes
	for _, node := range g.Nodes {
		if x, y, ok := h.GetComponentPosition(node.Id); ok {
			node.MoveTo(Position{X: x, Y: y})
		}
	}

	// Count crossings using the graph's edge intersection detection
	crossings := 0
	for i := 0; i < len(g.Edges)-1; i++ {
		for j := i + 1; j < len(g.Edges); j++ {
			if g.Edges[i].Intersects(g.Edges[j]) {
				crossings++
			}
		}
	}

	return crossings
}

// applyLayoutFromGraph extracts positions from the graph and stores them in h.Layout.
func (l *Layouter) applyLayoutFromGraph(h *hpsf.HPSF, g *Graph) {
	if h.Layout == nil {
		h.Layout = &hpsf.Layout{}
	}

	components := make([]hpsf.LayoutComponent, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		lc := hpsf.LayoutComponent{
			Name: node.Id,
			Position: &hpsf.Pos{
				X: node.Rect.Position.X,
				Y: node.Rect.Position.Y,
			},
		}
		// Preserve size if it was set (non-default)
		if node.Rect.Size.W != 0 || node.Rect.Size.H != 0 {
			lc.Size = &hpsf.Siz{
				W: node.Rect.Size.W,
				H: node.Rect.Size.H,
			}
		}
		components = append(components, lc)
	}
	h.Layout.Components = components
}
