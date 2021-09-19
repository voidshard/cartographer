/* Here we deal with cutting rectangles indicated by two points in to
polygons of various configurations.
*/
package geo

import (
	"github.com/voidshard/cartographer/pkg/shapes"
)

func Indent(size float64, a, b *shapes.Point) ([]*shapes.Point, []*shapes.Point) {
	/*
	   a----+
	   |    |
	   |    |
	   |    |
	   +----b

	   a----+
	   |+--+|
	   ||  ||
	   |+--+|
	   +----b
	*/
	out := []*shapes.Point{
		a,
		shapes.Pt(b.X, a.Y),
		b,
		shapes.Pt(a.X, b.Y),
	}
	return out, []*shapes.Point{
		shapes.Pt(out[0].X+size, out[0].Y+size),
		shapes.Pt(out[1].X-size, out[1].Y+size),
		shapes.Pt(out[2].X-size, out[2].Y-size),
		shapes.Pt(out[3].X+size, out[3].Y-size),
	}
}

func SplitNorthSouth(a, b *shapes.Point) ([]*shapes.Point, []*shapes.Point) {
	/*
	   a-----+
	   |     |
	   |     |
	   |     |
	   +-----b

	   a-----+
	   |     |
	   +-----+
	   |     |
	   +-----b
	*/
	halfy := (b.Y - a.Y) / 2
	return []*shapes.Point{
			a,
			shapes.Pt(b.X, a.Y),
			shapes.Pt(b.X, a.Y+halfy),
			shapes.Pt(a.X, a.Y+halfy),
		}, []*shapes.Point{
			shapes.Pt(a.X, a.Y+halfy),
			shapes.Pt(b.X, a.Y+halfy),
			b,
			shapes.Pt(a.X, b.Y),
		}
}

func SplitEastWest(a, b *shapes.Point) ([]*shapes.Point, []*shapes.Point) {
	/*
	   a-----+
	   |     |
	   |     |
	   |     |
	   +-----b

	   a--+--+
	   |  |  |
	   |  |  |
	   |  |  |
	   +--+--b
	*/
	halfx := (b.X - a.X) / 2
	return []*shapes.Point{
			a,
			shapes.Pt(a.X+halfx, a.Y),
			shapes.Pt(a.X+halfx, b.Y),
			shapes.Pt(a.X, b.Y),
		}, []*shapes.Point{
			shapes.Pt(a.X+halfx, a.Y),
			shapes.Pt(a.X+halfx, b.Y),
			b,
			shapes.Pt(b.X, a.Y),
		}
}

func CornerTL(size float64, a, b *shapes.Point) ([]*shapes.Point, []*shapes.Point) {
	/*
	   a------+
	   |      |
	   |      |
	   |      |
	   +------b

	   a--0---1
	   |  |   |
	   6--7   2
	   |      |
	   5--4---b

	*/

	tri := []*shapes.Point{
		shapes.Pt(a.X+size, a.Y),
		shapes.Pt(b.X, a.Y),
		shapes.Pt(b.X, a.Y+size), // * bonus
		b,
		shapes.Pt(a.X+size, b.Y), // * bonus
		shapes.Pt(a.X, b.Y),
		shapes.Pt(a.X, a.Y+size),
		shapes.Pt(a.X+size, a.Y+size),
	}
	return []*shapes.Point{a, tri[0], tri[7], tri[6]}, tri
}

func CornerTR(size float64, a, b *shapes.Point) ([]*shapes.Point, []*shapes.Point) {
	/*
	   a------+
	   |      |
	   |      |
	   |      |
	   +------b

	   a--1---+
	   |  |   |
	   7  2---3
	   |      |
	   6--5---b

	*/
	tri := []*shapes.Point{
		a,
		shapes.Pt(b.X-size, a.Y),
		shapes.Pt(b.X-size, a.Y+size),
		shapes.Pt(b.X, a.Y+size),
		b,
		shapes.Pt(b.X-size, b.Y), // * bonus
		shapes.Pt(a.X, b.Y),
		shapes.Pt(a.X, a.Y+size), // * bonus
	}
	return []*shapes.Point{tri[1], shapes.Pt(b.X, a.Y), tri[3], tri[2]}, tri
}

func CornerBL(size float64, a, b *shapes.Point) ([]*shapes.Point, []*shapes.Point) {
	/*
	   a------+
	   |      |
	   |      |
	   |      |
	   +------b

	   a--1---2
	   |      |
	   7--6   3
	   |  |   |
	   +--5---b

	*/
	tri := []*shapes.Point{
		a,
		shapes.Pt(a.X+size, a.Y), // * bonus
		shapes.Pt(b.X, a.Y),
		shapes.Pt(b.X, b.Y-size), // * bonus
		b,
		shapes.Pt(a.X+size, b.Y),
		shapes.Pt(a.X+size, b.Y-size),
		shapes.Pt(a.X, b.Y-size),
	}
	return []*shapes.Point{tri[7], tri[6], tri[5], shapes.Pt(a.X, b.Y)}, tri
}

func CornerBR(size float64, a, b *shapes.Point) ([]*shapes.Point, []*shapes.Point) {
	/*
	   a------+
	   |      |
	   |      |
	   |      |
	   +------b

	   a--1---2
	   |      |
	   7  4---3
	   |  |   |
	   6--5---b

	*/
	tri := []*shapes.Point{
		a,
		shapes.Pt(b.X-size, a.Y), // *bonus
		shapes.Pt(b.X, a.Y),
		shapes.Pt(b.X, b.Y-size),
		shapes.Pt(b.X-size, b.Y-size),
		shapes.Pt(b.X-size, b.Y),
		shapes.Pt(a.X, b.Y),
		shapes.Pt(a.X, b.Y-size), // *bonus
	}
	return []*shapes.Point{tri[4], tri[3], b, tri[5]}, tri
}
