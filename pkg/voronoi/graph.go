package voronoi

import (
	"github.com/voidshard/cartographer/pkg/shapes"
)

// Graph represents a voronoi diagram with cells, edges & verts.
type Graph struct {
	Poly  *shapes.Polygon
	Cells []*Cell
	Verts []*shapes.Point
	Edges []*Edge
}

//
func (d *Graph) ClosestVert(a *shapes.Point) *shapes.Point {
	_, v := d.closestVertInternal(a)
	return v
}

// closestVert returns the closest Vertex in the diagram to the given Point
// as well as it's index in our list of points
func (d *Graph) closestVertInternal(a *shapes.Point) (int, *shapes.Point) {
	if len(d.Verts) == 0 {
		return -1, nil
	} else if len(d.Verts) == 1 {
		return 0, d.Verts[0]
	}

	idx := 0
	pt := d.Verts[0]
	dist := a.DistPt(pt)

	for i := 1; i < len(d.Verts); i++ {
		px := d.Verts[i]
		dx := a.DistPt(px)

		if dx < dist {
			dist = dx
			pt = px
			idx = i
		}
	}

	return idx, pt
}

func (d *Graph) CellsWithVert(p *shapes.Point) []*Cell {
	results := []*Cell{}
	for _, cell := range d.Cells {
		for _, e := range cell.Edges {
			if p.Equals(e.Edge.PointA) || p.Equals(e.Edge.PointB) {
				results = append(results, cell)
				break
			}
		}
	}
	return results
}
