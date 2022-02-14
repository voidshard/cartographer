package landscape

import (
	perlin "github.com/voidshard/cartographer/pkg/perlin"
	"github.com/voidshard/cartographer/pkg/shapes"

	"math/rand"
	"sort"
)

// determineRivers determines where our rivers will be, we return a new heightmap
// & the map of rivers.
// Rivers are sufficiently complicated that they seem worth their own file ..
func determineRivers(hmap, sea, volc *MapImage, cfg *riverSettings, ls *lakeSettings) (*MapImage, []*MapImage, *MapImage, []*POI) {
	x, y := hmap.Dimensions()
	out := NewMapImage(x, y)
	out.SetBackground(0)

	rain := NewMapImage(x, y)
	rain.SetBackground(0)

	rivermaps := []*MapImage{}
	pois := []*POI{}

	if cfg.Number < 1 {
		return out, rivermaps, rain, pois
	}

	origins := riverOrigins(hmap, cfg) // places where a river might start
	shuffle(origins)

	rivers := 0 // rivers we've accepted
	lakes := 0  // lakes we've added

	minLakeRiverLen := int(ls.MinDistFromStart) + int(ls.MinDistFromEnd)

	for _, o := range origins {
		if volc.Value(o.X(), o.Y()) > 120 {
			// don't start rivers in volcanic land
			continue
		}

		// draw in the river, expanding the outline (& respecting other rivers)
		rvr, riverpois, rpath := drawRiver(hmap, out, rain, sea, volc, o, cfg)
		if len(rpath) > minLakeRiverLen && lakes < int(ls.Number) {
			// ie. as we get more lakes, new lakes become less likely
			if rand.Intn(int(ls.Number)) > lakes {
				continue
			}

			// we pick a random part of the river that is not too close
			// to the end nor the start
			idx := rand.Intn(len(rpath)-minLakeRiverLen) + int(ls.MinDistFromStart)

			size := fillLake(hmap, sea, out, rvr, volc, rpath[idx], ls)
			if size > 10 {
				// if they're too small we don't count them as lakes ..
				lakes++
				pois = append(pois, &POI{X: rpath[idx].X(), Y: rpath[idx].Y(), Type: LakeOrigin})
			}
		}

		pois = append(pois, riverpois...)
		rivermaps = append(rivermaps, rvr)

		rivers++
		if rivers >= int(cfg.Number) {
			break // we have the desired number
		}
	}

	return out, rivermaps, rain, pois
}

