package app

import (
	"io/ioutil"
	"fmt"
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/janivanecky/golib/font"
	"github.com/janivanecky/golib/graphics"
)


func getProgram(vertexShaderPath string, pixelShaderPath string) (graphics.Program, error) {
	vertexShaderData, err := ioutil.ReadFile(vertexShaderPath)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}
	
	vertexShader, err := graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}

	pixelShaderData, err := ioutil.ReadFile(pixelShaderPath)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}

	pixelShader, err := graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}

	program, err := graphics.GetProgram(vertexShader, pixelShader)
	if err != nil {
		fmt.Println(err)
		return graphics.Program(0), err
	}
	graphics.ReleaseShaders(vertexShader, pixelShader)

	return program, nil
}


func drawMesh(pipeline Pipeline, mesh graphics.Mesh, modelMatrix mgl32.Mat4, color mgl32.Vec4) {
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	graphics.DrawMesh(mesh)
}


func drawRect(pipeline Pipeline, pos mgl32.Vec2, size mgl32.Vec2, color mgl32.Vec4) {
	pipeline.SetUniform("color", color)
	
	x, y := pos[0], float32(screenHeight) - pos[1]
	modelMatrix := mgl32.Translate3D(x, y, 0).Mul4(
		mgl32.Scale3D(size[0], size[1], 1.0).Mul4(
			mgl32.Translate3D(0.5, -0.5, 0.0),
		),
	)
	pipeline.SetUniform("model_matrix", modelMatrix)
	graphics.DrawMesh(uiData.quad)
}

func drawText(pipeline Pipeline, text string, font *font.Font, position mgl32.Vec2, color mgl32.Vec4, origin mgl32.Vec2) {
	pipeline.SetUniform("color", color)
	
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
        sourceRect := mgl32.Vec4{relX,relY,relWidth,relHeight}
		pipeline.SetUniform("source_rect", sourceRect)

		currentX := x + glyphA.XOffset
		currentY := y + glyphA.YOffset
		modelMatrix := mgl32.Translate3D(float32(currentX), float32(screenHeight - currentY), 0).Mul4(
                mgl32.Scale3D(float32(glyphA.Width), float32(glyphA.Height), 1.0).Mul4(
					mgl32.Translate3D(0.5, -0.5, 0.0),
			),
		)
		pipeline.SetUniform("model_matrix", modelMatrix)
		
		graphics.DrawMesh(uiData.quad)
		
		x += float64(glyphA.Advance)
	}
}

func getKernels() [16]mgl32.Vec3 {
	var kernels [16]mgl32.Vec3
	for i := range kernels {
		kernels[i][0] = rand.Float32() * 2.0 - 1.0
		kernels[i][1] = rand.Float32() * 1.0
		kernels[i][2] = rand.Float32() * 2.0 - 1.0
		kernels[i] = kernels[i].Normalize()

		scale := float64(i) / 16.0
		scale = 0.1 + 0.9 * (scale * scale)
		kernels[i] = kernels[i].Mul(float32(scale))
	}
	return kernels
}

func getNoiseTex() [16]mgl32.Vec3 {
	var tex [16]mgl32.Vec3
	for i := range tex {
		azimuth := rand.Float64() * math.Pi * 2.0
		tex[i][0] = float32(math.Sin(azimuth))
		tex[i][1] = 0
		tex[i][2] = float32(math.Cos(azimuth))
	}
	return tex
}