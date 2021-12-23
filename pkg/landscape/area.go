package landscape

// Area represents a specific small area of the map
type Area struct {
	// 0-255, where bigger numbers are higher/wetter/hotter/geothermally active
	Height      uint8
	Rainfall    uint8
	Temperature uint8 // in degress c, offset so 100 => 0 degrees cel
	Volcanism   uint8

	// if the square contains fresh/swamp/salt water/lava
	Sea   bool
	River bool
	Swamp bool // as in swamp/stagnant water, generally implies a swamp biome
	Lava  bool

	// a general guess at the biome of the region as determined
	// by temperatue, elevation, rainfall, presense of fresh water etc
	Biome Biome

	// -- bonus fields --
	// lake implies river
	Lake bool
	// if river, we set a river ID else 0
	RiverID int
	// if lake, we set a lake ID (1 -> 254) else 0. Each river numbers it's own lakes
	LakeID int
}
