package app

import (
	"unsafe"
	
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/janivanecky/golib/graphics"
)

var scenePipeline Pipeline
var geometryPipeline Pipeline
var effectPipeline Pipeline
var shadingPipeline Pipeline
var blurPipeline Pipeline
var ssaoPipeline Pipeline

type ssaoSettings struct {
	noiseTexture graphics.Texture
	kernels [16]mgl32.Vec3
}
var ssaoData ssaoSettings

type sceneSettings struct {
	projectionMatrix mgl32.Mat4
	viewMatrix mgl32.Mat4
	lightPosition mgl32.Vec3
	cameraPosition mgl32.Vec3
}
var sceneData sceneSettings

var geometryBuffer graphics.Framebuffer
var sceneBuffer graphics.Framebuffer
var sceneBufferMS graphics.Framebuffer
var ssaoBuffer graphics.Framebuffer
var blurBuffer graphics.Framebuffer
var effectBuffer graphics.Framebuffer

var screenQuad graphics.Mesh

func initSceneRendering(windowWidth, windowHeight float64) {
	// Initialize scene rednering data
	sceneData = sceneSettings{
		mgl32.Perspective(mgl32.DegToRad(60.0), float32(windowWidth / windowHeight), 0.01, 500.0),
		mgl32.Ident4(),
		mgl32.Vec3{10, 20, 30},
		mgl32.Vec3{0,0,10},
	}
	
	scenePipeline = InitPipeline("shaders/geometry_vertex_shader.glsl", "shaders/pbr_pixel_shader.glsl")
	geometryPipeline = InitPipeline("shaders/geometry_vertex_shader.glsl", "shaders/geometry_pixel_shader.glsl")
	blurPipeline = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/blur_pixel_shader.glsl")
	effectPipeline = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/effect_pixel_shader.glsl")
	shadingPipeline = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/shading_pixel_shader.glsl")
	ssaoPipeline = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/ssao_pixel_shader.glsl")

	backbufferWidth, backbufferHeight := windowWidth, windowHeight

	// Set up necessary framebuffers
	sceneBuffer = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {"diffuse": []int32{gl.RGBA32F, 0}, "ambient": []int32{gl.RGBA32F, 1}}, true)
	sceneBufferMS = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 4,
		map[string][]int32 {"diffuse": []int32{gl.RGBA32F, 0}, "ambient": []int32{gl.RGBA32F, 1}}, true)

	geometryBuffer = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {
			"position": []int32{gl.RGBA32F, 0},
			"normal": []int32{gl.RGBA32F, 1},
		}, true)
		
	ssaoBuffer = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {"occlusion": []int32{gl.R32F, 0}}, false)

		
	blurBuffer = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {"occlusion": []int32{gl.R32F, 0}}, false)

	effectBuffer = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {"color": []int32{gl.RGBA8, 0}}, false)

	// SSAO data
	ssaoData.kernels = getKernels()
	noiseTexDataVec := getNoiseTex()
	noiseTexData := *(*[16 * 3]float32)(unsafe.Pointer(&noiseTexDataVec[0]))
	ssaoData.noiseTexture = graphics.GetTextureFloat32(4, 4, 3, noiseTexData[:], false)
	
	// Quad for blitting
	screenQuad = graphics.GetMesh(screenQuadVertices[:], quadIndices[:], []int{4,2})
}

func renderScene(targetBuffer graphics.Framebuffer) {
	// We got the cleaning done bitchez.
	bgColor := float32(0.9)
	graphics.SetFramebuffer(sceneBufferMS)
	graphics.SetFramebufferViewport(sceneBufferMS)
	graphics.ClearScreen(bgColor, bgColor, bgColor, 1.0)

	// Set up 3D scene rendering settings
	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	
	// Set up 3D scene rendering pipeline
	scenePipeline.Start()
	scenePipeline.SetUniform("projection_matrix", sceneData.projectionMatrix)
	scenePipeline.SetUniform("view_matrix", sceneData.viewMatrix)
	scenePipeline.SetUniform("roughness", float32(Roughness))
	scenePipeline.SetUniform("reflectivity", float32(Reflectivity))
	scenePipeline.SetUniform("direct_light_power", float32(DirectLight))
	scenePipeline.SetUniform("ambient_light_power", float32(AmbientLight))
	
	// Render meshes
	for _, meshEntity := range meshEntities {
		drawMesh(scenePipeline, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}

	// Geometry pass
	graphics.SetFramebuffer(geometryBuffer)
	graphics.SetFramebufferViewport(geometryBuffer)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	geometryPipeline.Start()
	geometryPipeline.SetUniform("projection_matrix", sceneData.projectionMatrix)
	geometryPipeline.SetUniform("view_matrix", sceneData.viewMatrix)
	
	// Render meshes
	for _, meshEntity := range meshEntities {
		drawMesh(geometryPipeline, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}

	// Set up 2D UI rendering settings
	gl.Disable(gl.DEPTH_TEST)
	
	graphics.BlitFramebufferAttachment(sceneBufferMS, sceneBuffer, "diffuse", "diffuse")
	graphics.BlitFramebufferAttachment(sceneBufferMS, sceneBuffer, "ambient", "ambient")
	
	graphics.SetFramebuffer(ssaoBuffer)
	graphics.SetFramebufferViewport(ssaoBuffer)
	width, height := graphics.GetFramebufferSize(ssaoBuffer)
	graphics.ClearScreen(0.0, 0.0, 1.0, 1.0)
	ssaoPipeline.Start()
	ssaoPipeline.SetUniform("projection_matrix", sceneData.projectionMatrix)
	ssaoPipeline.SetUniform("kernels", ssaoData.kernels[:])
	ssaoPipeline.SetUniform("ssao_radius", float32(SSAORadius))
	ssaoPipeline.SetUniform("ssao_range", float32(SSAORange))
	ssaoPipeline.SetUniform("ssao_range_boundary", float32(SSAOBoundary))
	ssaoPipeline.SetUniform("screen_size", mgl32.Vec2{float32(width), float32(height)})
	graphics.SetFramebufferTexture(geometryBuffer, "position", 0)
	graphics.SetFramebufferTexture(geometryBuffer, "normal", 1)
	graphics.SetTexture(ssaoData.noiseTexture, 2)
	graphics.DrawMesh(screenQuad)
	
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	graphics.SetFramebuffer(blurBuffer)
	graphics.SetFramebufferViewport(blurBuffer)
	blurPipeline.Start()
	blurPipeline.SetUniform("screen_size", mgl32.Vec2{float32(width), float32(height)})
	graphics.SetFramebufferTexture(ssaoBuffer, "occlusion", 0)
	width, height = graphics.GetFramebufferSize(blurBuffer)
	graphics.DrawMesh(screenQuad)

	graphics.SetFramebuffer(effectBuffer)
	graphics.SetFramebufferViewport(effectBuffer)
	shadingPipeline.Start()
	shadingPipeline.SetUniform("minWhite", float32(MinWhite))
	graphics.SetFramebufferTexture(sceneBuffer, "diffuse", 0)
	graphics.SetFramebufferTexture(sceneBuffer, "ambient", 1)
	graphics.SetFramebufferTexture(blurBuffer, "occlusion", 2)
	graphics.DrawMesh(screenQuad)

	graphics.SetFramebuffer(targetBuffer)
	graphics.SetFramebufferViewport(targetBuffer)
	effectPipeline.Start()
	graphics.SetFramebufferTexture(effectBuffer, "color", 0)
	graphics.DrawMesh(screenQuad)
}