package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"

	"./lib/graphics"
	"./lib/font"
	"./lib/ui"
	"./lib/input"
	gmath "./lib/math"

	"runtime"
	"syscall"
	"github.com/go-gl/gl/v4.1-core/gl"
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func vecFromPolarCoords(azimuth float64, polar float64, radius float64) gmath.Vec3 {
	result := gmath.Vec3{
		float32(math.Sin(polar) * math.Sin(azimuth) * radius),
		float32(math.Cos(polar) * radius),
		float32(math.Sin(polar) * math.Cos(azimuth) * radius),
	}
	return result
}

const WINDOW_WIDTH = 800
const WINDOW_HEIGHT = 600

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

var quad graphics.Mesh
var uiFont font.Font

func main() {
	window := graphics.GetWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "New fancy window")
	defer graphics.ReleaseWindow()
	//return
	input.Init(window)

	dll, err := syscall.LoadDLL("User32.dll")
	if err != nil {
		panic(err)
	}
	dpiForSystem, _ := dll.FindProc("GetDpiForSystem")
	dpi, errCode, _ := dpiForSystem.Call()
	if errCode > 0 {
		panic(errCode)
	}

	scale := float64(dpi) / 96.0
    
	truetypeBytes, err := ioutil.ReadFile("fonts/font.ttf")
	if err != nil {
		panic(err)
    }
    

    uiFont = font.GetFont(truetypeBytes, 40.0, scale)
    
	
	ui.Init(WINDOW_WIDTH, WINDOW_HEIGHT, uiFont)
	
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
    
    projectionMatrix = gmath.GetOrthographicProjectionGLRH(0.0, float64(WINDOW_WIDTH),
                                                           0.0, float64(WINDOW_HEIGHT),
                                                           10.0, -10.0)
    
    quad = graphics.GetMesh(quadVertices[:], quadIndices[:], []int{4, 2})

	var texture uint32
	// TODO: graphics create Texture function
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, 512, 512, 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(uiFont.Texture))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	aspectRatio := float64(WINDOW_WIDTH) / float64(WINDOW_HEIGHT)
	var viewMatrixUniform graphics.Uniform
	var projectionMatrixMeshUniform graphics.Uniform
	var program_mesh graphics.Program
	var viewMatrix gmath.Matrix4x4
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
	
	
	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:], []int{4, 4})

	polar := math.Pi / 2.0
	azimuth := 0.0
	radius := 5.0

	toggle := true
	val := 0.5

	start := time.Now()
	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		t := time.Now()
		elapsed := t.Sub(start)
		start = t
		fmt.Errorf("Dt", elapsed)

		// Let GLFW interface with the OS - not our job, right?
		glfw.PollEvents()
		input.Update(window)

		// Let's quit if user presses Esc, that cannot mean anything else.
		escState := window.GetKey(glfw.KeyEscape)
		if escState == glfw.Press {
			break
		}

		// Update mouse position and get position delta.
		dx, dy := input.GetMouseDeltaPosition()

		// We got the cleaning done bitchez.
		graphics.ClearScreen(0, 0, 0, 0)
		gl.Disable(gl.BLEND)
		gl.Enable(gl.DEPTH_TEST)

		graphics.SetProgram(program_mesh)

		camChanged := false
		uiResponsive := true
		if !ui.IsRegisteringInput {
			if input.IsMouseLeftButtonDown() {
				azimuth -= dx / 100.0
				polar -= dy / 100.0
				camChanged = true
				uiResponsive = false
			}
	
			mouseWheelDelta := input.GetMouseWheelDelta()
			if mouseWheelDelta != 0.0 {
				radius -= mouseWheelDelta / 2.0
				camChanged = true
			}
	
			if camChanged {
				camPosition := vecFromPolarCoords(azimuth, polar, radius)
				viewMatrix = gmath.GetLookAt(camPosition, gmath.Vec3{}, gmath.Vec3{0, 1, 0})
			}
		}
		graphics.SetUniformMatrix(viewMatrixUniform, viewMatrix)

		// Draw scene.
		projectionMatrix := gmath.GetPerspectiveProjectionGLRH(60.0*math.Pi/180.0, aspectRatio, 0.01, 100.0)
		graphics.SetUniformMatrix(projectionMatrixMeshUniform, projectionMatrix)
		graphics.DrawMesh(cube)


		panel := ui.StartPanel("Test panel", gmath.Vec2{})
		toggle, _ = panel.AddToggle("test", toggle)
		val, _ = panel.AddSlider("test2", val, 0, 1)
		panel.End()

		
		rectRenderingBuffer, textRenderingBuffer := ui.GetDrawData()
		
		for _, rectData := range rectRenderingBuffer {
			DrawRect(rectData.Position, rectData.Size, rectData.Color)
		}
		
		for _, textData := range textRenderingBuffer {
			DrawText(textData.Text, &uiFont, textData.Position, textData.Color, textData.Origin)
		}
		
		if !ui.IsRegisteringInput {
			ui.SetInputResponsive(uiResponsive)
		}
		DrawRect(gmath.Vec2{400,0}, gmath.Vec2{200,40}, gmath.Vec4{1,0,0,1})
		DrawText("TEST", &uiFont, gmath.Vec2{400,0}, gmath.Vec4{0,0,1,0}, gmath.Vec2{})

		ui.Clear()
		
		// Swappity-swap.
		window.SwapBuffers()
	}
}


func DrawText(text string, font *font.Font, position gmath.Vec2, color gmath.Vec4, origin gmath.Vec2) {
	graphics.SetProgram(program_text)
	graphics.SetUniformMatrix(projectionMatrixTextUniform, projectionMatrix)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	
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
			gmath.GetTranslation(currentX, WINDOW_HEIGHT - currentY, 0),
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

func DrawRect(pos gmath.Vec2, size gmath.Vec2, color gmath.Vec4) {
	graphics.SetProgram(program_rect)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	graphics.SetUniformMatrix(projectionMatrixRectUniform, projectionMatrix)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	
	graphics.SetUniformVec4(rectColorUniform, color)
	
	x, y := pos[0], float32(WINDOW_HEIGHT) - pos[1]
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