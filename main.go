package main


import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"time"
	"net/http"
	"encoding/json"
	"runtime"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/janivanecky/golib/graphics"
	"github.com/janivanecky/golib/font"
	"github.com/janivanecky/golib/ui"
	"github.com/janivanecky/golib/platform"
	"./app"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func vecFromPolarCoords(azimuth float64, polar float64, radius float64) mgl32.Vec3 {
	result := mgl32.Vec3{
		float32(math.Sin(polar) * math.Sin(azimuth) * radius),
		float32(math.Cos(polar) * radius),
		float32(math.Sin(polar) * math.Cos(azimuth) * radius),
	}
	return result
}

var windowWidth = 1400
var windowHeight = 700

func updateColorPalette(c chan []mgl32.Vec4) {
	newColors := getRandomColorPalette()
	c <- newColors
}

var responseValue = make(map[string]([5][]int))
var colors [5]mgl32.Vec4
var data = []byte (`{"model": "default"}`)
func getRandomColorPalette() [] mgl32.Vec4{
	res, err := http.Post("http://colormind.io/api/", "text/json", bytes.NewBuffer(data))
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &responseValue)
	if err != nil {
		panic(err)
	}

	colorInt := responseValue["result"]
	for i, color := range colorInt {
		colors[i] = mgl32.Vec4{
			float32(color[0]) / 255.0,
			float32(color[1]) / 255.0,
			float32(color[2]) / 255.0,
			float32(1.0),
		}
	}

	return colors[:]
}

type cell struct {
	position mgl32.Vec3
	scale mgl32.Vec3
	colorModifier float32
	colorIndex int
}

func generateCells(cells []cell) {
	for i := range cells {
		scaleX := rand.Float32() * 0.8 + 0.2
		scaleY := rand.Float32() * 0.2 + 0.8
		scaleZ := rand.Float32() * 0.8 + 0.2
		scaleY *= scaleY * 2.0 / 3.0
		scaleX *= scaleX * 4.0 / 3.0
		scaleZ *= scaleZ * 2.0
		scaleX *= 2.0
		scaleZ *= 2.0
		scaleY = scaleX + scaleZ
		
		angle := rand.Float64() * math.Pi * 2
		radius := rand.Float64() * 6.0 + 3.0
		radius *= radius
		polar := rand.NormFloat64() * 0.2 + math.Pi / 2.0
		x := float32(math.Sin(polar) * math.Sin(angle) * radius)
		y := float32(math.Cos(polar) * radius * 0.25)
		z := float32(math.Sin(polar) * math.Cos(angle) * radius)

		position := mgl32.Vec3{x,y,z}
		scale := mgl32.Vec3{scaleX, scaleY, scaleZ}
		colorModifier := rand.Float32() * 0.5 + 0.5
		colorIndex := rand.Int() % 5
		cells[i] = cell{position, scale, colorModifier, colorIndex}
	}
}

