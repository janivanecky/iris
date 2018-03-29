package main

import (
	"io/ioutil"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	
	"./lib/graphics"
	"./lib/font"
	"./lib/ui"
	"./lib/platform"
	"./app"
)

func vecFromPolarCoords(azimuth float64, polar float64, radius float64) mgl32.Vec3 {
	result := mgl32.Vec3{
		float32(math.Sin(polar) * math.Sin(azimuth) * radius),
		float32(math.Cos(polar) * radius),
		float32(math.Sin(polar) * math.Cos(azimuth) * radius),
	}
	return result
}

const windowWidth = 800
const windowHeight = 600

func main() {
	window := platform.GetWindow(windowWidth, windowHeight, "New fancy window")
	defer platform.ReleaseWindow()
	
	truetypeBytes, err := ioutil.ReadFile("fonts/font.ttf")
	if err != nil {
		panic(err)
    }
	
	scale := platform.GetWindowScaling()
    uiFont := font.GetFont(truetypeBytes, 20.0, scale)
	app.InitRendering(windowWidth, windowHeight, uiFont)
	
	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:], []int{4, 4})

	polar := math.Pi / 2.0
	azimuth := 0.0
	radius := 5.0

	toggle := true
	val := 0.5

	start := time.Now()

	camPosition := vecFromPolarCoords(azimuth, polar, radius)
	app.SetCameraPosition(camPosition)

	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		t := time.Now()
		_ = t.Sub(start)
		start = t

		platform.Update(window)

		// Let's quit if user presses Esc, that cannot mean anything else.
		if platform.IsEscPressed() {
			break
		}

		panel := ui.StartPanel("Test panel", mgl32.Vec2{})
		toggle, _ = panel.AddToggle("test", toggle)
		val, _ = panel.AddSlider("test2", val, 0, 1)
		panel.End()
		
		camChanged := false
		// Update mouse position and get position delta.
		dx, dy := platform.GetMouseDeltaPosition()
		if !ui.IsRegisteringInput {
			if platform.IsMouseLeftButtonDown() {
				azimuth -= dx / 100.0
				polar -= dy / 100.0
				camChanged = true
				ui.SetInputResponsive(false)
			} else {
				ui.SetInputResponsive(true)
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

		app.DrawRect(mgl32.Vec2{400,0}, mgl32.Vec2{200,40}, mgl32.Vec4{1,0,0,1})
		app.DrawText("TEST", &uiFont, mgl32.Vec2{400,0}, mgl32.Vec4{0,0,1,0}, mgl32.Vec2{})
		
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

