package landscape

import (
	perlin "github.com/voidshard/cartographer/pkg/perlin"
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

func determineGeothermal(hmap *MapImage, sealevel uint8, vs *volcSettings) (*MapImage, []*POI) {
	x, y := hmap.Dimensions()

	// new blank map
	vmap := NewMapImage(x, y)
	vmap.SetBackground(0)
	pois := []*POI{}

	// pick some places to place volcanoes
	origins := geothermalOrigins(hmap, vs)
	if len(origins) == 0 {
		return vmap, pois
	}

	pmap := &MapImage{im: perlin.Perlin(x, y, vs.Variance)}

	for _, volcano := range origins {
		pois = append(pois, &POI{X: volcano.X(), Y: volcano.Y(), Type: Volcano})

		pv := pmap.Value(volcano.X(), volcano.Y())

		vrMax := increment(pv, vs.VolcanicRedius)
		vrMin := decrement(pv, vs.VolcanicRedius)
		lvMax := increment(pv, vs.LavaRadius)
		lvMin := decrement(pv, vs.LavaRadius)

		seen := map[int]bool{}
		check := []*Pixel{volcano}

		for {
			if len(check) == 0 {
				break
			}

			me := check[len(check)-1]

			dist := volcano.Point.DistPt(me.Point)

			if me.V > lvMax || me.V < lvMin || dist >= vs.MaxRadius/2 { // VOLCANIC
				hv := hmap.Value(me.X(), me.Y())
				if hv < sealevel {
					// raise up land above sealevel
					notSealevel := 255 - sealevel
					hv = uint8(float64(hv)*float64(notSealevel)/255) + sealevel
				}
				hmap.SetValue(me.X(), me.Y(), hv)
				vmap.SetValue(me.X(), me.Y(), 120)
			} else { // LAVA
				hmap.SetValue(me.X(), me.Y(), decrement(hmap.Value(me.X(), me.Y()), 5))
				vmap.SetValue(me.X(), me.Y(), 255)
			}

			check = check[:len(check)-1] // slice off the last element
			for _, next := range pmap.Cardinals(me.X(), me.Y(), 1) {
				if next.V > vrMax || next.V < vrMin {
					vmap.SetValue(next.X(), next.Y(), 50)
					continue
				}

				idx := int(next.X())*y + int(next.Y())
				done, _ := seen[idx]
				if done { // we've done this so ignore
					continue
				}

				dist := volcano.Point.DistPt(next.Point)
				if dist > vs.MaxRadius { // too far from volc centre
					continue
				}

				seen[idx] = true
				check = append(check, next)
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
