package landscape

import (
	"image"
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

	// map of average temperature (no wind chill) where values
	// are in degrees c + 100 (so 100 => 0c, 120 => 20c, 60 => -40c)
	temperature *MapImage

	// map of average rainfall .. unsure on exactly what this means ..
	// but higher values imply more regular / more rain
	rainfall *MapImage
}

// Area represents a specific small area of the map
type Area struct {
	// 0-255, where higher is more/higher/better
	Height      uint8
	Rainfall    uint8
	Temperature uint8 // in degress c, offset so 100 => 0 degrees cel

	// if the square contains fresh/salt water
	Sea   bool
	River bool
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
	return &Area{
		Height:      l.height.Value(x, y),
		Rainfall:    l.rainfall.Value(x, y),
		Sea:         l.sea.Value(x, y) == 255,
		River:       l.rivers.Value(x, y) == 255,
		Temperature: l.temperature.Value(x, y),
	}
}

// DebugRender renders out the various maps we have to an
// OS temp dir.
func (w *Landscape) DebugRender() (string, error) {
	d := os.TempDir()

	for _, i := range []struct {
		Name string
		Img  image.Image
	}{
		{"height.png", w.height},
		{"sea.png", w.sea},
		{"rivers.png", w.rivers},
		{"temperature.png", w.temperature},
		{"rainfall.png", w.rainfall},
	} {
		err := savePng(filepath.Join(d, i.Name), i.Img)
		if err != nil {
			return d, err
		}
	}

	return d, nil
}
