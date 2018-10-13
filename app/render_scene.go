package app

import (
	"math"
	"math/rand"
	"unsafe"
	
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/janivanecky/golib/graphics"
)

// Pipelines used for 3D scene rendering.
var pipelinePBR Pipeline
var pipelinePBRInstanced Pipeline
var pipelineGeometry Pipeline
var pipelineGeometryInstanced Pipeline
var pipelineSSAO Pipeline
var pipelineBlur Pipeline
var pipelineBlit Pipeline
var pipelineShading Pipeline
var pipelineEffect Pipeline
var pipelineUI Pipeline

type SceneBuffers struct {
	// Framebuffers used for 3D scene rendering.
	bufferGeometry graphics.Framebuffer
	bufferScene graphics.Framebuffer
	bufferSceneMS graphics.Framebuffer
	bufferSSAO graphics.Framebuffer
	bufferBlur graphics.Framebuffer
	bufferShading graphics.Framebuffer
	bufferEffect graphics.Framebuffer
}

var fullBuffers SceneBuffers

// SSAO related data.
var ssaoNoiseTexture graphics.Texture
var ssaoKernels [16]mgl32.Vec3
const ssaoNoiseTextureSize = 4

// Scene rendering related data.
var sceneProjectionMatrix mgl32.Mat4
var sceneViewMatrix mgl32.Mat4
var sceneCameraPosition mgl32.Vec3

// Mesh quad spanning the whole screen, used for full-screen blitting.
var screenQuad graphics.Mesh

type RenderingSettings struct {
	DirectLight  float64
	AmbientLight float64

	Roughness    float64
	Reflectivity float64

	SSAORadius   float64
	SSAORange    float64
	SSAOBoundary float64

	MinWhite float64
}

var rendering *RenderingSettings

// Projection related info.
var screenWidth, screenHeight float64
var near, far float32 = 0.01, 500.0

func setUpBuffers(buffers *SceneBuffers, width, height int32) {
	buffers.bufferScene = graphics.GetFramebuffer(
		width, height, 1,
		map[string][]int32 {
			"direct": []int32{gl.RGBA32F, 0},
			"ambient": []int32{gl.RGBA32F, 1},
		}, true)
	buffers.bufferSceneMS = graphics.GetFramebuffer(
		width, height, 4,
		map[string][]int32 {
			"direct": []int32{gl.RGBA32F, 0},
			"ambient": []int32{gl.RGBA32F, 1},
		}, true)
	buffers.bufferGeometry = graphics.GetFramebuffer(
		width, height, 1,
		map[string][]int32 {
			"position": []int32{gl.RGBA32F, 0},
			"normal": []int32{gl.RGBA32F, 1},
		}, true)
	buffers.bufferSSAO = graphics.GetFramebuffer(
		width, height, 1,
		map[string][]int32 {"occlusion": []int32{gl.R32F, 0}}, false)
	buffers.bufferBlur = graphics.GetFramebuffer(
		width, height, 1,
		map[string][]int32 {"occlusion": []int32{gl.R32F, 0}}, false)
	buffers.bufferShading = graphics.GetFramebuffer(
		width, height, 1,
		map[string][]int32 {"color": []int32{gl.RGBA8, 0}}, false)
	buffers.bufferEffect = graphics.GetFramebuffer(
		width, height, 1,
		map[string][]int32 {"color": []int32{gl.RGBA8, 0}}, false)
	
}

var instanceModelBuffer graphics.InstanceBuffer
var instanceColorBuffer graphics.InstanceBuffer

func initSceneRendering(windowWidth, windowHeight float64, renderingSettings *RenderingSettings) {
	// Initialize 3D scene rendering pipelines.
	pipelinePBR = InitPipeline("shaders/geometry_vertex_shader.glsl", "shaders/pbr_pixel_shader.glsl")
	pipelinePBRInstanced = InitPipeline("shaders/geometry_vertex_shader_instanced.glsl", "shaders/pbr_pixel_shader.glsl")
	pipelineGeometry = InitPipeline("shaders/geometry_vertex_shader.glsl", "shaders/geometry_pixel_shader.glsl")
	pipelineGeometryInstanced = InitPipeline("shaders/geometry_vertex_shader_instanced.glsl", "shaders/geometry_pixel_shader.glsl")
	pipelineBlur = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/blur_pixel_shader.glsl")
	pipelineBlit = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/blit_pixel_shader.glsl")
	pipelineEffect = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/effect_pixel_shader.glsl")
	pipelineShading = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/shading_pixel_shader.glsl")
	pipelineSSAO = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/ssao_pixel_shader.glsl")
	pipelineUI = InitPipeline("shaders/geometry_vertex_shader.glsl", "shaders/flat_pixel_shader.glsl")

	// Set up framebuffers for rendering.
	backbufferWidth, backbufferHeight := windowWidth, windowHeight
	setUpBuffers(&fullBuffers, int32(backbufferWidth), int32(backbufferHeight))
	
	// Set up SSAO-related data.
	ssaoKernels = getSSAOKernels()
	noiseTexDataVec3 := getSSAONoiseTex()
	noiseTexDataFloat := *(*[ssaoNoiseTextureSize * ssaoNoiseTextureSize * 3]float32)(unsafe.Pointer(&noiseTexDataVec3[0]))
	ssaoNoiseTexture = graphics.GetTextureFloat32(ssaoNoiseTextureSize, ssaoNoiseTextureSize, 3, noiseTexDataFloat[:], false)
	
	// Set up matrices for scene rendering.
	sceneProjectionMatrix = mgl32.Perspective(mgl32.DegToRad(60.0), float32(windowWidth / windowHeight), near, far)
	sceneViewMatrix = mgl32.Ident4()
	sceneCameraPosition = mgl32.Vec3{0,0,10}
	
	// Set up blitting quad mesh.
	screenQuad = graphics.GetMesh(screenQuadVertices[:], quadIndices[:], []int{4,2})

	// Store window size.
	screenWidth, screenHeight = windowWidth, windowHeight

	rendering = renderingSettings

	instanceColorBuffer = graphics.GetInstanceBufferFloat32(4)
	instanceModelBuffer = graphics.GetInstanceBufferFloat32(16)
}

