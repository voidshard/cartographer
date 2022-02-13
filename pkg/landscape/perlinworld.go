package landscape

import (
	"sync"

	perlin "github.com/voidshard/cartographer/pkg/perlin"
)

// PerlinLandscape generates our maps from simple perlin noise & some basic math / combinations
func PerlinLandscape(cfg *Config) (*Landscape, error) {
	hmap := combine(
		weight(perlin.Perlin(int(cfg.Width), int(cfg.Height), cfg.Land.HeightVariance), 70),
		weight(perlin.Perlin(int(cfg.Width), int(cfg.Height), cfg.Land.MountainVariance), 30),
	)

	// modifies heightmap
	volc, pois := determineGeothermal(hmap, cfg.Sea.SeaLevel, cfg.Volcanic)

	// modifies heightmap
	sea := determineSea(hmap, cfg.Sea)

	// modifies heightmap
	// sadly, in order to run rivers to the sea, we have to know where the sea is
	// we also want to avoid running through lava
	rvrs, rivermaps, rpois := determineRivers(hmap, sea, volc, cfg.Rivers, cfg.Lakes)
	pois = append(pois, rpois...)

	wg := sync.WaitGroup{}
	wg.Add(5)

	plock := &sync.Mutex{}
	var temp *MapImage
	var rain *MapImage
	var swmp *MapImage

	go func() { // locate mountains
		defer wg.Done()
		mountains := findMountains(hmap)
		plock.Lock()
		defer plock.Unlock()
		pois = append(pois, mountains...)
	}()
	go func() {
		defer wg.Done()

		// we'll look at putting swamps at the ends of rivers
		plock.Lock()
		ends := []*POI{}
		for _, p := range pois {
			if p.Type == RiverEnd && int(cfg.Swamp.MaxHeight) > int(hmap.Value(p.X, p.Y)) {
				ends = append(ends, p)
			}
		}
		plock.Unlock()

		spois := []*POI{}
		swmp, spois = determineSwamp(hmap, rvrs, sea, cfg.Swamp, ends)

		plock.Lock()
		defer plock.Unlock()
		pois = append(pois, spois...)
	}()
	go func() {
		defer wg.Done()
		temp = determineTemp(hmap, volc, cfg.Sea.SeaLevel, cfg.Temp)
	}()
	go func() {
		defer wg.Done()
		rain = determineRainfall(hmap, cfg.Rain)
	}()
	wg.Wait()

	l := &Landscape{
		height:           hmap,
		sea:              sea,
		rivers:           rvrs,
		rivermaps:        rivermaps,
		temperature:      temp,
		rainfall:         rain,
		pointsOfInterest: pois,
		volcanic:         volc,
		swamp:            swmp,
	}

	// finally, using everything else, bucket areas into biomes
	l.determineBiomes(cfg)

	return l, nil
}
