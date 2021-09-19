package voronoi

import (
	"testing"

	"github.com/voidshard/cartographer/pkg/shapes"

	"github.com/stretchr/testify/assert"
)

func TestShortestPath(t *testing.T) {
	poly := shapes.NewPolygon([]*shapes.Point{
		shapes.Pt(0, 0),
		shapes.Pt(0, 1000),
		shapes.Pt(1000, 1000),
		shapes.Pt(1000, 0),
	})
	in := []*shapes.Point{
		shapes.Pt(200, 200),
		shapes.Pt(500, 300),
		shapes.Pt(800, 800),
	}

	gra, err := Compute(in, poly)

	assert.Nil(t, err)
	assert.NotNil(t, gra)

	verts := []float64{}
	for _, v := range gra.Verts {
		verts = append(verts, v.X, v.Y)
	}
	assert.Equal(t, []float64{
		150, 850,
		433, 0,
		1000, 340,
		0, 1000,
		0, 0,
		1000, 0,
		1000, 1000,
	}, verts)

	short, err := gra.ShortestPath(shapes.Pt(1, 1), shapes.Pt(999, 999), nil)
	assert.Nil(t, err)
	spath := []float64{}
	for _, v := range short {
		spath = append(spath, v.X, v.Y)
	}
	assert.Equal(t, []float64{0, 0, 0, 1000, 1000, 1000}, spath)

	short, err = gra.ShortestPath(shapes.Pt(0, 1000), shapes.Pt(150, 850), nil)
	assert.Nil(t, err)
	spath = []float64{}
	for _, v := range short {
		spath = append(spath, v.X, v.Y)
	}
	assert.Equal(t, []float64{0, 1000, 150, 850}, spath)
}
