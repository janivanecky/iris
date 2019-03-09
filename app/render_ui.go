package app

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	
	"../lib/font"
	"../lib/graphics"
)

// Pipelines used for UI rendering.
var uiTextPipeline graphics.Pipeline
var uiTexturedRectPipeline graphics.Pipeline
var uiRectPipeline graphics.Pipeline

// Additional settings/data necessary for UI rendering.
var uiQuad graphics.Mesh
var uiProjectionMatrix mgl32.Mat4

var quadVertices = [...]float32{
	0.0, -1.0, 0.0, 1.0,
	0.0, 0.0,
	0.0, 0.0, 0.0, 1.0,
	0.0, 1.0,
	1.0, 0.0, 0.0, 1.0,
	1.0, 1.0,
	1.0, -1.0, 0.0, 1.0,
	1.0, 0.0,
}

var quadIndices = [...]uint32{
	0, 1, 2,
	0, 2, 3,
}

// Cache for font textures.
var fontTextureCache map[*font.Font] graphics.Texture

// Structs for storing draw data.
type textData struct {
	font 	 *font.Font
	text 	 string
	position mgl32.Vec2
	color 	 mgl32.Vec4
	origin 	 mgl32.Vec2
}

type rectData struct {
	position mgl32.Vec2
	size 	 mgl32.Vec2
	color 	 mgl32.Vec4
}

type texturedRectData struct {
	position mgl32.Vec2
	size 	 mgl32.Vec2
	texture  graphics.Texture
	color    mgl32.Vec4
}

// Slices for storing draw data.
const noLayers = 2
var rectEntities 		 [][]rectData
var texturedRectEntities [][]texturedRectData
var textEntities		 [][]textData

// InitUIRendering initializes necessary objects for UI rendering.
func InitUIRendering(uiFont font.Font, windowWidth, windowHeight float64) {
	// Set up quad mesh used to display everything in UI.
	uiQuad = graphics.GetMesh(quadVertices[:], quadIndices[:], []int{4, 2})
	// For OpenGL NDC y-axis is positive upwards. We want to specify position
	// of UI objects with y-axis pointing downwards (0 is top of the screen).
	// To do that, we can just use negative direction of y-axis, which points
	// downwards and has 0 at the top. That means our projection matrix will
	// map (-windowHeight, 0) to NDC (-1, 1) y-axis and all the positions of
	// individual UI elements will be negated just before drawing (but user
	// will specify positive values, 0 meaning top of the screen).
	uiProjectionMatrix = mgl32.Ortho(0.0, float32(windowWidth),
									 -float32(windowHeight), 0,
									 10.0, -10.0)

	// Pipelines initialization.
	uiTextPipeline = graphics.GetPipeline(
		"shaders/text_vertex_shader.glsl",
		"shaders/text_pixel_shader.glsl")
	uiRectPipeline = graphics.GetPipeline(
		"shaders/rect_vertex_shader.glsl",
		"shaders/rect_pixel_shader.glsl")
	uiTexturedRectPipeline = graphics.GetPipeline(
		"shaders/text_vertex_shader.glsl",
		"shaders/rect_texture_pixel_shader.glsl")

	// Set up cache for font textures.
	fontTextureCache = make(map[*font.Font] graphics.Texture)

	// Initialize slices which will store data for draw calls.
	// Note that we'll have two layers at which the data is drawn,
	// hence slice of slices.
	rectEntities 	     = make([][]rectData, noLayers)
	texturedRectEntities = make([][]texturedRectData, noLayers)
	textEntities         = make([][]textData, noLayers)
	for i := 0; i < noLayers; i++ {
		rectEntities[i] 		= make([]rectData, 0, 100)
		texturedRectEntities[i] = make([]texturedRectData, 0, 100)
		textEntities[i] 		= make([]textData, 0, 100)
	}
}

// RenderUI sends commands to draw UI gathered from DrawUIXXX calls.
func RenderUI(targetBuffer graphics.Framebuffer) {
	// We need alpha rendering and no depth tests.
	graphics.EnableBlending()
	graphics.DisableDepthTest()

	// We're not doing gamma correction in shaders, so we have to set up SRGB rendering.
	graphics.EnableSRGBRendering()

	// Set targetBuffer for rendering.
	graphics.SetFramebuffer(targetBuffer)
	graphics.SetFramebufferViewport(targetBuffer)

	for i := 0; i < noLayers; i++ {
		// Render all the rectangle entities.
		uiRectPipeline.Start()
		uiRectPipeline.SetUniform("projection_matrix", uiProjectionMatrix)
		
		for _, rectEntity := range rectEntities[i] {
			drawRect(uiRectPipeline, rectEntity.position, rectEntity.size, rectEntity.color)
		}
	
		// Render all the textured rectangle entities.
		uiTexturedRectPipeline.Start()
		uiTexturedRectPipeline.SetUniform("projection_matrix", uiProjectionMatrix)
	
		for _, texturedRectEntity := range texturedRectEntities[i] {
			drawRectTexture(uiTexturedRectPipeline, texturedRectEntity.position, texturedRectEntity.size,
							texturedRectEntity.texture, texturedRectEntity.color)
		}
	
		// Render all the text entities.
		uiTextPipeline.Start()
		uiTextPipeline.SetUniform("projection_matrix", uiProjectionMatrix)
	
		for _, textEntity := range textEntities[i] {
			drawText(uiTextPipeline, textEntity.text, textEntity.font, textEntity.position, textEntity.color, textEntity.origin)
		}
	}
	

	// Disable SRGB rendering.
	// TODO: Ideally we want to revert to original SRGB rendering state instead of always disabling.
	graphics.DisableSRGBRendering()
	
	// Disable blending and enable depth test - "default" options.
	// TODO: Ideally we want to revert to original rendering state instead of setting to specific values.
	graphics.EnableDepthTest()
	graphics.DisableBlending()
}

