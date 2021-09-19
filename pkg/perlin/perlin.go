package gotile

import (
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"
)

// From stack overflow I believe, can't recall the original

const PI = 3.1415926535
const ITTERATIONS = 2

type vec2 struct {
	X, Y float32
}

type noise2DContext struct {
	rgradients   []vec2
	permutations []int
	gradients    [4]vec2
	origins      [4]vec2
}

func newNoise2DContext(seed int) *noise2DContext {
	rnd := rand.New(rand.NewSource(int64(seed)))

	n2d := new(noise2DContext)
	n2d.rgradients = make([]vec2, 256)
	n2d.permutations = rand.Perm(256)
	for i := range n2d.rgradients {
		n2d.rgradients[i] = random_gradient(rnd)
	}

	return n2d
}

// Perlin generates a single perlin noise map of size (fx,fy).
// The image is greyscale with colours 0-255 (0->black 255->white).
// The scale indicates how 'zoomed in' you wish the map to be with higher values
// being increasingly chaotic. Scale here is intended to be positive only, and we use
// it's absolute value.
func Perlin(fx, fy int, scale float64) *image.RGBA {
	x, y := sanitize(fx, fy, scale)

	noise := generate2DNoise(0, x, 0, y, ITTERATIONS, int(time.Now().UnixNano()))
	im := image.NewRGBA(image.Rect(0, 0, x, y))

	var max float32 = 0
	var min float32 = 1
	for dx := 0; dx < x; dx++ {
		for dy := 0; dy < y; dy++ {
			n := noise[(dy*x)+dx]
			if n > max {
				max = n
			}
			if n < min {
				min = n
			}
		}
	}

	for dx := 0; dx < x; dx++ {
		for dy := 0; dy < y; dy++ {
			n := (noise[(dy*x)+dx] - min) * (1 / (max - min))
			cv := uint8(n * 255)
			im.Set(dx, dy, color.RGBA{cv, cv, cv, 255})
		}
	}

	if fx != x || fy != y {
		return (resizeImage(float64(fx), float64(fy), im)).(*image.RGBA)
	}

	return im
}

func generate2DNoise(x, w, y, h, itterations, seed int) []float32 {
	dx := w - x
	dy := h - y

	n2d := newNoise2DContext(seed)
	pixels := make([]float32, w*h)

	for i := itterations; i > 0; i-- {
		for xi := 0; xi < dx-x; xi++ {
			for yi := 0; yi < dy-y; yi++ {
				v := n2d.Get(float32(x)+float32(xi)*0.1, float32(y)+float32(yi)*0.1)
				v = v*0.5 + 0.5
				pixels[yi*w+xi] = v
			}
		}
	}

	return pixels
}

func lerp(a, b, v float32) float32 {
	return a*(1-v) + b*v
}

func smooth(v float32) float32 {
	return v * v * (3 - 2*v)
}

func random_gradient(r *rand.Rand) vec2 {
	v := r.Float64() * PI * 2
	return vec2{
		float32(math.Cos(v)),
		float32(math.Sin(v)),
	}
}

func gradient(orig, grad, p vec2) float32 {
	sp := vec2{p.X - orig.X, p.Y - orig.Y}
	return grad.X*sp.X + grad.Y*sp.Y
}

func (n2d *noise2DContext) get_gradient(x, y int) vec2 {
	idx := n2d.permutations[x&255] + n2d.permutations[y&255]
	return n2d.rgradients[idx&255]
}

func (n2d *noise2DContext) get_gradients(x, y float32) {
	x0f := math.Floor(float64(x))
	y0f := math.Floor(float64(y))
	x0 := int(x0f)
	y0 := int(y0f)
	x1 := x0 + 1
	y1 := y0 + 1

	n2d.gradients[0] = n2d.get_gradient(x0, y0)
	n2d.gradients[1] = n2d.get_gradient(x1, y0)
	n2d.gradients[2] = n2d.get_gradient(x0, y1)
	n2d.gradients[3] = n2d.get_gradient(x1, y1)

	n2d.origins[0] = vec2{float32(x0f + 0.0), float32(y0f + 0.0)}
	n2d.origins[1] = vec2{float32(x0f + 1.0), float32(y0f + 0.0)}
	n2d.origins[2] = vec2{float32(x0f + 0.0), float32(y0f + 1.0)}
	n2d.origins[3] = vec2{float32(x0f + 1.0), float32(y0f + 1.0)}
}

func (n2d *noise2DContext) Get(x, y float32) float32 {
	p := vec2{x, y}
	n2d.get_gradients(x, y)
	v0 := gradient(n2d.origins[0], n2d.gradients[0], p)
	v1 := gradient(n2d.origins[1], n2d.gradients[1], p)
	v2 := gradient(n2d.origins[2], n2d.gradients[2], p)
	v3 := gradient(n2d.origins[3], n2d.gradients[3], p)
	fx := smooth(x - n2d.origins[0].X)
	vx0 := lerp(v0, v1, fx)
	vx1 := lerp(v2, v3, fx)
	fy := smooth(y - n2d.origins[0].Y)
	return lerp(vx0, vx1, fy)
}

// sanitize fixes weird input values & determines size of perlin map
func sanitize(x, y int, scale float64) (int, int) {
	if scale < 0 {
		scale = -1 * scale
	} else if scale == 0 {
		scale = 1
	}
	x = int(float64(x) * scale)
	y = int(float64(y) * scale)
	return x, y
}

// resizeImage to some desired size
func resizeImage(x, y float64, img image.Image) image.Image {
	return resize.Resize(
		uint(x),
		uint(y),
		img,
		resize.Lanczos3,
	)
}
