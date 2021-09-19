package shapes

import (
	"math"
)

// Stolen from https://github.com/kellydunn/golang-geo/blob/master/polygon.go

// A Polygon is carved out of a 2D plane by a set of (possibly disjoint) contours.
// It can thus contain holes, and can be self-intersecting.
type Polygon struct {
	Points []*Point
}

// NewPolygon: Creates and returns a new pointer to a Polygon
// composed of the passed in Points.  Points are
// considered to be in order such that the last point
// forms an edge with the first point.
func NewPolygon(Points []*Point) *Polygon {
	return &Polygon{Points: Points}
}

// Round all points of polygon in place
func (p *Polygon) Round() {
	for i, pnt := range p.Points {
		p.Points[i] = pnt.Round()
	}
}

// Scale returns a new polygon equal to the old polygon but scaled by some factor
func (p *Polygon) Scale(scale float64) *Polygon {
	points := make([]*Point, len(p.Points))
	for i, pnt := range p.Points {
		points[i] = Pt(pnt.X*scale, pnt.Y*scale)
	}
	return NewPolygon(points)
}

func (p *Polygon) Centre() *Point {
	x := 0.0
	y := 0.0

	for _, pnt := range p.Points {
		x += pnt.X
		y += pnt.Y
	}

	return Pt(x/float64(len(p.Points)), y/float64(len(p.Points)))
}

// Translate the polygon by x,y
func (p *Polygon) Translate(dx, dy float64) {
	for _, pnt := range p.Points {
		pnt.X += dx
		pnt.Y += dy
	}
}

// Floor keeps the same shape but moves all points as close to (0, 0) as possible.
// That is we "translate" the polygon to be as close to the origin as possible.
func (p *Polygon) Floor() (float64, float64) {
	if len(p.Points) == 0 {
		return 0, 0
	}

	topmost := p.Points[0]
	leftmost := p.Points[0]
	for i := 1; i < len(p.Points); i++ {
		pnt := p.Points[i]

		if pnt.Y < topmost.Y {
			topmost = pnt
		}
		if pnt.X < leftmost.X {
			leftmost = pnt
		}
	}

	defer p.Translate(-1*leftmost.X, -1*topmost.Y)
	return -1 * leftmost.X, -1 * topmost.Y
}

// BoundsArea returns an area that is equal to or greater than the area of Bounds.
// Note that this is the area of a rectangle containing the polygon and is only
// a rough approximation of the polygons actual area.
func (p *Polygon) BoundsArea() float64 {
	x0, y0, x1, y1 := p.Bounds()
	return (x1 - x0) * (y1 - y0)
}

// Bounds returns the highest & lowest x & y values from the Points in this polygon.
func (p *Polygon) Bounds() (x0, y0, x1, y1 float64) {
	if len(p.Points) == 0 {
		return
	}

	first := p.Points[0]

	x0 = first.X
	y0 = first.Y
	x1 = first.X
	y1 = first.Y

	for i := 1; i < len(p.Points); i++ {
		p := p.Points[i]
		if p.X < x0 {
			x0 = p.X
		} else if p.X > x1 {
			x1 = p.X
		}

		if p.Y < y0 {
			y0 = p.Y
		} else if p.Y > y1 {
			y1 = p.Y
		}
	}

	return
}

// Add: Appends the passed in contour to the current Polygon.
func (p *Polygon) Add(point *Point) {
	p.Points = append(p.Points, point)
}

// IsClosed returns whether or not the polygon is closed.
// TODO:  This can obviously be improved, but for now,
//        this should be sufficient for detecting if Points
//        are contained using the raycast algorithm.
func (p *Polygon) IsClosed() bool {
	if len(p.Points) < 3 {
		return false
	}
	return true
}

func (p *Polygon) ContainsAny(points ...*Point) bool {
	for _, pnt := range points {
		if p.Contains(pnt) {
			return true
		}
	}
	return false
}

func (p *Polygon) ContainsAll(points ...*Point) bool {
	for _, pnt := range points {
		if !p.Contains(pnt) {
			return false
		}
	}
	return true
}

// Contains returns whether or not the current Polygon contains the passed in Point.
func (p *Polygon) Contains(point *Point) bool {
	if !p.IsClosed() {
		return false
	}

	start := len(p.Points) - 1
	end := 0

	contains := p.intersectsWithRaycast(point, p.Points[start], p.Points[end])

	for i := 1; i < len(p.Points); i++ {
		if p.intersectsWithRaycast(point, p.Points[i-1], p.Points[i]) {
			contains = !contains
		}
	}

	return contains
}

// Using the raycast algorithm, this returns whether or not the passed in point
// Intersects with the edge drawn by the passed in start and end Points.
// Original implementation: http://rosettacode.org/wiki/Ray-casting_algorithm#Go
func (p *Polygon) intersectsWithRaycast(point *Point, start *Point, end *Point) bool {
	// Always ensure that the the first point
	// has a y coordinate that is less than the second point
	if start.Y > end.Y {

		// Switch the Points if otherwise.
		start, end = end, start

	}

	// Move the point's y coordinate
	// outside of the bounds of the testing region
	// so we can start drawing a ray
	for point.Y == start.Y || point.Y == end.Y {
		newLng := math.Nextafter(point.Y, math.Inf(1))
		point = Pt(point.X, newLng)
	}

	// If we are outside of the polygon, indicate so.
	if point.Y < start.Y || point.Y > end.Y {
		return false
	}

	if start.X > end.X {
		if point.X > start.X {
			return false
		}
		if point.X < end.X {
			return true
		}

	} else {
		if point.X > end.X {
			return false
		}
		if point.X < start.X {
			return true
		}
	}

	raySlope := (point.Y - start.Y) / (point.X - start.X)
	diagSlope := (end.Y - start.Y) / (end.X - start.X)

	return raySlope >= diagSlope
}
