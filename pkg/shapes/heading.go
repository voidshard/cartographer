package shapes

import (
	"fmt"
	"strings"
)

type Heading int

const (
	NORTH     Heading = 0
	NORTHEAST Heading = 1
	EAST      Heading = 2
	SOUTHEAST Heading = 3
	SOUTH     Heading = 4
	SOUTHWEST Heading = 5
	WEST      Heading = 6
	NORTHWEST Heading = 7
)

var headingStrings = map[Heading]string{
	NORTH:     "north",
	NORTHEAST: "northeast",
	EAST:      "east",
	SOUTHEAST: "southeast",
	SOUTH:     "south",
	SOUTHWEST: "southwest",
	WEST:      "west",
	NORTHWEST: "northwest",
}

var headingStringsInv = map[string]Heading{}

func init() {
	for k, v := range headingStrings {
		headingStringsInv[v] = k
	}
}

func ToHeadingStr(s string) (Heading, error) {
	val, ok := headingStringsInv[strings.ToLower(s)]
	if !ok {
		return NORTH, fmt.Errorf("no heading found for %s", s)
	}
	return val, nil
}

func ToHeadingInt(i int) Heading {
	switch i % 8 {
	case 0:
		return NORTH
	case 1:
		return NORTHEAST
	case 2:
		return EAST
	case 3:
		return SOUTHEAST
	case 4:
		return SOUTH
	case 5:
		return SOUTHWEST
	case 6:
		return WEST
	case 7:
		return NORTHWEST
	}
	return NORTH
}

func (h Heading) IsDiagonal() bool {
	switch h {
	case NORTHEAST, NORTHWEST, SOUTHEAST, SOUTHWEST:
		return true
	default:
		return false
	}
	return false
}

func (h Heading) Dist(a Heading) int {
	v := int(h) - int(a)
	if v < 0 {
		return -1 * v
	}
	if v > 4 {
		// 5-7 means that it must be closer going
		// clockwise crossing 0
		return 8 - v
	}
	return v
}

func (h Heading) RiseRun() (int, int) {
	switch h {
	case NORTH:
		return 0, -1
	case NORTHEAST:
		return 1, -1
	case EAST:
		return 1, 0
	case SOUTHEAST:
		return 1, 1
	case SOUTH:
		return 0, 1
	case SOUTHWEST:
		return -1, 1
	case WEST:
		return -1, 0
	case NORTHWEST:
		return -1, -1
	}
	// should never happen but :shrug:
	return 0, -1
}

func (h Heading) Left() Heading {
	v := int(h) - 1
	if v < 0 {
		v = 8 + v
	}
	return ToHeadingInt(v)
}

func (h Heading) Right() Heading {
	return ToHeadingInt(int(h) + 1)
}

func (h Heading) String() string {
	s := headingStrings[h]
	return s
}
