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
var textEntities []textData

type meshData struct {
	mesh graphics.Mesh
	modelMatrix mgl32.Mat4
	color mgl32.Vec4
}
var meshEntities []meshData

type rectData struct {
	position mgl32.Vec2
	size mgl32.Vec2
	color mgl32.Vec4
}
var rectEntities []rectData

var screenWidth float64
var screenHeight float64

var Roughness = 1.0
var Reflectivity = 0.05
var SSAORadius = 0.5
var SSAORange = 3.0
var SSAOBoundary = 1.0
var Color = mgl32.Vec4{1,1,0,1};
var DirectLight = 0.5
var AmbientLight = 0.75
var MinWhite = 8.0

var screenBuffer graphics.Framebuffer

func InitRendering(windowWidth float64, windowHeight float64, uiFont font.Font) {
	// Set up libraries
	graphics.Init()
	ui.Init(windowWidth, windowHeight, &uiFont)

	// Set up entity rendering lists
	rectEntities = make([]rectData, 0, 100)
	textEntities = make([]textData, 0, 100)
	meshEntities = make([]meshData, 0, 100)

	// Fetch screen size
	screenWidth = windowWidth
	screenHeight = windowHeight

	initSceneRendering(windowWidth, windowHeight)
	initUIRendering(uiFont, windowWidth, windowHeight)
	
	// Framebuffer setup
	screenBuffer = graphics.GetFramebufferDefault()
}

func Render() {
	renderScene(screenBuffer)
	renderUI(screenBuffer)
	
	// Clear rendering lists + UI
	rectEntities = rectEntities[:0]
	meshEntities = meshEntities[:0]
	textEntities = textEntities[:0]
	
	ui.Clear()
}

func SetCameraPosition(position mgl32.Vec3, up mgl32.Vec3) {
	sceneData.cameraPosition = position
	sceneData.viewMatrix = mgl32.LookAtV(position, mgl32.Vec3{}, up)
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
