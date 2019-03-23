package main

import (
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"time"

	"runtime"

	"github.com/go-gl/mathgl/mgl32"

	"./app"
	"./lib/font"
	"./lib/graphics"
	"./lib/platform"
	"./lib/ui"
)

// Circle constants
const circleWidth 	   = 0.75
const circleWidthHover = 1.5
const uiFadeOutTime    = 0.5
var   uiColor 		   = mgl32.Vec4{0.0, 0.0, 0.0, 0.4}
var   uiColorHover 	   = mgl32.Vec4{0.0, 0.0, 0.0, 0.8}
var   uiColorInactive  = mgl32.Vec4{0.0, 0.0, 0.0, 0.01}

// Screenshot constants
const screenshotTextDuration 	 = 1.75
const screenshotTextFadeDuration = 1.0

// Settings bar constants
const settingsBarWidth 		 = 500.0
const settingsBarWidthHidden = 50.0
var settingsBarColor 				  = mgl32.Vec4{1, 1, 1, 0.8}
var settingsBarColorHover 			  = mgl32.Vec4{1, 1, 1, 1.0}
var settingsDeleteButtonColor 		  = mgl32.Vec4{1, 0, 0, 0.5}
var settingsDeleteButtonColorHover    = mgl32.Vec4{1, 0, 0, 0.9}
var settingsDeleteButtonColorInactive = mgl32.Vec4{0, 0, 0, 0}


func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// APP UI
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

func isInRect(position mgl32.Vec2, rectPosition mgl32.Vec2, rectSize mgl32.Vec2) bool {
	if position[0] >= rectPosition[0] && position[0] <= rectPosition[0]+rectSize[0] &&
		position[1] >= rectPosition[1] && position[1] <= rectPosition[1]+rectSize[1] {
		return true
	}
	return false
}

// TODO: move, refactor
const far, near = 500.0, 0.01

func GetWorldRay(screenX, screenY float64,
	screenWidth, screenHeight float64,
	viewMatrix, projectionMatrix mgl32.Mat4) (mgl32.Vec4, mgl32.Vec4) {
	projViewMatrix := projectionMatrix.Mul4(viewMatrix)
	invProjViewMatrix := projViewMatrix.Inv()

	relX := float32(screenX/screenWidth*2.0 - 1.0)
	relY := -float32(screenY/screenHeight*2.0 - 1.0)

	vFar := mgl32.Vec4{relX, relY, 1.0, 1.0}
	vFar = vFar.Mul(far)
	vFar = invProjViewMatrix.Mul4x1(vFar)

	vNear := mgl32.Vec4{relX, relY, -1.0, 1.0}
	vNear = vNear.Mul(near)
	vNear = invProjViewMatrix.Mul4x1(vNear)

	vDiff := vFar.Sub(vNear)
	vDiff = vDiff.Normalize()
	return vNear, vDiff
}

// APP RENDER
func drawCells(cells []app.Cell, cellsSettings app.CellSettings, mesh graphics.Mesh) {
	matrices := app.GetCellModelMatrices(cells, cellsSettings.RadiusMin, cellsSettings.RadiusMax, cellsSettings.PolarStd,
		cellsSettings.PolarMean, cellsSettings.HeightRatio, cellsSettings.Count)
	colors := app.GetCellColors(cells, cellsSettings.Colors, cellsSettings.Count)
	app.DrawMeshInstanced(mesh, matrices, colors, cellsSettings.Count)
}

