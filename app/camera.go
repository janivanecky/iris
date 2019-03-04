package app

import (
	"math"
	"github.com/go-gl/mathgl/mgl32"
	
	"../lib/platform"
	"../lib/ui"
)

// Constants.
const cameraSpeed = 5.0
const keyPressHeightChangeStep = 0.25
const keyHoldHeightChangeStep = 0.1
const angleChangeStep = 0.01
const radiusChangeStep = 10.0

// Camera represents in-scene camera which smoothly moves
// around specific point in space, looking at that point.
type Camera struct {
	radius FloatParameter
	azimuth RadianParameter
	polar RadianParameter
	height FloatParameter
}

// GetCamera returns initialized Camera.
func GetCamera(radius, azimuth, polar, height float64) Camera {
	return Camera {
		radius: FloatParameter{radius, radius},
		azimuth: RadianParameter{azimuth, azimuth},
		polar: RadianParameter{polar, polar},
		height: FloatParameter{height, height},
	}
}

// SetStateWithTransition sets camera to smoothly move to other location in space.
func (cam *Camera) SetStateWithTransition(radius, azimuth, polar, height float64) {
	cam.radius.Target = radius
	cam.azimuth.Target = azimuth
	cam.polar.Target = polar
	cam.height.Target = height
}

// GetState returns current camera position in space (in polar coordinates).
func (cam *Camera) GetState() (radius, azimuth, polar, height float64) {
	return cam.radius.Target, cam.azimuth.Target, cam.polar.Target, cam.height.Target
}

// Update updates camera position in time.
func (cam *Camera) Update(dt float64) {
	// Update target height based on input. On the key press height
	// is adjusted in bigger step than on holding the key.
	if platform.IsKeyPressed(platform.KeyUp) {
		cam.height.Target += keyPressHeightChangeStep
	} else if platform.IsKeyDown(platform.KeyUp) {
		cam.height.Target += keyHoldHeightChangeStep
	}
	if platform.IsKeyPressed(platform.KeyDown) {
		cam.height.Target -= keyPressHeightChangeStep
	} else if platform.IsKeyDown(platform.KeyDown) {
		cam.height.Target -= keyHoldHeightChangeStep
	}

	// Update camera angles based on mouse input. Note that left alt key
	// has to be pressed for the camera to move.
	dx, dy := platform.GetMouseDeltaPosition()
	// TODO: remove dependencies on ui lib?
	if !ui.IsRegisteringInput && platform.IsKeyDown(platform.KeyLeftAlt) {
		if platform.IsMouseLeftButtonDown() {
			cam.azimuth.Target -= dx * angleChangeStep
			cam.polar.Target -= dy * angleChangeStep
			cam.polar.Target = clamp(cam.polar.Target, 0, math.Pi)
			// TODO: remove dependencies on ui lib?
			ui.SetInputResponsive(false)
		} else {
			ui.SetInputResponsive(true)
		}
	}

	// Update camera radius.
	mouseWheelDelta := platform.GetMouseWheelDelta()
	cam.radius.Target -= mouseWheelDelta * radiusChangeStep

	// Update camera state so the current state is closer to the target state.
	cam.azimuth.Update(dt, cameraSpeed)
	cam.polar.Update(dt, cameraSpeed)
	cam.polar.Val = clamp(cam.polar.Val, 0, math.Pi)
	cam.radius.Update(dt, cameraSpeed)
	cam.height.Update(dt, cameraSpeed)
}

// GetViewMatrix returns matrix transforming world space to camera coordinate system.
func (cam *Camera) GetViewMatrix() mgl32.Mat4 {
	// position is camera position relative to target.
	position := vecFromPolarCoords(cam.azimuth.Val, cam.polar.Val, cam.radius.Val)
	target := mgl32.Vec3{0, float32(cam.height.Val), 0.0}
	// Up is y-axis by default, in case camera is positioned close to y-axis, up
	// vector will be in x-z plane, in direction opposite to camera position.
	up := mgl32.Vec3{0, 1, 0}
	if cam.polar.Val < 0.1 {
		up = vecFromPolarCoords(cam.azimuth.Val + math.Pi, math.Pi / 2.0, 1.0)
	}
	cameraPosition := position.Add(target)
	viewMatrix := mgl32.LookAtV(cameraPosition, target, up)
	return viewMatrix
}

//
// Helper functions
//

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