func renderScene(meshEntities []meshData, meshEntitiesInstanced []meshDataInstanced) graphics.Framebuffer {
	buffers := &fullBuffers
	bgColor := float32(0.9)

	// Set up 3D rendering settings - no blending and depth test.
	graphics.DisableBlending()
	graphics.EnableDepthTest()
	
	// First we render the direct and indirect lighting multi-sampled.
	graphics.SetFramebuffer(buffers.bufferSceneMS)
	graphics.SetFramebufferViewport(buffers.bufferSceneMS)
	graphics.ClearScreen(bgColor, bgColor, bgColor, 1.0)

	pipelinePBR.Start()
	pipelinePBR.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelinePBR.SetUniform("view_matrix", sceneViewMatrix)
	pipelinePBR.SetUniform("roughness", float32(rendering.Roughness))
	pipelinePBR.SetUniform("reflectivity", float32(rendering.Reflectivity))
	pipelinePBR.SetUniform("direct_light_power", float32(rendering.DirectLight))
	pipelinePBR.SetUniform("ambient_light_power", float32(rendering.AmbientLight))
	
	for _, meshEntity := range meshEntities {
		drawMesh(pipelinePBR, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}

	pipelinePBRInstanced.Start()
	pipelinePBRInstanced.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelinePBRInstanced.SetUniform("view_matrix", sceneViewMatrix)
	pipelinePBRInstanced.SetUniform("roughness", float32(rendering.Roughness))
	pipelinePBRInstanced.SetUniform("reflectivity", float32(rendering.Reflectivity))
	pipelinePBRInstanced.SetUniform("direct_light_power", float32(rendering.DirectLight))
	pipelinePBRInstanced.SetUniform("ambient_light_power", float32(rendering.AmbientLight))

	for _, meshEntity := range meshEntitiesInstanced {
		drawMeshInstanced(pipelinePBRInstanced, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color, meshEntity.count)
	}

	// Since color rendering was multisampled, we need to resolve into non-MS framebuffer for it to be used
	// later as a texture.
	graphics.BlitFramebufferAttachment(buffers.bufferSceneMS, buffers.bufferScene, "direct", "direct")
	graphics.BlitFramebufferAttachment(buffers.bufferSceneMS, buffers.bufferScene, "ambient", "ambient")

	// Next we render into geometry buffer (position + normal)
	graphics.SetFramebuffer(buffers.bufferGeometry)
	graphics.SetFramebufferViewport(buffers.bufferGeometry)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)

	pipelineGeometry.Start()
	pipelineGeometry.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelineGeometry.SetUniform("view_matrix", sceneViewMatrix)
	
	for _, meshEntity := range meshEntities {
		drawMesh(pipelineGeometry, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}

	pipelineGeometryInstanced.Start()
	pipelineGeometryInstanced.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelineGeometryInstanced.SetUniform("view_matrix", sceneViewMatrix)

	//for _, meshEntity := range meshEntitiesInstanced {
	//}

	// SSAO computation. 
	graphics.SetFramebuffer(buffers.bufferSSAO)
	graphics.SetFramebufferViewport(buffers.bufferSSAO)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(buffers.bufferGeometry, "position", 0)
	graphics.SetFramebufferTexture(buffers.bufferGeometry, "normal", 1)
	graphics.SetTexture(ssaoNoiseTexture, 2)

	width, height := graphics.GetFramebufferSize(buffers.bufferSSAO)
	pipelineSSAO.Start()
	pipelineSSAO.SetUniform("screen_size", mgl32.Vec2{float32(width), float32(height)})
	pipelineSSAO.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelineSSAO.SetUniform("kernels", ssaoKernels[:])
	pipelineSSAO.SetUniform("ssao_radius", float32(rendering.SSAORadius))
	pipelineSSAO.SetUniform("ssao_range", float32(rendering.SSAORange))
	pipelineSSAO.SetUniform("ssao_range_boundary", float32(rendering.SSAOBoundary))
	
	graphics.DrawMesh(screenQuad)

	// Blur SSAO computed occlusion.
	graphics.SetFramebuffer(buffers.bufferBlur)
	graphics.SetFramebufferViewport(buffers.bufferBlur)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(buffers.bufferSSAO, "occlusion", 0)
	
	width, height = graphics.GetFramebufferSize(buffers.bufferBlur)
	pipelineBlur.Start()
	pipelineBlur.SetUniform("screen_size", mgl32.Vec2{float32(width), float32(height)})

	graphics.DrawMesh(screenQuad)

	// Deffered shading pass.
	graphics.SetFramebuffer(buffers.bufferShading)
	graphics.SetFramebufferViewport(buffers.bufferShading)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(buffers.bufferScene, "direct", 0)
	graphics.SetFramebufferTexture(buffers.bufferScene, "ambient", 1)
	graphics.SetFramebufferTexture(buffers.bufferBlur, "occlusion", 2)
	
	pipelineShading.Start()
	pipelineShading.SetUniform("minWhite", float32(rendering.MinWhite))
	
	graphics.DrawMesh(screenQuad)
	
	// Post processing effect pass.
	graphics.SetFramebuffer(buffers.bufferEffect)
	graphics.SetFramebufferViewport(buffers.bufferEffect)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(buffers.bufferShading, "color", 0)
	
	pipelineEffect.Start()
	
	graphics.DrawMesh(screenQuad)

	return buffers.bufferEffect
}

func drawMesh(pipeline Pipeline, mesh graphics.Mesh, modelMatrix mgl32.Mat4, color mgl32.Vec4) {
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	graphics.DrawMesh(mesh)
}

func drawMeshInstanced(pipeline Pipeline, mesh graphics.Mesh, modelMatrix []mgl32.Mat4, color []mgl32.Vec4, count int32) {
	graphics.UpdateInstanceBuffer(instanceModelBuffer, int(count), modelMatrix)
	graphics.UpdateInstanceBuffer(instanceColorBuffer, int(count), color)
	graphics.DrawMeshInstanced(mesh, count, []graphics.InstanceBuffer{instanceModelBuffer, instanceColorBuffer}, []uint32{2, 6})
}

func getSSAOKernels() [16]mgl32.Vec3 {
	var kernels [16]mgl32.Vec3
	for i := range kernels {
		// Generate random vector on the surface of unit hemisphere.
		kernels[i][0] = rand.Float32() * 2.0 - 1.0
		kernels[i][1] = rand.Float32() * 1.0
		kernels[i][2] = rand.Float32() * 2.0 - 1.0
		kernels[i] = kernels[i].Normalize()

		// Scale vector so it fills hemipshere's volume.
		scale := float64(i) / 16.0
		scale = 0.1 + 0.9 * (scale * scale)
		kernels[i] = kernels[i].Mul(float32(scale))
	}
	return kernels
}

func getSSAONoiseTex() [ssaoNoiseTextureSize * ssaoNoiseTextureSize]mgl32.Vec3 {
	var tex [ssaoNoiseTextureSize * ssaoNoiseTextureSize]mgl32.Vec3
	for i := range tex {
		azimuth := rand.Float64() * math.Pi * 2.0
		
		tex[i][0] = float32(math.Sin(azimuth))
		tex[i][1] = 0
		tex[i][2] = float32(math.Cos(azimuth))
	}
	return tex
}

func GetSceneBuffer() ([]byte, int32, int32) {
	buffer := graphics.GetFramebufferPixels(fullBuffers.bufferEffect, "color")
	width, height :=  graphics.GetFramebufferSize(fullBuffers.bufferEffect)
	return buffer, width, height
}

func GetSceneBufferTexture() graphics.Texture {
	return graphics.GetFramebufferTexture(fullBuffers.bufferEffect, "color")
}

func GetWorldRay(screenX, screenY float64) (mgl32.Vec4, mgl32.Vec4) {
	projViewMatrix := sceneProjectionMatrix.Mul4(sceneViewMatrix)
	invProjViewMatrix := projViewMatrix.Inv()

	relX := float32(screenX / screenWidth * 2.0 - 1.0)
	relY := -float32(screenY / screenHeight * 2.0 - 1.0)

	vFar := mgl32.Vec4{relX, relY, 1.0, 1.0}
	vFar = vFar.Mul(far)
    vFar = invProjViewMatrix.Mul4x1(vFar)

	vNear := mgl32.Vec4{relX, relY, -1.0, 1.0}
	vNear = vNear.Mul(near)
	vNear = invProjViewMatrix.Mul4x1(vNear)
	
	vDiff := vFar.Sub(vNear)
	vDiff = vDiff.Normalize()
	return vNear, vDiff
}