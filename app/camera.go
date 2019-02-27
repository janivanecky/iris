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
	targetRadius, radius float64
	targetAzimuth, azimuth float64 
	targetPolar, polar float64
	targetHeight, height float64
}

// GetCamera returns initialized Camera.
func GetCamera(radius, azimuth, polar, height float64) Camera {
	return Camera {
		targetRadius: radius, radius: radius,
		targetAzimuth: azimuth, azimuth: azimuth,
		targetPolar: polar, polar: polar,
		targetHeight: height, height: height,
	}
}

// SetStateWithTransition sets camera to smoothly move to other location in space.
func (cam *Camera) SetStateWithTransition(radius, azimuth, polar, height float64) {
	cam.targetRadius = radius
	cam.targetAzimuth = azimuth
	cam.targetPolar = polar
	cam.targetHeight = height
}

// GetState returns current camera position in space (in polar coordinates).
func (cam *Camera) GetState() (radius, azimuth, polar, height float64) {
	return cam.targetRadius, cam.targetAzimuth, cam.targetPolar, cam.targetHeight
}

// Update updates camera position in time.
func (cam *Camera) Update(dt float64) {
	// Update target height based on input. On the key press height
	// is adjusted in bigger step than on holding the key.
	if platform.IsKeyPressed(platform.KeyUp) {
		cam.targetHeight += keyPressHeightChangeStep
	} else if platform.IsKeyDown(platform.KeyUp) {
		cam.targetHeight += keyHoldHeightChangeStep
	}
	if platform.IsKeyPressed(platform.KeyDown) {
		cam.targetHeight -= keyPressHeightChangeStep
	} else if platform.IsKeyDown(platform.KeyDown) {
		cam.targetHeight -= keyHoldHeightChangeStep
	}

	// Update camera angles based on mouse input. Note that left alt key
	// has to be pressed for the camera to move.
	dx, dy := platform.GetMouseDeltaPosition()
	// TODO: remove dependencies on ui lib?
	if !ui.IsRegisteringInput && platform.IsKeyDown(platform.KeyLeftAlt) {
		if platform.IsMouseLeftButtonDown() {
			cam.targetAzimuth -= dx * angleChangeStep
			cam.targetPolar -= dy * angleChangeStep
			cam.targetPolar = clamp(cam.targetPolar, 0, math.Pi)
			// TODO: remove dependencies on ui lib?
			ui.SetInputResponsive(false)
		} else {
			ui.SetInputResponsive(true)
		}
	}

	// Update camera radius.
	mouseWheelDelta := platform.GetMouseWheelDelta()
	cam.targetRadius -= mouseWheelDelta * radiusChangeStep

	// Update camera state so the current state is closer to the target state.
	cam.azimuth, cam.targetAzimuth = updateAngles(cam.azimuth, cam.targetAzimuth, dt)
	cam.polar, cam.targetPolar = updateAngles(cam.polar, cam.targetPolar, dt)
	cam.polar = clamp(cam.polar, 0, math.Pi)
	cam.radius += (cam.targetRadius - cam.radius) * dt * cameraSpeed
	cam.height += (cam.targetHeight - cam.height) * dt * cameraSpeed
}

// GetViewMatrix returns matrix transforming world space to camera coordinate system.
func (cam *Camera) GetViewMatrix() mgl32.Mat4 {
	// position is camera position relative to target.
	position := vecFromPolarCoords(cam.azimuth, cam.polar, cam.radius)
	target := mgl32.Vec3{0, float32(cam.height), 0.0}
	// Up is y-axis by default, in case camera is positioned close to y-axis, up
	// vector will be in x-z plane, in direction opposite to camera position.
	up := mgl32.Vec3{0, 1, 0}
	if cam.polar < 0.1 {
		up = vecFromPolarCoords(cam.azimuth + math.Pi, math.Pi / 2.0, 1.0)
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

func updateAngles(current, target, dt float64) (newCurrent float64, newTarget float64) {
	// These corrections to target make sure that we're going to update in direction of shortest path.
	if target - current > math.Pi {
		target -= math.Pi * 2.0
	} else if current - target > math.Pi {
		target += math.Pi * 2.0
	}

	// Update current angle so it's closer to target angle.
	current += (target - current) * dt * cameraSpeed

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
