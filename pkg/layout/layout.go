package layout

/*
  This implements an automatic layout for the graph.
  It works like this:

  First we do a cycle detection pass, to ensure there are no cycles in the graph. Cycles
  are never valid in Pipeline Builder, so we can assume the graph is a DAG. We do a
  topological sort to order the nodes, and then we can use that in later steps.

  Layout goes from left to right, starting with nodes that have no inputs.
  Each node is placed in a column, with spacing between columns.
  Nodes in a column are stacked vertically, with spacing between nodes.

  Nodes are assigned to columns based on their distance from input nodes.
  The longest distance from any input node determines the column.
  Columns are evenly spaced horizontally, with a fixed horizontal spacing determined by the widest node in the graph plus a horizontal margin.

  The vertical position of nodes in a column is called the row.
  The total number of rows in the graph is determined by the maximum number of nodes in the longest column.
  The left edges of all nodes in a column are aligned.
  Rows are evenly spaced vertically, with a fixed vertical spacing determined by the tallest node in the graph plus a vertical margin.

  Both row and column spacing is a multiple of the snap grid size, which is a constant.

  The initial ordering of nodes within a column is determined by a sort, where the sort key is the concatenation of:
  - The lowest row value of the source nodes connected to the node's inputs (or 0 if no inputs)
  - The lowest port index of the source nodes connected to the node's inputs (or 0 if no inputs)
  - The node ID (to ensure a stable sort)

  We can detect edge crossings and try to reduce them by swapping nodes within a column.
  To detect edge crossings, we look at each pair of edges. We use the positioning of the nodes
  to calculate proxy connection points (they don't have to be exact, just ordered correctly).
  We can then do standard cross products to see if the line segments of the edges intersect.

  We can try to reduce edge crossings by swapping nodes within a column.

  We do this in 2 passes:
  * First, we do a barycentric pass, which attempts to put nodes close to their neighbor (connected) nodes by sweeping
    back and forth across the graph, reorganizing columns.
  * Then, we iterate over all pairs of nodes in a column, and calculate the number of edge crossings before and after
    the swap. If the swap reduces the number of edge crossings, we perform the swap. We repeat this process until no more
	swaps can be made that reduce edge crossings, or we reach an iteration limit.

  Finally, we try to minimize edge lengths by moving nodes up or down within their column.
  We do this by iterating over all nodes in a column, and calculating the total edge length before and after moving the node up or down one row.
  If moving the node up or down reduces the total edge length, we perform the move.
  We repeat this process until no more moves can be made that reduce edge lengths.

  This is a greedy algorithm and may not produce the optimal layout, but it should be good enough for most cases.
  The layout algorithm is O(N^2) in the number of nodes, which should be acceptable for most graphs
  in Pipeline Builder.
*/

import (
	"errors"
	"sort"
)

// Error values returned by layout.
var (
	ErrCycleDetected = errors.New("graph contains a cycle (layout requires a DAG)")
)

// LayoutOption allows callers to tweak behavior.
type LayoutOption func(*layoutConfig)

type layoutConfig struct {
	OptimizeCrossings bool
	OptimizeLengths   bool
	SnapGridSize      int
	HSeparation       int
	VSeparation       int
	MaxSwapIterations int
}

func defaultLayoutConfig() *layoutConfig {
	return &layoutConfig{
		OptimizeCrossings: true,
		OptimizeLengths:   true,
		SnapGridSize:      10,
		HSeparation:       40,
		VSeparation:       30,
		MaxSwapIterations: 50,
	}
}

// DisableCrossingOpt disables the crossing-reduction heuristic.
func DisableCrossingOpt() LayoutOption { return func(c *layoutConfig) { c.OptimizeCrossings = false } }

// DisableLengthOpt disables the edge-length minimization heuristic.
func DisableLengthOpt() LayoutOption { return func(c *layoutConfig) { c.OptimizeLengths = false } }

