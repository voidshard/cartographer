package landscape

import (
	"bytes"
	"image"
	"image/png"
	"io/ioutil"
	"math"
)

// decrement uint8 with min value of 0
func decrement(v, i uint8) uint8 {
	if v >= i {
		return v - i
	}
	return 0
}

// increment uin8 capping the max value to 255
func increment(v, i uint8) uint8 {
	if int(v)+int(i) > 255 {
		return 255
	}
	return v + i
}

// savePng file
func savePng(path string, in image.Image) error {
	buff := new(bytes.Buffer)
	err := png.Encode(buff, in)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, buff.Bytes(), 0666)
}

// toUint8 forces a float into a uint8 (enforcing max & min values)
func toUint8(f float64) uint8 {
	f = math.Round(f)
	if f < 0 {
		f = 0
	}
	if f > 255 {
		f = 255
	}
	return uint8(f)
}
