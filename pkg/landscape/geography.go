package landscape

import (
	perlin "github.com/voidshard/cartographer/pkg/perlin"

	"math"
)

// findMountains discovers all mountains (over some height).
// Nb. these might also be volcanoes
func findMountains(hmap *MapImage) []*POI {
	mountains := []*POI{}
	eachPixel(hmap, func(dx, dy int, c uint8) {
		if c >= 240 {
			mountains = append(mountains, &POI{X: dx, Y: dy, Type: Mountain})
		}
	})
	return mountains
}

// determineSwamp picks a few areas and figures out where might make sense to have swamps.
// Swamps are areas within some height bounds radiating out from the end section(s) of a river.
// Technically we can have swamp areas in other places .. but this is more straightforward to
// reason about.
// Nb. we probably should figure out areas based on 'drainage' eg. above sea level, near water,
// reasonably flat / bowl shaped ..
// Currently areas are "swamp" or not
func determineSwamp(hmap, rivers, sea *MapImage, ss *swampSettings, riverends []*POI) (*MapImage, []*POI) {
	if ss.Radius < 1 {
		ss.Radius = 1
	}
	x, y := hmap.Dimensions()

	// new blank map
	smap := NewMapImage(x, y)
	smap.SetBackground(0)
	pois := []*POI{}

	if len(riverends) == 0 {
		return smap, pois
	}

	per := &MapImage{im: perlin.Perlin(x, y, ss.Variance)}

	for i, start := range riverends {
		if i > int(ss.Number) {
			break
		}
		oh := int(hmap.Value(start.X, start.Y))
		pois = append(pois, &POI{X: start.X, Y: start.Y, Type: Swamp})

		for dy := start.Y - int(ss.Radius)/2; dy < start.Y+int(ss.Radius)/2; dy++ {
			for dx := start.X - int(ss.Radius)/2; dx < start.X+int(ss.Radius)/2; dx++ {
				h := int(hmap.Value(dx, dy))
				if h > int(ss.MaxHeight) {
					// swamps must be below max height
					continue
				}
				if h >= oh+int(ss.DeltaHeight) || h < oh-int(ss.DeltaHeight) {
					// swamp areas must be within some height of each other
					// (reasonably flat)
					continue
				}

				if sea.Value(dx, dy) != 0 || rivers.Value(dx, dy) != 0 {
					// skip sea/river
					continue
				}

				pv := per.Value(dx, dy)
				if pv%5 == 0 {
					// swamp water
					smap.SetValue(dx, dy, 255)
				} else {
					// swampy land
					smap.SetValue(dx, dy, 120)
				}
			}
		}

	}

	return smap, pois
}

// determineGeothermal picks some areas to be the centres of volcanoes & describes
// a region around them as geothermally active.
// Areas are indicated on a scale from 0-255 as to how geothermal they are
// (whatever that means).
func determineGeothermal(hmap *MapImage, vs *volcSettings) (*MapImage, []*POI) {
	if vs.Radius < 1 {
		vs.Radius = 1
	}

	x, y := hmap.Dimensions()

	// new blank map
	vmap := NewMapImage(x, y)
	vmap.SetBackground(0)
	pois := []*POI{}

	// find where we can put our areas
	origins := geothermalOrigins(hmap, vs)
	if len(origins) == 0 {
		return vmap, pois
	}

	pmap := &MapImage{im: perlin.Perlin(x, y, vs.Variance)}

	for _, volcano := range origins {
		pois = append(pois, &POI{X: volcano.X(), Y: volcano.Y(), Type: Volcano})

		lavamap := &MapImage{im: perlin.Perlin(x, y, vs.Variance*1.5)}

		// describe square around origin
		for dy := volcano.Y() - int(vs.Radius)/2; dy < volcano.Y()+int(vs.Radius)/2; dy++ {
			for dx := volcano.X() - int(vs.Radius)/2; dx < volcano.X()+int(vs.Radius)/2; dx++ {
				current := vmap.Value(dx, dy)

				pv := pmap.Value(dx, dy)
				if lavamap.Value(dx, dy) > vs.LavaOver {
					pv = 255
					hmap.SetValue(dx, dy, decrement(hmap.Value(dx, dy), 5))
				}

				dist := math.Sqrt(
					math.Pow(float64(volcano.X())-float64(dx), 2) + math.Pow(float64(volcano.Y())-float64(dy), 2),
				)

				// drastically cut geothermal activity as we radiate outwards
				if dist > vs.Radius {
					continue
				} else if dist > vs.Radius/2 {
					pv /= 6
				} else if dist > vs.Radius/3 && pv != 255 {
					pv /= 4
				} else if dist > vs.Radius/4 && pv != 255 {
					pv /= 2
				}

				// allow for volcanic areas to overlap
				if pv > current {
					vmap.SetValue(dx, dy, pv)
				}
			}
		}

	}

	return vmap, pois
}

// geothermalOrigins figures out where we can put volcanoes.
// Note that we actually could put these at any height .. even if it ended
// up at sealevel it could simply be a caldera with no volcanic cone.
// Even beneath the sea wouldn't be strange
func geothermalOrigins(hmap *MapImage, cfg *volcSettings) []*Pixel {
	return origins(
		hmap,
		cfg.OriginMinDist,
		int(cfg.Number),
		220, // we'll try to get volcanoes in the mountains
		255,
		90, // but actually pretty much anywhere is ok
	)
}

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
