package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"

	"github.com/alecthomas/kong"

	perlin "github.com/voidshard/cartographer/pkg/perlin"
)

const desc = `Mix tool composites images together via perlin noise with given weights.

Eg: ./mix -o newimage.png ~/image.01.png=0.5 ~/image.02.png=0.5
Creates an image that uses 50% pixels from both inputs in a random fashion following perlin noise.
We can mix any number of images similarly.

A special image name 'nil' mixes in 'nil' or transparent pixels. 

Nb.
- We anticipate that input images are all the same size. The behaviour of this tool is undefined if not.
- We require input floats to be between 0-1 exclusive.
`

var cli struct {
	// input images
	Input map[string]float64 `arg type:"file:" help:"input images & their weights"`

	// name of output image
	Output string `short:"o" default:"out.png" help:"output name"`
}

// helper struct to hold key, value, image values in order
type norm struct {
	Key   string
	Value float64
	Im    image.Image
}

func main() {
	kong.Parse(
		&cli,
		kong.Name("mix"),
		kong.Description(desc),
	)

	// read images
	total := 0.0
	x := 0
	y := 0
	in := map[string]image.Image{}
	for k, v := range cli.Input {
		if v <= 0 || v >= 1 {
			panic(fmt.Sprintf("floats required to be between 0-1 exclusive"))
		}
		if k == "nil" {
			in[k] = nil
			total += v
			continue
		}

		f, err := os.Open(k)
		if err != nil {
			panic(err)
		}

		i, err := decode(f)
		if err != nil {
			panic(err)
		}

		in[k] = i
		x = int(i.Bounds().Max.X - i.Bounds().Min.X)
		y = int(i.Bounds().Max.Y - i.Bounds().Min.Y)
		total += v
	}

	// normalise images
	prev := 0.0
	norms := []*norm{}
	for k, v := range cli.Input {
		// nb. this adds input images in any order .. but we don't care about
		// that, so long as we iterate them in order from here on out
		// (hence transfering to a list)
		i, _ := in[k]
		norms = append(norms, &norm{
			Key:   k,
			Value: (255 * v / total) + prev,
			Im:    i,
		})
		prev = norms[len(norms)-1].Value
	}
	if len(in) < 2 {
		panic(fmt.Sprintf("require at least 2 input images, have %d", len(in)))
	}
	norms[len(norms)-1].Value = 255 // no matter what, last image will trump all

	// merge images
	p := perlin.Perlin(x, y, 1)
	out := image.NewRGBA(image.Rect(0, 0, x, y))

	for dx := 0; dx < x; dx++ {
		for dy := 0; dy < y; dy++ {
			// value of perlin map
			ruint, _, _, _ := p.At(dx, dy).RGBA()

			for _, nimg := range norms {
				if float64(ruint/255) <= nimg.Value {
					if nimg.Im != nil {
						// copy pixel from this image
						out.Set(dx, dy, nimg.Im.At(dx, dy))
					}
					break
				}
			}
		}
	}

	// and save the output
	err := savePng(cli.Output, out)
	if err != nil {
		panic(err)
	}
	fmt.Println("wrote", cli.Output)
}

// decode reads an image
func decode(in io.Reader) (image.Image, error) {
	im, _, err := image.Decode(in)
	return im, err
}

// savePng to disk
func savePng(fpath string, in image.Image) error {
	buff := new(bytes.Buffer)
	err := png.Encode(buff, in)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fpath, buff.Bytes(), 0644)
}
