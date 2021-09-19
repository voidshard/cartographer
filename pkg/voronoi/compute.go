package voronoi

import (
	"fmt"

	"github.com/pzsz/voronoi"

	"github.com/voidshard/cartographer/pkg/geo"
	"github.com/voidshard/cartographer/pkg/shapes"
)

// safeCompute wraps calling pzsz voronoi .. which can panic on some arrangements of points
func safeCompute(errors chan<- error, verts []voronoi.Vertex, bbox voronoi.BBox) *voronoi.Diagram {
	defer close(errors)
	defer func() {
		if r := recover(); r != nil {
			errors <- fmt.Errorf("caught panic in voronoi.ComputeDiagram: %s", r)
		}
	}()
	// this func can panic given certain arrangements of points so we'll catch the panic
	// and return this as an error :(
	return voronoi.ComputeDiagram(verts, bbox, true)
}

// Compute a voronoi graph with cells at the given points
func Compute(pts []*shapes.Point, poly *shapes.Polygon) (*Graph, error) {
	verts := make([]voronoi.Vertex, len(pts))
	for i, pt := range pts {
		pt = pt.Round()
		verts[i] = voronoi.Vertex{X: pt.X, Y: pt.Y}
	}

	x0, y0, x1, y1 := poly.Bounds()
	bbox := voronoi.NewBBox(x0, x1, y0, y1)

	errs := make(chan error)
	done := make(chan bool)

	var finalErr error

	go func() {
		for err := range errs {
			if err != nil {
				if finalErr != nil {
					finalErr = fmt.Errorf("%w %v", finalErr, err)
				} else {
					finalErr = err
				}
			}
		}
		done <- true
	}()

	diag := safeCompute(errs, verts, bbox)
	<-done
	if finalErr != nil {
		return nil, finalErr
	}
	return newGraph(diag, poly), nil
}

// newGraph builds our graph wrapper from a voronoi lib diagram
func newGraph(diag *voronoi.Diagram, poly *shapes.Polygon) *Graph {
	d := &Graph{
		Poly:  poly,
		Cells: []*Cell{},
		Verts: []*shapes.Point{},
		Edges: []*Edge{},
	}

	pntMap := map[float64]map[float64]int{}

	getPoint := func(x, y float64) (int, *shapes.Point) {
		pt := shapes.Pt(x, y).Round()

		xdata, ok := pntMap[pt.X]
		if !ok {
			xdata = map[float64]int{}
		}

		index, ok := xdata[pt.Y]
		if ok {
			return index, pt
		}

		index = len(d.Verts)
		xdata[pt.Y] = index
		pntMap[pt.X] = xdata
		d.Verts = append(d.Verts, pt)

		return index, pt
	}

	edgemap := map[*voronoi.Edge]int{}
	cellmap := map[*voronoi.Cell]int{}

	for i, edge := range diag.Edges {
		idxA, ptA := getPoint(edge.Va.X, edge.Va.Y)
		idxB, ptB := getPoint(edge.Vb.X, edge.Vb.Y)

		newEdge := &Edge{
			ID:       i,
			PointAID: idxA,
			PointBID: idxB,
			PointA:   ptA,
			PointB:   ptB,
		}
		edgemap[edge] = i
		d.Edges = append(d.Edges, newEdge)
	}

	for i, cell := range diag.Cells {
		c := &Cell{
			ID:    i,
			Site:  shapes.Pt(cell.Site.X, cell.Site.Y).Round(),
			Edges: []*HalfEdge{},
		}

		pts := []*shapes.Point{}
		for _, half := range cell.Halfedges {
			myedge := d.Edges[edgemap[half.Edge]]
			c.Edges = append(c.Edges, &HalfEdge{Edge: myedge, Angle: half.Angle})

			_, pa := getPoint(half.Edge.Va.X, half.Edge.Va.Y)
			_, pb := getPoint(half.Edge.Vb.X, half.Edge.Vb.Y)
			pts = append(pts, pa, pb)
		}

		c.Bounds = shapes.NewPolygon(geo.SortPoints(c.Site, geo.ConvexHull(pts)))
		d.Cells = append(d.Cells, c)
		cellmap[cell] = i
	}

	for i, edge := range diag.Edges {
		myedge := d.Edges[i]

		if edge.LeftCell != nil {
			left := cellmap[edge.LeftCell]
			myedge.LeftCell = d.Cells[left]
		}

		if edge.RightCell != nil {
			right := cellmap[edge.RightCell]
			myedge.RightCell = d.Cells[right]
		}
	}

	return d
}
