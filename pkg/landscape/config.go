package landscape

type Config struct {
	// base map width
	Width uint
	// base map height
	Height uint

	Rain   *rainfallSettings
	Temp   *tempSettings
	Rivers *riverSettings
	Land   *landSettings
	Sea    *seaSettings
}

// tempSettings
type tempSettings struct {
	// in c, where 100 => 0c
	EquatorAverageTemp uint8
	PoleAverageTemp    uint8
	EquatorWidth       float64
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

func DefaultConfig() *Config {
	return &Config{
		Width:  500,
		Height: 500,
		Rain: &rainfallSettings{
			RainfallVariance: 0.03,
		},
		Temp: &tempSettings{
			EquatorAverageTemp: 140,  // 40c
			PoleAverageTemp:    60,   // -40c
			EquatorWidth:       0.05, // % of height
		},
		Rivers: &riverSettings{
			Number:                  30,
			OriginMinDist:           70,
			ForceNorthSouthSections: true,
		},
		Land: &landSettings{
			HeightVariance:   0.03, // base heightmap
			MountainVariance: 0.10, // extra roughness
		},
		Sea: &seaSettings{
			SeaLevel: 115,
		},
	}
}
