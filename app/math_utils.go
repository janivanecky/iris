package app

import (
	"math"
	"github.com/go-gl/mathgl/mgl32"
)


func vecFromPolarCoords(azimuth float64, polar float64, radius float64) mgl32.Vec3 {
	result := mgl32.Vec3{
		float32(math.Sin(polar) * math.Sin(azimuth) * radius),
		float32(math.Cos(polar) * radius),
		float32(math.Sin(polar) * math.Cos(azimuth) * radius),
	}
	return result
}

func clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(val, max))
}
