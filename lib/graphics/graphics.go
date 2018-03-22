package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"strings"
	"fmt"
	gmath "../math"
)

// Create window with OpenGL context.
func GetWindow(width int, height int, title string) *glfw.Window  {
	// Initialize GLFW context
	err := glfw.Init()
	if err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create our new fancy window
	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		panic(err)
	}

	// God knows why this is necessary
	window.MakeContextCurrent()	
	glfw.SwapInterval(0)

	// Create an OpenGL context
	err = gl.Init()
	if err != nil {
		panic(err)
	}
	
	// Set up useful settings
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.FRAMEBUFFER_SRGB)

	return window
}

// Release GLFW context.
func ReleaseWindow() {
	glfw.Terminate()
}

type ShaderType int

const (
	VERTEX_SHADER ShaderType = iota
	PIXEL_SHADER = iota
)

type Shader uint32
type Program uint32
type Uniform int32
type Mesh struct {
	vao uint32
	indexCount int32	
}

var shaderTypeToGL = map[ShaderType]uint32 {
	VERTEX_SHADER: gl.VERTEX_SHADER,
	PIXEL_SHADER: gl.FRAGMENT_SHADER,
}

func GetShader(shaderSource string, shaderType ShaderType) (Shader, error) {
	shaderSourceString, freeShaderSourceString := gl.Strs(shaderSource + "\x00")
	shader := gl.CreateShader(shaderTypeToGL[shaderType]);
	gl.ShaderSource(shader, 1, shaderSourceString, nil);
	freeShaderSourceString()
	gl.CompileShader(shader);
	
	var shaderCompileStatus int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &shaderCompileStatus);
	if shaderCompileStatus == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength + 1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return Shader(0), fmt.Errorf("Failed to compile shader: %v", log)
	}
	return Shader(shader), nil
}

func GetProgram(shaders ...Shader) (Program, error) {
	program := gl.CreateProgram()

	for _, shader := range shaders {
		gl.AttachShader(program, uint32(shader))
	}
	gl.LinkProgram(program)
	
	var programLinkingStatus int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &programLinkingStatus)
	if programLinkingStatus == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength + 1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return Program(0), fmt.Errorf("failed to link program: %v", log)
	}
	return Program(program), nil
}

func SetProgram(program Program) {
	gl.UseProgram(uint32(program))
}

func ReleaseShaders(shaders ...Shader) {
	for _, shader := range shaders {
		gl.DeleteShader(uint32(shader))
	}
}

func GetUniform(program Program, uniformName string) Uniform {
	location := gl.GetUniformLocation(uint32(program), gl.Str(uniformName + "\x00"))
	return Uniform(location)
}

func SetUniformMatrix(uniform Uniform, matrix gmath.Matrix4x4) {
	gl.UniformMatrix4fv(int32(uniform), 1, false, &matrix[0][0])
}

func SetUniformVec3(uniform Uniform, v gmath.Vec3) {
	gl.Uniform3fv(int32(uniform), 1, &v[0])
	
}

func SetUniformVec4(uniform Uniform, v gmath.Vec4) {
	gl.Uniform4fv(int32(uniform), 1, &v[0])
	
}

func GetMesh(vertices []float32, indices []uint32, vertexAttribs []int) Mesh {
	var cubeVao uint32
	gl.GenVertexArrays(1, &cubeVao)
	gl.BindVertexArray(cubeVao)

	var cubeVbo uint32
	gl.GenBuffers(1, &cubeVbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices) * 4, gl.Ptr(&vertices[0]), gl.STATIC_DRAW)
	
	var stride int32 = 0
	for _, vertexAttrib := range vertexAttribs {
		stride += int32(vertexAttrib * 4)
	}

	offset := 0
	for i, vertexAttrib := range vertexAttribs {
		index := uint32(i)
		gl.VertexAttribPointer(index, int32(vertexAttrib), gl.FLOAT, false, stride, gl.PtrOffset(offset));
		gl.EnableVertexAttribArray(index);  
		offset += vertexAttrib * 4
	}
	
	var cubeIbo uint32
	gl.GenBuffers(1, &cubeIbo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, cubeIbo);
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices) * 4, gl.Ptr(&indices[0]), gl.STATIC_DRAW);

	return Mesh{cubeVao, int32(len(indices))}
}

func DrawMesh(mesh Mesh) {
	gl.BindVertexArray(mesh.vao)
	gl.DrawElements(gl.TRIANGLES, mesh.indexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
}

func ClearScreen(r float32, g float32, b float32, a float32) {
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}