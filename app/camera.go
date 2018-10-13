package app

import (
	"math"
	"github.com/go-gl/mathgl/mgl32"
	
	"github.com/janivanecky/golib/platform"
	"github.com/janivanecky/golib/ui"
)


type Camera struct {
	TargetRadius, TargetAzimuth, TargetPolar float64
	radius, azimuth, polar float64

	TargetHeight float64
	height float64

	speed float64
}

func GetCamera(radius, azimuth, polar, speed float64) Camera {
	return Camera {
		TargetRadius: radius,
		TargetAzimuth: azimuth,
		TargetPolar: polar,
		speed: speed,
	}
}

func (cam *Camera) UpdateFully() {
	cam.radius = cam.TargetRadius
	cam.azimuth = cam.TargetAzimuth
	cam.polar = cam.TargetPolar
	cam.height = cam.TargetHeight
}

func (cam *Camera) Update(dt float64) {
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

	// Update mouse position and get position delta.
	dx, dy := platform.GetMouseDeltaPosition()
	if !ui.IsRegisteringInput && platform.IsKeyDown(platform.KeyLeftAlt) {
		if platform.IsMouseLeftButtonDown() {
			cam.TargetAzimuth -= dx / 100.0
			cam.TargetPolar -= dy / 100.0
			cam.TargetPolar = math.Max(cam.TargetPolar, 0)
			cam.TargetPolar = math.Min(cam.TargetPolar, math.Pi)
			ui.SetInputResponsive(false)
		} else {
			ui.SetInputResponsive(true)
		}
			
	}
	mouseWheelDelta := platform.GetMouseWheelDelta()
	if mouseWheelDelta != 0.0 {
		cam.TargetRadius -= mouseWheelDelta / 2.0 * 20.0
	}
	
	cam.azimuth += (cam.TargetAzimuth - cam.azimuth) * dt * cam.speed
	if cam.azimuth > math.Pi * 2.0 && cam.TargetAzimuth > math.Pi * 2.0 {
		cam.azimuth -= math.Pi * 2.0
		cam.TargetAzimuth -= math.Pi * 2.0
	}
	if cam.azimuth < 0 && cam.TargetAzimuth < 0 {
		cam.azimuth += math.Pi * 2.0
		cam.TargetAzimuth += math.Pi * 2.0
	}

	if cam.TargetPolar - cam.polar > math.Pi {
		cam.TargetPolar -= math.Pi * 2.0
	} else if cam.polar - cam.TargetPolar > math.Pi {
		cam.TargetPolar += math.Pi * 2.0
	}
	cam.polar += (cam.TargetPolar - cam.polar) * dt * cam.speed
	if cam.polar > math.Pi * 2.0 && cam.TargetPolar > math.Pi * 2.0 {
		cam.polar -= math.Pi * 2.0
		cam.TargetPolar -= math.Pi * 2.0
	}
	if cam.polar < 0 && cam.TargetPolar < 0 {
		cam.polar += math.Pi * 2.0
		cam.TargetPolar += math.Pi * 2.0
	}
	cam.polar = math.Max(cam.polar, 0)
	cam.polar = math.Min(cam.polar, math.Pi)

	cam.radius += (cam.TargetRadius - cam.radius) * dt * cam.speed
	cam.height += (cam.TargetHeight - cam.height) * dt * cam.speed
}

func (cam *Camera) SetAzimuth(azimuth float64) {
	cam.TargetAzimuth = azimuth
	if cam.TargetAzimuth - cam.azimuth > math.Pi {
		cam.TargetAzimuth -= math.Pi * 2.0
	} else if cam.azimuth - cam.TargetAzimuth > math.Pi {
		cam.TargetAzimuth += math.Pi * 2.0
	}
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

func vecFromPolarCoords(azimuth float64, polar float64, radius float64) mgl32.Vec3 {
	result := mgl32.Vec3{
		float32(math.Sin(polar) * math.Sin(azimuth) * radius),
		float32(math.Cos(polar) * radius),
		float32(math.Sin(polar) * math.Cos(azimuth) * radius),
	}
	return result
}