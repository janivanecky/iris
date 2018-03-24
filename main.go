package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"

	"./lib/graphics"
	_ "./lib/font"
	"./lib/ui"
	"./lib/input"
	gmath "./lib/math"

	"runtime"

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


func main() {
	window := graphics.GetWindow(800, 600, "New fancy window")
	defer graphics.ReleaseWindow()
	//return
	input.Init(window)

	ui.Init()
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, 512, 512, 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(ui.GetTex()))
	gl.GenerateMipmap(gl.TEXTURE_2D)


	
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

		projectionMatrix := gmath.GetPerspectiveProjectionGLRH(60.0*math.Pi/180.0, 800.0/600.0, 0.01, 1000.0)
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
				viewMatrix = gmath.GetLookAt(camPosition, gmath.Vec3{0, 0, 0}, gmath.Vec3{0, 1, 0})
			}
		}
		graphics.SetUniformMatrix(viewMatrixUniform, viewMatrix)

		// Draw scene.
		projectionMatrix := gmath.GetPerspectiveProjectionGLRH(60.0*math.Pi/180.0, 800.0 / 600.0, 0.01, 100.0)
		graphics.SetUniformMatrix(projectionMatrixMeshUniform, projectionMatrix)
		graphics.DrawMesh(cube)

		panel := ui.StartPanel("Test panel", gmath.Vec2{0,0})
		toggle, _ = panel.AddToggle("test", toggle)
		val, _ = panel.AddSlider("test2", val, 0, 1)
		panel.End()

		ui.Present()

		if !ui.IsRegisteringInput {
			ui.SetInputResponsive(uiResponsive)
		}

		// Swappity-swap.
		window.SwapBuffers()
	}
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
