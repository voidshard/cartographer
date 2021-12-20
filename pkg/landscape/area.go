package landscape

// Area represents a specific small area of the map
type Area struct {
	// 0-255, where bigger numbers are higher/wetter/hotter
	Height      uint8
	Rainfall    uint8
	Temperature uint8 // in degress c, offset so 100 => 0 degrees cel
	Volcanism   uint8

	// if the square contains fresh/salt water/lava
	Sea   bool
	River bool
	Lake  bool // lake implies river
	Swamp bool // swamp water
	Lava  bool

	Biome Biome

	// if river, we set a river ID else 0
	RiverID int
	// if lake, we set a lake ID (1 -> 254) else 0. Each river numbers it's own lakes
	LakeID int
}
