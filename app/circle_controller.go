package app

import (
	"math"
	"github.com/go-gl/mathgl/mgl32"

	"../lib/platform"
)

// Constants.
const controllerWidth         = 0.75
const controllerWidthHover    = 1.5
var   controllerColor         = mgl32.Vec4{0.0, 0.0, 0.0, 0.4}
var   controllerColorHover    = mgl32.Vec4{0.0, 0.0, 0.0, 0.8}
var   controllerColorInactive = mgl32.Vec4{0.0, 0.0, 0.0, 0.01}

const widthUpdateSpeed 		  = 10.0
const radiusUpdateSpeed		  = 15.0
const colorUpdateSpeed 		  = 4.0
const angleUpdateSpeed 		  = 6.0

const controllerMeshSegmentCount = 128
const controllerMeshRadialLength = math.Pi / 4.0

// CircleController represents UI element which can be used to change size of circle in 3D scene.
type CircleController struct {
	Color       ColorParameter
	Width       FloatParameter
	Radius      FloatParameter
	Angle       RadianParameter
	active, hot bool
}

// GetCircleController returns initialized CircleController.
func GetCircleController(radius, angle float64) CircleController {
	return CircleController{
		ColorParameter{controllerColor, controllerColor},
		FloatParameter{controllerWidth, controllerWidth},
		FloatParameter{radius, radius},
		RadianParameter{angle, angle},
		false, false,
	}
}

// Update handles input and smooth changes in controller's state.
func (controller *CircleController) Update(dt float64, mouseXWorld, mouseZWorld float64, minRadius, maxRadius float64, hidden bool, inputActive bool) {
	// Get mouse position in polar coordinates in XZ plane.
	// We need it to check if mouse is positioned over controller.
	mouseAngle := math.Atan2(mouseXWorld, mouseZWorld)
	mouseRadius := math.Sqrt(mouseXWorld * mouseXWorld + mouseZWorld * mouseZWorld)
	hover := math.Abs(controller.Radius.Val - mouseRadius) < 4.0

	// Update hot/active status of the controller.
	controller.hot = hover
	if controller.active && !platform.IsMouseLeftButtonDown() && inputActive {
		controller.active = false
	}
	if controller.hot && platform.IsMouseLeftButtonPressed() && inputActive {
		controller.active = true
	}

	// Dependent on hot/active/hidden states, set color and width targets.
	if controller.hot || controller.active {
		controller.Color.Target = controllerColorHover
		controller.Width.Target = controllerWidthHover
	} else {
		if hidden {
			controller.Color.Target = controllerColorInactive
		} else {
			controller.Color.Target = controllerColor
		}
		controller.Width.Target = controllerWidth
	}

	// If controller is active, we want it to have the same radius as mouse's radius in polar coords.
	// It's angle follows mouse angle all the time.
	if controller.active {
		controller.Radius.Target = math.Max(minRadius, math.Min(mouseRadius, maxRadius))
	}
	controller.Angle.Target = mouseAngle

	// Update controller's parameters.
	controller.Width.Update(dt, widthUpdateSpeed)
	controller.Radius.Update(dt, radiusUpdateSpeed)
	controller.Color.Update(dt, colorUpdateSpeed)
	controller.Angle.Update(dt, angleUpdateSpeed)
}

// GetMeshData returns vertex and index arrays for circle controller's mesh.
func (controller *CircleController) GetMeshData() ([]float32, []uint32) {
	vertices := make([]float32, 0)
	indices := make([]uint32, 0)

	for i := 0; i < controllerMeshSegmentCount; i++ {
		// Get current segment's angle.
		segmentPart := float64(i) / float64(controllerMeshSegmentCount)
		angle := segmentPart * controllerMeshRadialLength + controller.Angle.Val - controllerMeshRadialLength / 2.0

		// Get xyz coordinates for point on inner and outer side of the segment.
		xInner := float32(math.Sin(angle) * (controller.Radius.Val - controller.Width.Val))
		zInner := float32(math.Cos(angle) * (controller.Radius.Val - controller.Width.Val))
		xOuter := float32(math.Sin(angle) * (controller.Radius.Val + controller.Width.Val))
		zOuter := float32(math.Cos(angle) * (controller.Radius.Val + controller.Width.Val))
		y := float32(0.0)

		// Add vertices for inner and outer points of the segment to vertex array.
		// Note that this also adds normal to the vertex array, pointing in direction of y-axis.
		vertices = append(vertices, xInner, y, zInner, 1.0)
		vertices = append(vertices, 0.0, 1.0, 0.0, 1.0)
		vertices = append(vertices, xOuter, y, zOuter, 1.0)
		vertices = append(vertices, 0.0, 1.0, 0.0, 1.0)

		if i < controllerMeshSegmentCount - 1 {
			// Store indices for two triangles forming a quad. The quad connects vertices of the
			// current segment with vertices of the following segment.
			index := uint32(2 * i)
			indices = append(indices, index, index+1, index+3)
			indices = append(indices, index, index+3, index+2)
		}
	}

	return vertices, indices
}