// WithSnapGridSize sets the grid size for snapping coordinates (default: 10).
func WithSnapGridSize(size int) LayoutOption { return func(c *layoutConfig) { c.SnapGridSize = size } }

// WithHSeparation sets the horizontal separation between columns (default: 40).
// It should be a multiple of the snap grid size for best results.
func WithHSeparation(separation int) LayoutOption {
	return func(c *layoutConfig) { c.HSeparation = separation }
}

// WithVSeparation sets the vertical separation between rows (default: 30).
// It should be a multiple of the snap grid size for best results.
func WithVSeparation(separation int) LayoutOption {
	return func(c *layoutConfig) { c.VSeparation = separation }
}

// WithMaxSwapIterations sets the maximum iterations for crossing reduction (default: 50).
func WithMaxSwapIterations(max int) LayoutOption {
	return func(c *layoutConfig) { c.MaxSwapIterations = max }
}

// AutoLayout computes an automatic layout for the graph. It is idempotent given the same node set & edges.
// Steps (all executed only if graph is a DAG):
//  1. Cycle detection (topological ordering). Returns ErrCycleDetected if a cycle is found.
//  2. Column assignment (longest distance from any source).
//  3. Initial ordering within columns.
//  4. Position assignment (snap to grid).
//  5. Crossing reduction (optional heuristic).
//  6. Edge length minimization (optional heuristic).
//
// Future work: expose margins & snap size as options; support SCC condensation instead of hard error on cycles.
func (g *Graph) AutoLayout(opts ...LayoutOption) error {
	if len(g.Nodes) == 0 { // nothing to do
		return nil
	}

	cfg := defaultLayoutConfig()
	for _, o := range opts {
		o(cfg)
	}

	// 1. Topological order (cycle detection)
	order, err := g.topologicalOrder()
	if err != nil {
		return err
	}

	// 2. Column assignment (longest distance from any source) using topo order
	colIndex := g.assignColumns(order)

	// 3. Group nodes per column & compute max node width/height for spacing
	columns, _, _ := g.groupColumnsAndSizes(colIndex)

	// 4. Initial row ordering within each column (before heuristics)
	rowIndex := g.orderRows(columns, colIndex)

	// 5. Reduce edge crossings (barycentric + greedy swaps + two-column coordinated swaps)
	if cfg.OptimizeCrossings {
		g.reduceCrossings(columns, colIndex, rowIndex, cfg)
	}

	// 6. Assign concrete positions snapped to grid
	g.assignPositions(columns, colIndex, rowIndex, cfg)

	// 7. Column vertical shift optimization to tighten incoming edge lengths.
	// This never introduces crossings since it preserves relative node order within columns.
	if cfg.OptimizeLengths {
		g.shiftColumnsForIncoming(columns, colIndex, cfg)
	}

	return nil
}

// topologicalOrder returns a topological ordering of nodes or an error if a cycle is detected.
func (g *Graph) topologicalOrder() ([]*Node, error) {
	// Build in-degree and adjacency list.
	indegree := make(map[*Node]int, len(g.Nodes))
	adj := make(map[*Node][]*Node, len(g.Nodes))
	for _, n := range g.Nodes {
		indegree[n] = 0
	}
	for _, e := range g.Edges {
		src := e.From.Node
		dst := e.To.Node
		if src == nil || dst == nil { // malformed edge; ignore silently for now
			continue
		}
		if src == dst { // self-loop => cycle
			return nil, ErrCycleDetected
		}
		adj[src] = append(adj[src], dst)
		indegree[dst]++
	}

	// Queue of zero in-degree nodes.
	queue := make([]*Node, 0, len(g.Nodes))
	for n, d := range indegree {
		if d == 0 {
			queue = append(queue, n)
		}
	}

	order := make([]*Node, 0, len(g.Nodes))
	for len(queue) > 0 {
		// pop last (stack semantics ok; ordering arbitrary but deterministic given Go map iteration is randomized across runs, but if deterministic order is required we could sort IDs.)
		n := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		order = append(order, n)
		for _, m := range adj[n] {
			indegree[m]--
			if indegree[m] == 0 {
				queue = append(queue, m)
			}
		}
	}

	if len(order) != len(g.Nodes) {
		return nil, ErrCycleDetected
	}

	return order, nil
}

