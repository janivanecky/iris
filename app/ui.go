package app

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v4.1-core/gl"
	
	"github.com/janivanecky/golib/font"
	"github.com/janivanecky/golib/graphics"
	"github.com/janivanecky/golib/ui"
)

var textPipeline Pipeline
var rectPipeline Pipeline

type uiSettings struct {
	font font.Font
	fontTexture graphics.Texture
	quad graphics.Mesh
	projectionMatrix mgl32.Mat4
}
var uiData uiSettings

func initUIRendering(uiFont font.Font, windowWidth, windowHeight float64) {
	// Initialize UI rendering data
	uiData = uiSettings{
		uiFont,
		graphics.GetTexture(512, 512, 1, uiFont.Texture, false),
		graphics.GetMesh(quadVertices[:], quadIndices[:], []int{4, 2}),
		mgl32.Ortho(0.0, float32(windowWidth), 0.0, float32(windowHeight), 10.0, -10.0),
	}

	textPipeline = InitPipeline("shaders/text_vertex_shader.glsl", "shaders/text_pixel_shader.glsl")
	rectPipeline = InitPipeline("shaders/rect_vertex_shader.glsl", "shaders/rect_pixel_shader.glsl")
}

func renderUI(targetBuffer graphics.Framebuffer) {
	graphics.SetFramebuffer(targetBuffer)
	gl.Enable(gl.FRAMEBUFFER_SRGB)
	
	// Get UI rendering buffers
	rectRenderingBuffer, textRenderingBuffer := ui.GetDrawData()
	
	// Set up 2D rect rendering pipeline
	rectPipeline.Start()
	rectPipeline.SetUniform("projection_matrix", uiData.projectionMatrix)
	
	// Render UI rects
	for _, rectData := range rectRenderingBuffer {
		drawRect(rectPipeline, rectData.Position, rectData.Size, rectData.Color)
	}
	
	// Render the rest of rects
	for _, rectEntity := range rectEntities {
		drawRect(rectPipeline, rectEntity.position, rectEntity.size, rectEntity.color)
	}

	// Set up 2D text rendering pipeline
	textPipeline.Start()
	graphics.SetTexture(uiData.fontTexture, 0)
	textPipeline.SetUniform("projection_matrix", uiData.projectionMatrix)
	
	// Render UI text
	for _, textData := range textRenderingBuffer {
		drawText(textPipeline, textData.Text, &(uiData.font), textData.Position, textData.Color, textData.Origin)
	}
	
	// Render the rest of texts
	for _, textEntity := range textEntities {
		drawText(textPipeline, textEntity.text, textEntity.font, textEntity.position, textEntity.color, textEntity.origin)
	}
	gl.Disable(gl.FRAMEBUFFER_SRGB)
}