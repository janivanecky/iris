package app

import (
	gmath "../lib/math"
	"../lib/font"
	"../lib/graphics"
	"../lib/ui"
	"github.com/go-gl/gl/v4.1-core/gl"
	"math"
	"io/ioutil"
	"fmt"
)

var program_text graphics.Program
var textModelMatrixUniform graphics.Uniform
var projectionMatrixTextUniform graphics.Uniform
var sourceRectUniform graphics.Uniform
var textColorUniform graphics.Uniform
var projectionMatrix gmath.Matrix4x4

var program_rect graphics.Program
var projectionMatrixRectUniform graphics.Uniform
var rectColorUniform graphics.Uniform
var rectModelMatrixUniform graphics.Uniform

var screenWidth float64 = 0.0
var screenHeight float64 = 0.0

var quad graphics.Mesh
var aspectRatio float64 = 1.0

var viewMatrixUniform graphics.Uniform
var projectionMatrixMeshUniform graphics.Uniform
var program_mesh graphics.Program

var fontTexture graphics.Texture
var fontUI font.Font

func InitRendering(windowWidth float64, windowHeight float64, uiFont font.Font) {
	rectEntities = make([]rectData, 0, 100)
	textEntities = make([]textData, 0, 100)
	meshEntities = make([]meshData, 0, 100)
	
	graphics.Init()
	fontUI = uiFont
	ui.Init(windowWidth, windowHeight, uiFont)
	fontTexture = graphics.GetTexture(512, 512, uiFont.Texture)
	
	screenWidth = windowWidth
	screenHeight = windowHeight

	// Text rendering shaders
	vertexShaderData, err := ioutil.ReadFile("shaders/text_vertex_shader.glsl")
	vertexShader, err := graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
	if err != nil {
		fmt.Println(err)
	}

	pixelShaderData, err := ioutil.ReadFile("shaders/text_pixel_shader.glsl")
	pixelShader, err := graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
	if err != nil {
		fmt.Println(err)
	}

	program_text, err = graphics.GetProgram(vertexShader, pixelShader)
	if err != nil {
		fmt.Println(err)
	}

	graphics.ReleaseShaders(vertexShader, pixelShader)
	graphics.SetProgram(program_text)

	projectionMatrixTextUniform = graphics.GetUniform(program_text, "projection_matrix")
	sourceRectUniform = graphics.GetUniform(program_text, "source_rect")
	textModelMatrixUniform = graphics.GetUniform(program_text, "model_matrix")
	textColorUniform = graphics.GetUniform(program_text, "color")

	// Rect rendering shaders
	vertexShaderData, err = ioutil.ReadFile("shaders/rect_vertex_shader.glsl")
	vertexShader, err = graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
	if err != nil {
		fmt.Println(err)
	}

	pixelShaderData, err = ioutil.ReadFile("shaders/rect_pixel_shader.glsl")
	pixelShader, err = graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
	if err != nil {
		fmt.Println(err)
	}

	program_rect, err = graphics.GetProgram(vertexShader, pixelShader)
	if err != nil {
		fmt.Println(err)
	}

	graphics.ReleaseShaders(vertexShader, pixelShader)
	graphics.SetProgram(program_rect)

	projectionMatrixRectUniform = graphics.GetUniform(program_rect, "projection_matrix")
	rectColorUniform = graphics.GetUniform(program_rect, "color")
	rectModelMatrixUniform = graphics.GetUniform(program_rect, "model_matrix")
    
    projectionMatrix = gmath.GetOrthographicProjectionGLRH(0.0, screenWidth,
                                                           0.0, screenHeight,
														   10.0, -10.0)

	quad = graphics.GetMesh(quadVertices[:], quadIndices[:], []int{4, 2})





	aspectRatio = screenWidth / screenHeight
	
	{
		// Vertex shader
		vertexShaderData, err := ioutil.ReadFile("shaders/simple_vertex_shader.glsl")
		vertexShader, err := graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
		if err != nil {
			fmt.Println(err)
		}

		pixelShaderData, err := ioutil.ReadFile("shaders/simple_pixel_shader.glsl")
		pixelShader, err := graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
		if err != nil {
			fmt.Println(err)
		}

		program_mesh, err = graphics.GetProgram(vertexShader, pixelShader)
		if err != nil {
			fmt.Println(err)
		}

		graphics.ReleaseShaders(vertexShader, pixelShader)
		graphics.SetProgram(program_mesh)

		projectionMatrix := gmath.GetPerspectiveProjectionGLRH(60.0*math.Pi/180.0, aspectRatio, 0.01, 1000.0)
		projectionMatrixMeshUniform = graphics.GetUniform(program_mesh, "projection_matrix")
		graphics.SetUniformMatrix(projectionMatrixMeshUniform, projectionMatrix)

		viewMatrix = gmath.GetTranslation(0.0, 0.0, -5.0)
		viewMatrixUniform = graphics.GetUniform(program_mesh, "view_matrix")
		graphics.SetUniformMatrix(viewMatrixUniform, viewMatrix)

		lightPos := gmath.Vec3{10, 20, 30}
		lightPositionUniform := graphics.GetUniform(program_mesh, "light_position")
		graphics.SetUniformVec3(lightPositionUniform, lightPos)
	}
}

type textData struct {
	font *font.Font
	text string
	pos gmath.Vec2
	color gmath.Vec4
	origin gmath.Vec2
}

var textEntities []textData

func DrawText(text string, font *font.Font, position gmath.Vec2, color gmath.Vec4, origin gmath.Vec2) {
	textEntities = append(textEntities, textData{font, text, position, color, origin})
}

