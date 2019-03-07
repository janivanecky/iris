package app

import (
	"math"
	"math/rand"
	"github.com/go-gl/mathgl/mgl32"
)

// Constants.
const minColorMultiplier = 0.5
const maxColorMultiplier = 1.5
const minNormalizedScaleX = 0.2
const minNormalizedScaleZ = 0.2
const squaredScaleX = 2.75
const squaredScaleZ = 4.0

// Cell represents single element in the image.
type Cell struct {
	polar, azimuth, radius float64
	scale                mgl32.Vec3
	colorMultiplier      float32
	colorIndex           int
}

// GenerateCells initializes array of cells with random values.
func GenerateCells(cells []Cell) {
	for i := range cells {
		// Get scale vector.
		scaleX := rand.Float32() * (1.0 - minNormalizedScaleX) + minNormalizedScaleX
		scaleZ := rand.Float32() * (1.0 - minNormalizedScaleZ) + minNormalizedScaleZ
		scaleX *= scaleX * squaredScaleX
		scaleZ *= scaleZ * squaredScaleZ
		scaleY := (scaleX + scaleZ) / 2.0
		scale := mgl32.Vec3{scaleX, scaleY, scaleZ}

		// Get polar coordinates.
		polar := rand.NormFloat64()
		azimuth := rand.Float64() * math.Pi * 2
		radius := rand.Float64()

		// Get random parameters for colors.
		colorMultiplier := rand.Float32() * (maxColorMultiplier - minColorMultiplier) + minColorMultiplier
		colorIndex := rand.Int()
		cells[i] = Cell{polar, azimuth, radius, scale, colorMultiplier, colorIndex}
	}
}

// GetCellModelMatrices returns an array of model matrices, each transforming a single cell into world space.
func GetCellModelMatrices(cells []Cell, radiusMin, radiusMax, polarStd, polarMean, heightRatio float64, count int) []mgl32.Mat4{
	// TODO: This conversion should happen on UI side, not here
	radiusMin = radiusMin + 4.0
	radiusMax = radiusMax - 4.0

	matrices := make([]mgl32.Mat4, count)
	
	for i, cell := range cells[:count] {
		// Get position in cartesian coordinates.
		polar := cell.polar * polarStd + polarMean
		azimuth := cell.azimuth
		radius := cell.radius * (radiusMax - radiusMin) + radiusMin
		position := vecFromPolarCoords(azimuth, polar, radius)

		// Construct model matrix.
		scaleMatrix := mgl32.Scale3D(cell.scale[0], cell.scale[1], cell.scale[2])
		modelMatrix := mgl32.LookAt(position[0], position[1], position[2], 0, 0, 0, 0, 1, 0).Inv().Mul4(scaleMatrix)

		matrices[i] = modelMatrix
	}
	return matrices
}

// GetCellColors returns an array of color vectors, each for a single cell.
func GetCellColors(cells []Cell, colorPalette []mgl32.Vec4, count int) []mgl32.Vec4{
	colors := make([]mgl32.Vec4, count)
	colorPaletteSize := len(colorPalette)

	for i, cell := range cells[:count] {
		colorIndex := cell.colorIndex % colorPaletteSize
		colors[i] = colorPalette[colorIndex].Mul(cell.colorMultiplier)
	}
	return colors
}