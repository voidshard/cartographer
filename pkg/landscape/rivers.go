package landscape

import (
	"github.com/voidshard/cartographer/pkg/geo"
	"github.com/voidshard/cartographer/pkg/shapes"

	"math"
	"math/rand"
	"sort"
)

const (
	riverstep = 10
)

// determineRivers determines where our rivers will be, we return a new heightmap
// & the map of rivers.
// Rivers are sufficiently complicated that they seem worth their own file ..
func determineRivers(hmap *MapImage, sealevel uint8, cfg *riverSettings) *MapImage {
	x, y := hmap.Dimensions()
	out := NewMapImage(x, y)
	out.SetBackground(0)

	if cfg.Number < 1 {
		return out
	}

	origins := riverOrigins(hmap, cfg) // places where a river might start
	shuffle(origins)

	rivers := 0 // rivers we've accepted

	for _, o := range origins {
		// rough outline of where the river might go (downhill until we find sea)
		outline := roughPath(hmap, sealevel, o)
		if len(outline) < 2 {
			continue // too short
		}

		// amass outline of points into river -> lake -> river -> river -> lake etc
		path := expandLakes(outline)

		// forcibly adjust
		if cfg.ForceNorthSouthSections {
			path = ensureNSFlow(path)
		}

		// draw in the river, expanding the outline
		rvr := drawRiver(hmap, out, path)

		// ensure waterfalls are north-south facing
		if cfg.ForceNorthSouthSections {
			ensureNSFall(hmap, out, rvr, path)
		}

		rivers++
		if rivers >= int(cfg.Number) {
			break // we have the desired number
		}
	}

	return out
}

// ensureNSFall smooths a river so that falls in river height are
// in the N/S orientation.
func ensureNSFall(hmap, out, rvr *MapImage, path [][]*Pixel) {
	if len(path) < 3 {
		return
	}

	end := path[len(path)-1][0]
	riverbed := hmap.Value(end.X(), end.Y())

	for i := len(path) - 1; i > 0; i-- {
		// walk river backwards
		// nb. this means that "from" should be downhill from "next"
		this := path[i]
		next := path[i-1]
		from := this[0]
		to := next[0]

		if len(next) > 1 && len(this) == 1 { // we're exiting a lake
			sort.Slice(
				next,
				func(i, j int) bool {
					return next[i].Point.DistPt(from.Point) < next[j].Point.DistPt(from.Point)
				},
			)
			to = next[0]
		}

		fromV := hmap.Value(from.X(), from.Y())
		if fromV < riverbed {
			riverbed = fromV
		}
		toV := hmap.Value(to.X(), to.Y())

		// ok fine, we need to lower the height
		if len(next) > 1 && len(this) == 1 { // we're exiting a lake
			merge := false
			lake := allPointsInPolygon(hmap, next)

			for _, p := range lake {
				hmap.SetValue(p.X(), p.Y(), riverbed)
			}

			if merge {
				return
			}
		}

		pts := geo.PointsAlongLine(to.Point.Round(), from.Point.Round())
		hmap.SetValue(to.X(), to.Y(), riverbed)
		hmap.SetValue(from.X(), from.Y(), riverbed)
		for _, p := range pts {
			hmap.SetValue(int(p.X), int(p.Y), riverbed)
		}

		if from.X() == to.X() && from.Y() != to.Y() {
			riverbed = toV
			hmap.SetValue(to.X(), to.Y(), riverbed)
		}
	}
}

