package main

import (
	"strings"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"sort"
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

type CellSettings struct {
	PolarStd, PolarMean    float64
	RadiusMin, RadiusMax   float64
	HeightRatio            float64
	Count				   int
}

type AppSettings struct {
	Cells         CellSettings
	Rendering     app.RenderingSettings
	Camera        app.Camera
	Colors        []mgl32.Vec4
}

func copySettings(settings *AppSettings) AppSettings {
	newSettings := AppSettings{}
	newSettings.Cells = settings.Cells
	newSettings.Rendering = settings.Rendering
	newSettings.Camera = settings.Camera
	newSettings.Colors = make([]mgl32.Vec4, len(settings.Colors))
	copy(newSettings.Colors, settings.Colors)
	return newSettings
}

var defaultSettings = AppSettings{
	Cells: CellSettings{
		PolarStd: 0.00, PolarMean: math.Pi / 2.0,
		RadiusMin: 3.0, RadiusMax: 15.0,
		HeightRatio: 1.0,
		Count: 5000,
	},

	Rendering: app.RenderingSettings{
		DirectLight:  0.5,
		AmbientLight: 0.75,

		Roughness:    1.0,
		Reflectivity: 0.05,

		SSAORadius:   0.5,
		SSAORange:    3.0,
		SSAOBoundary: 1.0,

		MinWhite: 8.0,
	},

	Camera: app.GetCamera(100.0, 0.0, 0.0, 5.0),

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
func loadSettings(path string) AppSettings {
	settings := defaultSettings
	newColors := make([]mgl32.Vec4, len(settings.Colors))
	copy(newColors, settings.Colors)
	settings.Colors = newColors
	serializedSettings, err := ioutil.ReadFile(path)
	if err != nil {
		return settings//panic(err)
	}
	err = json.Unmarshal(serializedSettings, &settings)
	if err != nil {
		panic(err)
	}
	return settings
}

func saveSettings(path string, settings AppSettings) {
	serializedSettings, err := json.Marshal(settings)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path, serializedSettings, 0644)
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
		}
	}

	return vertices, indices
}

type Parameter interface {
	Update(dt, speed float64)
}

type ColorParameter struct {
	val, target mgl32.Vec4
}

func (color *ColorParameter) Update(dt, speed float64) {
	color.val[0] += (color.target[0] - color.val[0]) * float32(dt * speed)
	color.val[1] += (color.target[1] - color.val[1]) * float32(dt * speed)
	color.val[2] += (color.target[2] - color.val[2]) * float32(dt * speed)
	color.val[3] += (color.target[3] - color.val[3]) * float32(dt * speed)
}

type FloatParameter struct {
	val, target float64
}

func (value *FloatParameter) Update(dt, speed float64) {
	value.val += (value.target - value.val) * dt * speed
}

type RadianParameter struct {
	val, target float64
}

