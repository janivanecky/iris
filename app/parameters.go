package app

import (
	"math"	

	"github.com/go-gl/mathgl/mgl32"
)

type ColorParameter struct {
	Val, Target mgl32.Vec4
}

func (p *ColorParameter) Update(dt, speed float64) {
	p.Val[0] += (p.Target[0] - p.Val[0]) * float32(dt * speed)
	p.Val[1] += (p.Target[1] - p.Val[1]) * float32(dt * speed)
	p.Val[2] += (p.Target[2] - p.Val[2]) * float32(dt * speed)
	p.Val[3] += (p.Target[3] - p.Val[3]) * float32(dt * speed)
}

type FloatParameter struct {
	Val, Target float64
}

func (p *FloatParameter) Update(dt, speed float64) {
	p.Val += (p.Target - p.Val) * dt * speed
}

type RadianParameter struct {
	Val, Target float64
}

func (p *RadianParameter) Update(dt, speed float64) {
	// These corrections to target make sure that we're going to update in direction of shortest path.
	if p.Target - p.Val > math.Pi {
		p.Target -= math.Pi * 2.0
	} else if p.Val - p.Target > math.Pi {
		p.Target += math.Pi * 2.0
	}

	// Update current angle so it's closer to target angle.
	p.Val += (p.Target - p.Val) * dt * speed

	// These corrections normalize to [0, 2 * pi] range (if both current value and target are outside).
	if p.Val > math.Pi * 2.0 && p.Target > math.Pi * 2.0 {
		p.Val -= math.Pi * 2.0
		p.Target -= math.Pi * 2.0
	} else if p.Val < 0 && p.Target < 0 {
		p.Val += math.Pi * 2.0
		p.Target += math.Pi * 2.0
	}
}
