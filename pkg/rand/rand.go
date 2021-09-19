package rand

import (
	"github.com/voidshard/cartographer/pkg/shapes"
	"github.com/voidshard/cartographer/pkg/voronoi"
	"log"
	"math/rand"
	"time"
)

var (
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Int wraps math/rand Rand.Intn
func Int(a int) int {
	return rng.Intn(a)
}

//
func Voronoi(sites int, mindist float64, poly *shapes.Polygon) *voronoi.Graph {
	for {
		pnts := PointsMinDist(sites, mindist, poly)

		// our voronoi lib under some odd circumstances panics when
		// handling certain configurations of points if we managed to
		// form such a case then we'll try again with a new
		// set of points (this should be rare)
		gra, err := voronoi.Compute(pnts, poly)
		if err != nil {
			log.Printf("failed to create voronoi diagram %v\n", err)
			continue
		}

		return gra
	}
}

// PointsMinDist returns at most `sites` points at least `mindist` apart
// within `poly`
func PointsMinDist(sites int, mindist float64, poly *shapes.Polygon) []*shapes.Point {
	x0, y0, x1, y1 := poly.Bounds()

	pts := []*shapes.Point{}

	for i := 0; i < sites; i++ {
		pt := shapes.Pt(rng.Float64()*(x1-x0)+x0, rng.Float64()*(y1-y0)+y0)
		if !poly.Contains(pt) {
			continue
		}

		if len(pts) == 0 {
			pts = append(pts, pt)
		}

		tooclose := false
		for _, other := range pts {
			if other.DistPt(pt) < mindist {
				tooclose = true
				break
			}
		}

		if !tooclose {
			pts = append(pts, pt)
		}
	}

	return pts
}

// Points returns at most `sites` within `poly` indented by approximately
// `indent` (that is, points will not be placed around the edges)
func Points(sites int, indent int, poly *shapes.Polygon) []*shapes.Point {
	hbuff := float64(indent / 2)

	x0, y0, x1, y1 := poly.Bounds()

	pts := []*shapes.Point{}
	for i := 0; i < sites; i++ {
		pt := shapes.Pt(
			rng.Float64()*(x1-x0-float64(indent))+x0+hbuff,
			rng.Float64()*(y1-y0-float64(indent))+y0+hbuff,
		)
		if poly.Contains(pt) {
			pts = append(pts, pt)
		}
	}

	return pts
}
