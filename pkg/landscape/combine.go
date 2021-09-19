package landscape

import (
	"image"
)

// imageCombiner is a simple util for adding images together with some
// weights.
type imageCombiner struct {
	ix int
	iy int

	images  []image.Image
	weights []float64
}

func weight(im image.Image, w float64) func(*imageCombiner) {
	return func(c *imageCombiner) {
		rect := im.Bounds()
		dx := rect.Max.X - rect.Min.X
		dy := rect.Max.Y - rect.Min.Y
		if dx > c.ix {
			c.ix = dx
		}
		if dy > c.iy {
			c.iy = dy
		}

		c.images = append(c.images, im)
		c.weights = append(c.weights, w)
	}
}

func combine(in ...func(*imageCombiner)) *MapImage {
	ic := &imageCombiner{
		images:  []image.Image{},
		weights: []float64{},
	}

	for _, dw := range in {
		dw(ic)
	}

	// we'll use this to normalise
	sumOfWeights := 0.0
	for _, w := range ic.weights {
		sumOfWeights += w
	}

	// and now we can build the final image
	final := NewMapImage(ic.ix, ic.iy)

	for dx := 0; dx < ic.ix; dx++ {
		for dy := 0; dy < ic.iy; dy++ {
			value := 0.0
			for i, im := range ic.images {
				c, _, _, _ := im.At(dx, dy).RGBA()
				value += (float64(c/255) * (ic.weights[i] / sumOfWeights))
			}
			final.SetValue(dx, dy, toUint8(value))
		}
	}

	return final
}
