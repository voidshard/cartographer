package landscape

import ()

// mutateImage runs over an image and returns a new image
func mutateImage(im *MapImage, op func(int, int, uint8) uint8) *MapImage {
	x, y := im.Dimensions()
	out := NewMapImage(x, y)
	eachPixel(im, func(dx, dy int, c uint8) {
		out.SetValue(dx, dy, op(dx, dy, c))
	})
	return out
}

// between returns an image with white (255) pixels marked where values are
// between the given values (gte >=, lte <=)
func between(im *MapImage, gte, lte uint8) *MapImage {
	return mutateImage(im, func(dx, dy int, in uint8) uint8 {
		if in >= gte && in <= lte {
			return 255
		}
		return 0
	})
}
