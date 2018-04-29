package app

import (
	"fmt"
	
	"github.com/go-gl/mathgl/mgl32"
	"github.com/janivanecky/golib/graphics"
)

type Pipeline struct {
	program graphics.Program
	uniforms map[string]graphics.Uniform
}

func (pipeline *Pipeline) AddUniforms(uniforms ... string) {
	graphics.SetProgram(pipeline.program)
	for _, uniform := range uniforms {
		pipeline.uniforms[uniform] = graphics.GetUniform(pipeline.program, uniform)
	}
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

func InitPipeline(vertexShaderFile, pixelShaderFile string) Pipeline {
	program, err := getProgram(vertexShaderFile, pixelShaderFile)
	if err != nil {
		fmt.Println(err)
	}
	var pipeline Pipeline
	pipeline.program = program
	pipeline.uniforms = make(map[string]graphics.Uniform)
	return pipeline
}