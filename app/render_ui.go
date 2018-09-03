package app

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	
	"github.com/janivanecky/golib/font"
	"github.com/janivanecky/golib/graphics"
)

// Pipelines used for UI rendering.
var uiTextPipeline Pipeline
var uiRectPipeline Pipeline

// Additional settings/data necessary for UI rendering.
const uiFontTextureSize = 1024
var uiQuad graphics.Mesh
var uiProjectionMatrix mgl32.Mat4
var uiScreenHeight float64

var textureCache map[*font.Font] graphics.Texture

func initUIRendering(uiFont font.Font, windowWidth, windowHeight float64) {
	// UI settings set up.
	uiScreenHeight = windowHeight
	uiQuad = graphics.GetMesh(quadVertices[:], quadIndices[:], []int{4, 2})
	uiProjectionMatrix = mgl32.Ortho(0.0, float32(windowWidth), 0.0, float32(windowHeight), 10.0, -10.0)

	// Pipelines initialization
	uiTextPipeline = InitPipeline("shaders/text_vertex_shader.glsl", "shaders/text_pixel_shader.glsl")
	uiRectPipeline = InitPipeline("shaders/rect_vertex_shader.glsl", "shaders/rect_pixel_shader.glsl")

	// Set up texture cache.
	textureCache = make(map[*font.Font] graphics.Texture)
}

func renderUI(targetBuffer graphics.Framebuffer, textEntities []textData, rectEntities []rectData) {
	// We need alpha rendering and no depth tests.
	graphics.EnableBlending()
	graphics.DisableDepthTest()

	// We're not doing gamma correction in shaders, so we have to set up SRGB rendering.
	graphics.EnableSRGBRendering()
	graphics.SetFramebuffer(targetBuffer)
	
	// Set up 2D rect rendering pipeline.
	uiRectPipeline.Start()
	uiRectPipeline.SetUniform("projection_matrix", uiProjectionMatrix)
	
	// Render all the rectangle entities.
	for _, rectEntity := range rectEntities {
		drawRect(uiRectPipeline, rectEntity.position, rectEntity.size, rectEntity.color)
	}

	// Set up 2D text rendering pipeline,
	uiTextPipeline.Start()
	uiTextPipeline.SetUniform("projection_matrix", uiProjectionMatrix)
	
	// Render all the text entities.
	for _, textEntity := range textEntities {
		drawText(uiTextPipeline, textEntity.text, textEntity.font, textEntity.position, textEntity.color, textEntity.origin)
	}
	
	// Disable SRGB rendering.
	// TODO: Ideally we want to revert to original SRGB rendering state instead of always disabling.
	graphics.DisableSRGBRendering()
	
	// Disable blending and enable depth test - "default" options.
	// TODO: Ideally we want to revert to original rendering state instead of setting to specific values.
	graphics.EnableDepthTest()
	graphics.DisableBlending()
}

/*
The final model matrix consist of 3 matrices T2 * S.
S  - scales the quad to (sx, sy) dimensions.
T2  -positions the quad on the screen.
*/
func getModelMatrix(x, y, sx, sy float32) mgl32.Mat4 {
	modelMatrix := mgl32.Translate3D(x, y, 0).Mul4(
		mgl32.Scale3D(sx, sy, 1.0),
	)
	return modelMatrix
}

func drawRect(pipeline Pipeline, pos mgl32.Vec2, size mgl32.Vec2, color mgl32.Vec4) {
	// Set up transformation parameters.
	x, y := pos[0], float32(uiScreenHeight) - pos[1]
	sx, sy := size[0], size[1]
	modelMatrix := getModelMatrix(x, y, sx, sy)

	// Render rectangle.
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	graphics.DrawMesh(uiQuad)
}

func getGlyphSourceRect(glyph font.Glyph, textureSize float64) mgl32.Vec4{
	relX := float32(glyph.X) / float32(textureSize)
	relY := 1.0 - float32(glyph.Y + glyph.BitmapHeight) / float32(textureSize)
	relWidth := float32(glyph.BitmapWidth) / float32(textureSize)
	relHeight := float32(glyph.BitmapHeight) / float32(textureSize)
	return mgl32.Vec4{relX,relY,relWidth,relHeight}
}

func getFontTexture(font *font.Font) graphics.Texture {
	texture, ok := textureCache[font]
	if !ok {
		texture = graphics.GetTexture(uiFontTextureSize, uiFontTextureSize, 1, font.Texture, true)
		textureCache[font] = texture
	}
	return texture
}

func drawText(pipeline Pipeline, text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2) {
	fontTexture := getFontTexture(font)
	graphics.SetTexture(fontTexture, 0)
	
	// Set up starting position for text rendering.
	width, height := font.GetStringWidth(text), font.RowHeight
	currentX := math.Floor(float64(position[0]) - float64(width) * float64(origin[0]))
	currentY := math.Floor(float64(position[1]) + font.TopPad - float64(height) * float64(origin[1]))
	
	// Color is the same for all the letters, so we'll set it before the loop.
	pipeline.SetUniform("color", color)

	for i, char := range text {
		if i > 0 {
			kerning := float64(font.GetKerning(char, rune(text[i-1])))
			currentX += kerning
		}
		glyph := font.Glyphs[char]
		
		// Get transformation parameters.
		x := float32(currentX + glyph.XOffset)
		y := float32(uiScreenHeight - (currentY + float64(glyph.YOffset)))
		sx := float32(glyph.Width)
		sy := float32(glyph.Height)
		modelMatrix := getModelMatrix(x, y, sx, sy)
		pipeline.SetUniform("model_matrix", modelMatrix)
		
		// Get source rectangle for the font texture.
		sourceRect := getGlyphSourceRect(glyph, uiFontTextureSize)
		pipeline.SetUniform("source_rect", sourceRect)

		// Draw a single letter.
		graphics.DrawMesh(uiQuad)
		
		// Update horizontal rendering position.
		currentX += float64(glyph.Advance)
	}
}