// ensureNSFlow essentially we do some river path fixing to make sure
// it inclues areas where it flows north-south / south-north.
func ensureNSFlow(path [][]*Pixel) [][]*Pixel {
	valid := 0
	candidates := []int{}

	// count sections where we go n-s already
	// amass sections where we could insert a n-s step
	for i := 1; i < len(path); i++ {
		prev := path[i-1]
		me := path[i]

		if len(prev) > 1 || len(me) > 1 {
			continue
		}

		if prev[0].Point.X == me[0].Point.X && prev[0].Point.Y != me[0].Point.Y {
			valid += 1
			continue
		}

		candidates = append(candidates, i)
	}
	if len(candidates) == 0 {
		// no where to add section sorry
		return path
	}

	target := (len(path) / 5) + rand.Intn(3) - valid
	if target <= 0 {
		// path seems legit
		return path
	}
	if len(candidates) > target {
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
	}

	// figure out where we should insert paths
	insert := map[int]*Pixel{}
	for _, cindex := range candidates {
		prev := path[cindex-1][0]
		next := path[cindex][0]
		this := pix(prev.X(), (prev.Y()+next.Y())/2, prev.V)
		insert[cindex] = this
	}

	// and now we can insert
	fixed := [][]*Pixel{}
	for index, current := range path {
		add, ok := insert[index]
		if ok {
			fixed = append(fixed, []*Pixel{add})
		}
		fixed = append(fixed, current)
	}

	return fixed
}

func drawRiver(hmap, out *MapImage, path [][]*Pixel) *MapImage {
	x, y := hmap.Dimensions()
	rvr := NewMapImage(x, y)

	riverbed := decrement(hmap.Value(path[0][0].X(), path[0][0].Y()), 2)

	for i := 1; i < len(path); i++ {
		this := path[i-1]
		next := path[i]
		from := this[0]
		to := next[0]

		if len(this) > 1 && len(next) == 1 {
			// moving from lake into a river again
			// unlike in the other cases, we need to work out the exit lake pixel
			sort.Slice(
				this,
				func(i, j int) bool { return this[i].Point.DistPt(to.Point) < this[j].Point.DistPt(to.Point) },
			)
			from = this[0]

			merge := false
			lake := allPointsInPolygon(hmap, this)
			lowest := decrement(min(lake), 1)
			if lowest < riverbed {
				riverbed = lowest
			}

			for _, p := range lake {
				isRiver := out.Value(p.X(), p.Y()) == 255
				isThisRiver := rvr.Value(p.X(), p.Y()) == 255
				if isRiver && !isThisRiver {
					// we've merged with another river
					merge = true
				}

				rvr.SetValue(p.X(), p.Y(), 255)
				out.SetValue(p.X(), p.Y(), 255)

				// TODO: improve to have uneven lake bed
				hmap.SetValue(p.X(), p.Y(), riverbed)
			}

			if merge {
				// we don't need an exit river from the lake (since our lake hit another river)
				return rvr
			}
		}

		rvr.SetValue(from.X(), from.Y(), 255)
		out.SetValue(from.X(), from.Y(), 255)
		hmap.SetValue(from.X(), from.Y(), riverbed)
		pts := geo.PointsAlongLine(from.Point.Round(), to.Point.Round())
		for _, p := range pts {
			isRiver := out.Value(int(p.X), int(p.Y)) == 255
			isThisRiver := rvr.Value(int(p.X), int(p.Y)) == 255
			if isRiver && !isThisRiver {
				return rvr // we've merged with another river
			}

			// record the river
			rvr.SetValue(int(p.X), int(p.Y), 255)
			out.SetValue(int(p.X), int(p.Y), 255)

			h := decrement(hmap.Value(int(p.X), int(p.Y)), 1)
			if h < riverbed {
				riverbed = h
			}
			hmap.SetValue(int(p.X), int(p.Y), riverbed)
		}

		// fill river point backwards just to be sure our points line up
		for i, p := range geo.PointsAlongLine(to.Point.Round(), from.Point.Round()) {
			if i > 1 {
				break
			}

			// record the river
			rvr.SetValue(int(p.X), int(p.Y), 255)
			out.SetValue(int(p.X), int(p.Y), 255)
		}

		rvr.SetValue(to.X(), to.Y(), 255)
		out.SetValue(to.X(), to.Y(), 255)
		hmap.SetValue(to.X(), to.Y(), riverbed)
	}

	return rvr
}

func allPointsInPolygon(hmap *MapImage, in []*Pixel) []*Pixel {
	pts := []*Pixel{}
	for _, p := range in {
		pts = append(pts, hmap.Nearby(p.X(), p.Y(), riverstep-1, true)...)
	}
	return pts
}

