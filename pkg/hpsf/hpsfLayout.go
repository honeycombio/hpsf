package hpsf

// NodeSize contains default width and height for nodes in the layout
type NodeSize struct {
	Width  int
	Height int
}

// DefaultNodeSize returns the default node size for layout
func DefaultNodeSize() NodeSize {
	return NodeSize{Width: 100, Height: 50}
}

// GetComponentPosition retrieves the layout position for a component.
// Returns (0, 0, false) if the component has no layout information.
func (h *HPSF) GetComponentPosition(componentName string) (x, y int, ok bool) {
	if h.Layout == nil {
		return 0, 0, false
	}

	for _, lc := range h.Layout.Components {
		if lc.Name == componentName && lc.Position != nil {
			return lc.Position.X, lc.Position.Y, true
		}
	}

	return 0, 0, false
}

// GetComponentSize retrieves the layout size for a component.
// Returns (0, 0, false) if the component has no size information.
func (h *HPSF) GetComponentSize(componentName string) (width, height int, ok bool) {
	if h.Layout == nil {
		return 0, 0, false
	}

	for _, lc := range h.Layout.Components {
		if lc.Name == componentName && lc.Size != nil {
			return lc.Size.W, lc.Size.H, true
		}
	}

	return 0, 0, false
}
