package main

import (
	_ "fmt"
	"io/ioutil"
	"math"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"

	"./lib/graphics"
	"./lib/font"
	"./lib/ui"
	"./lib/platform"
	"./app"
	gmath "./lib/math"

	"runtime"
)

// TODO: move to platform(?)
func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func vecFromPolarCoords(azimuth float64, polar float64, radius float64) gmath.Vec3 {
	result := gmath.Vec3{
		float32(math.Sin(polar) * math.Sin(azimuth) * radius),
		float32(math.Cos(polar) * radius),
		float32(math.Sin(polar) * math.Cos(azimuth) * radius),
	}
	return result
}

const WINDOW_WIDTH = 800
const WINDOW_HEIGHT = 600

func main() {
	window := platform.GetWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "New fancy window")
	defer platform.ReleaseWindow()
	
	truetypeBytes, err := ioutil.ReadFile("fonts/font.ttf")
	if err != nil {
		panic(err)
    }
	
	scale := platform.GetWindowScaling()
    uiFont := font.GetFont(truetypeBytes, 20.0, scale)
	app.InitRendering(WINDOW_WIDTH, WINDOW_HEIGHT, uiFont)
	
	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:], []int{4, 4})

	polar := math.Pi / 2.0
	azimuth := 0.0
	radius := 5.0

	toggle := true
	val := 0.5

	start := time.Now()

	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		t := time.Now()
		_ = t.Sub(start)
		start = t

		// Let GLFW interface with the OS - not our job, right?
		platform.Update(window)

		// Let's quit if user presses Esc, that cannot mean anything else.
		escState := window.GetKey(glfw.KeyEscape)
		if escState == glfw.Press {
			break
		}

		// Update mouse position and get position delta.
		dx, dy := platform.GetMouseDeltaPosition()

		camChanged := false
		uiResponsive := true

		if !ui.IsRegisteringInput {
			if platform.IsMouseLeftButtonDown() {
				azimuth -= dx / 100.0
				polar -= dy / 100.0
				camChanged = true
				uiResponsive = false
			}
	
			mouseWheelDelta := platform.GetMouseWheelDelta()
			if mouseWheelDelta != 0.0 {
				radius -= mouseWheelDelta / 2.0
				camChanged = true
			}
	
			if camChanged {
				camPosition := vecFromPolarCoords(azimuth, polar, radius)
				app.SetCameraPosition(camPosition)
			}
		}

		panel := ui.StartPanel("Test panel", gmath.Vec2{})
		toggle, _ = panel.AddToggle("test", toggle)
		val, _ = panel.AddSlider("test2", val, 0, 1)
		panel.End()
		
		if !ui.IsRegisteringInput {
			ui.SetInputResponsive(uiResponsive)
		}

		app.DrawRect(gmath.Vec2{400,0}, gmath.Vec2{200,40}, gmath.Vec4{1,0,0,1})
		app.DrawText("TEST", &uiFont, gmath.Vec2{400,0}, gmath.Vec4{0,0,1,0}, gmath.Vec2{})
		
		app.DrawMesh(cube)
		app.Render()
		
		// Swappity-swap.
		window.SwapBuffers()
	}
}



var cubeVertices = [...]float32{
	// FRONT
	-0.5, -0.5, 0.5, 1.0,
	0.0, 0.0, 1.0, 0.0,
	-0.5, 0.5, 0.5, 1.0,
	0.0, 0.0, 1.0, 0.0,
	0.5, 0.5, 0.5, 1.0,
	0.0, 0.0, 1.0, 0.0,
	0.5, -0.5, 0.5, 1.0,
	0.0, 0.0, 1.0, 0.0,

	// LEFT
	-0.5, -0.5, 0.5, 1.0,
	-1.0, 0.0, 0.0, 0.0,
	-0.5, 0.5, 0.5, 1.0,
	-1.0, 0.0, 0.0, 0.0,
	-0.5, 0.5, -0.5, 1.0,
	-1.0, 0.0, 0.0, 0.0,
	-0.5, -0.5, -0.5, 1.0,
	-1.0, 0.0, 0.0, 0.0,

	// RIGHT
	0.5, -0.5, -0.5, 1.0,
	1.0, 0.0, 0.0, 0.0,
	0.5, 0.5, -0.5, 1.0,
	1.0, 0.0, 0.0, 0.0,
	0.5, 0.5, 0.5, 1.0,
	1.0, 0.0, 0.0, 0.0,
	0.5, -0.5, 0.5, 1.0,
	1.0, 0.0, 0.0, 0.0,

	// TOP
	-0.5, 0.5, 0.5, 1.0,
	0.0, 1.0, 0.0, 0.0,
	-0.5, 0.5, -0.5, 1.0,
	0.0, 1.0, 0.0, 0.0,
	0.5, 0.5, -0.5, 1.0,
	0.0, 1.0, 0.0, 0.0,
	0.5, 0.5, 0.5, 1.0,
	0.0, 1.0, 0.0, 0.0,

	// BOTTOM
	-0.5, -0.5, -0.5, 1.0,
	0.0, -1.0, 0.0, 0.0,
	-0.5, -0.5, 0.5, 1.0,
	0.0, -1.0, 0.0, 0.0,
	0.5, -0.5, 0.5, 1.0,
	0.0, -1.0, 0.0, 0.0,
	0.5, -0.5, -0.5, 1.0,
	0.0, -1.0, 0.0, 0.0,

	// BACK
	0.5, -0.5, -0.5, 1.0,
	0.0, 0.0, -1.0, 0.0,
	0.5, 0.5, -0.5, 1.0,
	0.0, 0.0, -1.0, 0.0,
	-0.5, 0.5, -0.5, 1.0,
	0.0, 0.0, -1.0, 0.0,
	-0.5, -0.5, -0.5, 1.0,
	0.0, 0.0, -1.0, 0.0,
}

var cubeIndices = [...]uint32{
	0, 1, 2,
	0, 2, 3,

	4, 5, 6,
	4, 6, 7,

	8, 9, 10,
	8, 10, 11,

	12, 13, 14,
	12, 14, 15,

	16, 17, 18,
	16, 18, 19,

	20, 21, 22,
	20, 22, 23,
}

