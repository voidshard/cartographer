package landscape

type Config struct {
	// base map width
	Width uint
	// base map height
	Height uint

	Lakes    *lakeSettings
	Rain     *rainfallSettings
	Temp     *tempSettings
	Rivers   *riverSettings
	Land     *landSettings
	Sea      *seaSettings
	Volcanic *volcSettings
	Swamp    *swampSettings
	Biome    *biomeSettings
}

type lakeSettings struct {
	// variance used in radius calculation
	Variance float64

	// range of values about the centre that are set as lake.
	// Higher values imply wider less windy lakes.
	// Low values produce streams, disconnected ponds / marshland
	// type terrain.
	Radius uint8

	// max number of lakes (best effort)
	Number uint

	// max size of a lake -- how far it can extend from the centre.
	// We make it increasingly less likely that lakes will extend over this
	SoftMaxRadius float64

	// hard max, lakes cannot go over this dist from the origin
	HardMaxRadius float64

	// lakes must form at least this far from the river origin
	MinDistFromStart uint

	// lakes must form at least this far from the river end
	MinDistFromEnd uint
}

type volcSettings struct {
	// variance used in radius calculation
	Variance float64

	// how far volcano (centres) have to be from each other
	OriginMinDist float64

	// range of values about the centre that imply lava.
	// higher -> more lava
	LavaRadius uint8

	// range of values about the centre that imply volcanic land.
	// Should be > LavaRadius
	// higher values -> more volcanic land around lava edges
	VolcanicRedius uint8

	// max number of volcanoes (best effort)
	Number uint

	// max dist from the volcano centre lava / volcanic land can extend
	MaxRadius float64
}

type swampSettings struct {
	// number of swamp regions (max)
	Number uint

	// swamps must exist below this height
	MaxHeight uint8

	// max height +/- across a swamp
	Radius uint8

	// variance used in swamp outlines
	Variance float64
}

// tempSettings
type tempSettings struct {
	// in c, where 100 => 0c
	EquatorAverageTemp uint8
	PoleAverageTemp    uint8
	EquatorWidth       float64
	Variance           float64
}

type rainfallSettings struct {
	RainfallVariance float64
}

type riverSettings struct {
	// number of rivers (max), we do not guarantee this many
	Number uint

	// min dist between river origin points
	OriginMinDist float64

	// every river should have at least one section where it flows north/south
	// we also ensure the major drops in river height are at these points (ie
	// waterfalls will occur only north-south or south-north facing).
	// This is essentially a hack to aid applications where maps require features to
	// be straight on (facing the user) or away.
	ForceNorthSouthSections bool

	// TurnChance is how likely a river is to change direction on a given pixel
	TurnChance float64
}

type seaSettings struct {
	// sections of the map we consider below sea level
	SeaLevel uint8
}

type landSettings struct {
	// base height variance, higher numbers makes everything more chaotic
	HeightVariance float64

	// mountains are added to the above heightmap, higher numbes -> chaos
	MountainVariance float64
}

type biomeSettings struct {
	// any temperature at or below this is auto frozen unless sea or volcanic
	FrozenTemp uint8

	// A desert is declared if rainfall too low & temp is too high.
	// Technically temp is less important here.
	DesertTemp uint8
	DesertRain uint8

	// Tundra is somewhere reasonably cold, windswept & lowish on rainfall.
	// We require it's temp & rainfall at or below these values
	TundraTemp    uint8
	TundraRainMin uint8
	TundraRainMax uint8

	// tropical forest requires a min temp & rainfall
	ForestTropicalTemp uint8
	ForestTropicalRain uint8

	// temperate forest requires a min temp & rainfall
	// ... higher temps indicate a tropical forest.
	ForestTemperateTemp uint8
	ForestTemperateRain uint8

	// Within this distance from a river/lake we consider the land
	// well watered (ie. the same as high rainfall)
	FreshWaterRadius uint

	// height over which we call the mountain biome
	MountainHeight uint

	// hight over which we call the highlands biome
	HighlandsHeight uint
}

func DefaultConfig() *Config {
	return &Config{
		Width:  1000,
		Height: 1000,
		Biome: &biomeSettings{
			FrozenTemp:          70,
			DesertTemp:          20,
			DesertRain:          40,
			TundraTemp:          80,
			ForestTropicalTemp:  130,
			ForestTropicalRain:  205,
			ForestTemperateRain: 100,
			ForestTemperateTemp: 105,
			FreshWaterRadius:    10,
			MountainHeight:      210,
			HighlandsHeight:     170,
		},
		Lakes: &lakeSettings{
			Variance:         0.23,
			Radius:           30,
			Number:           4,
			SoftMaxRadius:    40,
			HardMaxRadius:    60,
			MinDistFromStart: 15,
			MinDistFromEnd:   30,
		},
		Rain: &rainfallSettings{
			RainfallVariance: 0.03,
		},
		Temp: &tempSettings{
			EquatorAverageTemp: 140,  // 40c
			PoleAverageTemp:    60,   // -40c
			EquatorWidth:       0.05, // % of height
			Variance:           0.03,
		},
		Rivers: &riverSettings{
			Number:                  50,
			OriginMinDist:           70,
			ForceNorthSouthSections: true,
			TurnChance:              0.4,
		},
		Land: &landSettings{
			HeightVariance:   0.03, // base heightmap
			MountainVariance: 0.10, // extra roughness
		},
		Sea: &seaSettings{
			SeaLevel: 115,
		},
		Volcanic: &volcSettings{
			Variance:       0.6,
			LavaRadius:     18,
			VolcanicRedius: 30,
			OriginMinDist:  30,
			Number:         5,
			MaxRadius:      60,
		},
		Swamp: &swampSettings{
			Number:    25,
			MaxHeight: 185,
			Radius:    20,
			Variance:  0.8,
		},
	}
}
