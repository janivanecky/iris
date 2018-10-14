package app

import (
	"math"
	"encoding/json"
	"github.com/go-gl/mathgl/mgl32"
	
	"github.com/janivanecky/golib/platform"
	"github.com/janivanecky/golib/ui"
)


type Camera struct {
	TargetRadius, radius float64
	TargetAzimuth, azimuth float64 
	TargetPolar, polar float64
	TargetHeight, height float64
	speed float64
}

func GetCamera(radius, azimuth, polar, height, speed float64) Camera {
	return Camera {
		TargetRadius: radius, radius: radius,
		TargetAzimuth: azimuth, azimuth: azimuth,
		TargetPolar: polar, polar: polar,
		TargetHeight: height, height: height,
		speed: speed,
	}
}

type _Camera Camera
func (cam *Camera) UnmarshalJSON(b []byte) error {
	// This _Camera trickery prevents infinite recursion.
	_cam := _Camera{}
	err := json.Unmarshal(b, &_cam)
	if err != nil {
		return err
	}

	// When deserializing, we want the camera to be in the target position
	// already. We need to set the values manually, because current position
	// is not serialized (because current position describes application state,
	// not visualization parameters).
	cam.radius 		  = _cam.TargetRadius
	cam.TargetRadius  = _cam.TargetRadius
	cam.azimuth 	  = _cam.TargetAzimuth
	cam.TargetAzimuth = _cam.TargetAzimuth
	cam.polar 		  = _cam.TargetPolar
	cam.TargetPolar   = _cam.TargetPolar
	cam.height 		  = _cam.TargetHeight
	cam.TargetHeight  = _cam.TargetHeight

	return nil
}

// Struct "methods"

func (cam *Camera) Update(dt float64) {
	// Update target camera parameters based on input
	if platform.IsKeyPressed(platform.KeyUp) {
		cam.TargetHeight += 0.25
	} else if platform.IsKeyDown(platform.KeyUp) {
		cam.TargetHeight += 0.1
	}
	if platform.IsKeyPressed(platform.KeyDown) {
		cam.TargetHeight -= 0.25
	} else if platform.IsKeyDown(platform.KeyDown) {
		cam.TargetHeight -= 0.1
	}
	dx, dy := platform.GetMouseDeltaPosition()
	if !ui.IsRegisteringInput && platform.IsKeyDown(platform.KeyLeftAlt) {
		if platform.IsMouseLeftButtonDown() {
			cam.TargetAzimuth -= dx / 100.0
			cam.TargetPolar -= dy / 100.0
			cam.TargetPolar = clamp(cam.TargetPolar, 0, math.Pi)
			ui.SetInputResponsive(false)
		} else {
			ui.SetInputResponsive(true)
		}
	}
	mouseWheelDelta := platform.GetMouseWheelDelta()
	if mouseWheelDelta != 0.0 {
		cam.TargetRadius -= mouseWheelDelta / 2.0 * 20.0
	}

	// Update camera parameters
	cam.azimuth, cam.TargetAzimuth = updateAngles(cam.azimuth, cam.TargetAzimuth, cam.speed, dt)
	cam.polar, cam.TargetPolar = updateAngles(cam.polar, cam.TargetPolar, cam.speed, dt)
	cam.polar = clamp(cam.polar, 0, math.Pi)
	cam.radius += (cam.TargetRadius - cam.radius) * dt * cam.speed
	cam.height += (cam.TargetHeight - cam.height) * dt * cam.speed
}

func (cam *Camera) GetPosition() mgl32.Vec3 {
	return vecFromPolarCoords(cam.azimuth, cam.polar, cam.radius)
}

func (cam *Camera) GetTarget() mgl32.Vec3 {
	return mgl32.Vec3{0, float32(cam.height), 0.0}
}

func (cam *Camera) GetUp() mgl32.Vec3 {
	up := mgl32.Vec3{0, 1, 0}
	if cam.polar < 0.1 {
		up = vecFromPolarCoords(cam.azimuth + math.Pi, math.Pi / 2.0, 1.0)
	}
	return up
}

// Helper functions.

func vecFromPolarCoords(azimuth float64, polar float64, radius float64) mgl32.Vec3 {
	result := mgl32.Vec3{
		float32(math.Sin(polar) * math.Sin(azimuth) * radius),
		float32(math.Cos(polar) * radius),
		float32(math.Sin(polar) * math.Cos(azimuth) * radius),
	}
	return result
}

func updateAngles(current, target, speed, dt float64) (newCurrent float64, newTarget float64) {
	// These corrections to target make sure that we're going to update in direction of shortest path.
	if target - current > math.Pi {
		target -= math.Pi * 2.0
	} else if current - target > math.Pi {
		target += math.Pi * 2.0
	}
	current += (target - current) * dt * speed
	// These corrections normalize to [0, 2 * pi] range (if both current value and target are outside).
	if current > math.Pi * 2.0 && target > math.Pi * 2.0 {
		current -= math.Pi * 2.0
		target -= math.Pi * 2.0
	} else if current < 0 && target < 0 {
		current += math.Pi * 2.0
		target += math.Pi * 2.0
	}
	return current, target
}

func clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(val, max))
}
