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

type meshDataInstanced struct {
	mesh graphics.Mesh
	modelMatrix []mgl32.Mat4
	color []mgl32.Vec4
	count int32
}

type rectData struct {
	position mgl32.Vec2
	size mgl32.Vec2
	color mgl32.Vec4
}

type rectTextureData struct {
	position mgl32.Vec2
	size mgl32.Vec2
	texture graphics.Texture
	color mgl32.Vec4
}

var meshEntities []meshData
var meshEntitiesInstanced []meshDataInstanced
var meshEntitiesUI []meshData
var rectEntities [][]rectData
var rectTextureEntities [][]rectTextureData
var textEntities [][]textData

var screenBuffer graphics.Framebuffer

func InitRendering(windowWidth float64, windowHeight float64, uiFont font.Font, uiFontTitle font.Font, renderingSettings *RenderingSettings) {
	// Set up libraries
	graphics.Init()
	ui.Init(windowWidth, windowHeight, &uiFont, &uiFontTitle)

	// Set up entity rendering lists
	rectEntities = make([][]rectData, 2)
	rectTextureEntities = make([][]rectTextureData, 2)
	textEntities = make([][]textData, 2)
	for i := 0; i < 2; i++ {
		rectEntities[i] = make([]rectData, 0, 100)
		rectTextureEntities[i] = make([]rectTextureData, 0, 100)
		textEntities[i] = make([]textData, 0, 100)
	}
	meshEntities = make([]meshData, 0, 100)
	meshEntitiesInstanced = make([]meshDataInstanced, 0, 10)
	meshEntitiesUI = make([]meshData, 0, 100)

	initSceneRendering(windowWidth, windowHeight, renderingSettings	)
	initUIRendering(uiFont, windowWidth, windowHeight)
	
	// Framebuffer setup
	screenBuffer = graphics.GetFramebufferDefault()
}

func Render() {
	targetBuffer := renderScene(meshEntities, meshEntitiesInstanced)

	// Blit to screen
	graphics.SetFramebuffer(screenBuffer)
	graphics.SetFramebufferViewport(screenBuffer)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(targetBuffer, "color", 0)
	
	pipelineBlit.Start()
	
	graphics.DrawMesh(screenQuad)

	// Draw in-scene UI on top
	graphics.EnableBlending()
	graphics.DisableDepthTest()
	pipelineUI.Start()
	pipelineUI.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelineUI.SetUniform("view_matrix", sceneViewMatrix)
	for _, meshEntity := range meshEntitiesUI {
		drawMesh(pipelineUI, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}
	
	// Get UI rendering buffers
	rectRenderingBuffer, textRenderingBuffer := ui.GetDrawData()
	for _, rect := range rectRenderingBuffer {
		rectEntities[0] = append(rectEntities[0], rectData{rect.Position, rect.Size, rect.Color})
	}

	for _, text := range textRenderingBuffer {
		font := (*text.Font).(*font.Font)
		textEntities[0] = append(textEntities[0], textData{font, text.Text, text.Position, text.Color, text.Origin})
	}
	
	renderUI(screenBuffer, textEntities[0], rectEntities[0], rectTextureEntities[0])
	renderUI(screenBuffer, textEntities[1], rectEntities[1], rectTextureEntities[1])

	// Clear rendering lists + UI
	meshEntities = meshEntities[:0]
	meshEntitiesInstanced = meshEntitiesInstanced[:0]
	meshEntitiesUI = meshEntitiesUI[:0]

	rectEntities[0] = rectEntities[0][:0]
	textEntities[0] = textEntities[0][:0]
	rectTextureEntities[0] = rectTextureEntities[0][:0]
	rectEntities[1] = rectEntities[1][:0]
	textEntities[1] = textEntities[1][:0]
	rectTextureEntities[1] = rectTextureEntities[1][:0]
	
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

func DrawMeshInstanced(mesh graphics.Mesh, modelMatrix []mgl32.Mat4, color []mgl32.Vec4, count int) {
	meshEntitiesInstanced = append(meshEntitiesInstanced, meshDataInstanced{mesh, modelMatrix, color, int32(count)})
}

func DrawMeshUI(mesh graphics.Mesh, modelMatrix mgl32.Mat4, color mgl32.Vec4) {
	meshEntitiesUI = append(meshEntitiesUI, meshData{mesh, modelMatrix, color})
}

func DrawRect(pos mgl32.Vec2, size mgl32.Vec2, color mgl32.Vec4, layer int) {
	rectEntities[layer] = append(rectEntities[layer], rectData{pos, size, color})
}

func DrawRectTextured(pos mgl32.Vec2, size mgl32.Vec2, texture graphics.Texture, color mgl32.Vec4, layer int) {
	rectTextureEntities[layer] = append(rectTextureEntities[layer], rectTextureData{pos, size, texture, color})
}

func DrawText(text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2, layer int) {
	textEntities[layer] = append(textEntities[layer], textData{font, text, position, color, origin})
}
