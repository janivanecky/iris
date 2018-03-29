package app

import (
	"fmt"
	"io/ioutil"
	"math"
	
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v4.1-core/gl"

	"../lib/font"
	"../lib/graphics"
	"../lib/ui"
)

type pipelineData struct {
	textProgram graphics.Program
	textProjectionMatrixUniform graphics.Uniform
	textModelMatrixUniform graphics.Uniform
	textSourceRectUniform graphics.Uniform
	textColorUniform graphics.Uniform
	
	rectProgram graphics.Program
	rectProjectionMatrixUniform graphics.Uniform
	rectModelMatrixUniform graphics.Uniform
	rectColorUniform graphics.Uniform

	sceneViewMatrixUniform graphics.Uniform
	sceneProjectionMatrixUniform graphics.Uniform
	sceneLightPositionUniform graphics.Uniform
	sceneProgram graphics.Program
}
var pipelines pipelineData

type textData struct {
	font *font.Font
	text string
	pos mgl32.Vec2
	color mgl32.Vec4
	origin mgl32.Vec2
}
var textEntities []textData

type meshData struct {
	mesh graphics.Mesh
}
var meshEntities []meshData

type rectData struct {
	pos mgl32.Vec2
	size mgl32.Vec2
	color mgl32.Vec4
}
var rectEntities []rectData

type uiSettings struct {
	font font.Font
	fontTexture graphics.Texture
	quad graphics.Mesh
	projectionMatrix mgl32.Mat4
}
var uiData uiSettings

type sceneSettings struct {
	projectionMatrix mgl32.Mat4
	viewMatrix mgl32.Mat4
	lightPosition mgl32.Vec3
}
var sceneData sceneSettings

var screenWidth float64
var screenHeight float64

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

	// Initialize UI rendering data
	uiData = uiSettings{
		uiFont,
		graphics.GetTexture(512, 512, uiFont.Texture),
		graphics.GetMesh(quadVertices[:], quadIndices[:], []int{4, 2}),
		//mgl32.GetOrthographicProjectionGLRH(0.0, screenWidth, 0.0, screenHeight, 10.0, -10.0),
		mgl32.Ortho(0.0, float32(screenWidth), 0.0, float32(screenHeight), 10.0, -10.0),
	}

	// Initialize scene rednering data
	sceneData = sceneSettings{
		mgl32.Perspective(mgl32.DegToRad(60.0), float32(screenWidth / screenHeight), 0.01, 100.0),
		mgl32.Ident4(),
		mgl32.Vec3{10, 20, 30},
	}
	
	// Text rendering program setup
	var err error
	pipelines.textProgram, err = getProgram("shaders/text_vertex_shader.glsl", "shaders/text_pixel_shader.glsl")
	if err != nil {
		fmt.Println(err)
	}

	graphics.SetProgram(pipelines.textProgram)
	pipelines.textProjectionMatrixUniform = graphics.GetUniform(pipelines.textProgram, "projection_matrix")
	pipelines.textModelMatrixUniform = graphics.GetUniform(pipelines.textProgram, "model_matrix")
	pipelines.textSourceRectUniform = graphics.GetUniform(pipelines.textProgram, "source_rect")
	pipelines.textColorUniform = graphics.GetUniform(pipelines.textProgram, "color")

	// Rect rendering program setup
	pipelines.rectProgram, err = getProgram("shaders/rect_vertex_shader.glsl", "shaders/rect_pixel_shader.glsl")
	if err != nil {
		fmt.Println(err)
	}

	graphics.SetProgram(pipelines.rectProgram)
	pipelines.rectProjectionMatrixUniform = graphics.GetUniform(pipelines.rectProgram, "projection_matrix")
	pipelines.rectModelMatrixUniform = graphics.GetUniform(pipelines.rectProgram, "model_matrix")
	pipelines.rectColorUniform = graphics.GetUniform(pipelines.rectProgram, "color")
    
	// 3D rendering program setup
	pipelines.sceneProgram, err = getProgram("shaders/simple_vertex_shader.glsl", "shaders/simple_pixel_shader.glsl")
	if err != nil {
		fmt.Println(err)
	}

	graphics.SetProgram(pipelines.sceneProgram)
	pipelines.sceneProjectionMatrixUniform = graphics.GetUniform(pipelines.sceneProgram, "projection_matrix")
	pipelines.sceneViewMatrixUniform = graphics.GetUniform(pipelines.sceneProgram, "view_matrix")
	pipelines.sceneLightPositionUniform = graphics.GetUniform(pipelines.sceneProgram, "light_position")
}