func (value *RadianParameter) Update(dt, speed float64) {
	if value.target - value.val > math.Pi {
		value.target -= math.Pi * 2.0
	} else if value.val - value.target > math.Pi {
		value.target += math.Pi * 2.0
	}
	value.val += (value.target - value.val) * dt * speed
	if value.val > math.Pi * 2.0 && value.target > math.Pi * 2.0 {
		value.val -= math.Pi * 2.0
		value.target -= math.Pi * 2.0
	}
	if value.val < 0 && value.target < 0 {
		value.val += math.Pi * 2.0
		value.target += math.Pi * 2.0
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

func drawCells(cells []cell, cellsSettings CellSettings, colors []mgl32.Vec4, mesh graphics.Mesh) {
	radiusMin := radiusToRadiusMin(cellsSettings.RadiusMin)
	radiusMax := radiusToRadiusMax(cellsSettings.RadiusMax)
	polarStd := cellsSettings.PolarStd
	polarMean := cellsSettings.PolarMean
	heightRatio := cellsSettings.HeightRatio

	matrices := make([]mgl32.Mat4, cellsSettings.Count)
	colorsInstanced := make([]mgl32.Vec4, cellsSettings.Count)
	
	for i, cell := range cells[:cellsSettings.Count] {
		color := colors[cell.colorIndex].Mul(cell.colorModifier)

		polar := cell.polar * polarStd + polarMean
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

		matrices[i] = modelMatrix
		colorsInstanced[i] = color

	}

	app.DrawMeshInstanced(mesh, matrices, colorsInstanced, cellsSettings.Count)
}

func uiItemControl(hover, hot, active bool) (bool, bool) {
	if hover {
		hot = true
	} else {
		hot = false
	}
	if active {
		if !platform.IsMouseLeftButtonDown() {
			active = false
		}
	} else if hot {
		if platform.IsMouseLeftButtonPressed() {
			active = true
		}
	}
	return hot, active
}

func main() {
	os.Mkdir(SAVES_DIR, os.ModeDir)
	files, err := ioutil.ReadDir(SAVES_DIR)
	
	settingsMap := make(map[int]AppSettings, 0)
	settingsNameList := make([]int, 0)
	settingsList := make([]AppSettings, 0)
	maxSaveNum := 0
	for _, file := range files {
		fileNameParts := strings.Split(file.Name(), "_")
		fileNumString := "0"
		if len(fileNameParts) > 1 {
			fileNumString = fileNameParts[1]
		}
		fileNum, _ := strconv.Atoi(fileNumString)
		
		if maxSaveNum < fileNum {
			maxSaveNum = fileNum
		}
		settingsMap[fileNum] = loadSettings(SAVES_DIR + "/" + file.Name())
		settingsNameList = append(settingsNameList, fileNum)
	}
	sort.Ints(settingsNameList)
	for _, settingName := range settingsNameList {
		settingsList = append(settingsList, settingsMap[settingName])
	}
	
	settings := loadSettings("settings")
	settingsTextures := make([]graphics.Texture, 0)

	var windowWidth = 800
	var windowHeight = 600
	windowWidth, windowHeight = platform.GetMonitorResolution()
	window := platform.GetWindow(windowWidth, windowHeight, "New fancy window", true)
	defer platform.ReleaseWindow()

	// Create screenshot dir
	SCREENSHOT_DIR := "screenshots"
	os.Mkdir(SCREENSHOT_DIR, os.ModeDir)
	files, err = ioutil.ReadDir(SCREENSHOT_DIR)
	maxScreenshotNum := 0
	for _, file := range files {
		fileNumString := strings.Split(file.Name(), ".")[0]
		fileNum, _ := strconv.Atoi(fileNumString)

		if maxScreenshotNum < fileNum {
			maxScreenshotNum = fileNum
		}
	}

	truetypeTitleBytes, err := ioutil.ReadFile("fonts/Montserrat-Regular.ttf")
	if err != nil {
		panic(err)
	}

	truetypeNormalBytes, err := ioutil.ReadFile("fonts/Montserrat-Regular.ttf")
	if err != nil {
		panic(err)
	}

	scale := platform.GetWindowScaling()
	uiFont := font.GetFont(truetypeNormalBytes, 20.0, scale)
	uiFontTitle := font.GetFont(truetypeTitleBytes, 34.0, scale)
	infoFont := font.GetFont(truetypeNormalBytes, 32.0, scale)
	app.InitRendering(float64(windowWidth), float64(windowHeight), uiFont, uiFontTitle, &settings.Rendering)
	camera := settings.Camera
	camera.UpdateFully()
	app.SetCamera(&camera)

	trashIconFile, err := os.Open("trash.png")
    if err != nil {
        panic(err)
    }
    defer trashIconFile.Close()

    trashIconImg, _, err := image.Decode(trashIconFile)
    if err != nil {
        panic(err)
	}
	
	trashIconImgData := trashIconImg.(*image.NRGBA)
	width, height := trashIconImgData.Bounds().Max.X, trashIconImgData.Bounds().Max.Y
	invertedBytes := make([]byte, len(trashIconImgData.Pix))
	for i := 0; i < int(height); i++ {
		srcRowStart := i * trashIconImgData.Stride
		srcRowEnd := (i + 1) * trashIconImgData.Stride
		dstRowStart := (int(height) - 1 - i) * trashIconImgData.Stride
		dstRowEnd := (int(height) - i) * trashIconImgData.Stride
		copy(invertedBytes[dstRowStart:dstRowEnd], trashIconImgData.Pix[srcRowStart:srcRowEnd])
	}
	
	trashIconTexture := graphics.GetTexture(width, height, 4, invertedBytes, true)

	colorChannel := make(chan []mgl32.Vec4, 1)

	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:], []int{4, 4})
	cells := make([]cell, 10000)
	generateCells(cells)

	// Circle constants
	circleWidth := 0.75
	circleWidthHover := 1.5
	circleColor := mgl32.Vec4{0.0, 0.0, 0.0, 0.4}
	circleColorHover := mgl32.Vec4{0.0, 0.0, 0.0, 0.8}
	circleColorFade := mgl32.Vec4{0.0, 0.0, 0.0, 0.01}
	circleFadeOutTime := 0.5
	screenshotTextDuration := 1.75
	screenshotTextFadeDuration := 1.0

	countSliderBgSize := mgl32.Vec2{100.0, float32(windowHeight) * 0.7}
	countSliderBgPos := mgl32.Vec2{float32(windowWidth) - countSliderBgSize[0] - 100, float32(windowHeight) * 0.15}
	
	// Count controller parameters
	countSliderColor := ColorParameter{circleColor, circleColor}
	countSliderValue := FloatParameter{float64(settings.Cells.Count), float64(settings.Cells.Count)}

	// Inner controller circle parameters
	innerCircleColor := ColorParameter{circleColor, circleColor}
	innerCircleWidth := FloatParameter{circleWidth, circleWidth}
	innerCircleRadius := FloatParameter{settings.Cells.RadiusMin, settings.Cells.RadiusMin}
	innerCircleArch := RadianParameter{0, 0}
	
	// Outer controller circle parameters
	outerCircleColor := ColorParameter{circleColor, circleColor}
	outerCircleWidth := FloatParameter{circleWidth, circleWidth}
	outerCircleRadius := FloatParameter{settings.Cells.RadiusMax, settings.Cells.RadiusMax}
	outerCircleArch := RadianParameter{0, 0}

	// Load bar parameters
	loadBarWidth := 500.0
	loadBarHideWidth := 50.0
	loadBarStart := FloatParameter{0,0}
	deleteBarColor := mgl32.Vec4{1, 0, 0, 0.5}
	deleteBarColorHover := mgl32.Vec4{1, 0, 0, 0.9}
	deleteBarColorHidden := mgl32.Vec4{0,0,0,0}
	
	loadBarColor := ColorParameter{circleColor, circleColor}
	
	loadBarDeleteButtonColors := make([]ColorParameter, 0)
	for range settingsList {
		loadBarDeleteButtonColors = append(loadBarDeleteButtonColors, ColorParameter{deleteBarColorHidden, deleteBarColorHidden})
	}
	
	settingsColor := mgl32.Vec4{1,1,1,0.8}
	settingsColorHover := mgl32.Vec4{1,1,1,1.0}
	loadBarSettingsColors := make([]ColorParameter, 0)
	for range settingsList {
		loadBarSettingsColors = append(loadBarSettingsColors, ColorParameter{settingsColor, settingsColor})
	}
	loadBarHide := FloatParameter{-loadBarWidth + loadBarHideWidth,-loadBarWidth + loadBarHideWidth}

	// Colors parameters
	colorsParams := make([]ColorParameter, 0)
	for _, color := range settings.Colors {
		colorsParams = append(colorsParams, ColorParameter{color, color})
	}

	// Inner circle mesh
	circleVertices, circleIndices := getCircle(innerCircleRadius.val, innerCircleWidth.val, innerCircleArch.val)
	circleInner := graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
	
	// Outer circle mesh
	circleVertices, circleIndices = getCircle(outerCircleRadius.val, outerCircleWidth.val, outerCircleArch.val)
	circleOuter := graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
	
	// Runtime variables
	showUI := false
	pickerStates := make([]bool, len(settings.Colors))
	
	start := time.Now()
	timeSinceMouseMovement := 0.0
	screenshotTextTimer := 0.0
	
	countSliderHot, countSliderActive := false, false
	circleInnerHot, circleInnerActive := false, false
	circleOuterHot, circleOuterActive := false, false

	for i, _ := range settingsList {
		drawCells(cells, settingsList[i].Cells, settingsList[i].Colors, cube)

		settingsList[i].Camera.UpdateFully()
		app.SetCamera(&settingsList[i].Camera)
		app.Render()
		
		imageBytes, imageWidth, imageHeight := app.GetSceneBuffer()
		texture := graphics.GetTexture(int(imageWidth), int(imageHeight), 4, []uint8(imageBytes), true)
		settingsTextures = append(settingsTextures, texture)
	}

	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		now := time.Now()
		dt := now.Sub(start).Seconds()
		start = now

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
		app.DrawText(fpsString, &infoFont, mgl32.Vec2{float32(windowWidth) - 10, float32(windowHeight) - 10}, mgl32.Vec4{0, 0, 0, 0.8}, mgl32.Vec2{1, 1}, 0)
		//app.DrawText("IRIS", &infoFont, mgl32.Vec2{float32(windowWidth - 10), 10}, mgl32.Vec4{0.0, 0.0, 0.0, 0.8}, mgl32.Vec2{1, 0}, 0)
		
		// Show screenshot text.
		screenshotTextPart := screenshotTextTimer / screenshotTextFadeDuration
		alpha := float32(math.Sqrt(screenshotTextPart))
		if alpha > 1.0 {
			alpha = 1.0
		}
		if alpha > 0.0 {
			app.DrawText("IMAGE SAVED", &infoFont, mgl32.Vec2{float32(windowWidth) / 2.0, float32(windowHeight) - 10}, mgl32.Vec4{0.0, 0.0, 0.0, alpha * 0.8}, mgl32.Vec2{0.5, 1.0}, 0)
		}
		if screenshotTextTimer > 0.0 {
			screenshotTextTimer -= dt
		}

		mdx, mdy := platform.GetMouseDeltaPosition()
		if math.Abs(mdx) < 0.001 && math.Abs(mdy) < 0.001 {
			timeSinceMouseMovement += dt
		} else {
			timeSinceMouseMovement = 0.0
		}

		inactiveUIColor := circleColor
		if timeSinceMouseMovement > circleFadeOutTime {
			inactiveUIColor = circleColorFade
		}
		
		mouseX, mouseY := platform.GetMousePosition()
		rS, rD := app.GetWorldRay(mouseX, mouseY)
		s := 0.0 - rS.Y() / rD.Y()
		pos := rS.Add(rD.Mul(s))
		radius := math.Sqrt(float64(pos.X() * pos.X() + pos.Z() * pos.Z()))

		if true {
			loadBarPos := mgl32.Vec2{float32(loadBarHide.val), 0}
			loadBarSize := mgl32.Vec2{float32(loadBarWidth), float32(windowHeight)}

			aspectRatio := float32(windowHeight) / float32(windowWidth)
			settingsSize := mgl32.Vec2{loadBarSize[0] - 20, (loadBarSize[0] - 20) * aspectRatio}

			saveSize := mgl32.Vec2{settingsSize[0], 50}
			
			if isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, loadBarPos, loadBarSize) {
				loadBarHide.target = 0.0
				scrollDelta := platform.GetMouseWheelDelta()
				loadBarStart.target += scrollDelta * 50.0
				if loadBarStart.target > 0 {
					loadBarStart.target = 0.0
				}

				maxHeight := 10 + float64(len(settingsList)) * (float64(settingsSize[1]) + 10) + float64(saveSize[1] + 10)
				lowerBorderMax := math.Max(0.0, float64(windowHeight) - maxHeight)
				if float64(windowHeight) - (loadBarStart.target + maxHeight) > lowerBorderMax {
					loadBarStart.target = float64(windowHeight) - lowerBorderMax - maxHeight
				}
				loadBarColor.target = circleColorHover
			} else {
				loadBarHide.target = -loadBarWidth + loadBarHideWidth
				loadBarColor.target = inactiveUIColor
			}

			app.DrawRect(loadBarPos, loadBarSize, loadBarColor.val, 0)

			for i := range loadBarDeleteButtonColors {
				loadBarDeleteButtonColors[i].Update(dt, 8.0)
				loadBarSettingsColors[i].Update(dt, 8.0)
			}
			loadBarStart.Update(dt, 10.0)
			loadBarHide.Update(dt, 10.0)
			loadBarColor.Update(dt, 4.0)

			{
				savePos := mgl32.Vec2{
					(loadBarSize[0] - saveSize[0]) * 0.5 + loadBarPos[0] + (loadBarPos[0] / float32(loadBarWidth)) * float32(loadBarHideWidth), 10 + float32(loadBarStart.val),
				}
				saveColor := mgl32.Vec4{0,1,0.5,0.8}

				if isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, savePos, saveSize) {
					saveColor[3] = 1.0
					if platform.IsMouseLeftButtonPressed() {
						imageBytes, imageWidth, imageHeight := app.GetSceneBuffer()
						texture := graphics.GetTexture(int(imageWidth), int(imageHeight), 4, []uint8(imageBytes), true)
						settingsTextures = append(settingsTextures, texture)
						settingsNameList = append(settingsNameList, maxSaveNum)

						maxSaveNum++
						settingsName := "settings_" + strconv.Itoa(maxSaveNum)
						newSettings := copySettings(&settings)
						newSettings.Camera = camera
						settingsList = append(settingsList, newSettings)
						loadBarDeleteButtonColors = append(loadBarDeleteButtonColors, ColorParameter{deleteBarColorHidden, deleteBarColorHidden})
						loadBarSettingsColors = append(loadBarSettingsColors, ColorParameter{deleteBarColorHidden, deleteBarColorHidden})
						path := SAVES_DIR + "/" + settingsName
						saveSettings(path, newSettings)
					}
				} 
				app.DrawRect(savePos, saveSize, saveColor, 0)

				textPos := mgl32.Vec2{savePos[0] + saveSize[0] * 0.5, savePos[1] + saveSize[1] * 0.5}
				app.DrawText("SAVE", &infoFont, textPos, mgl32.Vec4{0, 0, 0, 0.6}, mgl32.Vec2{0.5,0.5}, 1)
			}

			toRemove := -1
			for i := range settingsList {
				settingsPos := mgl32.Vec2{
					(loadBarSize[0] - settingsSize[0]) * 0.5 + loadBarPos[0] + (loadBarPos[0] / float32(loadBarWidth)) * float32(loadBarHideWidth),
					float32(len(settingsList) - i) * 10 + float32(len(settingsList) - 1 - i) * settingsSize[1] + float32(loadBarStart.val)+ 10 + saveSize[1],
				}

				if settingsPos[0] + settingsSize[0] < 0.0 {
					break
				}
				
				settingsColor[3] = 0.8
				deleteButtonSize := mgl32.Vec2{70, settingsSize[1]}
				deleteButtonPos := mgl32.Vec2{settingsPos[0] + settingsSize[0] - deleteButtonSize[0], settingsPos[1]}
				
				deleteButtonColor := loadBarDeleteButtonColors[i].val
				if isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, settingsPos, settingsSize) {
					loadBarSettingsColors[i].target = settingsColorHover
					if isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, deleteButtonPos, deleteButtonSize) {
						loadBarDeleteButtonColors[i].target = deleteBarColorHover
						if platform.IsMouseLeftButtonPressed() {
							toRemove = i
						}
					} else {
						loadBarDeleteButtonColors[i].target = deleteBarColor
						if platform.IsMouseLeftButtonPressed() {
							settings = copySettings(&settingsList[i])
							camera.TargetRadius = settings.Camera.TargetRadius
							camera.TargetPolar = settings.Camera.TargetPolar
							camera.SetAzimuth(settings.Camera.TargetAzimuth)
							camera.TargetHeight = settings.Camera.TargetHeight
							outerCircleRadius.target = settings.Cells.RadiusMax
							countSliderValue.target = float64(settings.Cells.Count)
							innerCircleRadius.target = settings.Cells.RadiusMin
							for i := range colorsParams {
								colorsParams[i].target = settings.Colors[i]
							}
						}
					}
				} else {
					loadBarDeleteButtonColors[i].target = deleteBarColorHidden
					loadBarSettingsColors[i].target = settingsColor
				}
				app.DrawRect(deleteButtonPos, deleteButtonSize, deleteButtonColor, 1)

				trashCanSize := mgl32.Vec2{50, 50}
				trashCanPos := mgl32.Vec2{deleteButtonPos[0] + deleteButtonSize[0] * 0.5 - trashCanSize[0] * 0.5, deleteButtonPos[1] + deleteButtonSize[1] * 0.5 - trashCanSize[1] * 0.5}
				app.DrawRectTextured(trashCanPos, trashCanSize, trashIconTexture, mgl32.Vec4{1,1,1,deleteButtonColor[3]}, 1)

				texture := settingsTextures[i]
				app.DrawRectTextured(settingsPos, settingsSize, texture, loadBarSettingsColors[i].val, 0)

				textPos := mgl32.Vec2{settingsPos[0] + 10, settingsPos[1] + 5}
				app.DrawText("#" + strconv.Itoa(i), &infoFont, textPos, mgl32.Vec4{0, 0, 0, 0.8}, mgl32.Vec2{0,0}, 0)
			}

			if toRemove >= 0 {
				settingsList = append(settingsList[:toRemove], settingsList[toRemove + 1:]...)
				settingsTextures = append(settingsTextures[:toRemove], settingsTextures[toRemove + 1:]...)
				loadBarDeleteButtonColors = append(loadBarDeleteButtonColors[:toRemove], loadBarDeleteButtonColors[toRemove + 1:]...)
				loadBarSettingsColors = append(loadBarSettingsColors[:toRemove], loadBarSettingsColors[toRemove + 1:]...)
				filePath := SAVES_DIR + "/settings_" + strconv.Itoa(settingsNameList[toRemove])
				os.Remove(filePath)
				settingsNameList = append(settingsNameList[:toRemove], settingsNameList[toRemove + 1:]...)
			}
		}

		mouseAngle := math.Atan2(float64(pos.X()), float64(pos.Z()))
		innerCircleArch.target = mouseAngle
		outerCircleArch.target = mouseAngle
		{
			circleInnerHover := math.Abs(innerCircleRadius.val - radius) < 4.0
			circleInnerHot, circleInnerActive = uiItemControl(circleInnerHover, circleInnerHot, circleInnerActive)
			if circleInnerHot || circleInnerActive {
				innerCircleColor.target = circleColorHover
				innerCircleWidth.target = circleWidthHover
			} else {
				innerCircleColor.target = inactiveUIColor
				innerCircleWidth.target = circleWidth
			}
			if circleInnerActive {
				innerCircleRadius.target = math.Min(radius, outerCircleRadius.target)
			}
			settings.Cells.RadiusMin = innerCircleRadius.val

			innerCircleWidth.Update(dt, 10.0)
			innerCircleRadius.Update(dt, 15.0)
			innerCircleColor.Update(dt, 4.0)
			innerCircleArch.Update(dt, 6.0)

			circleVertices, circleIndices := getCircle(innerCircleRadius.val, innerCircleWidth.val, innerCircleArch.val)
			circleInner = graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})

			modelMatrix := mgl32.Ident4()
			app.DrawMeshUI(circleInner, modelMatrix, innerCircleColor.val)
		}
		{
			circleOuterHover :=  math.Abs(outerCircleRadius.val - radius) < 4.0
			circleOuterHot, circleOuterActive = uiItemControl(circleOuterHover, circleOuterHot, circleOuterActive)
			if circleOuterHot || circleOuterActive {
				outerCircleColor.target = circleColorHover
				outerCircleWidth.target = circleWidthHover
			} else {
				outerCircleColor.target = inactiveUIColor
				outerCircleWidth.target = circleWidth
			}
			if circleOuterActive {
				outerCircleRadius.target = math.Max(radius, innerCircleRadius.target)
			}
			settings.Cells.RadiusMax = outerCircleRadius.val

			outerCircleWidth.Update(dt, 10.0)
			outerCircleRadius.Update(dt, 15.0)
			outerCircleColor.Update(dt, 4.0)
			outerCircleArch.Update(dt, 6.0)

			circleVertices, circleIndices = getCircle(outerCircleRadius.val, outerCircleWidth.val, outerCircleArch.val)
			circleOuter = graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
			
			modelMatrix := mgl32.Ident4()
			app.DrawMeshUI(circleOuter, modelMatrix, outerCircleColor.val)
		}
		{
			countSliderHover := isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, countSliderBgPos, countSliderBgSize)
			countSliderHot, countSliderActive = uiItemControl(countSliderHover, countSliderHot, countSliderActive)
			if countSliderHot || countSliderActive {
				countSliderColor.target = circleColorHover
			} else {
				countSliderColor.target = inactiveUIColor
			}
			if countSliderActive {
				portion := 1.0 - (float32(mouseY) - countSliderBgPos[1]) / countSliderBgSize[1]
				if portion < 0.0 {
					portion = 0.0
				} else if portion > 1.0 {
					portion = 1.0
				}
				countSliderValue.target = float64(portion) * 10000.0
			}
			settings.Cells.Count = int(countSliderValue.val)
			if settings.Cells.Count > 10000 {
				settings.Cells.Count = 10000
			}
			portion := float32(countSliderValue.val) / 10000.0

			countSliderColor.Update(dt, 5.0)
			countSliderValue.Update(dt, 15.0)

			countSliderSize := mgl32.Vec2{countSliderBgSize[0], countSliderBgSize[1] * portion}
			countSliderPos := mgl32.Vec2{countSliderBgPos[0], countSliderBgPos[1] + countSliderBgSize[1] - countSliderSize[1]}

			app.DrawRect(countSliderBgPos, countSliderBgSize, countSliderColor.val, 0)
			app.DrawRect(countSliderPos, countSliderSize, countSliderColor.val, 0)
		}

		select {
		case newColors := <-colorChannel:
			for i := range colorsParams {
				colorsParams[i].target = newColors[i]
			}
		default:
		}
		for i := range settings.Colors {
			settings.Colors[i] = colorsParams[i].val
			colorsParams[i].Update(dt, 5.0)
		}

		drawCells(cells, settings.Cells, settings.Colors, cube)

		app.SetCamera(&camera)
		camera.Update(dt)
		app.Render()

		// Swappity-swap.
		window.SwapBuffers()
	}
	saveSettings("settings", settings)
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
