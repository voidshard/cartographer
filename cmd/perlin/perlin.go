package main

import (
	"fmt"
	gen "github.com/voidshard/cartographer/pkg/landscape"
)

func main() {
	// Generates random map (landscape) using perlin noise

	cfg := gen.DefaultConfig()

	w, err := gen.PerlinLandscape(cfg)
	if err != nil {
		panic(err)
	}

	folder, err := w.DebugRender()
	if err != nil {
		panic(err)
	}

	fmt.Println("rendered to", folder)
}
