package layout

import "testing"

// helper to construct a node at an explicit position with a given size and port counts
func nodeAt(id string, x, y, w, h, in, out int) *Node {
	n := &Node{Id: id, Rect: Rect{Position: Position{X: x, Y: y}, Size: Size{W: w, H: h}}}
	for i := 1; i <= in; i++ {
		p := &Port{Node: n, Index: i}
		n.Inputs = append(n.Inputs, p)
	}
	for i := 1; i <= out; i++ {
		p := &Port{Node: n, Index: i}
		n.Outputs = append(n.Outputs, p)
	}
	return n
}

// TestEdgeIntersects covers several geometric cases:
//  1. Proper intersection (X) of two edges.
//  2. Parallel / non‑intersecting edges.
//  3. Edges sharing the same start port (should NOT count as intersection).
//  4. Edges sharing the same end port (should NOT count as intersection).
//  5. Edges whose infinite lines would intersect but the finite segments do not (should be false).
func TestEdgeIntersects(t *testing.T) {
	// Common node sizing so port centers are at mid-height: single port => y = nodeY + h/2
	const w, h = 100, 40

	t.Run("proper crossing", func(t *testing.T) {
		// Layout:
		// Left column: S1 (y=0), S2 (y=200)
		// Right column: T1 (y=0), T2 (y=200)
		// Edges: S1->T2 and S2->T1 should cross.
		S1 := nodeAt("S1", 0, 0, w, h, 0, 1)
		S2 := nodeAt("S2", 0, 200, w, h, 0, 1)
		T1 := nodeAt("T1", 200, 0, w, h, 1, 0)
		T2 := nodeAt("T2", 200, 200, w, h, 1, 0)
		e1 := &Edge{From: &Port{Node: S1, Index: 1}, To: &Port{Node: T2, Index: 1}}
		e2 := &Edge{From: &Port{Node: S2, Index: 1}, To: &Port{Node: T1, Index: 1}}
		if !e1.Intersects(e2) {
			// Log coordinates for debugging
			p1a := e1.From.Position(PortTypeOutput)
			p1b := e1.To.Position(PortTypeInput)
			p2a := e2.From.Position(PortTypeOutput)
			p2b := e2.To.Position(PortTypeInput)
			t.Fatalf("expected intersection: e1 (%v->%v) e2 (%v->%v)", p1a, p1b, p2a, p2b)
		}
	})

	t.Run("non crossing parallel", func(t *testing.T) {
		S1 := nodeAt("S1", 0, 0, w, h, 0, 1)
		S2 := nodeAt("S2", 0, 200, w, h, 0, 1)
		T1 := nodeAt("T1", 200, 0, w, h, 1, 0)
		T2 := nodeAt("T2", 200, 200, w, h, 1, 0)
		e1 := &Edge{From: &Port{Node: S1, Index: 1}, To: &Port{Node: T1, Index: 1}}
		e2 := &Edge{From: &Port{Node: S2, Index: 1}, To: &Port{Node: T2, Index: 1}}
		if e1.Intersects(e2) {
			p1a := e1.From.Position(PortTypeOutput)
			p1b := e1.To.Position(PortTypeInput)
			p2a := e2.From.Position(PortTypeOutput)
			p2b := e2.To.Position(PortTypeInput)
			t.Fatalf("did not expect intersection: e1 (%v->%v) e2 (%v->%v)", p1a, p1b, p2a, p2b)
		}
	})

	t.Run("shared start port", func(t *testing.T) {
		S := nodeAt("S", 0, 100, w, h, 0, 1)
		A := nodeAt("A", 200, 0, w, h, 1, 0)
		B := nodeAt("B", 200, 200, w, h, 1, 0)
		// Both edges originate from identical port S.1
		e1 := &Edge{From: &Port{Node: S, Index: 1}, To: &Port{Node: A, Index: 1}}
		e2 := &Edge{From: &Port{Node: S, Index: 1}, To: &Port{Node: B, Index: 1}}
		if e1.Intersects(e2) {
			p1a := e1.From.Position(PortTypeOutput)
			p1b := e1.To.Position(PortTypeInput)
			p2a := e2.From.Position(PortTypeOutput)
			p2b := e2.To.Position(PortTypeInput)
			t.Fatalf("edges sharing start port should not intersect: e1 (%v->%v) e2 (%v->%v)", p1a, p1b, p2a, p2b)
		}
	})

	t.Run("shared end port", func(t *testing.T) {
		T := nodeAt("T", 200, 100, w, h, 1, 0)
		A := nodeAt("A", 0, 0, w, h, 0, 1)
		B := nodeAt("B", 0, 200, w, h, 0, 1)
		e1 := &Edge{From: &Port{Node: A, Index: 1}, To: &Port{Node: T, Index: 1}}
		e2 := &Edge{From: &Port{Node: B, Index: 1}, To: &Port{Node: T, Index: 1}}
		if e1.Intersects(e2) {
			p1a := e1.From.Position(PortTypeOutput)
			p1b := e1.To.Position(PortTypeInput)
			p2a := e2.From.Position(PortTypeOutput)
			p2b := e2.To.Position(PortTypeInput)
			t.Fatalf("edges sharing end port should not intersect: e1 (%v->%v) e2 (%v->%v)", p1a, p1b, p2a, p2b)
		}
	})

	t.Run("would intersect if extended", func(t *testing.T) {
		// Two segments whose infinite lines cross outside the span of both segments.
		// Segment1: (S1->T1) rising; Segment2: (S2->T2) falling; gap between x ranges.
		S1 := nodeAt("S1", 0, 0, w, h, 0, 1)     // port ~ (100,20)
		T1 := nodeAt("T1", 140, 60, w, h, 1, 0)  // port ~ (140,80)
		S2 := nodeAt("S2", 260, 60, w, h, 0, 1)  // port ~ (360,80)
		T2 := nodeAt("T2", 300, -20, w, h, 1, 0) // port ~ (300,0)
		e1 := &Edge{From: &Port{Node: S1, Index: 1}, To: &Port{Node: T1, Index: 1}}
		e2 := &Edge{From: &Port{Node: S2, Index: 1}, To: &Port{Node: T2, Index: 1}}
		if e1.Intersects(e2) {
			p1a := e1.From.Position(PortTypeOutput)
			p1b := e1.To.Position(PortTypeInput)
			p2a := e2.From.Position(PortTypeOutput)
			p2b := e2.To.Position(PortTypeInput)
			t.Fatalf("segments whose lines intersect only if extended should not count: e1 (%v->%v) e2 (%v->%v)", p1a, p1b, p2a, p2b)
		}
	})

	t.Run("axis crossing vs axis segment disjoint", func(t *testing.T) {
		// e1 crosses the X axis; e2 lies on the X axis but does not include the crossing point.
		// e1 endpoints y: -30 and +30 so it crosses at y=0 around x=200.
		// e2 spans along y=0 but starts to the right of that crossing (x>400), so no intersection.
		S1 := nodeAt("S1", 0, -50, w, h, 0, 1)   // center y = -30
		T1 := nodeAt("T1", 300, 10, w, h, 1, 0)  // center y = 30; crossing at ~x=200
		S2 := nodeAt("S2", 350, -20, w, h, 0, 1) // center y = 0, start x ~ 450
		T2 := nodeAt("T2", 500, -20, w, h, 1, 0) // center y = 0, end x ~ 500
		e1 := &Edge{From: &Port{Node: S1, Index: 1}, To: &Port{Node: T1, Index: 1}}
		e2 := &Edge{From: &Port{Node: S2, Index: 1}, To: &Port{Node: T2, Index: 1}}
		if e1.Intersects(e2) {
			p1a := e1.From.Position(PortTypeOutput)
			p1b := e1.To.Position(PortTypeInput)
			p2a := e2.From.Position(PortTypeOutput)
			p2b := e2.To.Position(PortTypeInput)
			t.Fatalf("expected no intersection: e1 crosses axis at x~200, e2 spans x~450-500; e1 (%v->%v) e2 (%v->%v)", p1a, p1b, p2a, p2b)
		}
	})
}
