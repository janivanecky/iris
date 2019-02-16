package platform

import (
	//"syscall"
	"runtime"

	"github.com/go-gl/glfw/v3.2/glfw"
)

var windowScale = 1.0
func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()

	// Initialize GLFW context
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	windowScale = GetWindowScaling()
}


func GetWindowScaling() float64 {
	//dll, err := syscall.LoadDLL("User32.dll")
	//if err != nil {
	//	return 1.0
	//}
	//dpiForSystem, _ := dll.FindProc("GetDpiForSystem")
	//dpi, errCode, _ := dpiForSystem.Call()
	//if errCode > 0 {
	//	return 1.0
	//}
	//scale := float64(dpi) / 96.0
	return 2.0
	//return scale
}

// GetMonitorResolution returns current monitor's resolution in pixels.
// Note that it is DPI scale adjusted, so if DPI scaling is set to 2.0,
// returned resolution will be half of the actual physical resolution.
// You can check DPI scaling value by calling GetWindowScaling().
func GetMonitorResolution() (int, int) {
	monitor := glfw.GetPrimaryMonitor()
	videoMode := monitor.GetVideoMode()
	width := int(float64(videoMode.Width) / windowScale)
	height := int(float64(videoMode.Height) / windowScale)
	return width, height
}

// GetWindow creates a window and returns *glfw.Window pointer.
// This includes setting up OpenGL context (4.1) via GLFW.
func GetWindow(width int, height int, title string, fullscreen bool)  *glfw.Window  {
	// Set GLFW window flags
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Decorated, glfw.False);
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)

	// Create scaled window
	scaledWidth  := int(windowScale * float64(width))
	scaledHeight := int(windowScale * float64(height))
	var monitor *glfw.Monitor
	if fullscreen {
		monitor = glfw.GetPrimaryMonitor()
	}
	window, err := glfw.CreateWindow(scaledWidth, scaledHeight, title, monitor, nil)
	if err != nil {
		panic(err)
	}
	// These are necessary for proper OpenGL support
	window.MakeContextCurrent()	
	glfw.SwapInterval(1)

	// TODO(jan): Ditch(?)
	initInput(window)
	window.SetPos(500, 100)

	return window
}

// ReleaseWindow releases GLFW context.
func ReleaseWindow() {
	// Release GLFW context.
	glfw.Terminate()
}
