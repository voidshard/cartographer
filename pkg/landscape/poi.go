package landscape

type PointType string

const (
	RiverOrigin PointType = "river-origin"
	RiverEnd    PointType = "river-end"
	LakeOrigin  PointType = "lake-origin"
	LakeEnd     PointType = "lake-end"
	Volcano     PointType = "volcano"
	Swamp       PointType = "swamp"
)

// POI `PointOfInterest`
type POI struct {
	X    int
	Y    int
	Type PointType
}