// assignColumns computes the column index for each node: longest distance from any source.
// Returns a map of node->column index.
func (g *Graph) assignColumns(order []*Node) map[*Node]int {
	col := make(map[*Node]int, len(order))
	// For faster predecessor lookup, build incoming edges list.
	incoming := make(map[*Node][]*Node, len(order))
	for _, e := range g.Edges {
		src := e.From.Node
		dst := e.To.Node
		if src == nil || dst == nil {
			continue
		}
		incoming[dst] = append(incoming[dst], src)
	}
	for _, n := range order {
		maxPred := 0
		if preds, ok := incoming[n]; ok {
			for _, p := range preds {
				if c := col[p] + 1; c > maxPred {
					maxPred = c
				}
			}
		}
		col[n] = maxPred
	}
	return col
}

// groupColumnsAndSizes organizes nodes into columns based on computed column index and
// returns: slice of columns (each a slice of *Node) and the maximum width & height seen.
func (g *Graph) groupColumnsAndSizes(col map[*Node]int) ([][]*Node, int, int) {
	maxCol := 0
	maxW := 0
	maxH := 0
	for n, c := range col {
		if c > maxCol {
			maxCol = c
		}
		w := n.Rect.Size.W
		h := n.Rect.Size.H
		if w > maxW {
			maxW = w
		}
		if h > maxH {
			maxH = h
		}
	}
	cols := make([][]*Node, maxCol+1)
	for n, c := range col {
		cols[c] = append(cols[c], n)
	}
	return cols, maxW, maxH
}

// orderRows determines an initial row ordering within each column.
// Strategy:
//
//	Process columns left-to-right; for each column we sort that column's nodes by a key:
//	  (min predecessor row, min predecessor output port index, node.Id)
//	Nodes with no predecessors get key (0, 0, Id).
//	Row indices are local to a column (they restart at 0 for each column) which matches test expectations.
//
// Returns a map of node -> rowIndex (local to its column).
func (g *Graph) orderRows(columns [][]*Node, col map[*Node]int) map[*Node]int {
	row := make(map[*Node]int, len(g.Nodes))

	// Precompute incoming edges per node for efficiency.
	incoming := make(map[*Node][]*Edge, len(g.Nodes))
	for _, e := range g.Edges {
		if e == nil || e.From == nil || e.To == nil || e.From.Node == nil || e.To.Node == nil {
			continue
		}
		incoming[e.To.Node] = append(incoming[e.To.Node], e)
	}

	// Iterate columns left to right so predecessor rows are known.
	for _, colNodes := range columns {
		if len(colNodes) == 0 {
			continue
		}
		// Build sortable slice with computed keys.
		// Use average (barycentric) position of predecessors instead of minimum
		// to get a better initial ordering that considers all incoming edges.
		type keyed struct {
			n           *Node
			avgPredRow  float64
			minPredPort int
		}
		keyedNodes := make([]keyed, 0, len(colNodes))
		for _, n := range colNodes {
			avgRow := 0.0
			minPort := 0
			if inc, ok := incoming[n]; ok && len(inc) > 0 {
				totalRow := 0.0
				countRow := 0
				const big = int(^uint(0) >> 1) // max int
				minPort = big
				for _, e := range inc {
					src := e.From.Node
					if src == nil {
						continue
					}
					r, okR := row[src]
					if okR {
						totalRow += float64(r)
						countRow++
					}
					if e.From.Index < minPort {
						minPort = e.From.Index
					}
				}
				if countRow > 0 {
					avgRow = totalRow / float64(countRow)
				}
				if minPort == big {
					minPort = 0
				}
			}
			keyedNodes = append(keyedNodes, keyed{n: n, avgPredRow: avgRow, minPredPort: minPort})
		}
		sort.SliceStable(keyedNodes, func(i, j int) bool {
			if keyedNodes[i].avgPredRow != keyedNodes[j].avgPredRow {
				return keyedNodes[i].avgPredRow < keyedNodes[j].avgPredRow
			}
			if keyedNodes[i].minPredPort != keyedNodes[j].minPredPort {
				return keyedNodes[i].minPredPort < keyedNodes[j].minPredPort
			}
			return keyedNodes[i].n.Id < keyedNodes[j].n.Id
		})
		// Assign row indices local to this column.
		for rIdx, k := range keyedNodes {
			row[k.n] = rIdx
		}
	}
	return row
}