func Render() {
	// We got the cleaning done bitchez.
	graphics.ClearScreen(0, 0, 0, 0)

	// Set up 3D scene rendering settings
	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	
	// Set up 3D scene rendering pipeline
	graphics.SetProgram(pipelines.sceneProgram)
	graphics.SetUniformMatrix(pipelines.sceneViewMatrixUniform, sceneData.viewMatrix)
	graphics.SetUniformMatrix(pipelines.sceneProjectionMatrixUniform, sceneData.projectionMatrix)
	graphics.SetUniformVec3(pipelines.sceneLightPositionUniform, sceneData.lightPosition)
	
	// Render meshes
	for _, meshEntity := range meshEntities {
		drawMesh(meshEntity.mesh)
	}

	// Set up 2D UI rendering settings
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	
	// Get UI rendering buffers
	rectRenderingBuffer, textRenderingBuffer := ui.GetDrawData()

	// Set up 2D rect rendering pipeline
	graphics.SetProgram(pipelines.rectProgram)
	graphics.SetUniformMatrix(pipelines.rectProjectionMatrixUniform, uiData.projectionMatrix)
	
	// Render UI rects
	for _, rectData := range rectRenderingBuffer {
		drawRect(rectData.Position, rectData.Size, rectData.Color)
	}
	
	// Render the rest of rects
	for _, rectEntity := range rectEntities {
		drawRect(rectEntity.pos, rectEntity.size, rectEntity.color)
	}

	// Set up 2D text rendering pipeline
	graphics.SetProgram(pipelines.textProgram)
	graphics.SetTexture(uiData.fontTexture, 0)
	graphics.SetUniformMatrix(pipelines.textProjectionMatrixUniform, uiData.projectionMatrix)
	
	// Render UI text
	for _, textData := range textRenderingBuffer {
		drawText(textData.Text, &(uiData.font), textData.Position, textData.Color, textData.Origin)
	}

	// Render the rest of texts
	for _, textEntity := range textEntities {
		drawText(textEntity.text, textEntity.font, textEntity.pos, textEntity.color, textEntity.origin)
	}
	
	// Clear rendering lists + UI
	rectEntities = rectEntities[:0]
	meshEntities = meshEntities[:0]
	textEntities = textEntities[:0]
	ui.Clear()
}

func DrawText(text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2) {
	textEntities = append(textEntities, textData{font, text, position, color, origin})
}

func drawText(text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2) {
	graphics.SetUniformVec4(pipelines.textColorUniform, color)
	
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
        sourceRect := mgl32.Vec4{relX,relY,relWidth,relHeight}
		graphics.SetUniformVec4(pipelines.textSourceRectUniform, sourceRect)

		currentX := x + glyphA.XOffset
		currentY := y + glyphA.YOffset
		modelMatrix := mgl32.Translate3D(float32(currentX), float32(screenHeight - currentY), 0).Mul4(
                mgl32.Scale3D(float32(glyphA.Width), float32(glyphA.Height), 1.0).Mul4(
					mgl32.Translate3D(0.5, -0.5, 0.0),
			),
		)
		graphics.SetUniformMatrix(pipelines.textModelMatrixUniform, modelMatrix)			
		
		graphics.DrawMesh(uiData.quad)
		
		x += float64(glyphA.Advance)
	}
}

func DrawMesh(mesh graphics.Mesh) {
	meshEntities = append(meshEntities, meshData{mesh})
}

func drawMesh(mesh graphics.Mesh) {
	graphics.DrawMesh(mesh)
}

func DrawRect(pos mgl32.Vec2, size mgl32.Vec2, color mgl32.Vec4) {
	rectEntities = append(rectEntities, rectData{pos, size, color})
}

func drawRect(pos mgl32.Vec2, size mgl32.Vec2, color mgl32.Vec4) {
	graphics.SetUniformVec4(pipelines.rectColorUniform, color)
		
	x, y := pos[0], float32(screenHeight) - pos[1]
	modelMatrix := mgl32.Translate3D(x, y, 0).Mul4(
			mgl32.Scale3D(size[0], size[1], 1.0).Mul4(
				mgl32.Translate3D(0.5, -0.5, 0.0),
		),
	)
	graphics.SetUniformMatrix(pipelines.rectModelMatrixUniform, modelMatrix)

	graphics.DrawMesh(uiData.quad)
}

func SetCameraPosition(position mgl32.Vec3) {
	sceneData.viewMatrix = mgl32.LookAtV(position, mgl32.Vec3{}, mgl32.Vec3{0, 1, 0})
}

func getProgram(vertexShaderPath string, pixelShaderPath string) (graphics.Program, error) {
	vertexShaderData, err := ioutil.ReadFile(vertexShaderPath)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}
	
	vertexShader, err := graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}

	pixelShaderData, err := ioutil.ReadFile(pixelShaderPath)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}

	pixelShader, err := graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}

	program, err := graphics.GetProgram(vertexShader, pixelShader)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}
	graphics.ReleaseShaders(vertexShader, pixelShader)

	return program, nil
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