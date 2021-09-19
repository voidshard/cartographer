package geo

import (
	"github.com/voidshard/cartographer/pkg/shapes"
	"math"
)

// More stack overflow credit

// QuadraticBezier return a list of points describing a quadratic that passes
// through the given points.
func QuadraticBezier(p0, p1, p2 *shapes.Point) []*shapes.Point {
	l := (math.Hypot(p1.X-p0.X, p1.Y-p0.Y) + math.Hypot(p2.X-p1.X, p2.Y-p1.Y))
	n := int(l + 0.5)
	if n < 4 {
		n = 4
	}
	d := float64(n) - 1
	result := make([]*shapes.Point, n)
	for i := 0; i < n; i++ {
		t := float64(i) / d
		x, y := quadratic(p0.X, p0.Y, p1.X, p1.Y, p2.X, p2.Y, t)
		result[i] = shapes.Pt(x, y)
	}
	return result
}

// CubicBezier return a list of points describing a cubic that passes
// through the given points.
func CubicBezier(p0, p1, p2, p3 *shapes.Point) []*shapes.Point {
	l := (math.Hypot(p1.X-p0.X, p1.Y-p0.Y) +
		math.Hypot(p2.X-p1.X, p2.Y-p1.Y) +
		math.Hypot(p3.X-p2.X, p3.Y-p2.Y))
	n := int(l + 0.5)
	if n < 4 {
		n = 4
	}
	d := float64(n) - 1
	result := make([]*shapes.Point, n)
	for i := 0; i < n; i++ {
		t := float64(i) / d
		x, y := cubic(p0.X, p0.Y, p1.X, p1.Y, p2.X, p2.Y, p3.X, p3.Y, t)
		result[i] = shapes.Pt(x, y)
	}
	return result
}

func cubic(x0, y0, x1, y1, x2, y2, x3, y3, t float64) (x, y float64) {
	u := 1 - t
	a := u * u * u
	b := 3 * u * u * t
	c := 3 * u * t * t
	d := t * t * t
	x = a*x0 + b*x1 + c*x2 + d*x3
	y = a*y0 + b*y1 + c*y2 + d*y3
	return
}

func quadratic(x0, y0, x1, y1, x2, y2, t float64) (x, y float64) {
	u := 1 - t
	a := u * u
	b := 2 * u * t
	c := t * t
	x = a*x0 + b*x1 + c*x2
	y = a*y0 + b*y1 + c*y2
	return
}
