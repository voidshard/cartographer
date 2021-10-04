package landscape

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Landscape represents some landmass(es) with associated
// maps. Images are currently greyscale uint8 (0-255) maps.
type Landscape struct {
	// map of height in m from the sea bottom.
	// Very roughly a point of height is around 63m, thus with sea level
	// at 115, 255 indicates (140 * 63) 8820m (Mt Everest)
	height *MapImage

	// map of sea/not sea where 255 => sea, 0 => not
	sea *MapImage

	// map of river (or lake) vs not where 255 => fresh water, 0 => not
	rivers *MapImage

	// maps where each draws out a single river, using 255 for river
	// and 1-254 for lake pixels
	rivermaps []*MapImage

	// map of average temperature (no wind chill) where values
	// are in degrees c + 100 (so 100 => 0c, 120 => 20c, 60 => -40c)
	temperature *MapImage

	// map of average rainfall .. unsure on exactly what this means ..
	// but higher values imply more regular / more rain
	rainfall *MapImage

	// map of swampy terrain. Technically this would be areas of low drainage
	// but high water
	swamp *MapImage

	// areas of geothermal activity
	volcanic *MapImage

	// interesting points on the map
	pointsOfInterest []*POI
}

// PointsOfInterest returns `POI` or `Points of Interest` - these
// are denoted via (X,Y) co-ords (in pixels) and a `PointType`
func (l *Landscape) PointsOfInterest() []*POI {
	return l.pointsOfInterest
}

// Dimensions returns the width & height of each map in pixels.
func (l *Landscape) Dimensions() (int, int) {
	return l.height.Dimensions()
}

// At returns the state of the land at a given place on the map.
// Nil is returned if an out-of-bounds area is requested.
func (l *Landscape) At(x, y int) *Area {
	maxx, maxy := l.Dimensions()
	if x < 0 || y < 0 || x >= maxx || y >= maxy {
		return nil
	}

	a := &Area{
		Height:      l.height.Value(x, y),
		Rainfall:    l.rainfall.Value(x, y),
		Sea:         l.sea.Value(x, y) == 255,
		River:       l.rivers.Value(x, y) == 255,
		Temperature: l.temperature.Value(x, y),
		Swamp:       l.swamp.Value(x, y) == 255,
		Volcanisim:  l.volcanic.Value(x, y),
	}
	if !a.River {
		return a
	}

	// figure out which river
	for i, m := range l.rivermaps {
		// note that;
		// 0 -> not a river
		// 1-254 -> lake id (on river)
		// 255 -> is a river
		rv := m.Value(x, y)
		if rv == 0 {
			// not a river
			continue
		}
		a.RiverID = i + 1 // reserve 0 as 'not a river id'
		if rv != 255 {
			// use first lake
			a.LakeID = int(rv)
			a.Lake = true
			break
		}
	}

	return a
}

// DebugRender renders out the various maps we have to an
// OS temp dir.
func (w *Landscape) DebugRender() (string, error) {
	d := os.TempDir()

	// write out main maps
	for _, i := range []struct {
		Name string
		Img  image.Image
	}{
		{"height.png", w.height},
		{"sea.png", w.sea},
		{"rivers.png", w.rivers},
		{"temperature.png", w.temperature},
		{"rainfall.png", w.rainfall},
		{"volcanism.png", w.volcanic},
		{"swamp.png", w.swamp},
	} {
		err := savePng(filepath.Join(d, i.Name), i.Img)
		if err != nil {
			return d, err
		}
	}

	// write out river maps
	for i := range w.rivermaps {
		err := savePng(filepath.Join(d, fmt.Sprintf("river.%d.png", i)), w.rivermaps[i])
		if err != nil {
			return d, err
		}
	}

	data, err := json.Marshal(w.pointsOfInterest)
	if err != nil {
		return d, err
	}
	err = ioutil.WriteFile(filepath.Join(d, "pois.json"), data, 0644)

	return d, err
}