// expandLakes makes wide sections of river pixels into 'lake' sections.
// Essentially if a river pixel is within `riverstep` of 3+ other river pixels then
// it is considered a lake. We amass nearby lake pixels together to form
// a large lake.
func expandLakes(path []*Pixel) [][]*Pixel {
	diag := math.Sqrt(math.Pow(float64(riverstep), 2) + math.Pow(float64(riverstep), 2))

	// map of all points nearby a given point
	neighbours := map[int][]int{}
	for i, p := range path {
		near := []int{}

		for j, other := range path {
			if i == j {
				continue
			}
			if p.Point.DistPt(other.Point) <= diag {
				near = append(near, j)
			}
		}

		neighbours[i] = near
	}

	res := [][]*Pixel{}
	seen := map[int]bool{} // pixels we're 100% done with
	for i, p := range path {
		done := seen[i]
		if done {
			continue
		}

		near := neighbours[i]
		if len(near) < 3 {
			// this is a standard river tile
			seen[i] = true
			res = append(res, []*Pixel{p})
			continue
		}

		// ok, it's a lake tile, we have to figure out all pixels in the lake ..
		checked := map[int]bool{i: true}  // marks if we've seen this neighbour in this loop
		adjacent := map[int]bool{i: true} // a neighbour we consider part of the lake
		stack := near
		for {
			if len(stack) < 1 {
				break
			}

			j := stack[0]

			done = seen[j]
			passed := checked[j]

			if !(done || passed) {
				near = neighbours[j]
				if len(near) > 2 {
					adjacent[j] = true
					stack = append(stack, near...)
				}
			}

			checked[j] = true

			// swap with end, shave off
			stack[0] = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		}

		lake := []*Pixel{p}
		seen[i] = true
		for j := range adjacent {
			seen[j] = true
			lake = append(lake, path[j])
		}
		res = append(res, lake)
	}

	return res
}

// roughPath returns a roughly 'downward' path to the sea from the origin
func roughPath(hmap *MapImage, sealevel uint8, o *Pixel) []*Pixel {
	path := []*Pixel{o}

	x, y := hmap.Dimensions()
	rvr := NewMapImage(x, y)
	rvr.SetValue(o.X(), o.Y(), 255)

	for {
		this := path[len(path)-1]
		if this.V < sealevel {
			break // river has reached the sea
		}

		crds := hmap.Cardinals(this.X(), this.Y(), riverstep)
		var next *Pixel
		for _, c := range crds {
			if rvr.Value(c.X(), c.Y()) == 255 {
				continue
			}

			if next == nil {
				next = c
				continue
			}
			if c.V >= next.V {
				continue
			}
			next = c
		}
		if next == nil {
			break
		}

		rvr.SetValue(next.X(), next.Y(), 255)
		path = append(path, next)
	}

	return path
}

// riverOrigins figures out where rivers can start
func riverOrigins(hmap *MapImage, cfg *riverSettings) []*Pixel {
	if cfg.OriginMinDist < 0 {
		cfg.OriginMinDist = 0
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

	// figure out all places we can start rivers
	var min uint8 = 220 // fairly high up
	var max uint8 = 240 // but not on a mountain peak ..
	origins := []*Pixel{}
	for {
		candidates := pixelsBetween(min, max, byheight)
		if len(candidates) == 0 { // no where looks good to start a river
			break
		}
		shuffle(candidates)

		for _, origin := range candidates {
			if len(origins) >= int(cfg.Number) {
				break // we've done all the rivers we were asked to
			}
			tooclose := false
			for _, other := range origins {
				if other.Point.DistPt(origin.Point) < cfg.OriginMinDist {
					tooclose = true
					break
				}
			}
			if tooclose {
				continue
			}
			origins = append(origins, origin)
		}
		if len(origins) >= int(cfg.Number) {
			break
		}

		max = min
		min = min - 20
		if min < 140 { // we're too low to realistically start a river
			break
		}
	}

	return origins
}
