package landscape

import (
	"github.com/voidshard/cartographer/pkg/shapes"

	"math/rand"
	"sort"
)

const (
	riverTurnChance = 0.3
)

// determineRivers determines where our rivers will be, we return a new heightmap
// & the map of rivers.
// Rivers are sufficiently complicated that they seem worth their own file ..
func determineRivers(hmap, sea, volc *MapImage, cfg *riverSettings) (*MapImage, []*MapImage, []*POI) {
	x, y := hmap.Dimensions()
	out := NewMapImage(x, y)
	out.SetBackground(0)

	rivermaps := []*MapImage{}
	pois := []*POI{}

	if cfg.Number < 1 {
		return out, rivermaps, pois
	}

	origins := riverOrigins(hmap, cfg) // places where a river might start
	shuffle(origins)

	rivers := 0 // rivers we've accepted

	for _, o := range origins {
		if volc.Value(o.X(), o.Y()) > 120 {
			// don't start rivers in volcanic land
			continue
		}

		// draw in the river, expanding the outline (& respecting other rivers)
		rvr, riverpois, _ := drawRiver(hmap, out, sea, volc, o, cfg)

		pois = append(pois, riverpois...)
		rivermaps = append(rivermaps, rvr)

		rivers++
		if rivers >= int(cfg.Number) {
			break // we have the desired number
		}
	}

	return out, rivermaps, pois
}

// drawRiver determines a path a river will follow.
// There's actually a lot of steps we want to do here, including smoothing the river path & surrounds,
// determining the direction of the river, ensuring it stops if / when it merges with another river etc.
// Rather than go over the river path multiple times (as previously) we're going to attempt to do this
// all at once & save on re-going over the path multiple times.
func drawRiver(hmap, out, sea, volc *MapImage, o *Pixel, cfg *riverSettings) (*MapImage, []*POI, []*Pixel) {
	x, y := hmap.Dimensions()

	pois := []*POI{&POI{X: o.X(), Y: o.Y(), Type: RiverOrigin}}
	rvr := NewMapImage(x, y)
	rvr.SetBackground(0)

	// kick off by decrementing the hight of our river
	riverbed := decrement(hmap.Value(o.X(), o.Y()), 5)
	hmap.SetValue(o.X(), o.Y(), riverbed)

	// meander randomly until we reach the sea or another river
	path := []*Pixel{o}

	// direction is tricky. We want the river to change diretion, but not to twist wholly around,
	// so we'll keep tabs on the currect direction & the starting direction.
	startingdir := shapes.ToHeadingInt(rand.Intn(8))
	if cfg.ForceNorthSouthSections {
		// because we don't want to force such a sharp turn (East / West -> North / South)
		if rand.Intn(2) == 1 {
			startingdir = shapes.NORTH
		} else {
			startingdir = shapes.SOUTH
		}
	}
	prevDir := startingdir.Right()
	if rand.Intn(2) == 1 { // handling for extending rivers travelling diagonally
		prevDir = startingdir.Left()
	}
	dir := startingdir
	missedDecrements := 0

	lastVolc := volc.Value(o.X(), o.Y())

	for {
		this := path[len(path)-1]

		possible := []*shapes.Heading{}
		for _, h := range []shapes.Heading{
			dir,
			dir.Left(),
			dir.Right(),
		} {
			if h.Dist(startingdir) > 2 {
				// turn is too sharp
				possible = append(possible, nil)
				continue
			}

			dx, dy := h.RiseRun()
			v := volc.Value(this.X()+dx, this.Y()+dy)
			if v > lastVolc {
				// reject flowing towards volcanoes
				possible = append(possible, nil)
				continue
			}

			possible = append(possible, &h)
		}

		// change direction left or right
		if possible[0] == nil && possible[1] == nil && possible[2] == nil {
			break // well, no where to go ..
		} else if possible[0] == nil || float64(rand.Intn(int(100*riverTurnChance))) < riverTurnChance {
			if possible[1] == nil && possible[2] != nil {
				prevDir = dir
				dir = *possible[2]
			} else if possible[1] != nil && possible[2] == nil {
				prevDir = dir
				dir = *possible[1]
			} else if possible[1] != nil && possible[2] != nil {
				prevDir = dir
				dir = *possible[rand.Intn(2)+1]
			}
		}

		// figure out the next piece of the river
		dx, dy := dir.RiseRun()
		next := pix(this.X()+dx, this.Y()+dy, 255)
		lastVolc = volc.Value(next.X(), next.Y())
		path = append(path, next)

		// decide new height of riverbed
		h := min(hmap.Nearby(next.X(), next.Y(), 1, true))
		if cfg.ForceNorthSouthSections && (dir == shapes.NORTH || dir == shapes.SOUTH) {
			// if we're forcing n/s and we're going n/s then we'll decrement as
			// far as we need to
			h = decrement(h, 1+uint8(missedDecrements))
			missedDecrements = 0
		} else if cfg.ForceNorthSouthSections && !(dir == shapes.NORTH || dir == shapes.SOUTH) {
			// if we're forcing n/s and we're not going n/s then we'll record that
			// we wanted to drop the riverbed but couldn't
			missedDecrements++
		} else {
			// otherwise, each step of the river we'll decrement by 1
			h = decrement(h, 1)
		}
		if h < riverbed {
			riverbed = h
		}

		// if going diagonally, we'll add a pixel of thickness
		// so our river isn't blocky
		if dir.IsDiagonal() {
			px, py := prevDir.RiseRun()
			rvr.SetValue(this.X()+px, this.Y()+py, 255)
			out.SetValue(this.X()+px, this.Y()+py, 255)
			hmap.SetValue(this.X()+px, this.Y()+py, riverbed)
		}

		// record the river
		rvr.SetValue(next.X(), next.Y(), 255)
		out.SetValue(next.X(), next.Y(), 255)
		hmap.SetValue(next.X(), next.Y(), riverbed)

		// figure out if we're merging with another river
		touchesRiver := pixelsBetween(
			255, 255,
			out.Nearby(next.X(), next.Y(), 1, true),
		)
		nearme := rvr.Nearby(next.X(), next.Y(), 1, true)
		touchesThisRiver := pixelsBetween(1, 255, nearme)
		merge := false
		if len(touchesRiver) > len(touchesThisRiver) {
			merge = true
		}

		// if we are merging, flesh out the river around that so we merge gracefully
		if merge {
			for _, ground := range pixelsBetween(0, 0, nearme) {
				rvr.SetValue(ground.X(), ground.Y(), 255)
				out.SetValue(ground.X(), ground.Y(), 255)
				hmap.SetValue(ground.X(), ground.Y(), riverbed)
			}
			break
		}

		if next.X() < 0 || next.Y() < 0 || next.X() > x || next.Y() > y {
			// we've left the map
			break
		}
		if sea.Value(next.X(), next.Y()) == 255 {
			// we're in the sea
			break
		}
	}

	end := path[len(path)-1]
	pois = append(pois, &POI{X: end.X(), Y: end.Y(), Type: RiverEnd})

	return rvr, pois, path
}

