package app

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	
	"../lib/font"
	"../lib/graphics"
)

// Pipelines used for UI rendering.
var uiTextPipeline graphics.Pipeline
var uiRectTexturePipeline graphics.Pipeline
var uiRectPipeline graphics.Pipeline

// Additional settings/data necessary for UI rendering.
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
	uiTextPipeline = graphics.GetPipeline("shaders/text_vertex_shader.glsl", "shaders/text_pixel_shader.glsl")
	uiRectPipeline = graphics.GetPipeline("shaders/rect_vertex_shader.glsl", "shaders/rect_pixel_shader.glsl")
	uiRectTexturePipeline = graphics.GetPipeline("shaders/text_vertex_shader.glsl", "shaders/rect_texture_pixel_shader.glsl")

	// Set up texture cache.
	textureCache = make(map[*font.Font] graphics.Texture)
}

func renderUI(targetBuffer graphics.Framebuffer, textEntities []textData, rectEntities []rectData, rectTextureEntities []rectTextureData) {
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

	uiRectTexturePipeline.Start()
	uiRectTexturePipeline.SetUniform("projection_matrix", uiProjectionMatrix)

	// Render all textured rectangle entities.
	for _, rectTextureEntity := range rectTextureEntities {
		drawRectTexture(uiRectTexturePipeline, rectTextureEntity.position, rectTextureEntity.size, rectTextureEntity.texture, rectTextureEntity.color)
	}

	// Set up 2D text rendering pipeline.
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
The final model matrix is M; M = T * S.
S - scales the quad to (sx, sy) dimensions.
T - positions the quad on the screen.
*/
func getModelMatrix(x, y, sx, sy float32) mgl32.Mat4 {
	modelMatrix := mgl32.Translate3D(x, y, 0).Mul4(
		mgl32.Scale3D(sx, sy, 1.0),
	)
	return modelMatrix
}

func drawRect(pipeline graphics.Pipeline, pos mgl32.Vec2, size mgl32.Vec2, color mgl32.Vec4) {
	// Set up transformation parameters.
	x, y := pos[0], float32(uiScreenHeight) - pos[1]
	sx, sy := size[0], size[1]
	modelMatrix := getModelMatrix(x, y, sx, sy)

	// Render rectangle.
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	graphics.DrawMesh(uiQuad)
}

func drawRectTexture(pipeline graphics.Pipeline, pos mgl32.Vec2, size mgl32.Vec2, texture graphics.Texture, color mgl32.Vec4) {
	// Set up transformation parameters.
	x, y := pos[0], float32(uiScreenHeight) - pos[1]
	sx, sy := size[0], size[1]
	modelMatrix := getModelMatrix(x, y, sx, sy)

	// Render rectangle.
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	pipeline.SetUniform("source_rect", mgl32.Vec4{0,0,1,1})
	graphics.SetTexture(texture, 0)
	graphics.DrawMesh(uiQuad)
}

func getGlyphSourceRect(glyph font.Glyph, textureWidth, textureHeight int) mgl32.Vec4{
	relX := float32(glyph.X) / float32(textureWidth)
	relY := 1.0 - float32(glyph.Y + glyph.BitmapHeight) / float32(textureHeight)
	relWidth := float32(glyph.BitmapWidth) / float32(textureWidth)
	relHeight := float32(glyph.BitmapHeight) / float32(textureHeight)
	return mgl32.Vec4{relX,relY,relWidth,relHeight}
}

func getFontTexture(font *font.Font) graphics.Texture {
	texture, ok := textureCache[font]
	if !ok {
		texture = graphics.GetTextureUint8(font.TextureWidth, font.TextureHeight, 1, font.Texture, true)
		textureCache[font] = texture
	}
	return texture
}

func drawText(pipeline graphics.Pipeline, text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2) {
	fontTexture := getFontTexture(font)
	graphics.SetTexture(fontTexture, 0)
	
	// Set up starting position for text rendering.
	width, height := font.GetStringWidth(text), font.RowHeight
	currentX := math.Floor(float64(position[0]) - float64(width) * float64(origin[0]))
	currentY := math.Floor(float64(position[1]) - float64(height) * float64(origin[1]))
	
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
		sourceRect := getGlyphSourceRect(glyph, font.TextureWidth, font.TextureHeight)
		pipeline.SetUniform("source_rect", sourceRect)

		// Draw a single letter.
		graphics.DrawMesh(uiQuad)
		
		// Update horizontal rendering position.
		currentX += float64(glyph.Advance)
	}
}