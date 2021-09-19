package voronoi

import (
	"fmt"
	"math"

	"github.com/RyanCarrier/dijkstra"

	"github.com/voidshard/cartographer/pkg/shapes"
)

// Weight defines some weight for path "length" between two points
type Weight func(x, y *shapes.Point) int64

// DistanceWeight is a weight for path x->y by how far apart they are
func DistanceWeight(x, y *shapes.Point) int64 {
	return int64(math.Round(x.DistPt(y)))
}

// NoWeight makes all paths the same weight value
func NoWeight(_, _ *shapes.Point) int64 {
	return 0
}

// LongestPath returns the longest path between two points using
// Dijkstra's algorithm (along graph verts).
func (d *Graph) LongestPath(x, y *shapes.Point, w Weight) ([]*shapes.Point, error) {
	xi, _ := d.closestVertInternal(x)
	yi, _ := d.closestVertInternal(y)

	ph := d.newDijkstra(w)
	best, err := ph.Longest(xi, yi)
	if err != nil {
		return nil, err
	}

	pts := []*shapes.Point{}
	for _, i := range best.Path {
		pts = append(pts, d.Verts[i])
	}

	return pts, nil
}

// ShortestPath returns the shortest path between two points using
// Dijkstra's algorithm (along graph verts).
func (d *Graph) ShortestPath(x, y *shapes.Point, w Weight) ([]*shapes.Point, error) {
	xi, _ := d.closestVertInternal(x)
	yi, _ := d.closestVertInternal(y)

	ph := d.newDijkstra(w)
	best, err := ph.Shortest(xi, yi)
	if err != nil {
		return nil, fmt.Errorf("%w between verts %d->%d", err, xi, yi)
	}

	pts := []*shapes.Point{}
	for _, i := range best.Path {
		pts = append(pts, d.Verts[i])
	}

	return pts, nil
}

func (d *Graph) newDijkstra(w Weight) *dijkstra.Graph {
	if w == nil {
		w = DistanceWeight
	}

	ph := dijkstra.NewGraph()

	for i := range d.Verts {
		ph.AddVertex(i)
	}

	for _, e := range d.Edges {
		ph.AddArc(e.PointAID, e.PointBID, w(e.PointA, e.PointB))
		ph.AddArc(e.PointBID, e.PointAID, w(e.PointB, e.PointA))
	}

	return ph
}
