package voronoi

import (
	"github.com/voidshard/cartographer/pkg/shapes"
)

type HalfEdge struct {
	Edge  *Edge
	Angle float64
}

type Edge struct {
	ID int

	LeftCell  *Cell
	RightCell *Cell

	PointAID int
	PointBID int
	PointA   *shapes.Point
	PointB   *shapes.Point
}
