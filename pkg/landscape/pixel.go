package landscape

import (
	"github.com/voidshard/cartographer/pkg/shapes"
	"math/rand"
)

// Pixel is a wrapper struct for the value at some location
type Pixel struct {
	Point *shapes.Point
	V     uint8
}

func (p *Pixel) X() int {
	return int(p.Point.X)
}

func (p *Pixel) Y() int {
	return int(p.Point.Y)
}

func pix(x, y int, v uint8) *Pixel {
	return &Pixel{
		Point: shapes.Pt(float64(x), float64(y)),
		V:     v,
	}
}

func pixelsBetween(gte, lte uint8, in []*Pixel) []*Pixel {
	result := []*Pixel{}
	for _, p := range in {
		if p.V >= gte && p.V <= lte {
			result = append(result, p)
		}
	}
	return result
}

func shuffle(in []*Pixel) {
	rand.Shuffle(len(in), func(i, j int) {
		in[i], in[j] = in[j], in[i]
	})
}

// any runs func on all inputs & returns if any return true
func any(n []*Pixel, fn func(*Pixel) bool) bool {
	for _, v := range n {
		if fn(v) {
			return true
		}
	}
	return false
}

// eachPixel runs some func for each pixel in the image
func eachPixel(im *MapImage, op func(dx, dy int, c uint8)) {
	x, y := im.Dimensions()
	for dx := 0; dx < x; dx++ {
		for dy := 0; dy < y; dy++ {
			op(dx, dy, im.Value(dx, dy))
		}
	}
}

//
func min(in []*Pixel) uint8 {
	if len(in) == 0 {
		return 0
	}
	low := in[0].V
	if len(in) == 1 {
		return low
	}
	for _, p := range in[1:] {
		if p.V < low {
			low = p.V
		}
	}
	return low
}
