package graphics

import (
	"fmt"
	"io/ioutil"
	
	"github.com/go-gl/mathgl/mgl32"
)

// Pipeline is higher-level abstraction unifying Program and Uniform values.
type Pipeline struct {
	program Program
	uniforms map[string] Uniform
}

// GetPipeline returns initialized Pipeline consisting of vertex shader
// and pixel shader stages.
func GetPipeline(vertexShaderFile, pixelShaderFile string) Pipeline {
	// Get vertex shader source code.
	vertexShaderData, err := ioutil.ReadFile(vertexShaderFile)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}

	// Compile and get vertex shader.
	vertexShader, err := GetShader(string(vertexShaderData), VertexShader)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}

	// Get pixel shader source code.
	pixelShaderData, err := ioutil.ReadFile(pixelShaderFile)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}

	// Compile and get pixel shader.
	pixelShader, err := GetShader(string(pixelShaderData), PixelShader)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}

	// Create a program from vertex and pixel shaders.
	program, err := GetProgram(vertexShader, pixelShader)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}

	// Cleanup.
	ReleaseShaders(vertexShader, pixelShader)

	return Pipeline{program, make(map[string]Uniform)}
}

// SetUniform sets value of uniform variable.
func (pipeline *Pipeline) SetUniform(uniform string, value interface{}) {
	// Get uniform variable's location in program from cache.
	uniformLocation, ok := pipeline.uniforms[uniform]
	if !ok {
		uniformLocation = GetUniform(pipeline.program, uniform)
		pipeline.uniforms[uniform] = uniformLocation
	}

	// Set the value based on value type.
	switch value.(type) {
	case float32:
		number := value.(float32)
		SetUniformFloat(uniformLocation, number)
	case mgl32.Vec2:
		vector := value.(mgl32.Vec2)
		SetUniformVec2(uniformLocation, vector)
	case mgl32.Vec3:
		vector := value.(mgl32.Vec3)
		SetUniformVec3(uniformLocation, vector)
	case []mgl32.Vec3:
		vectorSlice := value.([]mgl32.Vec3)
		SetUniformVec3A(uniformLocation, vectorSlice)
	case mgl32.Vec4:
		vector := value.(mgl32.Vec4)
		SetUniformVec4(uniformLocation, vector)
	case mgl32.Mat4:
		matrix := value.(mgl32.Mat4)
		SetUniformMatrix(uniformLocation, matrix)
	}
}

// Start binds the Pipeline's program for use.
func (pipeline *Pipeline) Start() {
	SetProgram(pipeline.program)
}
