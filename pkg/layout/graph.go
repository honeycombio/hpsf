package layout

const (
	portMargin = 10 // minimum distance from node corners
)

type PortType string

const (
	PortTypeInput  PortType = "input"
	PortTypeOutput PortType = "output"
)

type RenderType string

const (
	RenderTypeDOT RenderType = "dot"
	RenderTypeSVG RenderType = "svg"
)

type RenderOption struct {
	Key   string
	Value string
}

type Position struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
}

type Size struct {
	W int `yaml:"w"`
	H int `yaml:"h"`
}

type Rect struct {
	Position Position `yaml:"position"`
	Size     Size     `yaml:"size"`
}

type Node struct {
	Id      string  `yaml:"id"`
	Style   string  `yaml:"style,omitempty"`
	Rect    Rect    `yaml:"rect"`
	Inputs  []*Port `yaml:"inputs,omitempty"`
	Outputs []*Port `yaml:"outputs,omitempty"`
}

func (n *Node) MoveTo(p Position) {
	n.Rect.Position = p
}

// Port represents an input or output port of a Node. Inputs are always on the left, outputs on the right.
// Ports are identified by their index, which is unique per node and port type, and indices go from top to bottom
// in natural order (1, 2, ...). Index is optional if there is only one port of a given type.
type Port struct {
	Node  *Node `yaml:"node,omitempty"`
	Index int   `yaml:"index"`
}

// Position returns the (x,y) center coordinates of this port based on the owning node's
// rectangle and the ordering of ports of the same type. It mirrors the logic in
// Graph.portPosition but is instance-based for convenience. Returns (0,0) if
// the node or its sizing info is missing, or if the port cannot be located.
func (p *Port) Position(portType PortType) Position {
	if p == nil || p.Node == nil {
		return Position{0, 0}
	}
	n := p.Node
	x := float64(n.Rect.Position.X)
	y := float64(n.Rect.Position.Y)
	w := float64(n.Rect.Size.W)
	h := float64(n.Rect.Size.H)
	var ports []*Port
	if portType == PortTypeInput {
		ports = n.Inputs
	} else {
		ports = n.Outputs
	}
	if len(ports) == 0 {
		return Position{0, 0}
	}
	idx := -1
	for i, pt := range ports {
		if pt.Index == p.Index {
			idx = i
			break
		}
	}
	if idx == -1 {
		return Position{0, 0}
	}
	available := h - 2*portMargin
	spacing := available / float64(len(ports)+1)
	py := y + portMargin + spacing*float64(idx+1)
	px := x
	if portType == PortTypeOutput {
		px = x + w
	}
	return Position{X: int(px), Y: int(py)}
}

// Convenience helpers
func (p *Port) IsInput() bool  { return p != nil && p.Node != nil && containsPort(p.Node.Inputs, p) }
func (p *Port) IsOutput() bool { return p != nil && p.Node != nil && containsPort(p.Node.Outputs, p) }

func containsPort(sl []*Port, target *Port) bool {
	for _, p := range sl {
		if p == target {
			return true
		}
	}
	return false
}

// Edges always go from an output port to an input port.
type Edge struct {
	From *Port `yaml:"from"`
	To   *Port `yaml:"to"`
}

// Intersects determines whether this edge geometrically crosses another edge (other) using
// straight-line segments between port centers. It excludes cases where the edges share
// any endpoint node (touching at a node is not considered a crossing) and ignores
// collinear overlaps. Assumes node positions and sizes are already assigned.
func (e *Edge) Intersects(other *Edge) bool {
	if e == nil || other == nil || e.From == nil || e.To == nil || other.From == nil || other.To == nil {
		return false
	}
	// Exclude only if they share an identical endpoint port (same node pointer AND same port index)
	samePort := func(a, b *Port) bool {
		if a == nil || b == nil || a.Node == nil || b.Node == nil {
			return false
		}
		return a.Node == b.Node && a.Index == b.Index
	}
	if samePort(e.From, other.From) || samePort(e.From, other.To) || samePort(e.To, other.From) || samePort(e.To, other.To) {
		return false
	}
	spFrom := e.From.Position(PortTypeOutput)
	spTo := e.To.Position(PortTypeInput)
	opFrom := other.From.Position(PortTypeOutput)
	opTo := other.To.Position(PortTypeInput)

	cross := func(o, a, b Position) int {
		return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
	}
	d1 := cross(spFrom, spTo, opFrom)
	d2 := cross(spFrom, spTo, opTo)
	d3 := cross(opFrom, opTo, spFrom)
	d4 := cross(opFrom, opTo, spTo)
	if d1 == 0 && d2 == 0 && d3 == 0 && d4 == 0 { // collinear
		return false
	}
	return (d1*d2) < 0 && (d3*d4) < 0
}

type Graph struct {
	Nodes []*Node `yaml:"nodes"`
	Edges []*Edge `yaml:"edges"`
}

func (g *Graph) AddNode(n *Node) {
	g.Nodes = append(g.Nodes, n)
}

func (g *Graph) AddEdge(e *Edge) {
	g.Edges = append(g.Edges, e)
}

func (g *Graph) FindNodeById(id string) *Node {
	for _, n := range g.Nodes {
		if n.Id == id {
			return n
		}
	}
	return nil
}

func (g *Graph) FindPort(nodeId string, portType PortType, index int) *Port {
	n := g.FindNodeById(nodeId)
	if n == nil {
		return nil
	}
	var ports []*Port
	if portType == PortTypeInput {
		ports = n.Inputs
	} else {
		ports = n.Outputs
	}
	for _, p := range ports {
		if p.Index == index {
			return p
		}
	}
	return nil
}
