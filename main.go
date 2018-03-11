package main

import (
	"fmt"
	"math"
	"time"
	gmath "./lib/math"
	"github.com/go-gl/glfw/v3.2/glfw"
	"./lib/graphics"
    "io/ioutil"
)

const simpleVertexShaderText = `
#version 330 core
in vec4 in_position;
in vec4 in_normal;

out vec4 normal;

uniform mat4 projection_matrix;
uniform mat4 view_matrix;

void main()
{
	gl_Position = projection_matrix * view_matrix * in_position;
	normal = in_normal;
}
` + "\x00"

const simplePixelShaderText = `
#version 330 core
out vec4 out_color;
in vec4 normal;

uniform vec3 light_position;

void main()
{
	vec3 light_direction = normalize(light_position);
	float d = clamp(dot(light_direction, normal.xyz), 0, 1);
	out_color.xyz = vec3(0,1,0) * d;
	out_color.a = 1.0f;
}
` + "\x00"

var quadVertices = [...]float32 {
	-0.5, -0.5, 0.0, 1.0,
	-0.5, 0.5, 0.0, 1.0,
	0.5, 0.5, 0.0, 1.0,
	0.5, -0.5, 0.0, 1.0,
}

var cubeVertices = [...]float32 {
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

var quadIndices = [...]uint32 {
	0, 1, 2,
	0, 2, 3,
}

var cubeIndices = [...]uint32 {
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

func main() {
	window := graphics.GetWindow(1920, 1080, "New fancy window")
	defer graphics.ReleaseWindow()

	var viewMatrixUniform graphics.Uniform
	{
		// Vertex shader
		vertexShaderData, err := ioutil.ReadFile("shaders/simple_vertex_shader.glsl")
		vertexShader, err := graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
		if err != nil{
			fmt.Println(err)
		}
		
		pixelShaderData, err := ioutil.ReadFile("shaders/simple_pixel_shader.glsl")
		pixelShader, err := graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
		if err != nil{
			fmt.Println(err)
		}
		
		program, err := graphics.GetProgram(vertexShader, pixelShader);
		if err != nil{
			fmt.Println(err)
		}

		graphics.ReleaseShaders(vertexShader, pixelShader)
		graphics.SetProgram(program)

		projectionMatrix := gmath.GetPerspectiveProjectionGLRH(60.0 * math.Pi / 180.0, 1920.0 / 1080.0, 0.01, 10.0)
		projectionMatrixUniform := graphics.GetUniform(program, "projection_matrix")
		graphics.SetUniformMatrix(projectionMatrixUniform, projectionMatrix)

		viewMatrix := gmath.GetTranslation(0.0, 0.0, -5.0)
		viewMatrixUniform = graphics.GetUniform(program, "view_matrix")
		graphics.SetUniformMatrix(viewMatrixUniform, viewMatrix)

		lightPos := gmath.Vec3{10, 20, 30};
		lightPositionUniform := graphics.GetUniform(program, "light_position")
		graphics.SetUniformVec3(lightPositionUniform, lightPos)
	}

	//quad := graphics.GetMesh(quadVertices[:], quadIndices[:])
	cube := graphics.GetMesh(cubeVertices[:], cubeIndices[:])

	polar := math.Pi / 2.0
	azimuth := 0.0
	radius := 5.0

	mx, my := -1.0, -1.0

	start := time.Now()
	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		t := time.Now()
		elapsed := t.Sub(start)
		start = t
		fmt.Println("Dt", elapsed)
		
		// Let GLFW interface with the OS - not our job, right?
		glfw.PollEvents()

		// Let's quit if user presses Esc, that cannot mean anything else.
		escState := window.GetKey(glfw.KeyEscape)
		if escState == glfw.Press {
			break
		}

		x, y := window.GetCursorPos()
		dx, dy := 0.0, 0.0
		if mx > 0.0 && my > 0.0 {
			dx, dy = x - mx, y - my
		} 
		mx, my = x, y
		lButtonState := window.GetMouseButton(glfw.MouseButtonLeft);
		if lButtonState == glfw.Press {
			azimuth -= dx / 100.0
			polar -= dy / 100.0

			camPosition := gmath.Vec3{
				float32(math.Sin(polar) * math.Sin(azimuth) * radius),
				float32(math.Cos(polar) * radius),
				float32(math.Sin(polar) * math.Cos(azimuth) * radius),
			}
			viewMatrix := gmath.GetLookAt(camPosition, gmath.Vec3{0,0,0}, gmath.Vec3{0,1,0})
			graphics.SetUniformMatrix(viewMatrixUniform, viewMatrix)
		}

		// We got the cleaning done bitchez.
		graphics.ClearScreen(0,0,0,0);

		graphics.DrawMesh(cube)
		// Do OpenGL stuff.
		window.SwapBuffers()
	}
}
