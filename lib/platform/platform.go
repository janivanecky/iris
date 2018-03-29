package platform

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"syscall"
	"runtime"
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

var windowScale float64 = 1.0

func GetWindowScaling() float64 {
	dll, err := syscall.LoadDLL("User32.dll")
	if err != nil {
		return 1.0
	}
	dpiForSystem, _ := dll.FindProc("GetDpiForSystem")
	dpi, errCode, _ := dpiForSystem.Call()
	if errCode > 0 {
		return 1.0
	}
	scale := float64(dpi) / 96.0
	return scale
}

func GetWindow(width int, height int, title string)  *glfw.Window  {
	// Initialize GLFW context
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4);

	// Create our new fancy window
	windowScale = GetWindowScaling()
	scaledWidth, scaledHeight := int(windowScale * float64(width)), int(windowScale * float64(height))
	window, err := glfw.CreateWindow(scaledWidth, scaledHeight, title, nil, nil)
	if err != nil {
		panic(err)
	}

	// God knows why this is necessary
	window.MakeContextCurrent()	
	glfw.SwapInterval(0)

	initInput(window)

	return window
}

// Release GLFW context.
func ReleaseWindow() {
	glfw.Terminate()
}