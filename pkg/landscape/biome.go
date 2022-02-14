package landscape

import (
	"image/color"

	"golang.org/x/image/colornames"
)

//
type Biome string

var (
	Frozen          Biome = "frozen" // very cold and almost entirely barren
	Desert          Biome = "desert" // hot & barren
	Swampland       Biome = "swamp"
	Volcanic        Biome = "volcanic"
	Sea             Biome = "sea"
	Mountainous     Biome = "mountainous"
	Tundra          Biome = "tundra"    // cold scrublands
	Lowlands        Biome = "lowlands"  // between coast / highlands
	Highlands       Biome = "highlands" // between mountainous & lowlands
	ForestTemperate Biome = "forest-temperate"
	ForestTropical  Biome = "forest-tropical"

	frozenColor    color.Color = colornames.Antiquewhite
	desertColor    color.Color = colornames.Brown
	swampColor     color.Color = colornames.Purple
	volcColor      color.Color = colornames.Crimson
	seaColor       color.Color = colornames.Aqua
	mountColor     color.Color = colornames.Slategrey
	tundraColor    color.Color = colornames.Turquoise
	lowColor       color.Color = colornames.Gold
	highColor      color.Color = colornames.Yellow
	forestTmpColor color.Color = colornames.Lightgreen
	forestTrpColor color.Color = colornames.Green

	biomecmap = map[Biome]color.Color{
		Frozen:          frozenColor,
		Desert:          desertColor,
		Swampland:       swampColor,
		Volcanic:        volcColor,
		Sea:             seaColor,
		Mountainous:     mountColor,
		Tundra:          tundraColor,
		Lowlands:        lowColor,
		Highlands:       highColor,
		ForestTemperate: forestTmpColor,
		ForestTropical:  forestTrpColor,
	}
	biomermap = map[color.Color]Biome{}
)

func init() {
	// reverse the map
	for k, v := range biomecmap {
		biomermap[v] = k
	}
}

func toBiome(c color.Color) Biome {
	b, ok := biomermap[c]
	if ok {
		return b
	}
	return Lowlands // shrug
}

func (l *Landscape) determineBiomes(cfg *Config) {
	x, y := l.rivers.Dimensions()
	out := NewMapImage(x, y)

	eachPixel(out, func(dx, dy int, c uint8) {
		if l.volcanic.Value(dx, dy) > 0 {
			out.Set(dx, dy, volcColor)
			return
		}

		height := l.height.Value(dx, dy)
		temp := l.temperature.Value(dx, dy)
		// drop temperature by 1 degree per 5 height units we climb
		aboveSea := decrement(height, cfg.Sea.SeaLevel)
		temp = decrement(temp, aboveSea/5)

		rain := l.rainfall.Value(dx, dy)

		if temp <= cfg.Biome.FrozenTemp {
			out.Set(dx, dy, frozenColor)
		} else if l.sea.Value(dx, dy) == 255 {
			out.Set(dx, dy, seaColor)
		} else if l.swamp.Value(dx, dy) == 255 || l.swamp.Value(dx, dy) == 120 {
			// swamp takes precedence over tundra as slightly warmer tundra appears
			// swamp-ish (eg "plains tundra") .. or water logged regions with no trees
			// since their roots cannot sink into permafrost .. though during the summer
			// months enough water does melt to pool in shallow bogs.
			// https://www.youtube.com/watch?v=NL95ehsFb-4
			out.Set(dx, dy, swampColor)
		} else if temp <= cfg.Biome.TundraTemp {
			// tundra tends to be a thin band of nearly perma frozen land
			out.Set(dx, dy, tundraColor)
		} else if temp >= cfg.Biome.DesertTemp && rain <= cfg.Biome.DesertRain {
			out.Set(dx, dy, desertColor)
		} else if temp >= cfg.Biome.ForestTropicalTemp && rain >= cfg.Biome.ForestTropicalRain {
			out.Set(dx, dy, forestTrpColor)
		} else if temp >= cfg.Biome.ForestTemperateTemp && rain >= cfg.Biome.ForestTemperateRain {
			out.Set(dx, dy, forestTmpColor)
		} else if uint(height) >= cfg.Biome.MountainHeight {
			out.Set(dx, dy, mountColor)
		} else if uint(height) >= cfg.Biome.HighlandsHeight {
			out.Set(dx, dy, highColor)
		} else {
			out.Set(dx, dy, lowColor)
		}

	})

	l.biomes = out
}
