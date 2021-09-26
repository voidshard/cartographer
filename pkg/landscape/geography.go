package landscape

import (
	perlin "github.com/voidshard/cartographer/pkg/perlin"
)

// determineRainfall returns rainfall 0-255
// TODO; include rain shadowing, consider prevailing winds
func determineRainfall(hmap *MapImage, rs *rainfallSettings) *MapImage {
	x, y := hmap.Dimensions()
	return &MapImage{im: perlin.Perlin(x, y, rs.RainfallVariance)}
}

// determineTemp returns a map of average temperatures in degrees Celcius,
// where 100 => 0 degress (putting our min at -100c and max 155c .. seems
// wide enough for an earth like planet)
//
// Assumptions for our rough calculations
// treelinebackpacker.com/2013/05/06/calculate-temperature-change-with-elevation/
// assuming humidty and .. a lot of factors .. we guesstimate 1c per 100 metres
//
// Mt Everest (at uint 255) is ~8800m above sea level.
//
// Deepest point in Java Trench assumed ~7200m (uint 0)
//
// Thus every point in our heightmap is ~63 metres.
//
// This means we should lose 1c in temp from sealevel as we climb every 2 pts
// of height. Well, more like 3c per 5 points but .. whatever.
func determineTemp(hm *MapImage, sealevel uint8, cfg *tempSettings) *MapImage {
	_, y := hm.Dimensions()
	equator := y / 2

	// how wide the equator 'band' is
	band := cfg.EquatorWidth * float64(y)

	// difference in temp per pixel as we move out from the equator
	dty := float64(cfg.EquatorAverageTemp-cfg.PoleAverageTemp) / ((float64(y) - band) / 2)

	return mutateImage(hm, func(dx, dy int, z uint8) uint8 {
		temp := cfg.EquatorAverageTemp

		dueToY := 0.0
		if dy > equator {
			dueToY = float64(dy - equator)
		} else {
			dueToY = float64(equator - dy)
		}
		if dueToY <= band {
			// we're inside the equator
			return temp
		}

		return decrement(temp, toUint8(dueToY*dty))
	})
}

// determineSea returns all areas that should be regarded as sea.
// - Any pixel on the map edge beneath sea level is sea
// - Any pixel adjacent to a sea pixel that is below sea level is also sea.
// nb; this meas we can have areas of lowlands below sea level that are
// not sea -- this is intentional & actually the case in some parts of
// the world.
func determineSea(hm *MapImage, cfg *seaSettings) *MapImage {
	x, y := hm.Dimensions()
	level := cfg.SeaLevel
	sea := NewMapImage(x, y)
	sea.SetBackground(0)

	// stack of pixels to expand sea from
	todo := []*Pixel{}

	// go around the edges & find initial sea tiles
	for dx := 0; dx < x; dx++ {
		if hm.Value(dx, 0) <= level {
			sea.SetValue(dx, 0, 255)
			todo = append(todo, hm.Pixel(dx, 0))
		}
		if hm.Value(dx, y-1) <= level {
			sea.SetValue(dx, y-1, 255)
			todo = append(todo, hm.Pixel(dx, y-1))
		}
	}
	for dy := 0; dy < y; dy++ {
		if hm.Value(0, dy) <= level {
			sea.SetValue(0, dy, 255)
			todo = append(todo, hm.Pixel(0, dy))
		}
		if hm.Value(x-1, dy) <= level {
			sea.SetValue(x-1, dy, 255)
			todo = append(todo, hm.Pixel(x-1, dy))
		}
	}

	// expand sea tiles into neighbouring tiles
	for {
		if len(todo) == 0 {
			break
		}

		p := todo[0]
		for _, n := range hm.Nearby(p.X(), p.Y(), 1, false) {
			if n.V > level {
				continue
			}

			if sea.Value(n.X(), n.Y()) == 255 { // it's sea already
				continue
			}

			sea.SetValue(n.X(), n.Y(), 255) // set as sea
			todo = append(todo, n)
		}

		// switch last into the first place, carve off the last
		todo[0] = todo[len(todo)-1]
		todo = todo[:len(todo)-1]
	}

	return sea
}
