package geo

import (
	"github.com/voidshard/cartographer/pkg/shapes"
	"math"
	"sort"
)

func Rect(a, b *shapes.Point) []*shapes.Point {
	return []*shapes.Point{a, shapes.Pt(b.X, a.Y), b, shapes.Pt(a.X, b.Y)}
}

func PolysOverlap(a, b *shapes.Polygon) bool {
	for i, p := range a.Points {
		if b.Contains(p) {
			return true
		}

		var prev *shapes.Point
		if i == 0 {
			prev = a.Points[len(a.Points)-1]
		} else {
			prev = a.Points[i-1]
		}

		if CrossesPolygon(b, prev, p) || CrossesPolygon(b, p, prev) {
			return true
		}
	}

	for i, p := range b.Points {
		if a.Contains(p) {
			return true
		}

		var prev *shapes.Point
		if i == 0 {
			prev = b.Points[len(b.Points)-1]
		} else {
			prev = b.Points[i-1]
		}

		if CrossesPolygon(a, prev, p) || CrossesPolygon(a, p, prev) {
			return true
		}
	}

	return false
}

func CentrePolygon(poly *shapes.Polygon, pt *shapes.Point) {
	current := poly.Centre()
	poly.Translate(pt.X-current.X, pt.Y-current.Y)
}

// return point on polygon that intersects the line starting at `start` and going towards `direction`
// `nil` indicates that the two do not intersect.
func IntersectionWithPolygon(poly *shapes.Polygon, start, direction *shapes.Point) *shapes.Point {
	desired := angleBetweenPoints(start, direction)

	var left *shapes.Point
	var right *shapes.Point

	// for every pair of points along the polygon, figure out whether our desired angle is between
	// the two points (left, right)
	for i := 0; i < len(poly.Points); i++ {
		var this *shapes.Point
		if i > 0 {
			this = poly.Points[i-1]
		} else {
			this = poly.Points[len(poly.Points)-1]
		}

		next := poly.Points[i]

		thisAngle := angleBetweenPoints(start, this)
		nextAngle := angleBetweenPoints(start, next)

		// TODO think about angles of 0 ..
		if thisAngle == desired {
			return this
		} else if nextAngle == desired {
			return next
		} else if (thisAngle < desired && nextAngle > desired) || (thisAngle > desired && nextAngle < desired) {
			left = this
			right = next
			break
		}
	}

	if left == nil || right == nil {
		return nil // there is no itersection
	}

	dist := left.DistPt(right)

	closestPt := left
	closestAngle := math.Abs(desired - angleBetweenPoints(start, left))

	// assuming we weren't lucky enough to find a suitable point already, we'll start at one point
	// and inch along the line between (left, right) to find the closest angle we can, and thus,
	// the most ideal point ..
	for i := 1.0; i < dist; i++ {
		pt := PointAlongLine(left, right, i)
		angleDist := math.Abs(desired - angleBetweenPoints(start, pt))
		if angleDist < closestAngle {
			closestAngle = angleDist
			closestPt = pt
		}
	}

	return closestPt
}

// Return whether the line from start -> end crosses over the given polygon.
// Note this is more subtle that poly.Contains(start || end) as it catches
//
//  start -----> *---* ---> end
//               |   |
//               *---*
//               poly
//
func CrossesPolygon(poly *shapes.Polygon, start, end *shapes.Point) bool {
	inter := IntersectionWithPolygon(poly, start, end)
	if inter == nil {
		// they dont interest at all so .. no problems
		return false
	}

	distEnd := start.DistPt(end)
	distInt := start.DistPt(inter)

	return distInt < distEnd
}

func AnyPolyContains(polys []*shapes.Polygon, pnt *shapes.Point) bool {
	for _, p := range polys {
		if p.Contains(pnt) {
			return true
		}
	}
	return false
}

func SortPoints(centre *shapes.Point, pts []*shapes.Point) []*shapes.Point {
	// TODO: horribly inefficient but .. works
	v := map[float64]*shapes.Point{}
	keys := []float64{}

	a := centre

	for i := 0; i < len(pts); i++ {
		b := pts[i]
		angle := math.Atan2(a.Y-b.Y, a.X-b.X)

		keys = append(keys, angle)
		v[angle] = b
	}

	sort.Float64s(keys)

	result := []*shapes.Point{}
	for _, key := range keys {
		result = append(result, v[key])
	}

	return result
}

// PointAlongLine return a point `dist` along the line from a towards b.
func PointAlongLine(a, b *shapes.Point, dist float64) *shapes.Point {
	dx, dy := ChangeAlongLine(a, b)
	return shapes.Pt(a.X+(dx*dist), a.Y+(dy*dist))
}

// PointsAlongLine applies PointsAlongLine with dist of +1 increments
// to list out points from a -> b
func PointsAlongLine(a, b *shapes.Point) []*shapes.Point {
	dist := a.DistPt(b)
	if dist <= 1 { // close enough
		return []*shapes.Point{a, b}
	}

	pts := []*shapes.Point{a}
	mv := 1.0
	for {
		next := PointAlongLine(a, b, mv)
		newDist := next.DistPt(b)
		if newDist >= dist || newDist < 1.0 {
			break
		}
		pts = append(pts, next)
		dist = newDist
		mv++
	}
	pts = append(pts, b)

	return pts
}

// changeAlongLine returns dx, dy value between (-1, 1) for x & y directions
func ChangeAlongLine(a, b *shapes.Point) (float64, float64) {
	dx := b.X - a.X
	dy := b.Y - a.Y

	total := math.Abs(dx) + math.Abs(dy)
	dx /= total
	dy /= total

	return dx, dy
}

// Rotate a around b by some degrees counter clockwise.
func Rotate(a, b *shapes.Point, deg float64) *shapes.Point {
	// stackoverflow.com/questions/34372480/rotate-point-about-another-point-in-degrees-python
	radians := deg * (math.Pi / 180.0)

	return shapes.Pt(
		b.X+math.Cos(radians)*(a.X-b.X)-math.Sin(radians)*(a.Y-b.Y),
		b.Y+math.Sin(radians)*(a.X-b.X)+math.Cos(radians)*(a.Y-b.Y),
	).Round()
}

func angleBetweenPoints(a, b *shapes.Point) float64 {
	/*
			Because we've reversed the direction of Y, we get an angle like this
				 stackoverflow.com/questions/59903415/how-do-i-calculate-the-angle-between-two-points

				          -270    -180    -90
					    +      +      +

				        360 +             + 0

					    +      +      + 90
					   270    180
		     We correct to the normal angle ( 0 -> 360 counter clockwise from the +x axis)
	*/
	dx := b.X - a.X
	dy := b.Y - a.Y
	radians := math.Atan2(dy, dx)
	degrees := (radians * 360.0) / math.Pi

	for {
		if degrees > 360.0 {
			degrees -= 360.0
			continue
		}
		if degrees < -360.0 {
			degrees += 360.0
			continue
		}
		break
	}

	if degrees == 0 {
		return 0
	} else if degrees < 0 {
		return (-1 * degrees) / 2
	} else {
		return 180.0 + (360-degrees)/2
	}
}