// assignPositions sets concrete X,Y coordinates for each node based on its column & row indices.
// Horizontal spacing is uniform across all columns based on the widest node.
// Vertical spacing is calculated per column based on the tallest node in each column,
// allowing columns with smaller nodes to be more compact.
// Row index is local to a column (restarts each column).
func (g *Graph) assignPositions(columns [][]*Node, col map[*Node]int, row map[*Node]int, cfg *layoutConfig) {
	if len(g.Nodes) == 0 {
		return
	}

	// Snap spacing itself to grid so aligned columns/rows remain on grid after multiplication.
	snap := func(v int) int {
		if v%cfg.SnapGridSize == 0 {
			return v
		}
		return ((v / cfg.SnapGridSize) + 1) * cfg.SnapGridSize
	}

	// Compute max width for horizontal spacing (uniform across all columns)
	maxW := 0
	for _, n := range g.Nodes {
		if n.Rect.Size.W > maxW {
			maxW = n.Rect.Size.W
		}
	}
	hSpace := snap(maxW + cfg.HSeparation)

	// Compute vertical spacing per column based on the tallest node in each column
	colVSpace := make([]int, len(columns))
	for colIdx, colNodes := range columns {
		maxH := 0
		for _, n := range colNodes {
			if n.Rect.Size.H > maxH {
				maxH = n.Rect.Size.H
			}
		}
		colVSpace[colIdx] = snap(maxH + cfg.VSeparation)
	}

	// Assign positions
	for _, n := range g.Nodes {
		c := col[n]
		r := row[n]
		x := c * hSpace
		y := r * colVSpace[c]
		// already snapped because spacing multiples are snapped, but ensure safety
		x = snap(x)
		y = snap(y)
		n.MoveTo(Position{X: x, Y: y})
	}
}

// findDownstreamNodes finds all nodes that are downstream (reachable) from the given node.
// This is used for path-aware swapping to ensure that when we swap two nodes, we also
// swap their downstream dependencies to maintain path ordering.
func (g *Graph) findDownstreamNodes(start *Node, col map[*Node]int) []*Node {
	visited := make(map[*Node]bool)
	downstream := make([]*Node, 0)

	var dfs func(*Node)
	dfs = func(n *Node) {
		if visited[n] {
			return
		}
		visited[n] = true

		// Find all edges where this node is the source
		for _, e := range g.Edges {
			if e == nil || e.From == nil || e.To == nil || e.From.Node == nil || e.To.Node == nil {
				continue
			}
			if e.From.Node == n {
				downstream = append(downstream, e.To.Node)
				dfs(e.To.Node)
			}
		}
	}

	// Don't include the start node itself in the downstream list
	for _, e := range g.Edges {
		if e == nil || e.From == nil || e.To == nil || e.From.Node == nil || e.To.Node == nil {
			continue
		}
		if e.From.Node == start {
			downstream = append(downstream, e.To.Node)
			dfs(e.To.Node)
		}
	}

	return downstream
}

