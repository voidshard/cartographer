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

	wg := sync.WaitGroup{}

	var sea *MapImage
	var rvrs *MapImage
	var rivermaps []*MapImage

	plock := &sync.Mutex{}
	pois := []*POI{}

	wg.Add(2)
	go func() {
		// modifies heightmap
		defer wg.Done()
		sea = determineSea(hmap, cfg.Sea)
	}()
	go func() {
		defer wg.Done()
		rpois := []*POI{}
		rvrs, rivermaps, rpois = determineRivers(hmap, cfg.Sea.SeaLevel, cfg.Rivers)
		plock.Lock()
		defer plock.Unlock()
		pois = append(pois, rpois...)
	}()

	wg.Wait()
	wg.Add(5)

	var temp *MapImage
	var rain *MapImage
	var volc *MapImage
	var swmp *MapImage

	go func() {
		defer wg.Done()
		mountains := findMountains(hmap)
		plock.Lock()
		defer plock.Unlock()
		pois = append(pois, mountains...)
	}()
	go func() {
		defer wg.Done()
		vpois := []*POI{}
		volc, vpois = determineGeothermal(hmap, cfg.Volcanic)
		plock.Lock()
		defer plock.Unlock()
		pois = append(pois, vpois...)
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
		temp = determineTemp(hmap, cfg.Sea.SeaLevel, cfg.Temp)
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
