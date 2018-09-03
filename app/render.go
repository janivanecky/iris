package app

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/janivanecky/golib/font"
	"github.com/janivanecky/golib/graphics"
	"github.com/janivanecky/golib/ui"
)

type textData struct {
	font *font.Font
	text string
	position mgl32.Vec2
	color mgl32.Vec4
	origin mgl32.Vec2
}

type meshData struct {
	mesh graphics.Mesh
	modelMatrix mgl32.Mat4
	color mgl32.Vec4
}

type rectData struct {
	position mgl32.Vec2
	size mgl32.Vec2
	color mgl32.Vec4
}

var meshEntities []meshData
var rectEntities []rectData
var textEntities []textData

var screenBuffer graphics.Framebuffer

func InitRendering(windowWidth float64, windowHeight float64, uiFont font.Font, uiFontTitle font.Font) {
	// Set up libraries
	graphics.Init()
	ui.Init(windowWidth, windowHeight, &uiFont, &uiFontTitle)

	// Set up entity rendering lists
	rectEntities = make([]rectData, 0, 100)
	textEntities = make([]textData, 0, 100)
	meshEntities = make([]meshData, 0, 100)

	initSceneRendering(windowWidth, windowHeight)
	initUIRendering(uiFont, windowWidth, windowHeight)
	
	// Framebuffer setup
	screenBuffer = graphics.GetFramebufferDefault()
}

func Render() {
	renderScene(screenBuffer, meshEntities)
	
	// Get UI rendering buffers
	rectRenderingBuffer, textRenderingBuffer := ui.GetDrawData()
	for _, rect := range rectRenderingBuffer {
		rectEntities = append(rectEntities, rectData{rect.Position, rect.Size, rect.Color})
	}

	for _, text := range textRenderingBuffer {
		font := (*text.Font).(*font.Font)
		textEntities = append(textEntities, textData{font, text.Text, text.Position, text.Color, text.Origin})
	}
	
	renderUI(screenBuffer, textEntities, rectEntities)
	
	// Clear rendering lists + UI
	rectEntities = rectEntities[:0]
	meshEntities = meshEntities[:0]
	textEntities = textEntities[:0]
	
	ui.Clear()
}

func SetCamera(camera *Camera) {
	position := camera.GetPosition()
	target := camera.GetTarget()
	up := camera.GetUp()
	sceneCameraPosition = position.Add(target)
	sceneViewMatrix = mgl32.LookAtV(position.Add(target), target, up)
}

func DrawMesh(mesh graphics.Mesh, modelMatrix mgl32.Mat4, color mgl32.Vec4) {
	meshEntities = append(meshEntities, meshData{mesh, modelMatrix, color})
}

func DrawRect(pos mgl32.Vec2, size mgl32.Vec2, color mgl32.Vec4) {
	rectEntities = append(rectEntities, rectData{pos, size, color})
}

func DrawText(text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2) {
	textEntities = append(textEntities, textData{font, text, position, color, origin})
}
