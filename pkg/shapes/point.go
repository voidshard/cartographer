package shapes

import (
	"math"
)

type Point struct {
	X float64
	Y float64
}

func Pt(x, y float64) *Point {
	return &Point{X: x, Y: y}
}

func (p *Point) Equals(q *Point) bool {
	return p.X == q.X && p.Y == q.Y
}

func (p *Point) Round() *Point {
	p.X = math.Round(p.X)
	p.Y = math.Round(p.Y)
	return p
}

func (p *Point) DistPt(q *Point) float64 {
	return p.Dist(q.X, q.Y)
}

func (p *Point) Dist(x, y float64) float64 {
	return math.Sqrt(math.Pow(p.X-x, 2) + math.Pow(p.Y-y, 2))
}
