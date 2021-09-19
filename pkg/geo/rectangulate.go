package geo

import (
	"fmt"
	"github.com/fogleman/triangulate"
	"github.com/voidshard/cartographer/pkg/shapes"
	"math"
)

// Rectangulate breaks a polygon into as many squares as possible.
// Nb. this only uses existing points
func Rectangulate(poly *shapes.Polygon) []*shapes.Polygon {
	if len(poly.Points) < 4 {
		return []*shapes.Polygon{}
	}

	// find triangles
	result := triangulate.Polygon{Exterior: toRing(poly)}.Triangulate()

	// turn []triangulate.Triangle result into []*shapes.Polygon
	tris := []*shapes.Polygon{}
	for _, t := range result {
		pnts := []*shapes.Point{}
		for _, p := range []triangulate.Point{t.A, t.B, t.C} {
			pnts = append(pnts, shapes.Pt(p.X, p.Y))
		}
		tris = append(tris, shapes.NewPolygon(pnts))
	}

	return toRects(tris)
}

func toRing(poly *shapes.Polygon) triangulate.Ring {
	ring := []triangulate.Point{}
	for _, p := range poly.Points {
		ring = append(ring, triangulate.Point{X: p.X, Y: p.Y})
	}
	return ring
}

func toPoly(pnts []*shapes.Point) *shapes.Polygon {
	x0, y0, x1, y1 := shapes.NewPolygon(pnts).Bounds()
	centre := shapes.Pt((x0+x1)/2, (y0+y1)/2)
	return shapes.NewPolygon(SortPoints(centre, pnts))
}

func toRects(tris []*shapes.Polygon) []*shapes.Polygon {
	// ok we have triangles .. now to figure out which pairs make rectangles
	matched := map[int]bool{}
	polys := []*shapes.Polygon{}

	for i, a := range tris {
		if matched[i] {
			continue
		}
		for j, b := range tris {
			if i == j {
				continue
			}
			if matched[j] {
				continue
			}

			uniq, xcounts, ycounts := uniquePoints(a, b)

			if !(len(uniq) == 4 && len(xcounts) == 2 && len(ycounts) == 2) {
				// essentially "if not a rectangle"
				continue
			}

			matched[i] = true
			matched[j] = true
			polys = append(polys, toPoly(uniq))
			break
		}
	}

	return polys
}

func uniquePoints(a, b *shapes.Polygon) ([]*shapes.Point, map[int]int, map[int]int) {
	seen := map[string]*shapes.Point{}

	xvalues := map[int]int{}
	yvalues := map[int]int{}

	for _, p := range a.Points {
		x := int(math.Round(p.X))
		y := int(math.Round(p.Y))
		key := fmt.Sprintf("%d-%d", x, y)

		seen[key] = p
		xvalues[x] = xvalues[x] + 1
		yvalues[y] = yvalues[y] + 1
	}
	for _, p := range b.Points {
		x := int(math.Round(p.X))
		y := int(math.Round(p.Y))
		key := fmt.Sprintf("%d-%d", x, y)

		seen[key] = p
		xvalues[x] = xvalues[x] + 1
		yvalues[y] = yvalues[y] + 1
	}

	ret := []*shapes.Point{}
	for _, v := range seen {
		ret = append(ret, v)
	}

	return ret, xvalues, yvalues
}
