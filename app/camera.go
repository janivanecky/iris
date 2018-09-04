package app

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	
	"github.com/janivanecky/golib/platform"
	"github.com/janivanecky/golib/ui"
)

type Camera struct {
	Radius, Azimuth, Polar float64
	dRadius, dAzimuth, dPolar float64

	Height float64
	dHeight float64

	speed float64
}

func GetCamera(radius, azimuth, polar, speed float64) Camera {
	return Camera {
		Radius: radius,
		Azimuth: azimuth,
		Polar: polar,
		speed: speed,
	}
}

func (cam *Camera) Update(dt float64) {
	if platform.IsKeyPressed(platform.KeyUp) {
		cam.dHeight += 0.25
	} else if platform.IsKeyDown(platform.KeyUp) {
		cam.dHeight += 0.1
	}

	if platform.IsKeyPressed(platform.KeyDown) {
		cam.dHeight -= 0.25
	} else if platform.IsKeyDown(platform.KeyDown) {
		cam.dHeight -= 0.1
	}

	// Update mouse position and get position delta.
	dx, dy := platform.GetMouseDeltaPosition()
	if !ui.IsRegisteringInput && platform.IsKeyDown(platform.KeyLeftAlt) {
		if platform.IsMouseLeftButtonDown() {
			cam.dAzimuth -= dx / 100.0
			cam.dPolar -= dy / 100.0
			ui.SetInputResponsive(false)
		} else {
			ui.SetInputResponsive(true)
		}
			
	}
	mouseWheelDelta := platform.GetMouseWheelDelta()
	if mouseWheelDelta != 0.0 {
		cam.dRadius -= mouseWheelDelta / 2.0
	}

	cam.Azimuth += cam.dAzimuth * dt * cam.speed
	cam.Polar += cam.dPolar * dt * cam.speed * 2.0
	cam.Polar = math.Max(cam.Polar, 0)
	cam.Polar = math.Min(cam.Polar, math.Pi)
	cam.Radius += cam.dRadius * dt * cam.speed * 20.0
	cam.Height += cam.dHeight * dt * cam.speed * 10.0

	cam.dPolar *= 0.8
	cam.dAzimuth *= 0.8
	//cam.dAzimuth = math.Max(cam.dAzimuth, 0.1)
	cam.dRadius *= 0.9
	cam.dHeight *= 0.9
}

func (cam *Camera) GetPosition() mgl32.Vec3 {
	return vecFromPolarCoords(cam.Azimuth, cam.Polar, cam.Radius)
}

func (cam *Camera) GetTarget() mgl32.Vec3 {
	return mgl32.Vec3{0, float32(cam.Height), 0.0}
}

func (cam *Camera) GetUp() mgl32.Vec3 {
	up := mgl32.Vec3{0, 1, 0}
	if cam.Polar < 0.05 {
		up = vecFromPolarCoords(cam.Azimuth + math.Pi, math.Pi / 2.0, 1.0)
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