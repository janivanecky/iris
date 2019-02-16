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

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// APP UI
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

func getCircleMesh(radius float64, width float64, arch float64) ([]float32, []uint32) {
	pointCount := uint32(128)

	vertices := make([]float32, 0)
	indices := make([]uint32, 0)

	archLength := math.Pi / 4.0
	for i := uint32(0); i < pointCount; i++ {
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

func isInRect(position mgl32.Vec2, rectPosition mgl32.Vec2, rectSize mgl32.Vec2) bool {
	if position[0] >= rectPosition[0] && position[0] <= rectPosition[0] + rectSize[0] &&
	position[1] >= rectPosition[1] && position[1] <= rectPosition[1] + rectSize[1] {
		return true
	}
    return false
}

// UTILS
func InvertBytes (bytes []byte, rowLength, rowCount int) []byte{
	invertedBytes := make([]byte, len(bytes))
	for i := 0; i < rowCount; i++ {
		srcRowStart := i * rowLength
		srcRowEnd := (i + 1) * rowLength
		dstRowStart := (int(rowCount) - 1 - i) * rowLength
		dstRowEnd := (int(rowCount) - i) * rowLength
		copy(invertedBytes[dstRowStart:dstRowEnd], bytes[srcRowStart:srcRowEnd])
	}
	return invertedBytes
}

// APP RENDER
func drawCells(cells []app.Cell, cellsSettings app.CellSettings, mesh graphics.Mesh) {
	matrices := app.GetCellMatrices(cells, cellsSettings.RadiusMin, cellsSettings.RadiusMax, cellsSettings.PolarStd,
		cellsSettings.PolarMean,cellsSettings.HeightRatio, cellsSettings.Count)
	colors := app.GetCellColors(cells, cellsSettings.Colors, cellsSettings.Count)
	app.DrawMeshInstanced(mesh, matrices, colors, cellsSettings.Count)
}

func main() {
	settings, settingsCount := app.LoadSettings()
	settingsTextures := make([]graphics.Texture, 0)

	// WINDOW 
	var windowWidth = 800
	var windowHeight = 600
	//windowWidth, windowHeight = platform.GetMonitorResolution()
	window := platform.GetWindow(windowWidth, windowHeight, "New fancy window", false)
	defer platform.ReleaseWindow()
	
	// UI
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
	
	// RENDERING (tied to UI)
	app.InitRendering(float64(windowWidth), float64(windowHeight), uiFont, uiFontTitle, &settings.Rendering)

	// UI
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
	width, height := trashIconImgData.Bounds().Max.X, trashIconImgData.Bounds().Max.Y
	invertedBytes := InvertBytes(trashIconImgData.Pix, trashIconImgData.Stride, height)
	trashIconTexture := graphics.GetTextureUint8(width, height, 4, invertedBytes, true)
	
	// CELLS
	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:], []int{4, 4})
	cells := make([]app.Cell, 10000)
	app.GenerateCells(cells)
	colorChannel := make(chan []mgl32.Vec4, 1)

	// Colors parameters
	colorsParams := make([]ColorParameter, 0)
	for _, color := range settings.Cells.Colors {
		colorsParams = append(colorsParams, ColorParameter{color, color})
	}

	// CIRCLE
	// Circle constants
	circleWidth := 0.75
	circleWidthHover := 1.5
	uiColor := mgl32.Vec4{0.0, 0.0, 0.0, 0.4}
	uiColorHover := mgl32.Vec4{0.0, 0.0, 0.0, 0.8}
	uiColorInactive := mgl32.Vec4{0.0, 0.0, 0.0, 0.01}
	uiFadeOutTime := 0.5

	// Inner controller circle parameters
	innerCircle := app.GetCircle(uiColor, circleWidth, settings.Cells.RadiusMin, 0)
	circleInnerHot, circleInnerActive := false, false
	circleVertices, circleIndices := getCircleMesh(innerCircle.Radius.Val, innerCircle.Width.Val, innerCircle.Arc.Val)
	circleInner := graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})

	// Outer controller circle parameters
	outerCircle := app.GetCircle(uiColor, circleWidth, settings.Cells.RadiusMax, 0)
	circleOuterHot, circleOuterActive := false, false
	circleVertices, circleIndices = getCircleMesh(outerCircle.Radius.Val, outerCircle.Width.Val, outerCircle.Arc.Val)
	circleOuter := graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})

	// SCREENSHOTS
	screenshotTextDuration := 1.75
	screenshotTextFadeDuration := 1.0

	// COUNTS
	countSliderBgSize := mgl32.Vec2{100.0, float32(windowHeight) * 0.7}
	countSliderBgPos := mgl32.Vec2{float32(windowWidth) - countSliderBgSize[0] - 100, float32(windowHeight) * 0.15}
	
	// Count controller parameters
	countSliderColor := ColorParameter{uiColor, uiColor}
	countSliderValue := FloatParameter{float64(settings.Cells.Count), float64(settings.Cells.Count)}
	countSliderHot, countSliderActive := false, false

	// LOAD BAR
	// Load bar constants
	settingsColor := mgl32.Vec4{1,1,1,0.8}
	settingsColorHover := mgl32.Vec4{1,1,1,1.0}
	loadBarWidth := 500.0
	loadBarHideWidth := 50.0
	deleteBarColor := mgl32.Vec4{1, 0, 0, 0.5}
	deleteBarColorHover := mgl32.Vec4{1, 0, 0, 0.9}
	deleteBarColorHidden := mgl32.Vec4{0,0,0,0}
	
	// Load bar parameters
	loadBarStart := FloatParameter{0,0}
	loadBarColor := ColorParameter{uiColor, uiColor}
	loadBarDeleteButtonColors := make([]ColorParameter, 0)
	loadBarSettingsColors := make([]ColorParameter, 0)
	for i := 0; i < settingsCount; i++ {
		loadBarDeleteButtonColors = append(loadBarDeleteButtonColors, ColorParameter{deleteBarColorHidden, deleteBarColorHidden})
		loadBarSettingsColors = append(loadBarSettingsColors, ColorParameter{settingsColor, settingsColor})
	}
	loadBarHide := FloatParameter{-loadBarWidth + loadBarHideWidth,-loadBarWidth + loadBarHideWidth}

	// Runtime variables
	showUI := false
	pickerStates := make([]bool, len(settings.Cells.Colors))
	
	start := time.Now()
	timeSinceMouseMovement := 0.0
	screenshotTextTimer := 0.0
	
		// UI - depends on RENDERING
	for i := 0; i < settingsCount; i++ {
		settings := app.GetSettings(i)
		drawCells(cells, settings.Cells, cube)
		camera := app.GetCamera(settings.Camera.Radius, settings.Camera.Azimuth, settings.Camera.Polar, settings.Camera.Height, 5.0)
		app.SetCamera(camera)
		app.Render()
		
		imageBytes, imageWidth, imageHeight := app.GetSceneBuffer()
		texture := graphics.GetTextureUint8(int(imageWidth), int(imageHeight), 4, []uint8(imageBytes), true)
		settingsTextures = append(settingsTextures, texture)
	}

	// RENDERING
	camera := app.GetCamera(settings.Camera.Radius, settings.Camera.Azimuth, settings.Camera.Polar, settings.Camera.Height, 5.0)
	app.SetCamera(camera)

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
				colorsParams[i].target = newColors[i]
			}
		default:
		}
		for i := range settings.Cells.Colors {
			colorsParams[i].Update(dt, 5.0)
			settings.Cells.Colors[i] = colorsParams[i].val
		}
		// UI
		// Let's quit if user presses Esc, that cannot mean anything else.
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
			stride := len(imageBytes) / int(imageHeight)
			invertedBytes = InvertBytes(imageBytes, stride, int(imageHeight))
			img := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))
			img.Pix = invertedBytes

			app.SaveScreenshot(img)
		}

		// UI
		if showUI {
			// SSAO related settings.
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
			for i := range settings.Cells.Colors {
				pickerStates[i], _ = panel.AddColorPalette("Color"+strconv.Itoa(i), settings.Cells.Colors[i], pickerStates[i])
				if pickerStates[i] {
					settings.Cells.Colors[i], _ = panel.AddColorPicker("Pick"+strconv.Itoa(i), settings.Cells.Colors[i], false)
				}
			}
			panel.End()

		}
		fpsString := fmt.Sprintf("%d", fps)
		app.DrawText(fpsString, &infoFont, mgl32.Vec2{float32(windowWidth) - 10, float32(windowHeight) - 10}, mgl32.Vec4{0, 0, 0, 0.8}, mgl32.Vec2{1, 1}, 0)
		
		// Show screenshot text.
		screenshotTextPart := math.Min(screenshotTextTimer / screenshotTextFadeDuration, 1.0)
		alpha := math.Sqrt(screenshotTextPart)
		if alpha > 0.0 {
			app.DrawText("IMAGE SAVED", &infoFont, mgl32.Vec2{float32(windowWidth) / 2.0, float32(windowHeight) - 10}, mgl32.Vec4{0.0, 0.0, 0.0, float32(alpha) * 0.8}, mgl32.Vec2{0.5, 1.0}, 0)
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
		if timeSinceMouseMovement > uiFadeOutTime {
			inactiveUIColor = uiColorInactive
		}
		
		// LOAD BAR
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

			maxHeight := 10 + float64(settingsCount) * (float64(settingsSize[1]) + 10) + float64(saveSize[1] + 10)
			lowerBorderMax := math.Max(0.0, float64(windowHeight) - maxHeight)
			if float64(windowHeight) - (loadBarStart.target + maxHeight) > lowerBorderMax {
				loadBarStart.target = float64(windowHeight) - lowerBorderMax - maxHeight
			}
			loadBarColor.target = uiColorHover
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
					texture := graphics.GetTextureUint8(int(imageWidth), int(imageHeight), 4, []uint8(imageBytes), true)
					settingsTextures = append(settingsTextures, texture)
					loadBarDeleteButtonColors = append(loadBarDeleteButtonColors, ColorParameter{deleteBarColorHidden, deleteBarColorHidden})
					loadBarSettingsColors = append(loadBarSettingsColors, ColorParameter{deleteBarColorHidden, deleteBarColorHidden})

					settingsCount = app.SaveSettings(settings)
				}
			} 
			app.DrawRect(savePos, saveSize, saveColor, 0)

			textPos := mgl32.Vec2{savePos[0] + saveSize[0] * 0.5, savePos[1] + saveSize[1] * 0.5}
			app.DrawText("SAVE", &infoFont, textPos, mgl32.Vec4{0, 0, 0, 0.6}, mgl32.Vec2{0.5,0.5}, 1)
		}

		toRemove := -1
		for i := 0; i < settingsCount; i++ {
			settings := app.GetSettings(i)
			settingsPos := mgl32.Vec2{
				(loadBarSize[0] - settingsSize[0]) * 0.5 + loadBarPos[0] + (loadBarPos[0] / float32(loadBarWidth)) * float32(loadBarHideWidth),
				float32(settingsCount - i) * 10 + float32(settingsCount - 1 - i) * settingsSize[1] + float32(loadBarStart.val)+ 10 + saveSize[1],
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
						settings = app.GetSettings(i)
						camera.TargetRadius = settings.Camera.Radius
						camera.TargetPolar = settings.Camera.Polar
						camera.TargetAzimuth = settings.Camera.Azimuth
						camera.TargetHeight = settings.Camera.Height
						countSliderValue.target = float64(settings.Cells.Count)
						outerCircle.Radius.Target = settings.Cells.RadiusMax
						innerCircle.Radius.Target = settings.Cells.RadiusMin
						for i := range colorsParams {
							colorsParams[i].target = settings.Cells.Colors[i]
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
			settingsCount = app.DeleteSettings(toRemove)
			settingsTextures = append(settingsTextures[:toRemove], settingsTextures[toRemove + 1:]...)
			loadBarDeleteButtonColors = append(loadBarDeleteButtonColors[:toRemove], loadBarDeleteButtonColors[toRemove + 1:]...)
			loadBarSettingsColors = append(loadBarSettingsColors[:toRemove], loadBarSettingsColors[toRemove + 1:]...)
		}
		
		// CIRCLE
		// Calculate mouse position in world space (considering pos 0 on y-axis).
		rS, rD := app.GetWorldRay(mouseX, mouseY)
		posY := float32(0.0)
		s := posY - rS.Y() / rD.Y()
		pos := rS.Add(rD.Mul(s))

		mouseAngle := math.Atan2(float64(pos.X()), float64(pos.Z()))
		radius := math.Sqrt(float64(pos.X() * pos.X() + pos.Z() * pos.Z()))
		
		{
			hover := math.Abs(innerCircle.Radius.Val - radius) < 4.0
			circleInnerHot, circleInnerActive = uiItemControl(hover, circleInnerHot, circleInnerActive)
			// Update circle state.
			if circleInnerHot || circleInnerActive {
				innerCircle.Color.Target = uiColorHover
				innerCircle.Width.Target = circleWidthHover
			} else {
				innerCircle.Color.Target = inactiveUIColor
				innerCircle.Width.Target = circleWidth
			}
			if circleInnerActive {
				innerCircle.Radius.Target = math.Max(3.0, math.Min(radius, outerCircle.Radius.Target))
			}
			innerCircle.Arc.Target = mouseAngle
			innerCircle.Update(dt)
			
			circleVertices, circleIndices := getCircleMesh(innerCircle.Radius.Val, innerCircle.Width.Val, innerCircle.Arc.Val)
			circleInner = graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
			app.DrawMeshUI(circleInner, mgl32.Ident4(), innerCircle.Color.Val)
		}
		settings.Cells.RadiusMin = innerCircle.Radius.Val

		{
			hover := math.Abs(outerCircle.Radius.Val - radius) < 4.0
			circleOuterHot, circleOuterActive = uiItemControl(hover, circleOuterHot, circleOuterActive)
			// Update circle state.
			if circleOuterHot || circleOuterActive {
				outerCircle.Color.Target = uiColorHover
				outerCircle.Width.Target = circleWidthHover
			} else {
				outerCircle.Color.Target = inactiveUIColor
				outerCircle.Width.Target = circleWidth
			}
			if circleOuterActive {
				outerCircle.Radius.Target = math.Max(innerCircle.Radius.Target, math.Min(radius, 1000.0))
			}
			outerCircle.Arc.Target = mouseAngle
			outerCircle.Update(dt)
			
			circleVertices, circleIndices = getCircleMesh(outerCircle.Radius.Val, outerCircle.Width.Val, outerCircle.Arc.Val)
			circleOuter = graphics.GetMesh(circleVertices, circleIndices, []int{4, 4})
			app.DrawMeshUI(circleOuter, mgl32.Ident4(), outerCircle.Color.Val)
		}
		settings.Cells.RadiusMax = outerCircle.Radius.Val

		// COUNT
		{
			countSliderHover := isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, countSliderBgPos, countSliderBgSize)
			countSliderHot, countSliderActive = uiItemControl(countSliderHover, countSliderHot, countSliderActive)
			if countSliderHot || countSliderActive {
				countSliderColor.target = uiColorHover
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
				portion = float32(math.Max(0.0, math.Min(float64(portion), 1.0)))
				countSliderValue.target = float64(portion) * 10000.0
			}
			settings.Cells.Count = int(countSliderValue.val)
			if settings.Cells.Count > 10000 {
				settings.Cells.Count = 10000
			}
			countSliderColor.Update(dt, 5.0)
			countSliderValue.Update(dt, 15.0)
			
			portion := float32(countSliderValue.val) / 10000.0
			countSliderSize := mgl32.Vec2{countSliderBgSize[0], countSliderBgSize[1] * portion}
			countSliderPos := mgl32.Vec2{countSliderBgPos[0], countSliderBgPos[1] + countSliderBgSize[1] - countSliderSize[1]}

			app.DrawRect(countSliderBgPos, countSliderBgSize, countSliderColor.val, 0)
			app.DrawRect(countSliderPos, countSliderSize, countSliderColor.val, 0)
		}

		drawCells(cells, settings.Cells, cube)

		app.SetCamera(camera)
		camera.Update(dt)
		settings.Camera.Radius = camera.TargetRadius
		settings.Camera.Azimuth = camera.TargetAzimuth
		settings.Camera.Polar = camera.TargetPolar
		settings.Camera.Height = camera.TargetHeight
		app.Render()

		// Swappity-swap.
		window.SwapBuffers()
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