// riverOrigins figures out where rivers can start
func riverOrigins(hmap *MapImage, cfg *riverSettings) []*Pixel {
	return origins(
		hmap,
		cfg.OriginMinDist,
		int(cfg.Number),
		220, // we'd like rivers to start 220-240 height
		240,
		140, // but if we're desperate we'll take down to 140
	)
}

// origins picks places between given heights on the map some dist apart
func origins(hmap *MapImage, minDist float64, number int, omin, omax, minHeight uint8) []*Pixel {
	if minDist < 0 {
		minDist = 0
	}

	// all pixels, sorted by height
	byheight := []*Pixel{}
	eachPixel(hmap, func(dx, dy int, c uint8) {
		byheight = append(
			byheight,
			&Pixel{Point: shapes.Pt(float64(dx), float64(dy)), V: c},
		)
	})
	sort.Slice(
		byheight,
		func(i, j int) bool { return byheight[i].V > byheight[j].V },
	)

	// figure out all places we can start
	var min uint8 = omin
	var max uint8 = omax
	origins := []*Pixel{}
	for {
		candidates := pixelsBetween(min, max, byheight)
		if len(candidates) == 0 { // no where looks good to start a river
			break
		}
		shuffle(candidates)

		for _, origin := range candidates {
			if len(origins) >= number {
				break // we've done all the rivers we were asked to
			}
			tooclose := false
			for _, other := range origins {
				if other.Point.DistPt(origin.Point) < minDist {
					tooclose = true
					break
				}
			}
			if tooclose {
				continue
			}
			origins = append(origins, origin)
		}
		if len(origins) >= number {
			break
		}

		max = min
		min = min - 20
		if min < minHeight {
			break
		}
	}

	return origins
}
