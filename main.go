package main

import (
	"fmt"
	"strings"
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const simple_vertex_shader_text = `
#version 330 core
in vec4 in_position;

void main()
{
	gl_Position = in_position;
}
`

const simple_pixel_shader_text = `
#version 330 core
out vec4 out_color;

void main()
{
	out_color = vec4(0,1,0,1);
}
`

type Vec4 struct {
	x float32
	y float32
	z float32
	w float32
}

var quad_vertices = [...]Vec4 {
	{-0.5, -0.5, 0.5, 1.0},
	{-0.5, 0.5, 0.5, 1.0},
	{0.5, 0.5, 0.5, 1.0},
	{0.5, -0.5, 0.5, 1.0},
}

var quad_indices = [...]uint32 {
	0, 1, 2,
	0, 2, 3,
}

func main() {
	// Initialize GLFW context
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// Create our new fancy window
	window, err := glfw.CreateWindow(1920, 1080, "New fancy window", nil, nil)
	if err != nil {
		panic(err)
	}

	// God knows why this is necessary
	window.MakeContextCurrent()

	// Create an OpenGL context
	err = gl.Init()
	if err != nil {
		panic(err)
	}

	{
		// Vertex shader
		vertex_shader := gl.CreateShader(gl.VERTEX_SHADER);
		vertex_shader_source, free_vertex_shader := gl.Strs(simple_vertex_shader_text)
		gl.ShaderSource(vertex_shader, 1, vertex_shader_source, nil);
		free_vertex_shader()
	
		gl.CompileShader(vertex_shader);
		
		var vertex_shader_compile_status int32
		gl.GetShaderiv(vertex_shader, gl.COMPILE_STATUS, &vertex_shader_compile_status);
		if vertex_shader_compile_status == gl.FALSE {
			var log_length int32
			gl.GetShaderiv(vertex_shader, gl.INFO_LOG_LENGTH, &log_length)
	
			log := strings.Repeat("\x00", int(log_length + 1))
			gl.GetShaderInfoLog(vertex_shader, log_length, nil, gl.Str(log))
	
			fmt.Println("failed to compile vertex_shader: %v", log)
			return
		}
	
		pixel_shader := gl.CreateShader(gl.FRAGMENT_SHADER)
		pixel_shader_source, free_pixel_shader := gl.Strs(simple_pixel_shader_text)
		gl.ShaderSource(pixel_shader, 1, pixel_shader_source, nil);
		free_pixel_shader()
	
		gl.CompileShader(pixel_shader);
		
		var pixel_shader_compile_status int32
		gl.GetShaderiv(pixel_shader, gl.COMPILE_STATUS, &pixel_shader_compile_status);
		if pixel_shader_compile_status == gl.FALSE {
			var log_length int32
			gl.GetShaderiv(pixel_shader, gl.INFO_LOG_LENGTH, &log_length)
	
			log := strings.Repeat("\x00", int(log_length + 1))
			gl.GetShaderInfoLog(pixel_shader, log_length, nil, gl.Str(log))
	
			fmt.Println("failed to compile pixel_shader: %v", log)
			return
		}
		
		// shader Program
		program := gl.CreateProgram()
		gl.AttachShader(program, vertex_shader)
		gl.AttachShader(program, pixel_shader)
		gl.LinkProgram(program)
		// print linking errors if any
		var program_linking_status int32
		gl.GetProgramiv(program, gl.LINK_STATUS, &program_linking_status)
		if program_linking_status == gl.FALSE {
			var log_length int32
			gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &log_length)
	
			log := strings.Repeat("\x00", int(log_length + 1))
			gl.GetProgramInfoLog(program, log_length, nil, gl.Str(log))
	
			fmt.Println("failed to link program: %v", log)
			return
		}

		// delete the shaders as they're linked into our program now and no longer necessery
		gl.DeleteShader(vertex_shader)
		gl.DeleteShader(pixel_shader)

		gl.UseProgram(program)
	}

	var quad_vao uint32
	{
		gl.GenVertexArrays(1, &quad_vao)
		gl.BindVertexArray(quad_vao)

		var quad_vbo uint32
		gl.GenBuffers(1, &quad_vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, quad_vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(quad_vertices) * 4 * 4, gl.Ptr(&quad_vertices[0].x), gl.STATIC_DRAW)
		var position_size int32 = 4
		var position_size_in_bytes int32 = 4 * position_size
		gl.VertexAttribPointer(0, position_size, gl.FLOAT, false, position_size_in_bytes, gl.PtrOffset(0));
		gl.EnableVertexAttribArray(0);  
		
		var quad_ebo uint32
		gl.GenBuffers(1, &quad_ebo)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, quad_ebo);
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(quad_indices) * 4, gl.Ptr(&quad_indices[0]), gl.STATIC_DRAW);
	}
		
	gl.Disable(gl.DEPTH_TEST)

	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		// Let GLFW interface with the OS - not our job, right?
		glfw.PollEvents()

		// Let's quit if user presses Esc, that cannot mean anything else.
		esc_state := window.GetKey(glfw.KeyEscape)
		if esc_state == glfw.Press {
			break
		}

		// We got the cleaning done bitchez.
		gl.ClearColor(0, 0, 1, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.BindVertexArray(quad_vao)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.PtrOffset(0))
		// Do OpenGL stuff.
		window.SwapBuffers()
	}
}
