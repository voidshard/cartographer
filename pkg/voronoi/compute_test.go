package voronoi

import (
	"testing"

	"github.com/voidshard/cartographer/pkg/shapes"

	"github.com/stretchr/testify/assert"
)

func TestCompute(t *testing.T) {
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
	assert.Equal(t, poly, gra.Poly)

	assert.Equal(t, 3, len(gra.Cells))
	assert.Equal(t, 7, len(gra.Verts))
	assert.Equal(t, 9, len(gra.Edges))

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

	edges := []int{}
	cells := []int{}
	for i, edge := range gra.Edges {
		assert.Equal(t, i, edge.ID)
		edges = append(edges, edge.PointAID, edge.PointBID)

		lc := -1
		if edge.LeftCell != nil {
			lc = edge.LeftCell.ID
		}

		rc := -1
		if edge.RightCell != nil {
			rc = edge.RightCell.ID
		}

		cells = append(cells, lc, rc)
	}
	assert.Equal(t, []int{0, 1, 0, 2, 0, 3, 1, 4, 4, 3, 2, 5, 5, 1, 3, 6, 6, 2}, edges)
	assert.Equal(t, []int{0, 1, 1, 2, 2, 0, 0, -1, 0, -1, 1, -1, 1, -1, 2, -1, 2, -1}, cells)

	sites := []float64{}
	for i, cell := range gra.Cells {
		assert.Equal(t, i, cell.ID)
		sites = append(sites, cell.Site.X, cell.Site.Y)
	}
	assert.Equal(t, []float64{200, 200, 500, 300, 800, 800}, sites)
}
