package app

import (
	"fmt"
	"io/ioutil"
	
	"github.com/go-gl/mathgl/mgl32"
	"../lib/graphics"
)

type Pipeline struct {
	program graphics.Program
	uniforms map[string]graphics.Uniform
}

func (pipeline *Pipeline) SetUniform(uniform string, value interface{}) {
	uniformLocation, ok := pipeline.uniforms[uniform]
	if !ok {
		uniformLocation = graphics.GetUniform(pipeline.program, uniform)
		pipeline.uniforms[uniform] = uniformLocation
	}
	switch value.(type) {
	case float32:
		number := value.(float32)
		graphics.SetUniformFloat(uniformLocation, number)
	case mgl32.Vec2:
		vector := value.(mgl32.Vec2)
		graphics.SetUniformVec2(uniformLocation, vector)
	case mgl32.Vec3:
		vector := value.(mgl32.Vec3)
		graphics.SetUniformVec3(uniformLocation, vector)
	case []mgl32.Vec3:
		vectorSlice := value.([]mgl32.Vec3)
		graphics.SetUniformVec3A(uniformLocation, vectorSlice)
	case mgl32.Vec4:
		vector := value.(mgl32.Vec4)
		graphics.SetUniformVec4(uniformLocation, vector)
	case mgl32.Mat4:
		matrix := value.(mgl32.Mat4)
		graphics.SetUniformMatrix(uniformLocation, matrix)
	}
}

func (pipeline *Pipeline) Start() {
	graphics.SetProgram(pipeline.program)
}

func GetPipeline(vertexShaderFile, pixelShaderFile string) Pipeline {
	vertexShaderData, err := ioutil.ReadFile(vertexShaderFile)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}
	
	vertexShader, err := graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}

	pixelShaderData, err := ioutil.ReadFile(pixelShaderFile)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}

	pixelShader, err := graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}

	program, err := graphics.GetProgram(vertexShader, pixelShader)
	if err != nil {
		fmt.Println(err)
		return Pipeline{}
	}
	graphics.ReleaseShaders(vertexShader, pixelShader)
	return Pipeline{program, make(map[string]graphics.Uniform)}
}
