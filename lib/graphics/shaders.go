package graphics

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Uniform is a handle to OpenGL uniform value location.
type Uniform int32
// Shader is a handle to OpenGL shader.
type Shader uint32
// Program is a handle to OpenGL program.
type Program uint32
// ShaderType is an enum specifying type of shader - one of {VertexShader | PixelShader | GeometryShader}.
type ShaderType int
const (
	// VertexShader represents vertex shader type.
	VertexShader ShaderType = iota
	// PixelShader represents pixel shader type.
	PixelShader ShaderType = iota
	// GeometryShader represents geometry shader type.
	GeometryShader ShaderType = iota
)

var shaderTypeToGL = map[ShaderType]uint32 {
	VertexShader: gl.VERTEX_SHADER,
	PixelShader: gl.FRAGMENT_SHADER,
	GeometryShader: gl.GEOMETRY_SHADER,
}

// GetShader returns a handle to OpenGL Shader.
func GetShader(shaderSource string, shaderType ShaderType) (Shader, error) {
	// Need to add null terminated character.
	shaderSourceString, freeShaderSourceString := gl.Strs(shaderSource + "\x00")

	// Compile a shader.
	shader := gl.CreateShader(shaderTypeToGL[shaderType]);
	gl.ShaderSource(shader, 1, shaderSourceString, nil);
	gl.CompileShader(shader);
	freeShaderSourceString()

	// Handle compilation failure.
	var shaderCompileStatus int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &shaderCompileStatus);
	if shaderCompileStatus == gl.FALSE {
		// Get error log length.
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		// Get error log itself.
		log := strings.Repeat("\x00", int(logLength + 1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return Shader(0), fmt.Errorf("Failed to compile shader: %v", log)
	}
	return Shader(shader), nil
}

// GetProgram returns a handle to OpenGL Program.
func GetProgram(shaders ...Shader) (Program, error) {
	// Create an OpenGL program.
	program := gl.CreateProgram()

	// Attach shaders and link.
	for _, shader := range shaders {
		gl.AttachShader(program, uint32(shader))
	}
	gl.LinkProgram(program)
	
	// Handle linking failure.
	var programLinkingStatus int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &programLinkingStatus)
	if programLinkingStatus == gl.FALSE {
		// Get error log length.
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		// Get error log itself.
		log := strings.Repeat("\x00", int(logLength + 1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return Program(0), fmt.Errorf("Failed to link program: %v", log)
	}
	return Program(program), nil
}

// SetProgram sets specific program for use.
func SetProgram(program Program) {
	gl.UseProgram(uint32(program))
}

// ReleaseShaders releases multiple shaders.
func ReleaseShaders(shaders ...Shader) {
	for _, shader := range shaders {
		gl.DeleteShader(uint32(shader))
	}
}

// GetUniform returns Uniform handle for specific uniform value in program.
func GetUniform(program Program, uniformName string) Uniform {
	location := gl.GetUniformLocation(
		uint32(program),
		gl.Str(uniformName + "\x00")) // Need to add null terminating character to string.
	return Uniform(location)
}

// SetUniformMatrix sets uniform matrix value.
func SetUniformMatrix(uniform Uniform, matrix mgl32.Mat4) {
	gl.UniformMatrix4fv(int32(uniform), 1, false, &matrix[0])
}

// SetUniformFloat sets uniform float value.
func SetUniformFloat(uniform Uniform, v float32) {
	gl.Uniform1f(int32(uniform), v)
}

// SetUniformVec3 sets uniform Vec3 value.
func SetUniformVec3(uniform Uniform, v mgl32.Vec3) {
	gl.Uniform3fv(int32(uniform), 1, &v[0])
}

// SetUniformVec3A sets uniform array of Vec3 values.
func SetUniformVec3A(uniform Uniform, v []mgl32.Vec3) {
	gl.Uniform3fv(int32(uniform), int32(len(v)), &v[0][0])
}

// SetUniformVec2 sets uniform Vec2 value.
func SetUniformVec2(uniform Uniform, v mgl32.Vec2) {
	gl.Uniform2fv(int32(uniform), 1, &v[0])
}

// SetUniformVec4 sets uniform Vec4 value.
func SetUniformVec4(uniform Uniform, v mgl32.Vec4) {
	gl.Uniform4fv(int32(uniform), 1, &v[0])
}
