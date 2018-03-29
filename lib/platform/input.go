package platform

import (
	"github.com/go-gl/glfw/v3.2/glfw"
)

var mouseX, mouseY float64 = -1.0, -1.0
var dmouseX, dmouseY float64 = 0.0, 0.0
var mouseButtonDown bool = false
var mouseButtonPressed bool = false
var mouseWheelDelta float64 = 0.0
var keyPressed map[glfw.Key]bool

func scrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	mouseWheelDelta = yoff
}

func initInput(window *glfw.Window) {
	window.SetScrollCallback(scrollCallback)

	keyPressed = make(map[glfw.Key]bool)
}

func Update(window *glfw.Window) {
	glfw.PollEvents()
	
	mouseButtonPressed = false

	// Update mouse position and get position delta.
	x, y := window.GetCursorPos()
	x, y = x / windowScale, y / windowScale
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

	for key, _ := range keyPressed {
		keyPressed[key] = false
	}

	escState := window.GetKey(glfw.KeyEscape)
	if escState == glfw.Press {
		keyPressed[glfw.KeyEscape] = true
	}
}

func IsEscPressed() bool {
	return keyPressed[glfw.KeyEscape]
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