package voronoi

import (
	"github.com/voidshard/cartographer/pkg/shapes"
)

type Cell struct {
	ID     int
	Site   *shapes.Point
	Bounds *shapes.Polygon
	Edges  []*HalfEdge
}
