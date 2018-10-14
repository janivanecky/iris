package graphics

import (
	"fmt"
	"strings"
	
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)


type ShaderType int
type Uniform int32

const (
	VERTEX_SHADER ShaderType = iota
	PIXEL_SHADER ShaderType = iota
	GEOMETRY_SHADER ShaderType = iota
)

type Shader uint32
type Program uint32

var shaderTypeToGL = map[ShaderType]uint32 {
	VERTEX_SHADER: gl.VERTEX_SHADER,
	PIXEL_SHADER: gl.FRAGMENT_SHADER,
	GEOMETRY_SHADER: gl.GEOMETRY_SHADER,
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

func SetUniformMatrix(uniform Uniform, matrix mgl32.Mat4) {
	gl.UniformMatrix4fv(int32(uniform), 1, false, &matrix[0])
}

func SetUniformFloat(uniform Uniform, v float32) {
	gl.Uniform1f(int32(uniform), v)
}

func SetUniformVec3(uniform Uniform, v mgl32.Vec3) {
	gl.Uniform3fv(int32(uniform), 1, &v[0])
}

func SetUniformVec3A(uniform Uniform, v []mgl32.Vec3) {
	gl.Uniform3fv(int32(uniform), int32(len(v)), &v[0][0])
}

func SetUniformVec2(uniform Uniform, v mgl32.Vec2) {
	gl.Uniform2fv(int32(uniform), 1, &v[0])
}

func SetUniformVec4(uniform Uniform, v mgl32.Vec4) {
	gl.Uniform4fv(int32(uniform), 1, &v[0])
}
