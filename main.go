package main

import (
	"strings"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"runtime"

	"github.com/go-gl/mathgl/mgl32"

	"./app"
	"github.com/janivanecky/golib/font"
	"github.com/janivanecky/golib/graphics"
	"github.com/janivanecky/golib/platform"
	"github.com/janivanecky/golib/ui"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func updateColorPalette(c chan []mgl32.Vec4) {
	newColors := app.GetRandomColorPalette()
	if newColors != nil {
		c <- newColors
	}
}

type cell struct {
	polar, angle, radius float64
	scale                mgl32.Vec3
	colorModifier        float32
	colorIndex           int
}

type RenderingSettings struct {
	DirectLight  float64
	AmbientLight float64

	Roughness    float64
	Reflectivity float64

	SSAORadius   float64
	SSAORange    float64
	SSAOBoundary float64

	MinWhite float64
}

type CellSettings struct {
	PolarStd, PolarMean    float64
	RadiusMin, RadiusMax float64
	HeightRatio            float64
	Count				   int
}

type AppSettings struct {
	Cells         CellSettings
	Rendering     RenderingSettings
	Camera        app.Camera
	Colors        []mgl32.Vec4
}

var settings = AppSettings{
	Cells: CellSettings{
		PolarStd: 0.00, PolarMean: math.Pi / 2.0,
		RadiusMin: 3.0, RadiusMax: 6.0,
		HeightRatio: 1.0,
		Count: 5000,
	},

	Rendering: RenderingSettings{
		DirectLight:  0.5,
		AmbientLight: 0.75,

		Roughness:    1.0,
		Reflectivity: 0.05,

		SSAORadius:   0.5,
		SSAORange:    3.0,
		SSAOBoundary: 1.0,

		MinWhite: 8.0,
	},

	Camera: app.GetCamera(45.0, 0.0, math.Pi/5.0, 5.0),

	Colors: []mgl32.Vec4{
		mgl32.Vec4{24 / 255.0, 193 / 255.0, 236 / 255.0, 1.0},
		mgl32.Vec4{0 / 255.0, 185 / 255.0, 121 / 255.0, 1.0},
		mgl32.Vec4{236 / 255.0, 24 / 255.0, 97 / 255.0, 1.0},
		mgl32.Vec4{33 / 255.0, 73 / 255.0, 83 / 255.0, 1.0},
		mgl32.Vec4{194 / 255.0, 55 / 255.0, 48 / 255.0, 1.0},
	},
}

func generateCells(cells []cell) {
	for i := range cells {
		scaleX := rand.Float32()*0.8 + 0.2
		scaleZ := rand.Float32()*0.8 + 0.2
		scaleX *= scaleX * 4.0 / 3.0
		scaleZ *= scaleZ * 2.0
		scaleX *= 2.0
		scaleZ *= 2.0
		scaleY := scaleX + scaleZ

		polar := rand.NormFloat64()
		angle := rand.Float64() * math.Pi * 2

		radius := rand.Float64()

		//radius *= radius
		scale := mgl32.Vec3{scaleX, scaleY, scaleZ}
		colorModifier := rand.Float32()*0.5 + 0.5
		colorIndex := rand.Int() % 5
		cells[i] = cell{polar, angle, radius, scale, colorModifier, colorIndex}
	}
}

var SAVES_DIR = "saves"
func loadSettings() {
	serializedSettings, err := ioutil.ReadFile(SAVES_DIR + "/settings")
	if err != nil {
		return//panic(err)
	}
	err = json.Unmarshal(serializedSettings, &settings)
	if err != nil {
		panic(err)
	}
}

func saveSettings() {
	serializedSettings, err := json.Marshal(settings)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(SAVES_DIR + "/settings", serializedSettings, 0644)
	if err != nil {
		panic(err)
	}
}

func getCircle(radius float64, width float64, arch float64) ([]float32, []uint32) {
	pointCount := uint32(128)

	vertices := make([]float32, 0)
	indices := make([]uint32, 0)

	archLength := math.Pi / 4.0
	for i := uint32(0); i < pointCount; i++ {
		//angle := float64(i) * math.Pi * 2.0 / float64(pointCount)
		angle := float64(i) * archLength / float64(pointCount) + arch - archLength / 2.0
		xInner := float32(math.Sin(angle) * (radius - width))
		zInner := float32(math.Cos(angle) * (radius - width))
		xOuter := float32(math.Sin(angle) * (radius + width))
		zOuter := float32(math.Cos(angle) * (radius + width))
		y := float32(0.0)

		vertices = append(vertices, xInner, y, zInner, 1.0)
		vertices = append(vertices, 0.0, 1.0, 0.0, 1.0)
		vertices = append(vertices, xOuter, y, zOuter, 1.0)
		vertices = append(vertices, 0.0, 1.0, 0.0, 1.0)

		index := 2 * i
		if i < pointCount - 1 {
			indices = append(indices, index, index+1, index+3)
			indices = append(indices, index, index+3, index+2)
		} else {
			//indices = append(indices, index, index+1, 1)
			//indices = append(indices, index, 1, 0)
		}
	}

	return vertices, indices
}

type Circle struct {
	color mgl32.Vec4
	targetColor mgl32.Vec4
	width float64
	targetWidth float64
	radius float64
	targetRadius float64
	arch float64
	targetArch float64
}

func (c *Circle) getRadialDistanceToInnerCircle(radius float64) float64 {
	return math.Abs(radius - (c.radius - 4.0))
}

func (c *Circle) getRadialDistanceToOuterCircle(radius float64) float64 {
	return math.Abs(radius - (c.radius + 4.0))
}

func (c *Circle) getAsInnerCircle() ([]float32, []uint32) {
	radius := c.radius - 4.0
	circleVertices, circleIndices := getCircle(radius, c.width, c.arch)
	return circleVertices, circleIndices
}

func (c *Circle) getAsOuterCircle() ([]float32, []uint32) {
	radius := c.radius + 4.0
	circleVertices, circleIndices := getCircle(radius, c.width, c.arch)
	return circleVertices, circleIndices
}

func (c *Circle) setWidth(targetWidth float64) {
	c.targetWidth = targetWidth
}

func (c *Circle) setRadius(targetRadius float64) {
	c.targetRadius = targetRadius
}

func (c *Circle) setColor(targetColor mgl32.Vec4) {
	c.targetColor = targetColor
}

func (c *Circle) setArch(targetArch float64) {
	if targetArch - c.arch > math.Pi {
		targetArch -= math.Pi * 2.0
	} else if c.arch - targetArch > math.Pi {
		targetArch += math.Pi * 2.0
	}
	c.targetArch = targetArch
}

func (c *Circle) update(dt float64) {
	c.width += (c.targetWidth - c.width) * 10.0 * dt
	c.radius += (c.targetRadius - c.radius) * 15.0 * dt
	c.color[0] += (c.targetColor[0] - c.color[0]) * 5.0 * float32(dt)
	c.color[1] += (c.targetColor[1] - c.color[1]) * 5.0 * float32(dt)
	c.color[2] += (c.targetColor[2] - c.color[2]) * 5.0 * float32(dt)
	c.color[3] += (c.targetColor[3] - c.color[3]) * 5.0 * float32(dt)
	c.arch += (c.targetArch - c.arch) * math.Pi * 6.0 * dt
	if c.arch > math.Pi * 2.0 && c.targetArch > math.Pi * 2.0 {
		c.arch -= math.Pi * 2.0
		c.targetArch -= math.Pi * 2.0
	}
	if c.arch < 0 && c.targetArch < 0 {
		c.arch += math.Pi * 2.0
		c.targetArch += math.Pi * 2.0
	}
}

func radiusToRadiusMin(radius float64) float64 {
	return radius + 4.0
}

func radiusToRadiusMax(radius float64) float64 {
	return radius - 4.0
}

func isInRect(position mgl32.Vec2, rectPosition mgl32.Vec2, rectSize mgl32.Vec2) bool {
    if position[0] >= rectPosition[0] && position[0] <= rectPosition[0] + rectSize[0] &&
       position[1] >= rectPosition[1] && position[1] <= rectPosition[1] + rectSize[1] {
		   return true
	   }
    return false
}

func main() {
	loadSettings()
	var windowWidth = 800
	var windowHeight = 600
	windowWidth, windowHeight = platform.GetMonitorResolution()
	window := platform.GetWindow(windowWidth, windowHeight, "New fancy window", true)
	defer platform.ReleaseWindow()

	// Create screenshot dir
	SCREENSHOT_DIR := "screenshots"
	os.Mkdir(SCREENSHOT_DIR, os.ModeDir)
	files, err := ioutil.ReadDir(SCREENSHOT_DIR)
	maxScreenshotNum := 0
	for _, file := range files {
		fileNumString := strings.Split(file.Name(), ".")[0]
		fileNum, _ := strconv.Atoi(fileNumString)

		if maxScreenshotNum < fileNum {
			maxScreenshotNum = fileNum
		}
	}

	// Create save dir
	os.Mkdir(SAVES_DIR, os.ModeDir)
	files, err = ioutil.ReadDir(SAVES_DIR)

	//truetypeBytes, err := ioutil.ReadFile("fonts/DidactGothic-Regular.ttf")
	//truetypeBytes, err := ioutil.ReadFile("fonts/Montserrat-Regular.ttf")
	truetypeTitleBytes, err := ioutil.ReadFile("fonts/Roboto-Medium.ttf")
	if err != nil {
		panic(err)
	}
	truetypeNormalBytes, err := ioutil.ReadFile("fonts/Roboto-Regular.ttf")
	if err != nil {
		panic(err)
	}
	//truetypeBytes, err := ioutil.ReadFile("fonts/JuliusSansOne-Regular.ttf")

	scale := platform.GetWindowScaling()
	//uiFont := font.GetFont(truetypeNormalBytes, 22.0, scale)
	uiFont := font.GetFont(truetypeNormalBytes, 20.0, scale)
	uiFontTitle := font.GetFont(truetypeTitleBytes, 34.0, scale)
	infoFont := font.GetFont(truetypeNormalBytes, 32.0, scale)
	app.InitRendering(float64(windowWidth), float64(windowHeight), uiFont, uiFontTitle)
	settings.Camera.Polar = 0.0
	app.SetCamera(&settings.Camera)

	colorChannel := make(chan []mgl32.Vec4, 1)

	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:], []int{4, 4})
	cells := make([]cell, 10000)
	generateCells(cells)

	circleWidth := 0.75
	circleWidthHover := 1.5
	circleColor := mgl32.Vec4{0.0, 0.0, 0.0, 0.2}
	circleColorHover := mgl32.Vec4{0.0, 0.0, 0.0, 0.5}
	circleColorFade := mgl32.Vec4{0.0, 0.0, 0.0, 0.01}
	circleFadeOutTime := 0.5
	
	innerCircle := Circle{circleColor, circleColor, circleWidth, circleWidth, settings.Cells.RadiusMin, settings.Cells.RadiusMin, 0, 0}
	circleVertices, circleIndices := innerCircle.getAsInnerCircle()
	circleInner := graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
	outerCircle := Circle{circleColor, circleColor, circleWidth, circleWidth, settings.Cells.RadiusMax, settings.Cells.RadiusMax, 0, 0}
	circleVertices, circleIndices = outerCircle.getAsOuterCircle()
	circleOuter := graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
	
	pickerStates := make([]bool, len(settings.Colors))

	showUI := true

	start := time.Now()
	t := 0.0

	screenshotTextTimer := 0.0
	screenshotTextDuration := 1.75
	screenshotTextFadeDuration := 1.0

	circleInnerChanging := false
	circleOuterChanging := false

	timeSinceMouseMovement := 0.0

	countSliderColor := circleColor
	countSliderColorTarget := circleColor
	countSliderChanging := false

	// Start our fancy-shmancy loop
	for !window.ShouldClose() {

		now := time.Now()
		dt := now.Sub(start).Seconds()
		start = now

		t += dt

		fps := 0
		if dt > 0.0 {
			fps = int(1.0 / dt)
		}

		platform.Update(window)

		// Let's quit if user presses Esc, that cannot mean anything else.
		if platform.IsKeyPressed(platform.KeyEscape) {
			break
		}
		if platform.IsKeyPressed(platform.KeyF2) {
			showUI = !showUI
		}
		if platform.IsKeyPressed(platform.KeyR) {
			generateCells(cells)
		}
		if platform.IsKeyPressed(platform.KeyC) {
			go updateColorPalette(colorChannel)
		}
		if platform.IsKeyPressed(platform.KeyF10) {
			screenshotTextTimer = screenshotTextDuration
			
			imageBytes, imageWidth, imageHeight := app.GetSceneBuffer()
			fmt.Println(imageWidth, imageHeight, windowWidth, windowHeight)
			stride := len(imageBytes) / int(imageHeight)
			invertedBytes := make([]byte, len(imageBytes))
			for i := 0; i < int(imageHeight); i++ {
				srcRowStart := i * stride
				srcRowEnd := (i + 1) * stride
				dstRowStart := (int(imageHeight) - 1 - i) * stride
				dstRowEnd := (int(imageHeight) - i) * stride
				copy(invertedBytes[dstRowStart:dstRowEnd], imageBytes[srcRowStart:srcRowEnd])
			}

			img := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))
			img.Pix = invertedBytes

			maxScreenshotNum++
			f, err := os.Create(SCREENSHOT_DIR + "/" + strconv.Itoa(maxScreenshotNum) + ".jpg")
			if err != nil {
				panic(err)
			}

			defer f.Close()
			jpeg.Encode(f, img, nil)
		}

		if showUI {
			if false {
				// Scene control.
				panel := ui.StartPanel("Scene", mgl32.Vec2{10, 10}, 450)
				radiusMinChanged, radiusMaxChanged := false, false
				settings.Cells.RadiusMin, radiusMinChanged = panel.AddSlider("RadiusMin", settings.Cells.RadiusMin, 0, 100.0)
				if radiusMinChanged {
					circleVertices, circleIndices := getCircle(settings.Cells.RadiusMin - 2.0, circleWidth, 0.0)
					circleInner = graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
				}
				settings.Cells.RadiusMax, radiusMaxChanged = panel.AddSlider("RadiusMax", settings.Cells.RadiusMax, 0, 100.0)
				if radiusMaxChanged {
					circleVertices, circleIndices := getCircle(settings.Cells.RadiusMax + 2.0 + circleWidth, circleWidth, 0.0)
					circleOuter = graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
				}
				settings.Cells.PolarStd, _ = panel.AddSlider("PolarStd", settings.Cells.PolarStd, 0, math.Pi)
				settings.Cells.HeightRatio, _ = panel.AddSlider("HeightRatio", settings.Cells.HeightRatio, 0, 1.0)
	
				panel.End()
			}

			// SSAO related settings.
			//panel = ui.StartPanel("Rendering", mgl32.Vec2{10, panel.GetBottom() + 00}, 450)
			panel := ui.StartPanel("Rendering", mgl32.Vec2{10, 10}, 450)
			cellCountFloat, _ := panel.AddSlider("CellCount", float64(settings.Cells.Count), 0, 10000)
			settings.Cells.Count = int(cellCountFloat)
			settings.Rendering.DirectLight, _ = panel.AddSlider("DirectLight", settings.Rendering.DirectLight, 0, 5.0)
			settings.Rendering.AmbientLight, _ = panel.AddSlider("AmbientLight", settings.Rendering.AmbientLight, 0, 5.0)
			settings.Rendering.MinWhite, _ = panel.AddSlider("MinWhite", settings.Rendering.MinWhite, 0, 20.0)
			settings.Rendering.SSAORadius, _ = panel.AddSlider("SSAORadius", settings.Rendering.SSAORadius, 0, 1.0)
			settings.Rendering.SSAORange, _ = panel.AddSlider("SSAORange", settings.Rendering.SSAORange, 0, 10.0)
			settings.Rendering.SSAOBoundary, _ = panel.AddSlider("SSAOBoundary", settings.Rendering.SSAOBoundary, 0, 10.0)
			panel.End()

			// Colors/material related settings.
			nextWidth := panel.GetWidth()
			panel = ui.StartPanel("Material", mgl32.Vec2{float32(windowWidth) - 10 - nextWidth, 10}, float64(nextWidth))
			settings.Rendering.Roughness, _ = panel.AddSlider("Roughness", settings.Rendering.Roughness, 0, 1.0)
			settings.Rendering.Reflectivity, _ = panel.AddSlider("Reflectivity", settings.Rendering.Reflectivity, 0, 1.0)
			for i := range settings.Colors {
				pickerStates[i], _ = panel.AddColorPalette("Color"+strconv.Itoa(i), settings.Colors[i], pickerStates[i])
				if pickerStates[i] {
					settings.Colors[i], _ = panel.AddColorPicker("Pick"+strconv.Itoa(i), settings.Colors[i], false)
				}
			}
			panel.End()

		}
		fpsString := fmt.Sprintf("%d", fps)
		app.DrawText(fpsString, &infoFont, mgl32.Vec2{float32(windowWidth) - 10, float32(windowHeight) - 10}, mgl32.Vec4{0, 0, 0, 0.8}, mgl32.Vec2{1, 1})
		//app.DrawText("v0.1", &infoFont, mgl32.Vec2{float32(10), float32(windowHeight) - 10}, mgl32.Vec4{0.0,0.0,0.0,0.8}, mgl32.Vec2{0,1})
		app.DrawText("v0.1", &infoFont, mgl32.Vec2{float32(10), float32(windowHeight) - 10}, mgl32.Vec4{0.0, 0.0, 0.0, 0.8}, mgl32.Vec2{0, 1})
		
		// Show screenshot text.
		screenshotTextPart := screenshotTextTimer / screenshotTextFadeDuration
		alpha := float32(math.Sqrt(screenshotTextPart))
		if alpha > 1.0 {
			alpha = 1.0
		}
		if alpha > 0.0 {
			app.DrawText("IMAGE SAVED", &infoFont, mgl32.Vec2{float32(windowWidth) / 2.0, float32(windowHeight) - 10}, mgl32.Vec4{0.0, 0.0, 0.0, alpha * 0.8}, mgl32.Vec2{0.5, 1.0})
		}
		if screenshotTextTimer > 0.0 {
			screenshotTextTimer -= dt
		}

		settings.Camera.Update(dt)
		app.SetCamera(&settings.Camera)

		
		
		mouseX, mouseY := platform.GetMousePosition()
		rS, rD := app.GetWorldRay(mouseX, mouseY)
		s := 0.0 - rS.Y() / rD.Y()
		pos := rS.Add(rD.Mul(s))
		radius := math.Sqrt(float64(pos.X() * pos.X() + pos.Z() * pos.Z()))
		
		mdx, mdy := platform.GetMouseDeltaPosition()
		if math.Abs(mdx) < 0.001 && math.Abs(mdy) < 0.001 {
			timeSinceMouseMovement += dt
		} else {
			timeSinceMouseMovement = 0.0
		}

		if circleInnerChanging {
			if !platform.IsMouseLeftButtonDown() {
				circleInnerChanging = false
				innerCircle.setColor(circleColor)
				innerCircle.setWidth(circleWidth)
			}
			newRadius := math.Min(radiusToRadiusMin(radius), outerCircle.targetRadius)
			innerCircle.setRadius(newRadius)
		}
		if !circleInnerChanging {
			if innerCircle.getRadialDistanceToInnerCircle(radius) < 4.0{
				innerCircle.setColor(circleColorHover)
				innerCircle.setWidth(circleWidthHover)
				if  platform.IsMouseLeftButtonPressed() { 
					circleInnerChanging = true
				}
			}  else {
				if timeSinceMouseMovement > circleFadeOutTime {
					innerCircle.setColor(circleColorFade)
				} else {
					innerCircle.setColor(circleColor)
				}
				innerCircle.setWidth(circleWidth)
			}
		}
		settings.Cells.RadiusMin = innerCircle.radius
		innerCircle.update(dt)
		
		if circleOuterChanging {
			if !platform.IsMouseLeftButtonDown() {
				circleOuterChanging = false
				outerCircle.setColor(circleColor)
				outerCircle.setWidth(circleWidth)
			}
			newRadius := math.Max(radiusToRadiusMax(radius), innerCircle.targetRadius)
			outerCircle.setRadius(newRadius)
		}
		if !circleOuterChanging {
			if outerCircle.getRadialDistanceToOuterCircle(radius) < 4.0 {
				outerCircle.setColor(circleColorHover)
				outerCircle.setWidth(circleWidthHover)
				if platform.IsMouseLeftButtonPressed() {
					circleOuterChanging = true
				}
			} else {
				if timeSinceMouseMovement > circleFadeOutTime {
					outerCircle.setColor(circleColorFade)
				} else {
					outerCircle.setColor(circleColor)
				}
				outerCircle.setWidth(circleWidth)
			}
		}
		outerCircle.update(dt)
		settings.Cells.RadiusMax = outerCircle.radius

		//
		countSliderWidth := float32(100.0)
		countSliderBgSize := mgl32.Vec2{countSliderWidth, float32(windowHeight) * 0.7}
		countSliderBgPos := mgl32.Vec2{float32(windowWidth) - countSliderBgSize[0] - 100, float32(windowHeight) * 0.15}
		
		
		portion := float32(settings.Cells.Count) / 10000.0
		if countSliderChanging {
			portion = 1.0 - (float32(mouseY) - countSliderBgPos[1]) / countSliderBgSize[1]
			if portion < 0.0 {
				portion = 0.0
			} else if portion > 1.0 {
				portion = 1.0
			}
			settings.Cells.Count = int(portion * 10000.0)
			countSliderColorTarget = circleColorHover
			if !platform.IsMouseLeftButtonDown() {
				countSliderChanging = false
			}
		}
		if !countSliderChanging {
			if isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, countSliderBgPos, countSliderBgSize) {
				countSliderColorTarget = circleColorHover
				if platform.IsMouseLeftButtonPressed() {
					countSliderChanging = true
				}
			} else {
				countSliderColorTarget = circleColor
				if timeSinceMouseMovement > circleFadeOutTime {
					countSliderColorTarget = circleColorFade
				}
			}
		}
			
		countSliderSize := mgl32.Vec2{countSliderWidth, countSliderBgSize[1] * portion}
		countSliderPos := mgl32.Vec2{countSliderBgPos[0], countSliderBgPos[1] + countSliderBgSize[1] - countSliderSize[1]}

		countSliderColor[0] += (countSliderColorTarget[0] - countSliderColor[0]) * 5.0 * float32(dt)
		countSliderColor[1] += (countSliderColorTarget[1] - countSliderColor[1]) * 5.0 * float32(dt)
		countSliderColor[2] += (countSliderColorTarget[2] - countSliderColor[2]) * 5.0 * float32(dt)
		countSliderColor[3] += (countSliderColorTarget[3] - countSliderColor[3]) * 5.0 * float32(dt)

		app.DrawRect(countSliderBgPos, countSliderBgSize, countSliderColor)
		app.DrawRect(countSliderPos, countSliderSize, countSliderColor)
		
		//

		app.DirectLight = settings.Rendering.DirectLight
		app.AmbientLight = settings.Rendering.AmbientLight
		app.Roughness = settings.Rendering.Roughness
		app.Reflectivity = settings.Rendering.Reflectivity
		app.SSAORadius = settings.Rendering.SSAORadius
		app.SSAORange = settings.Rendering.SSAORange
		app.SSAOBoundary = settings.Rendering.SSAOBoundary
		app.MinWhite = settings.Rendering.MinWhite

		select {
		case settings.Colors = <-colorChannel:
		default:
		}

		radiusMin := settings.Cells.RadiusMin
		radiusMax := settings.Cells.RadiusMax
		polarStd := settings.Cells.PolarStd
		polarMean := settings.Cells.PolarMean
		heightRatio := settings.Cells.HeightRatio
		
		for _, cell := range cells[:settings.Cells.Count] {
			color := settings.Colors[cell.colorIndex].Mul(cell.colorModifier)

			polar := cell.polar*polarStd + polarMean
			angle := cell.angle

			radius := cell.radius*(radiusMax - radiusMin) + radiusMin

			x := float32(math.Sin(polar) * math.Sin(angle) * radius)
			y := float32(math.Cos(polar) * radius * heightRatio)
			z := float32(math.Sin(polar) * math.Cos(angle) * radius)

			modelMatrix := mgl32.LookAt(
				x, y, z,
				0, 0, 0,
				0, 1, 0,
			).Inv().Mul4(
				mgl32.Scale3D(cell.scale[0], cell.scale[1], cell.scale[2]),
			)

			app.DrawMesh(cube, modelMatrix, color)
		}

		mouseAngle := math.Atan2(float64(pos.X()), float64(pos.Z()))
		innerCircle.setArch(mouseAngle)
		outerCircle.setArch(mouseAngle)
		circleVertices, circleIndices := innerCircle.getAsInnerCircle()
		circleInner = graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})

		modelMatrix := mgl32.Ident4()
		app.DrawMeshUI(circleInner, modelMatrix, innerCircle.color)

		circleVertices, circleIndices = outerCircle.getAsOuterCircle()
		circleOuter = graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
		
		app.DrawMeshUI(circleOuter, modelMatrix, outerCircle.color)
		app.Render()

		// Swappity-swap.
		window.SwapBuffers()
	}
	saveSettings()
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
