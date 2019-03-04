package app

import (
	"github.com/go-gl/mathgl/mgl32"
)

// APP UI


type Circle struct {
	Color ColorParameter
	Width FloatParameter
	Radius FloatParameter
	Arc RadianParameter
}

func (circle *Circle) Update(dt float64) {
	circle.Width.Update(dt, 10.0)
	circle.Radius.Update(dt, 15.0)
	circle.Color.Update(dt, 4.0)
	circle.Arc.Update(dt, 6.0)
}

func GetCircle(color mgl32.Vec4, width, radius, angle float64) Circle {
	return Circle {
		ColorParameter{color, color},
		FloatParameter{width, width},
		FloatParameter{radius, radius},
		RadianParameter{angle, angle},
	}
}