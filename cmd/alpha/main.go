package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/alecthomas/kong"
)

const desc = `Alpha tool takes some input image & a region (x0,y0)->(x1,y1).

All colours in the given region are made fully transparent across the image.
`

var cli struct {
	//
	Input string `arg short:"i" type:"existingfile" help:"input image"`

	AlphaX0 int `arg long:"alpha-x0" help:"x0 of region designating colours to alpha"`
	AlphaY0 int `arg long:"alpha-y0" help:"y0 of region designating colours to alpha"`
	AlphaX1 int `arg long:"alpha-x1" help:"x1 of region designating colours to alpha"`
	AlphaY1 int `arg long:"alpha-y1" help:"y1 of region designating colours to alpha"`

	//
	Output string `short:"o" default:"out.png" help:"output name"`
}

func main() {
	kong.Parse(
		&cli,
		kong.Name("alpha"),
		kong.Description(desc),
	)
	fmt.Println("args", cli)

	f, err := os.Open(cli.Input)
	if err != nil {
		panic(err)
	}

	im, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	toalpha := coloursInRegion(im, cli.AlphaX0, cli.AlphaY0, cli.AlphaX1, cli.AlphaY1)
	fmt.Printf("colors to remove: %d\n", len(toalpha))
	if len(toalpha) < 1 {
		panic(fmt.Errorf("region must contain at least one colour"))
	}

	bnds := im.Bounds()
	out := image.NewRGBA(bnds)

	for dy := bnds.Min.Y; dy < bnds.Max.Y; dy++ {
		for dx := bnds.Min.X; dx < bnds.Max.X; dx++ {
			c := im.At(dx, dy)
			_, ok := toalpha[c]
			if ok {
				out.Set(dx, dy, color.RGBA{255, 255, 255, 0})
			} else {
				out.Set(dx, dy, c)
			}
		}
	}

	err = savePng(cli.Output, out)
	if err != nil {
		panic(err)
	}
}

// coloursInRegion returns a map of colours found in the given region
func coloursInRegion(im image.Image, x0, y0, x1, y1 int) map[color.Color]bool {
	cs := map[color.Color]bool{}

	for dy := y0; dy <= y1; dy++ {
		for dx := x0; dx <= x1; dx++ {
			cs[im.At(dx, dy)] = true
		}
	}

	return cs
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
