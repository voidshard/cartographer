package geo

import (
	r3 "github.com/golang/geo/r3"
	qh "github.com/markus-wa/quickhull-go"

	"github.com/voidshard/cartographer/pkg/shapes"
)

// ConvexHull returns the outer most points of the given set of points.
func ConvexHull(pts []*shapes.Point) []*shapes.Point {
	pointCloud := []r3.Vector{}
	for _, p := range pts {
		pointCloud = append(pointCloud, r3.Vector{X: p.X, Y: p.Y, Z: 0})
	}

	hull := qh.ConvexHull(pointCloud)

	verts := []*shapes.Point{}
	for _, p := range hull {
		verts = append(verts, shapes.Pt(p.X, p.Y))
	}

	return verts
}
