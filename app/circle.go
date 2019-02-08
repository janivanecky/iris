package app

import (
	"math"
	"github.com/go-gl/mathgl/mgl32"
)

// APP UI
type Parameter interface {
	Update(dt, speed float64)
}

type ColorParameter struct {
	Val, Target mgl32.Vec4
}

func (color *ColorParameter) Update(dt, speed float64) {
	color.Val[0] += (color.Target[0] - color.Val[0]) * float32(dt * speed)
	color.Val[1] += (color.Target[1] - color.Val[1]) * float32(dt * speed)
	color.Val[2] += (color.Target[2] - color.Val[2]) * float32(dt * speed)
	color.Val[3] += (color.Target[3] - color.Val[3]) * float32(dt * speed)
}

type FloatParameter struct {
	Val, Target float64
}

func (value *FloatParameter) Update(dt, speed float64) {
	value.Val += (value.Target - value.Val) * dt * speed
}

type RadianParameter struct {
	Val, Target float64
}

func (value *RadianParameter) Update(dt, speed float64) {
	if value.Target - value.Val > math.Pi {
		value.Target -= math.Pi * 2.0
	} else if value.Val - value.Target > math.Pi {
		value.Target += math.Pi * 2.0
	}
	value.Val += (value.Target - value.Val) * dt * speed
	if value.Val > math.Pi * 2.0 && value.Target > math.Pi * 2.0 {
		value.Val -= math.Pi * 2.0
		value.Target -= math.Pi * 2.0
	}
	if value.Val < 0 && value.Target < 0 {
		value.Val += math.Pi * 2.0
		value.Target += math.Pi * 2.0
	}
}


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