func drawText(text string, font *font.Font, position gmath.Vec2, color gmath.Vec4, origin gmath.Vec2) {
	graphics.SetProgram(program_text)
	graphics.SetTexture(fontTexture, 0)
	graphics.SetUniformMatrix(projectionMatrixTextUniform, projectionMatrix)
	
	width, height := font.GetStringWidth(text), font.RowHeight
	x := math.Floor(float64(position[0]) - width * float64(origin[0]))
	y := math.Floor(float64(position[1]) + font.TopPad - float64(height) * float64(origin[1]))
    texWidth := float32(512.0)
	for _, char := range text {
        glyphA := font.Glyphs[char]
		relX := float32(glyphA.X) / texWidth
		relY := 1.0 - float32(glyphA.Y + glyphA.BitmapHeight) / texWidth
		relWidth := float32(glyphA.BitmapWidth) / texWidth
		relHeight := float32(glyphA.BitmapHeight) / texWidth
        sourceRect := gmath.Vec4{relX,relY,relWidth,relHeight}
		graphics.SetUniformVec4(sourceRectUniform, sourceRect)
		graphics.SetUniformVec4(textColorUniform, color)

		currentX := x + glyphA.XOffset
		currentY := y + glyphA.YOffset
		modelMatrix := gmath.Matmul(
			gmath.GetTranslation(currentX, screenHeight - currentY, 0),
			gmath.Matmul(
                gmath.GetScale(glyphA.Width, glyphA.Height, 1.0),
				gmath.GetTranslation(0.5, -0.5, 0.0),
			),
		)
		graphics.SetUniformMatrix(textModelMatrixUniform, modelMatrix)			
		
		graphics.DrawMesh(quad)
		
		x += float64(glyphA.Advance)
	}
}

var viewMatrix gmath.Matrix4x4
func SetCameraPosition(position gmath.Vec3) {
	viewMatrix = gmath.GetLookAt(position, gmath.Vec3{}, gmath.Vec3{0, 1, 0})
}

type meshData struct {
	mesh graphics.Mesh
}

var meshEntities []meshData

func DrawMesh(mesh graphics.Mesh) {
	meshEntities = append(meshEntities, meshData{mesh})
}

func drawMesh(mesh graphics.Mesh) {
	graphics.SetUniformMatrix(viewMatrixUniform, viewMatrix)
	graphics.SetProgram(program_mesh)

	// Draw scene.
	projectionMatrix := gmath.GetPerspectiveProjectionGLRH(60.0*math.Pi/180.0, aspectRatio, 0.01, 100.0)
	graphics.SetUniformMatrix(projectionMatrixMeshUniform, projectionMatrix)

	graphics.SetUniformMatrix(viewMatrixUniform, viewMatrix)
	graphics.DrawMesh(mesh)
}

type rectData struct {
	pos gmath.Vec2
	size gmath.Vec2
	color gmath.Vec4
}

var rectEntities []rectData

func DrawRect(pos gmath.Vec2, size gmath.Vec2, color gmath.Vec4) {
	rectEntities = append(rectEntities, rectData{pos, size, color})
}

func drawRect(pos gmath.Vec2, size gmath.Vec2, color gmath.Vec4) {
	graphics.SetProgram(program_rect)
	graphics.SetUniformMatrix(projectionMatrixRectUniform, projectionMatrix)
	graphics.SetUniformVec4(rectColorUniform, color)
		
	x, y := pos[0], float32(screenHeight) - pos[1]
	modelMatrix := gmath.Matmul(
		gmath.GetTranslation(float64(x), float64(y), 0),
		gmath.Matmul(
			gmath.GetScale(float64(size[0]), float64(size[1]), 1.0),
			gmath.GetTranslation(0.5, -0.5, 0.0),
		),
	)
	graphics.SetUniformMatrix(rectModelMatrixUniform, modelMatrix)
		
	graphics.DrawMesh(quad)
}

func Render() {
	// We got the cleaning done bitchez.
	graphics.ClearScreen(0, 0, 0, 0)

	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	
	for _, meshEntity := range meshEntities {
		drawMesh(meshEntity.mesh)
	}

	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	
	rectRenderingBuffer, textRenderingBuffer := ui.GetDrawData()
	
	// UI Rendering
	for _, rectData := range rectRenderingBuffer {
		drawRect(rectData.Position, rectData.Size, rectData.Color)
	}
	
	for _, textData := range textRenderingBuffer {
		drawText(textData.Text, &fontUI, textData.Position, textData.Color, textData.Origin)
	}

	// Custom rect/text rendering
	for _, rectEntity := range rectEntities {
		drawRect(rectEntity.pos, rectEntity.size, rectEntity.color)
	}

	for _, textEntity := range textEntities {
		drawText(textEntity.text, textEntity.font, textEntity.pos, textEntity.color, textEntity.origin)
	}
	
	rectEntities = rectEntities[:0]
	meshEntities = meshEntities[:0]
	textEntities = textEntities[:0]
	ui.Clear()
}

var quadVertices = [...]float32{
	-0.5, -0.5, 0.0, 1.0,
	0.0, 0.0,
	-0.5, 0.5, 0.0, 1.0,
	0.0, 1.0,
	0.5, 0.5, 0.0, 1.0,
	1.0, 1.0,
	0.5, -0.5, 0.0, 1.0,
	1.0, 0.0,
}

var quadIndices = [...]uint32{
	0, 1, 2,
	0, 2, 3,
}