// fillLake draws in a lake given it's origin point (on some river).
// We're allowed to touch pixels adjacent to our own river (expanding it)
// but we can't join other rivers (because we'd then have a lake with
// more than one exit river .. which is really weird).
func fillLake(hmap, sea, rvrs, rvr, volc *MapImage, o *Pixel, ls *lakeSettings) int {
	x, y := hmap.Dimensions()

	pmap := &MapImage{im: perlin.Perlin(x, y, ls.Variance)}
	pv := pmap.Value(o.X(), o.Y())

	pmax := increment(pv, ls.Radius)
	pmin := decrement(pv, ls.Radius)

	seen := map[int]bool{}
	check := []*Pixel{o}
	lakeBed := hmap.Value(o.X(), o.Y())
	startVolc := volc.Value(o.X(), o.Y())

	size := 0

	for {
		if len(check) == 0 {
			break
		}

		me := check[len(check)-1]
		check = check[:len(check)-1] // slice off the last element

		rvr.SetValue(me.X(), me.Y(), 255)
		rvrs.SetValue(me.X(), me.Y(), 255)
		size++

		currentH := hmap.Value(me.X(), me.Y())
		if currentH > lakeBed {
			newh := decrement(currentH, uint8(rand.Intn(3)))
			hmap.SetValue(me.X(), me.Y(), newh)
		} else if currentH < lakeBed {
			lakeBed = currentH
		}

		candidates := []*Pixel{}
		for _, next := range pmap.Nearby(me.X(), me.Y(), 2, false) {
			// check / record we've been here
			idx := int(next.X())*y + int(next.Y())
			done, _ := seen[idx]
			if done {
				continue
			}
			seen[idx] = true

			// we have two 'reject' cases; where the whole set of pixels
			// are too close to something invalid (volcano, sea, another river),
			// and where the tile itself is invalid (dist, perlin mismatch).
			// If we reject a pixel we'll keep looking at further pixels.
			// If we 'greater' reject then this whole set will be thrown out.
			if sea.Value(next.X(), next.Y()) == 255 {
				candidates = []*Pixel{}
				break
			}
			if volc.Value(next.X(), next.Y()) > startVolc {
				candidates = []*Pixel{}
				break
			}
			if rvrs.Value(next.X(), next.Y()) == 255 && rvr.Value(next.X(), next.Y()) != 255 {
				// we're in a river that is *not* us
				candidates = []*Pixel{}
				break
			}

			// rather than hard stopping at the radius, we'll allow it to phase out
			dist := o.Point.DistPt(next.Point)
			if dist > ls.HardMaxRadius {
				continue
			}
			effectPv := next.V
			if dist > ls.SoftMaxRadius {
				effectPv = increment(effectPv, uint8(2*dist-ls.SoftMaxRadius))
			}
			if effectPv > pmax || effectPv < pmin {
				continue // perlin map says no
			}

			candidates = append(candidates, next)
		}

		check = append(check, candidates...)
	}

	return size
}

type candidate struct {
	H      shapes.Heading
	Reason string
}

func (c *candidate) Ok() bool {
	return c.Reason == ""
}

// drawRiver determines a path a river will follow.
// There's actually a lot of steps we want to do here, including smoothing the river path & surrounds,
// determining the direction of the river, ensuring it stops if / when it merges with another river etc.
// Rather than go over the river path multiple times (as previously) we're going to attempt to do this
// all at once & save on re-going over the path multiple times.
func drawRiver(hmap, out, rain, sea, volc *MapImage, o *Pixel, cfg *riverSettings) (*MapImage, []*POI, []*Pixel) {
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

	startVolc := volc.Value(o.X(), o.Y())

	for {
		this := path[len(path)-1]

		possible := []*candidate{
			&candidate{dir, ""},
			&candidate{dir.Left(), ""},
			&candidate{dir.Right(), ""},
		}

		for _, c := range possible {
			if c.H.Dist(startingdir) > 2 {
				// turn is too sharp
				c.Reason = "sharp"
				continue
			}

			dx, dy := c.H.RiseRun()
			v := volc.Value(this.X()+dx, this.Y()+dy)
			if v > startVolc {
				// reject flowing towards volcanoes
				c.Reason = "volcanic"
				continue
			}
		}

		// change direction left or right
		if !possible[0].Ok() && !possible[1].Ok() && !possible[2].Ok() {
			break // well, no where to go ..
		} else if !possible[0].Ok() || rand.Float64() < cfg.TurnChance {
			if !possible[1].Ok() && possible[2].Ok() {
				prevDir = dir
				dir = possible[2].H
			} else if possible[1].Ok() && !possible[2].Ok() {
				prevDir = dir
				dir = possible[1].H
			} else if possible[1].Ok() && possible[2].Ok() {
				prevDir = dir
				dir = possible[rand.Intn(2)+1].H
			} else if possible[0].Ok() {
				prevDir = dir
				dir = possible[0].H
			}
		}

		// figure out the next piece of the river
		dx, dy := dir.RiseRun()
		next := pix(this.X()+dx, this.Y()+dy, 255)
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

		for _, near := range rvr.Nearby(next.X(), next.Y(), 10, true) {
			// up rain / fresh water map since .. there's fresh water
			rain.SetValue(near.X(), near.Y(), 50)
		}

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