// countCrossings counts edge crossings.
//
// Only considers edges where the source column < target column. Multi-column span edges are treated the same.
// countCrossings computes the number of true geometric crossings between straight-line edge segments.
// Each edge is represented as a segment from the center of an output port on the source node's right edge
// to the center of an input port on the target node's left edge. We synthesize node positions from the
// provided column/row maps (so this can be used before actual positioning) using the same spacing logic
// as assignPositions (max node size + default margins). Port vertical offsets are simply the port index.
// We ignore intersections that only touch at endpoints (shared node or shared exact point) and only count
// proper interior segment intersections.
func (g *Graph) countCrossings(row map[*Node]int, col map[*Node]int) int {
	if len(g.Edges) < 2 {
		return 0
	}
	filtered := make([]*Edge, 0, len(g.Edges))
	for _, e := range g.Edges {
		if e == nil || e.From == nil || e.To == nil || e.From.Node == nil || e.To.Node == nil {
			continue
		}
		if col[e.From.Node] >= col[e.To.Node] { // only forward
			continue
		}
		filtered = append(filtered, e)
	}
	if len(filtered) < 2 {
		return 0
	}
	cnt := 0
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[i].Intersects(filtered[j]) {
				cnt++
			}
		}
	}
	return cnt
}

// reduceCrossings attempts to reduce edge crossings by swapping rows of nodes within each column.
// It mutates the row map in place and returns true if any swap was applied.
// This function runs multiple passes of different optimization strategies until no further
// improvement is possible or max iterations is reached.
func (g *Graph) reduceCrossings(columns [][]*Node, col map[*Node]int, row map[*Node]int, cfg *layoutConfig) bool {
	improved := false
	baseCross := g.countCrossings(row, col)
	if baseCross == 0 {
		return false
	}

	// Run multiple passes until we can't improve anymore
	maxPasses := 5
	for pass := 0; pass < maxPasses && baseCross > 0; pass++ {
		passImproved := false

		// Barycentric sweeps - only use predecessor-based (leftward) ordering
		// The successor-based (rightward) pass can undo good orderings
		for sweep := 0; sweep < 4; sweep++ {
			leftChanged := g.barycentricPass(columns, col, row, -1, cfg)
			if !leftChanged {
				break
			}
			newCross := g.countCrossings(row, col)
			if newCross < baseCross {
				baseCross = newCross
				improved = true
				passImproved = true
			}
			if baseCross == 0 {
				return true
			}
		}

		// Greedy pairwise swaps within single columns
		// We do multiple iterations per column until no more improvements or max iterations reached.
		// This is a local optimization and may not reach a global optimum.
		// We process columns independently, left to right, so earlier columns are fixed when processing later ones.
		// This is a heuristic; a more global approach could yield better results but would be more complex.
		for _, colNodes := range columns {
			if len(colNodes) < 2 {
				continue
			}
			iterations := 0
			for iterations < cfg.MaxSwapIterations {
				changed := false
				// Try all unordered pairs (i<j)
				for i := 0; i < len(colNodes)-1; i++ {
					n1 := colNodes[i]
					for j := i + 1; j < len(colNodes); j++ {
						n2 := colNodes[j]
						// Swap simulated by exchanging row indices
						r1, r2 := row[n1], row[n2]
						row[n1], row[n2] = r2, r1
						// Recompute positions to reflect swapped rows
						g.assignPositions(columns, col, row, cfg)
						newCross := g.countCrossings(row, col)
						if newCross < baseCross { // accept
							baseCross = newCross
							changed = true
							improved = true
							passImproved = true
						} else { // revert
							row[n1], row[n2] = r1, r2
							g.assignPositions(columns, col, row, cfg)
						}
						if baseCross == 0 {
							return true
						}
					}
				}
				if !changed {
					break
				}
				iterations++
			}
		}

		// Two-column coordinated swaps
		// If crossings remain, try swapping node pairs across adjacent column pairs.
		// This can find solutions that require coordinated swaps (e.g., swap A↔B in col i AND swap C↔D in col i+1).
		if baseCross > 0 {
			twoColChanged := g.tryTwoColumnSwaps(columns, col, row, &baseCross, cfg)
			if twoColChanged {
				improved = true
				passImproved = true
			}
		}

		// Column vertical shift optimization - run this even during crossing reduction
		// because shifting columns can sometimes resolve crossings by aligning nodes
		// with their predecessors. We need to check crossings again after shifting.
		if baseCross > 0 {
			// Assign positions first so shiftColumnsForIncoming has actual coordinates to work with
			g.assignPositions(columns, col, row, cfg)
			g.shiftColumnsForIncoming(columns, col, cfg)
			// After shifting, recount crossings - the shift might have resolved some
			newCross := g.countCrossings(row, col)
			if newCross < baseCross {
				baseCross = newCross
				improved = true
				passImproved = true
			}
		}

		// If this pass didn't improve anything, we're done
		if !passImproved {
			break
		}
	}

	return improved
}

