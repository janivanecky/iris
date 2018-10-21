package app

import (
	"math"
	"math/rand"
	"github.com/go-gl/mathgl/mgl32"
)

// CELLS
type Cell struct {
	polar, angle, radius float64
	scale                mgl32.Vec3
	colorModifier        float32
	colorIndex           int
}

func GenerateCells(cells []Cell) {
	for i := range cells {
		scaleX := rand.Float32()*0.8 + 0.2
		scaleZ := rand.Float32()*0.8 + 0.2
		scaleX *= scaleX * 4.0 / 3.0
		scaleZ *= scaleZ * 2.0
		scaleX *= 2.0
		scaleZ *= 2.0
		scaleY := scaleX + scaleZ

		polar := rand.NormFloat64()
		angle := rand.Float64() * math.Pi * 2

		radius := rand.Float64()

		//radius *= radius
		scale := mgl32.Vec3{scaleX, scaleY, scaleZ}
		colorModifier := rand.Float32()*0.5 + 0.5
		colorIndex := rand.Int() % 5
		cells[i] = Cell{polar, angle, radius, scale, colorModifier, colorIndex}
	}
}


func radiusToRadiusMin(radius float64) float64 {
	return radius + 4.0
}

func radiusToRadiusMax(radius float64) float64 {
	return radius - 4.0
}


func GetCellMatrices(cells []Cell, cellsSettings CellSettings) []mgl32.Mat4{
	radiusMin := radiusToRadiusMin(cellsSettings.RadiusMin)
	radiusMax := radiusToRadiusMax(cellsSettings.RadiusMax)
	polarStd := cellsSettings.PolarStd
	polarMean := cellsSettings.PolarMean
	heightRatio := cellsSettings.HeightRatio

	matrices := make([]mgl32.Mat4, cellsSettings.Count)
	
	for i, cell := range cells[:cellsSettings.Count] {
		polar := cell.polar * polarStd + polarMean
		angle := cell.angle

		radius := cell.radius*(radiusMax - radiusMin) + radiusMin

		x := float32(math.Sin(polar) * math.Sin(angle) * radius)
		y := float32(math.Cos(polar) * radius * heightRatio)
		z := float32(math.Sin(polar) * math.Cos(angle) * radius)

		modelMatrix := mgl32.LookAt(
			x, y, z,
			0, 0, 0,
			0, 1, 0,
		).Inv().Mul4(
			mgl32.Scale3D(cell.scale[0], cell.scale[1], cell.scale[2]),
		)

		matrices[i] = modelMatrix
	}
	return matrices
}

func GetCellColors(cells []Cell, cellsSettings CellSettings, palette []mgl32.Vec4) []mgl32.Vec4{
	colors := make([]mgl32.Vec4, cellsSettings.Count)
	for i, cell := range cells[:cellsSettings.Count] {
		colors[i] = palette[cell.colorIndex].Mul(cell.colorModifier)

	}
	return colors
}