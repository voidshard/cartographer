package landscape

import (
	"log"
	"sync"
	"time"

	perlin "github.com/voidshard/cartographer/pkg/perlin"
)

func timer(in string) func() {
	start := time.Now()

	return func() {
		end := time.Now()
		log.Println(in, "took", end.Sub(start))
	}
}

// PerlinLandscape generates our maps from simple perlin noise & some basic math / combinations
func PerlinLandscape(cfg *Config) (*Landscape, error) {
	t := timer("heightmap")
	hmap := combine(
		weight(perlin.Perlin(int(cfg.Width), int(cfg.Height), cfg.Land.HeightVariance), 70),
		weight(perlin.Perlin(int(cfg.Width), int(cfg.Height), cfg.Land.MountainVariance), 30),
	)
	t()

	// modifies heightmap
	t = timer("geothermal")
	// nb. geothermal outputs the temperature map because this greatly decreases
	// our later workload increasing temperature near volcanic land
	volc, temp, pois := determineGeothermal(hmap, cfg.Sea.SeaLevel, cfg.Volcanic)
	t()

	// modifies heightmap
	t = timer("sea")
	sea := determineSea(hmap, cfg.Sea)
	t()

	// modifies heightmap
	// sadly, in order to run rivers to the sea, we have to know where the sea is
	// we also want to avoid running through lava
	t = timer("rivers")
	rvrs, rivermaps, rain, rpois := determineRivers(hmap, sea, volc, cfg.Rivers, cfg.Lakes)
	pois = append(pois, rpois...)
	t()

	wg := sync.WaitGroup{}
	wg.Add(4)

	plock := &sync.Mutex{}
	var swmp *MapImage

	go func() { // locate mountains
		tm := timer("mountains")
		defer wg.Done()
		mountains := findMountains(hmap)
		plock.Lock()
		defer plock.Unlock()
		pois = append(pois, mountains...)
		tm()
	}()
	go func() {
		ts := timer("swamp")
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
		ts()
	}()
	go func() {
		tt := timer("temperature")
		defer wg.Done()
		determineTemp(hmap, temp, cfg.Sea.SeaLevel, cfg.Temp)
		tt()
	}()
	go func() {
		tr := timer("rainfall")
		defer wg.Done()
		determineRainfall(hmap, rain, cfg.Rain)
		tr()
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
	t = timer("biomes")
	l.determineBiomes(cfg)
	t()

	return l, nil
}