func main() {
	aspectRatio := float64(windowWidth) / float64(windowHeight)
	window := platform.GetWindow(windowWidth, windowHeight, "New fancy window")
	defer platform.ReleaseWindow()
	
	truetypeBytes, err := ioutil.ReadFile("fonts/DidactGothic-Regular.ttf")
	if err != nil {
		panic(err)
    }
	
	scale := platform.GetWindowScaling()
    uiFont := font.GetFont(truetypeBytes, 22.0, scale)
	app.InitRendering(float64(windowWidth), float64(windowHeight), uiFont)
	
	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:], []int{4, 4})

	polar := math.Pi / 5.0
	dPolar := 0.0
	azimuth := 0.0
	dAzimuth := 0.0
	radius := 45.0
	dRadius := 0.0

	colorChannel := make(chan []mgl32.Vec4, 1)
	colorful := []mgl32.Vec4 {
		mgl32.Vec4{24 / 255.0, 193 / 255.0, 236 / 255.0, 1.0},
		mgl32.Vec4{0 / 255.0, 185 / 255.0, 121 / 255.0, 1.0},
		mgl32.Vec4{236 / 255.0, 24 / 255.0, 97 / 255.0, 1.0},
		mgl32.Vec4{33 / 255.0, 73 / 255.0, 83 / 255.0, 1.0},
		mgl32.Vec4{194 / 255.0, 55 / 255.0, 48 / 255.0, 1.0},
	}
	
	cells := make([]cell, 2400)
	generateCells(cells)

	camPosition := vecFromPolarCoords(azimuth, polar, radius)
	app.SetCameraPosition(camPosition, mgl32.Vec3{0,1,0})
	rad := float32(1.0)
	pickerStates := make([]bool, len(colorful))

	showUI := true

	start := time.Now()
	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		t := time.Now()
		dt := t.Sub(start).Seconds()
		start = t
		fps := 0
		if dt > 0.0 {
			fps = int(1.0 / dt)
		}
		fpsString := fmt.Sprintf("%d", fps)
		app.DrawText(fpsString, &uiFont, mgl32.Vec2{float32(windowWidth), float32(windowHeight * 2)}, mgl32.Vec4{1,1,1,1}, mgl32.Vec2{1,1})

		platform.Update(window)

		// Let's quit if user presses Esc, that cannot mean anything else.
		if platform.IsEscDown() {
			break
		}
		if platform.IsF2Pressed() {
			showUI = !showUI
		}
		if platform.IsKeyPressed(platform.KeyR) {
			generateCells(cells)
		}

		if platform.IsKeyPressed(platform.KeyC) {
			go updateColorPalette(colorChannel)
		}

		if showUI {
			panel := ui.StartPanel("Control", mgl32.Vec2{10, 10}, 0)
			rad64 := float64(rad)
			rad64, _ = panel.AddSlider("Radius", rad64, 0, 5.0)
			rad = float32(rad64)
			app.DirectLight, _ = panel.AddSlider("DirectLight", app.DirectLight, 0, 5.0)
			app.AmbientLight, _ = panel.AddSlider("AmbientLight", app.AmbientLight, 0, 5.0)
			app.MinWhite, _ = panel.AddSlider("MinWhite", app.MinWhite, 0, 20.0)
			panel.End()
			
			panel = ui.StartPanel("SSAO", mgl32.Vec2{10, panel.GetBottom() + 10}, 0)
			app.SSAORadius, _ = panel.AddSlider("SSAORadius", app.SSAORadius, 0, 1.0)
			app.SSAORange, _ = panel.AddSlider("SSAORange", app.SSAORange, 0, 10.0)
			app.SSAOBoundary, _ = panel.AddSlider("SSAOBoundary", app.SSAOBoundary, 0, 10.0)
			panel.End()
	
			nextWidth := panel.GetWidth()
			panel = ui.StartPanel("Material", mgl32.Vec2{float32(windowWidth) - 10 - nextWidth, 10}, float64(nextWidth))
			app.Roughness, _ = panel.AddSlider("Roughness", app.Roughness, 0, 1.0)
			app.Reflectivity, _ = panel.AddSlider("Reflectivity", app.Reflectivity, 0, 1.0)
			for i := range colorful {
				pickerStates[i], _ = panel.AddColorPalette("Color" + strconv.Itoa(i), colorful[i], pickerStates[i])
				if pickerStates[i] {
					colorful[i], _ = panel.AddColorPicker("Pick" + strconv.Itoa(i), colorful[i], false)
				}
			}
			panel.End()
		}
		
		// Update mouse position and get position delta.
		dx, dy := platform.GetMouseDeltaPosition()
		if !ui.IsRegisteringInput {
			if platform.IsMouseLeftButtonDown() {
				dAzimuth -= dx / 100.0
				dPolar -= dy / 100.0
				ui.SetInputResponsive(false)
			} else {
				ui.SetInputResponsive(true)
			}
				
			mouseWheelDelta := platform.GetMouseWheelDelta()
			if mouseWheelDelta != 0.0 {
				dRadius -= mouseWheelDelta / 2.0
			}
		}
			
		speed := 5.0
		azimuth += dAzimuth * dt * speed
		polar += dPolar * dt * speed * aspectRatio
		polar = math.Max(polar, 0)
		radius += dRadius * dt * speed * 2.0

		dPolar *= 0.8
		dAzimuth *= 0.8
		dAzimuth = math.Max(dAzimuth, 0.1)
		dRadius *= 0.9
		
		camPosition := vecFromPolarCoords(azimuth, polar, radius)
		up := mgl32.Vec3{0, 1, 0}
		if camPosition.Normalize().Dot(up) > 0.9 {
			up = vecFromPolarCoords(azimuth + math.Pi, math.Pi / 2.0, 1.0)
		}
		app.SetCameraPosition(camPosition, up)
		
		for _, cell := range cells {
			color := colorful[cell.colorIndex].Mul(cell.colorModifier)
			position := cell.position.Mul(rad)
			modelMatrix := mgl32.LookAt(
				position[0], position[1], position[2],
				0, 0, 0,
				0, 1, 0,
				).Inv().Mul4(
					mgl32.Scale3D(cell.scale[0], cell.scale[1], cell.scale[2]),
				)
				
			app.DrawMesh(cube, modelMatrix, color)
		}
		app.Render()
			
		// Swappity-swap.
		window.SwapBuffers()
		select {
		case colorful = <-colorChannel:
		default:
		}
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

