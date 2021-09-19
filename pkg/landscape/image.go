package landscape

import (
	"image"
	"image/color"
)

type MapImage struct {
	im *image.RGBA
}

func NewMapImage(x, y int) *MapImage {
	return &MapImage{im: image.NewRGBA(image.Rect(0, 0, x, y))}
}

func (m *MapImage) SetBackground(v uint8) {
	eachPixel(m, func(dx, dy int, _ uint8) {
		m.SetValue(dx, dy, v)
	})
}

func (m *MapImage) Pixel(x, y int) *Pixel {
	return pix(x, y, m.Value(x, y))
}

func (m *MapImage) ColorModel() color.Model {
	return m.im.ColorModel()
}

func (m *MapImage) Bounds() image.Rectangle {
	return m.im.Bounds()
}

func (m *MapImage) At(x, y int) color.Color {
	return m.im.At(x, y)
}

func (m *MapImage) SetValue(x, y int, v uint8) {
	m.im.Set(x, y, color.RGBA{v, v, v, 255})
}

func (m *MapImage) Value(x, y int) uint8 {
	r, _, _, _ := m.At(x, y).RGBA()
	return uint8(r)
}

func (m *MapImage) Dimensions() (int, int) {
	rect := m.im.Bounds()
	return rect.Max.X - rect.Min.X, rect.Max.Y - rect.Min.Y
}

// Cardinals returns points in the cardinal directions of `radius` distance
// away. Nb, this is similar to Nearby but returns far fewer points
// (at most 8).
// We never return points off the map.
func (m *MapImage) Cardinals(dx, dy, radius int) []*Pixel {
	ns := []*Pixel{}
	if radius < 1 {
		return ns
	}

	x, y := m.Dimensions()
	for _, iy := range []int{dy - radius, dy, dy + radius} {
		for _, ix := range []int{dx - radius, dx, dx + radius} {
			if ix == dx && iy == dy {
				continue
			}
			if iy < 0 || iy >= y || ix < 0 || ix >= x {
				continue // off the map
			}

			ns = append(ns, pix(ix, iy, m.Value(ix, iy)))
		}
	}

	return ns
}

// Nearby returns all pixels nearby dx,dy within some radius.
// If inclusive is set then the point at dx,dy is returned too.
// We never return points off the map.
func (m *MapImage) Nearby(dx, dy, radius int, inclusive bool) []*Pixel {
	ns := []*Pixel{}

	if radius < 1 {
		if inclusive {
			ns = append(ns, pix(dx, dy, m.Value(dx, dy)))
		}

		return ns
	}

	x, y := m.Dimensions()
	for iy := dy - radius; iy <= dy+radius; iy++ {
		for ix := dx - radius; ix <= dx+radius; ix++ {
			if ix == dx && iy == dy && !inclusive {
				continue // is excluded from consideration
			}
			if iy < 0 || iy >= y || ix < 0 || ix >= x {
				continue // off the map
			}

			ns = append(ns, pix(ix, iy, m.Value(ix, iy)))
		}
	}

	return ns
}