// tryTwoColumnSwaps attempts coordinated swaps across pairs of adjacent columns.
// For each adjacent column pair (i, i+1), it tries swapping one pair of nodes in column i
// together with one pair of nodes in column i+1, accepting the swap if it reduces crossings.
// This can solve cases where single-column optimization gets stuck in local optima.
// Returns true if any improvement was made. Updates baseCross pointer with the new crossing count.
func (g *Graph) tryTwoColumnSwaps(columns [][]*Node, col map[*Node]int, row map[*Node]int, baseCross *int, cfg *layoutConfig) bool {
	improved := false

	for colIdx := 0; colIdx < len(columns)-1; colIdx++ {
		leftCol := columns[colIdx]
		rightCol := columns[colIdx+1]

		if len(leftCol) < 2 || len(rightCol) < 2 {
			continue
		}

		// Try all combinations of swaps in both columns
		for iL := 0; iL < len(leftCol)-1; iL++ {
			for jL := iL + 1; jL < len(leftCol); jL++ {
				nL1, nL2 := leftCol[iL], leftCol[jL]
				rL1, rL2 := row[nL1], row[nL2]

				for iR := 0; iR < len(rightCol)-1; iR++ {
					for jR := iR + 1; jR < len(rightCol); jR++ {
						nR1, nR2 := rightCol[iR], rightCol[jR]
						rR1, rR2 := row[nR1], row[nR2]

						// Try swapping both pairs simultaneously
						row[nL1], row[nL2] = rL2, rL1
						row[nR1], row[nR2] = rR2, rR1
						g.assignPositions(columns, col, row, cfg)
						newCross := g.countCrossings(row, col)

						if newCross < *baseCross {
							// Accept the coordinated swap
							*baseCross = newCross
							improved = true
							if *baseCross == 0 {
								return true
							}
						} else {
							// Revert both swaps
							row[nL1], row[nL2] = rL1, rL2
							row[nR1], row[nR2] = rR1, rR2
							g.assignPositions(columns, col, row, cfg)
						}
					}
				}
			}
		}
	}

	return improved
}

// barycentricPass reorders rows within each column based on the average (barycentric) row
// of neighboring nodes in the adjacent column (direction -1: predecessors / left, +1: successors / right).
// Returns true if any row index changed. It updates positions if changes occur.
func (g *Graph) barycentricPass(columns [][]*Node, col map[*Node]int, row map[*Node]int, direction int, cfg *layoutConfig) bool {
	changed := false
	for cIdx, colNodes := range columns {
		if len(colNodes) < 2 {
			continue
		}
		neighborCol := cIdx + direction
		if neighborCol < 0 || neighborCol >= len(columns) {
			continue
		}
		type scored struct {
			n   *Node
			avg float64
		}
		scoredNodes := make([]scored, 0, len(colNodes))
		for _, n := range colNodes {
			var total float64
			count := 0
			for _, e := range g.Edges {
				if e == nil || e.From == nil || e.To == nil || e.From.Node == nil || e.To.Node == nil {
					continue
				}
				if direction == -1 { // predecessors
					if e.To.Node == n && col[e.From.Node] == neighborCol {
						total += float64(row[e.From.Node])
						count++
					}
				} else { // successors
					if e.From.Node == n && col[e.To.Node] == neighborCol {
						total += float64(row[e.To.Node])
						count++
					}
				}
			}
			avg := float64(row[n])
			if count > 0 {
				avg = total / float64(count)
			}
			scoredNodes = append(scoredNodes, scored{n: n, avg: avg})
		}
		sort.SliceStable(scoredNodes, func(i, j int) bool {
			if scoredNodes[i].avg != scoredNodes[j].avg {
				return scoredNodes[i].avg < scoredNodes[j].avg
			}
			if row[scoredNodes[i].n] != row[scoredNodes[j].n] {
				return row[scoredNodes[i].n] < row[scoredNodes[j].n]
			}
			return scoredNodes[i].n.Id < scoredNodes[j].n.Id
		})
		for newR, s := range scoredNodes {
			if row[s.n] != newR {
				row[s.n] = newR
				changed = true
			}
		}
	}
	if changed {
		g.assignPositions(columns, col, row, cfg)
	}
	return changed
}