func main() {
	settings, settingsCount := app.LoadSettings()
	
	var windowWidth = 1600
	var windowHeight = 800
	//windowWidth, windowHeight = platform.GetMonitorResolution()
	window := platform.GetWindow(windowWidth, windowHeight, "New fancy window", false)
	defer platform.ReleaseWindow()

	// TODO: Maybe somehow encapsulate?
	var uiFont, uiFontTitle, infoFont font.Font
	{
		truetypeTitleBytes, err := ioutil.ReadFile("fonts/Montserrat-Regular.ttf")
		if err != nil {
			panic(err)
		}
		truetypeNormalBytes, err := ioutil.ReadFile("fonts/Montserrat-Regular.ttf")
		if err != nil {
			panic(err)
		}
		scale := platform.GetWindowScaling()
		uiFont = font.GetFont(truetypeNormalBytes, 20.0, scale)
		uiFontTitle = font.GetFont(truetypeTitleBytes, 34.0, scale)
		infoFont = font.GetFont(truetypeNormalBytes, 32.0, scale)
	}

	// Set up libraries
	{
		graphics.Init()
		ui.Init(float64(windowWidth), float64(windowHeight), &uiFont, &uiFontTitle)
	}

	// Init renderers.
	{
		app.InitUIRendering(uiFont, float64(windowWidth), float64(windowHeight))
		app.InitSceneRendering(float64(windowWidth), float64(windowHeight))
	}
	screenBuffer := graphics.GetFramebufferDefault()

	// Load cell mesh.
	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:], []int{4, 4})

	// Create cells array.
	cells := make([]app.Cell, 10000)
	app.GenerateCells(cells)

	// Create channel used to update asynchronously cell colors.
	colorChannel := make(chan []mgl32.Vec4, 1)

	// Colors parameters
	colorsParams := make([]app.ColorParameter, 0)
	for _, color := range settings.Cells.Colors {
		colorsParams = append(colorsParams, app.ColorParameter{color, color})
	}

	// Circle controllers
	innerCircleController := app.GetCircleController(uiColor, circleWidth, settings.Cells.RadiusMin, 0)
	outerCircleController := app.GetCircleController(uiColor, circleWidth, settings.Cells.RadiusMax, 0)

	// COUNTS
	// TODO: move to specific file
	countSliderBgSize := mgl32.Vec2{50.0, float32(windowHeight) * 0.7}
	countSliderBgPos := mgl32.Vec2{float32(windowWidth) - countSliderBgSize[0] - 50, float32(windowHeight) * 0.15}

	// Count controller parameters
	countSliderColor := app.ColorParameter{uiColor, uiColor}
	countSliderValue := app.FloatParameter{float64(settings.Cells.Count), float64(settings.Cells.Count)}
	countSliderHot, countSliderActive := false, false

	// TODO: Shouldn't be here probably?
	trashIconFile, err := os.Open("trash.png")
	if err != nil {
		panic(err)
	}
	trashIconImg, _, err := image.Decode(trashIconFile)
	if err != nil {
		panic(err)
	}
	trashIconFile.Close()
	trashIconImgData := trashIconImg.(*image.NRGBA)
	iconWidth, iconHeight := trashIconImgData.Bounds().Max.X, trashIconImgData.Bounds().Max.Y
	trashIconTexture := graphics.GetTextureUint8(iconWidth, iconHeight, 4, trashIconImgData.Pix, true)

	settingsBar := app.GetSettingsBar(infoFont, trashIconTexture, float32(windowHeight))

	// Runtime variables
	showUI := false
	pickerStates := make([]bool, len(settings.Cells.Colors))

	start := time.Now()
	timeSinceMouseMovement := 0.0
	screenshotTextTimer := 0.0

	// TODO: Should be here?
	const near, far float32 = 0.01, 500.0
	aspectRatio := float64(windowWidth)/float64(windowHeight)
	projectionMatrix := mgl32.Perspective(mgl32.DegToRad(60.0), float32(aspectRatio), near, far)

	// UI - depends on RENDERING
	for i := 0; i < settingsCount; i++ {
		settings := app.GetSettings(i)
		drawCells(cells, settings.Cells, cube)
		camera := app.GetCamera(settings.Camera.Radius, settings.Camera.Azimuth, settings.Camera.Polar, settings.Camera.Height)
		viewMatrix := camera.GetViewMatrix()

		app.RenderScene(screenBuffer, viewMatrix, projectionMatrix, &settings.Rendering)
		app.ResetScene()

		imageBytes, imageWidth, imageHeight := app.GetSceneBuffer()
		texture := graphics.GetTextureUint8(int(imageWidth), int(imageHeight), 4, []uint8(imageBytes), true)
		settingsBar.AddSettings(texture)
	}

	// RENDERING
	camera := app.GetCamera(settings.Camera.Radius, settings.Camera.Azimuth, settings.Camera.Polar, settings.Camera.Height)

	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		now := time.Now()
		dt := now.Sub(start).Seconds()
		start = now
		platform.Update(window)

		// UI
		fps := 0
		if dt > 0.0 {
			fps = int(1.0 / dt)
		}

		// CELLS
		if platform.IsKeyPressed(platform.KeyR) {
			app.GenerateCells(cells)
		}
		if platform.IsKeyPressed(platform.KeyC) {
			go func() {
				newColors := app.GetRandomColorPalette()
				if newColors != nil {
					colorChannel <- newColors
				}
			}()
		}
		select {
		case newColors := <-colorChannel:
			for i := range colorsParams {
				colorsParams[i].Target = newColors[i]
			}
		default:
		}
		for i := range settings.Cells.Colors {
			colorsParams[i].Update(dt, 5.0)
			settings.Cells.Colors[i] = colorsParams[i].Val
		}
		// UI
		if platform.IsKeyPressed(platform.KeyEscape) {
			break
		}
		if platform.IsKeyPressed(platform.KeyF2) {
			showUI = !showUI
		}

		// SCREENSHOTS
		if platform.IsKeyPressed(platform.KeyF10) {
			screenshotTextTimer = screenshotTextDuration

			imageBytes, imageWidth, imageHeight := app.GetSceneBuffer()
			img := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))
			img.Pix = imageBytes

			app.SaveScreenshot(img)
		}

		// UI
		if showUI {
			// SSAO related settings.
			panel := ui.StartPanel("Rendering", mgl32.Vec2{100, 10}, 450)
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
			for i := range settings.Cells.Colors {
				pickerStates[i], _ = panel.AddColorPalette("Color"+strconv.Itoa(i), settings.Cells.Colors[i], pickerStates[i])
				if pickerStates[i] {
					settings.Cells.Colors[i], _ = panel.AddColorPicker("Pick"+strconv.Itoa(i), settings.Cells.Colors[i], false)
				}
			}
			panel.End()
		}

		fpsString := fmt.Sprintf("%d", fps)
		app.DrawUIText(fpsString, &infoFont, mgl32.Vec2{float32(windowWidth) - 10, float32(windowHeight) - 10}, mgl32.Vec4{0, 0, 0, 0.8}, mgl32.Vec2{1, 1}, 0)

		// Show screenshot text.
		screenshotTextPart := math.Min(screenshotTextTimer/screenshotTextFadeDuration, 1.0)
		alpha := math.Sqrt(screenshotTextPart)
		if alpha > 0.0 {
			app.DrawUIText("IMAGE SAVED", &infoFont, mgl32.Vec2{float32(windowWidth) / 2.0, float32(windowHeight) - 10}, mgl32.Vec4{0.0, 0.0, 0.0, float32(alpha) * 0.8}, mgl32.Vec2{0.5, 1.0}, 0)
		}
		if screenshotTextTimer > 0.0 {
			screenshotTextTimer -= dt
		}

		// UI
		mdx, mdy := platform.GetMouseDeltaPosition()
		if math.Abs(mdx) < 0.001 && math.Abs(mdy) < 0.001 {
			timeSinceMouseMovement += dt
		} else {
			timeSinceMouseMovement = 0.0
		}

		mouseX, mouseY := platform.GetMousePosition()
		inactiveUIColor := uiColor
		hideUI := timeSinceMouseMovement > uiFadeOutTime
		if timeSinceMouseMovement > uiFadeOutTime {
			inactiveUIColor = uiColorInactive
		}
		action, index := settingsBar.Update(dt, float32(mouseX), float32(mouseY), hideUI)
		switch action {
		case app.SAVE:
			imageBytes, imageWidth, imageHeight := app.GetSceneBuffer()
			texture := graphics.GetTextureUint8(int(imageWidth), int(imageHeight), 4, []uint8(imageBytes), true)
			settingsBar.AddSettings(texture)

			settingsCount = app.SaveSettings(settings)
		case app.SELECT:
			settings = app.GetSettings(index)
			camera.SetStateWithTransition(settings.Camera.Radius, settings.Camera.Azimuth,
				settings.Camera.Polar, settings.Camera.Height)
			countSliderValue.Target = float64(settings.Cells.Count)
			outerCircleController.Radius.Target = settings.Cells.RadiusMax
			innerCircleController.Radius.Target = settings.Cells.RadiusMin
			for i := range colorsParams {
				colorsParams[i].Target = settings.Cells.Colors[i]
			}
		case app.DELETE:
			settingsCount = app.DeleteSettings(index)
			settingsBar.RemoveSettings(index)
		}

		viewMatrix := camera.GetViewMatrix()
		// Calculate mouse position in world space (considering pos 0 on y-axis).
		rS, rD := GetWorldRay(mouseX, mouseY, float64(windowWidth), float64(windowHeight), viewMatrix, projectionMatrix)
		posY := float32(0.0)
		s := posY - rS.Y()/rD.Y()
		pos := rS.Add(rD.Mul(s))

		{
			innerCircleController.Update(dt, float64(pos.X()), float64(pos.Z()), 3.0, outerCircleController.Radius.Target, hideUI)

			circleVertices, circleIndices := innerCircleController.GetMeshData()
			circleInner := graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
			app.DrawMeshSceneUI(circleInner, mgl32.Ident4(), innerCircleController.Color.Val)
		}
		
		{
			outerCircleController.Update(dt, float64(pos.X()), float64(pos.Z()), innerCircleController.Radius.Target, 1000.0, hideUI)
			
			circleVertices, circleIndices := outerCircleController.GetMeshData()
			circleOuter := graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
			app.DrawMeshSceneUI(circleOuter, mgl32.Ident4(), outerCircleController.Color.Val)
		}
		
		// TODO: Should be encapsulated.
		// COUNT
		{
			countSliderHover := isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, countSliderBgPos, countSliderBgSize)
			countSliderHot, countSliderActive = uiItemControl(countSliderHover, countSliderHot, countSliderActive)
			if countSliderHot || countSliderActive {
				countSliderColor.Target = uiColorHover
			} else {
				countSliderColor.Target = inactiveUIColor
			}
			if countSliderActive {
				portion := 1.0 - (float32(mouseY)-countSliderBgPos[1])/countSliderBgSize[1]
				if portion < 0.0 {
					portion = 0.0
				} else if portion > 1.0 {
					portion = 1.0
				}
				portion = float32(math.Max(0.0, math.Min(float64(portion), 1.0)))
				countSliderValue.Target = float64(portion) * 10000.0
			}
			settings.Cells.Count = int(countSliderValue.Val)
			if settings.Cells.Count > 10000 {
				settings.Cells.Count = 10000
			}
			countSliderColor.Update(dt, 5.0)
			countSliderValue.Update(dt, 15.0)

			portion := float32(countSliderValue.Val) / 10000.0
			countSliderSize := mgl32.Vec2{countSliderBgSize[0], countSliderBgSize[1] * portion}
			countSliderPos := mgl32.Vec2{countSliderBgPos[0], countSliderBgPos[1] + countSliderBgSize[1] - countSliderSize[1]}

			app.DrawUIRect(countSliderBgPos, countSliderBgSize, countSliderColor.Val, 0)
			app.DrawUIRect(countSliderPos, countSliderSize, countSliderColor.Val, 0)
		}

		drawCells(cells, settings.Cells, cube)
		viewMatrix = camera.GetViewMatrix()
		camera.Update(dt)
		app.RenderScene(screenBuffer, viewMatrix, projectionMatrix, &settings.Rendering)
		app.ResetScene()
		
		rectRenderingBuffer, textRenderingBuffer := ui.GetDrawData()
		for _, rect := range rectRenderingBuffer {
			app.DrawUIRect(rect.Position, rect.Size, rect.Color, 0)
		}
		
		for _, text := range textRenderingBuffer {
			font := (*text.Font).(*font.Font)
			app.DrawUIText(text.Text, font, text.Position, text.Color, text.Origin, 0)
		}
		
		app.RenderUI(screenBuffer)
		app.ResetUI()
		ui.Clear()
		
		// Swappity-swap.
		window.SwapBuffers()

		settings.Camera.Radius, settings.Camera.Azimuth, settings.Camera.Polar, settings.Camera.Height = camera.GetState()
		settings.Cells.RadiusMin = innerCircleController.Radius.Val
		settings.Cells.RadiusMax = outerCircleController.Radius.Val
	}
	app.SaveActiveSettings(settings)
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
