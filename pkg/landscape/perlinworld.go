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

	wg.Add(2)
	go func() {
		// modifies heightmap
		sea = determineSea(hmap, cfg.Sea)
		wg.Done()
	}()
	go func() {
		rvrs = determineRivers(hmap, cfg.Sea.SeaLevel, cfg.Rivers)
		wg.Done()
	}()

	wg.Wait()
	wg.Add(2)

	var temp *MapImage
	var rain *MapImage

	go func() {
		temp = determineTemp(hmap, cfg.Sea.SeaLevel, cfg.Temp)
		wg.Done()
	}()
	go func() {
		rain = determineRainfall(hmap, cfg.Rain)
		wg.Done()
	}()
	wg.Wait()

	vege := determineVegetation(sea, rvrs, temp, rain, cfg.Vegetation)

	return &Landscape{
		height:      hmap,
		sea:         sea,
		rivers:      rvrs,
		temperature: temp,
		rainfall:    rain,
		vegetation:  vege,
	}, nil
}
