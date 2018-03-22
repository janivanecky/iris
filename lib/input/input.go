package input

import (
	"github.com/go-gl/glfw/v3.2/glfw"
)

var mouseX, mouseY float64 = -1.0, -1.0
var dmouseX, dmouseY float64 = 0.0, 0.0
var mouseButtonDown bool = false
var mouseButtonPressed bool = false
var mouseWheelDelta float64 = 0.0

func scrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	mouseWheelDelta = yoff
}

func Init(window *glfw.Window) {
	window.SetScrollCallback(scrollCallback)
}

func Update(window *glfw.Window) {
	mouseButtonPressed = false

	// Update mouse position and get position delta.
	x, y := window.GetCursorPos()
	if mouseX > 0.0 && mouseY > 0.0 {
		dmouseX, dmouseY = x - mouseX, y - mouseY
	}
	mouseX, mouseY = x, y

	lButtonState := window.GetMouseButton(glfw.MouseButtonLeft)
	if lButtonState == glfw.Press {
		if !mouseButtonDown {
			mouseButtonPressed = true
		}
		mouseButtonDown = true
	} else {
		mouseButtonDown = false
	}
}

func GetMouseDeltaPosition() (float64, float64) {
	return dmouseX, dmouseY
}

func GetMousePosition() (float64, float64) {
	return mouseX, mouseY
}

func IsMouseLeftButtonPressed() bool {
	return mouseButtonPressed
}

func IsMouseLeftButtonDown() bool {
	return mouseButtonDown
}

func GetMouseWheelDelta() float64 {
	result := mouseWheelDelta
	mouseWheelDelta = 0.0
	return result
}