// shiftColumnsForIncoming iterates left-to-right over columns (starting at 1) and applies a vertical
// shift to every node in a column to minimize the total vertical distance of its incoming edges.
// For a column c, we gather all edges whose target is in column c. Let y_from_i be the Y coordinate
// of the source port center and y_to_i the current Y coordinate of the target port center. We want a
// delta added uniformly to all targets so as to minimize sum |y_from_i - (y_to_i + delta)|. The
// minimizing delta is the median of the set {y_from_i - y_to_i}. We snap the chosen shift to the
// nearest snapGridSize multiple to preserve grid alignment. After all columns are shifted, if any
// node ends up with negative Y we normalize by translating the entire graph downward so min Y = 0.
func (g *Graph) shiftColumnsForIncoming(columns [][]*Node, col map[*Node]int, cfg *layoutConfig) {
	if len(columns) == 0 {
		return
	}
	// Snap to nearest grid multiple (ties to up). We also treat shifts smaller than half a grid as zero.
	snap := func(v int) int {
		if v == 0 {
			return 0
		}
		neg := v < 0
		if neg {
			v = -v
		}
		rem := v % cfg.SnapGridSize
		base := v - rem
		if rem*2 >= cfg.SnapGridSize { // round up
			base += cfg.SnapGridSize
		}
		// If rounded value is less than half grid, treat as zero to avoid jitter on already aligned columns.
		if base < cfg.SnapGridSize/2 {
			base = 0
		}
		if neg {
			base = -base
		}
		return base
	}
	// Iterate columns 1..end (column 0 has no predecessors to align to)
	for c := 1; c < len(columns); c++ {
		colNodes := columns[c]
		if len(colNodes) == 0 {
			continue
		}
		// Single-node columns are eligible for shifting; aligning their lone target can still reduce edge length.
		diffs := make([]int, 0, 8)
		for _, n := range colNodes {
			for _, e := range g.Edges {
				if e == nil || e.From == nil || e.To == nil || e.From.Node == nil || e.To.Node == nil {
					continue
				}
				if e.To.Node != n {
					continue
				}
				// Only consider forward edges (sources in earlier columns) — by construction due to DAG
				if col[e.From.Node] >= c {
					continue
				}
				fromY := e.From.Position(PortTypeOutput).Y
				toY := e.To.Position(PortTypeInput).Y
				diffs = append(diffs, fromY-toY)
			}
		}
		if len(diffs) == 0 {
			continue // no incoming edges; nothing to align
		}
		sort.Ints(diffs)
		median := diffs[len(diffs)/2]
		shift := snap(median)
		if shift == 0 {
			continue
		}
		for _, n := range colNodes {
			pos := n.Rect.Position
			pos.Y += shift
			n.MoveTo(pos)
		}
	}
	// Normalize to non-negative Y if needed
	minY := 0
	for i, n := range g.Nodes {
		if i == 0 || n.Rect.Position.Y < minY {
			minY = n.Rect.Position.Y
		}
	}
	if minY < 0 {
		offset := -minY
		for _, n := range g.Nodes {
			pos := n.Rect.Position
			pos.Y += offset
			n.MoveTo(pos)
		}
	}
}