// ResetUI clears lists of meshes to draw.
// Should be called right after RenderUI().
func ResetUI() {
	for i := 0; i < noLayers; i++ {
		rectEntities[i] 		= rectEntities[i][:0]
		textEntities[i] 		= textEntities[i][:0]
		texturedRectEntities[i] = texturedRectEntities[i][:0]
	}
}

// DrawUIRect sets rectangle to be drawn in UI next frame.
func DrawUIRect(pos mgl32.Vec2, size mgl32.Vec2, color mgl32.Vec4, layer int) {
	rectEntities[layer] = append(rectEntities[layer], rectData{pos, size, color})
}

// DrawUIRectTextured sets textured rectangle to be drawn in UI next frame.
func DrawUIRectTextured(pos mgl32.Vec2, size mgl32.Vec2, texture graphics.Texture, color mgl32.Vec4, layer int) {
	texturedRectEntities[layer] = append(texturedRectEntities[layer], texturedRectData{pos, size, texture, color})
}

// DrawUIText sets text to be drawn in UI next frame.
func DrawUIText(text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2, layer int) {
	textEntities[layer] = append(textEntities[layer], textData{font, text, position, color, origin})
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
	x, y := pos[0], -pos[1]
	sx, sy := size[0], size[1]
	modelMatrix := getModelMatrix(x, y, sx, sy)

	// Render rectangle.
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	graphics.DrawMesh(uiQuad)
}

func drawRectTexture(pipeline graphics.Pipeline, pos mgl32.Vec2, size mgl32.Vec2, texture graphics.Texture, color mgl32.Vec4) {
	// Set up transformation parameters.
	x, y := pos[0], -pos[1]
	sx, sy := size[0], size[1]
	modelMatrix := getModelMatrix(x, y, sx, sy)

	// Render textured rectangle.
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	pipeline.SetUniform("source_rect", mgl32.Vec4{0,0,1,1})
	graphics.SetTexture(texture, 0)
	graphics.DrawMesh(uiQuad)
}

// getGlyphSourceRect converts absolute (pixel) values of glyph pos/size into texcoord space.
func getGlyphSourceRect(glyph font.Glyph, textureWidth, textureHeight int) mgl32.Vec4{
	relX := float32(glyph.X) / float32(textureWidth)
	relY := 1.0 - float32(glyph.Y + glyph.BitmapHeight) / float32(textureHeight)
	relWidth := float32(glyph.BitmapWidth) / float32(textureWidth)
	relHeight := float32(glyph.BitmapHeight) / float32(textureHeight)
	return mgl32.Vec4{relX, relY, relWidth, relHeight}
}

func getFontTexture(font *font.Font) graphics.Texture {
	texture, ok := fontTextureCache[font]
	if !ok {
		texture = graphics.GetTextureUint8(font.TextureWidth, font.TextureHeight, 1, font.Texture, true)
		fontTextureCache[font] = texture
	}
	return texture
}

func drawText(pipeline graphics.Pipeline, text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2) {
	// Set font texture.
	fontTexture := getFontTexture(font)
	graphics.SetTexture(fontTexture, 0)
	
	// Set up starting position for text rendering.
	width, height := font.GetStringWidth(text), font.RowHeight
	currentX := math.Floor(float64(position[0]) - float64(width) * float64(origin[0]))
	currentY := math.Floor(float64(position[1]) - float64(height) * float64(origin[1]))
	
	// Color is the same for all the letters, so we'll set it before the loop.
	pipeline.SetUniform("color", color)

	// Render each character.
	for i, char := range text {
		glyph := font.Glyphs[char]

		// Account for kerning.
		if i > 0 {
			currentX += float64(font.GetKerning(char, rune(text[i-1])))
		}
		
		// Get and set model transform matrix.
		x := float32(currentX + glyph.XOffset)
		y := -float32(currentY + float64(glyph.YOffset))
		sx := float32(glyph.Width)
		sy := float32(glyph.Height)
		modelMatrix := getModelMatrix(x, y, sx, sy)
		pipeline.SetUniform("model_matrix", modelMatrix)
		
		// Get source rectangle for the font texture.
		sourceRect := getGlyphSourceRect(glyph, font.TextureWidth, font.TextureHeight)
		pipeline.SetUniform("source_rect", sourceRect)

		// Draw a current letter.
		graphics.DrawMesh(uiQuad)
		
		// Update horizontal rendering position.
		currentX += float64(glyph.Advance)
	}
}