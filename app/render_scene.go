package app

import (
	"math"
	"math/rand"
	"unsafe"
	
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v4.1-core/gl"

	"../lib/graphics"
)

// Constants.
const ssaoNoiseTextureSize = 4
const noiseTexTexelCount = ssaoNoiseTextureSize * ssaoNoiseTextureSize * 3
const bgColor = float32(0.9)

// Pipelines used for 3D scene rendering.
var pipelinePBR graphics.Pipeline
var pipelinePBRInstanced graphics.Pipeline
var pipelineGeometry graphics.Pipeline
var pipelineGeometryInstanced graphics.Pipeline
var pipelineSSAO graphics.Pipeline
var pipelineBlur graphics.Pipeline
var pipelineShading graphics.Pipeline
var pipelineEffect graphics.Pipeline
var pipelineSceneUI graphics.Pipeline

// Framebuffers used for 3D scene rendering.
var bufferGeometry graphics.Framebuffer
var bufferLight graphics.Framebuffer
var bufferLightMS graphics.Framebuffer
var bufferSSAO graphics.Framebuffer
var bufferBlur graphics.Framebuffer
var bufferShading graphics.Framebuffer
var bufferEffect graphics.Framebuffer
var bufferSceneUI graphics.Framebuffer

// SSAO related data.
var ssaoNoiseTexture graphics.Texture
var ssaoKernels [16]mgl32.Vec3

// Mesh quad spanning the whole screen, used for full-screen blitting.
var screenQuad graphics.Mesh

var screenQuadVertices = [...]float32{
	-1.0, -1.0, 0.0, 1.0,
	0.0, 0.0,
	-1.0, 1.0, 0.0, 1.0,
	0.0, 1.0,
	1.0, 1.0, 0.0, 1.0,
	1.0, 1.0,
	1.0, -1.0, 0.0, 1.0,
	1.0, 0.0,
}

var screenQuadIndices = [...]uint32{
	0, 1, 2,
	0, 2, 3,
}

// Instance buffers for drawing cells.
var instanceModelBuffer graphics.InstanceBuffer
var instanceColorBuffer graphics.InstanceBuffer

// Structs for storing draw data.
type meshData struct {
	mesh 	    graphics.Mesh
	modelMatrix mgl32.Mat4
	color 	    mgl32.Vec4
}

type meshDataInstanced struct {
	mesh 	    graphics.Mesh
	modelMatrix []mgl32.Mat4
	color 	    []mgl32.Vec4
	count 		int32
}

// Slices for storing draw data.
var meshEntities 		  []meshData
var meshEntitiesSceneUI   []meshData
var meshEntitiesInstanced []meshDataInstanced

// InitSceneRendering initializes necessary objects for 3D scene rendering.
func InitSceneRendering(windowWidth, windowHeight float64) {
	// Initialize 3D scene rendering pipelines.
	pipelinePBR = graphics.GetPipeline(
		"shaders/geometry_vertex_shader.glsl",
		"shaders/pbr_pixel_shader.glsl")
	pipelinePBRInstanced = graphics.GetPipeline(
		"shaders/geometry_vertex_shader_instanced.glsl",
		"shaders/pbr_pixel_shader.glsl")
	pipelineGeometry = graphics.GetPipeline(
		"shaders/geometry_vertex_shader.glsl",
		"shaders/geometry_pixel_shader.glsl")
	pipelineGeometryInstanced = graphics.GetPipeline(
		"shaders/geometry_vertex_shader_instanced.glsl",
		"shaders/geometry_pixel_shader.glsl")
	pipelineBlur = graphics.GetPipeline(
		"shaders/blit_vertex_shader.glsl",
		"shaders/blur_pixel_shader.glsl")
	pipelineEffect = graphics.GetPipeline(
		"shaders/blit_vertex_shader.glsl",
		"shaders/effect_pixel_shader.glsl")
	pipelineShading = graphics.GetPipeline(
		"shaders/blit_vertex_shader.glsl",
		"shaders/shading_pixel_shader.glsl")
	pipelineSSAO = graphics.GetPipeline(
		"shaders/blit_vertex_shader.glsl",
		"shaders/ssao_pixel_shader.glsl")
	pipelineSceneUI = graphics.GetPipeline(
		"shaders/geometry_vertex_shader.glsl",
		"shaders/flat_pixel_shader.glsl")

	// Set up framebuffers for rendering.
	backbufferWidth, backbufferHeight := int32(windowWidth), int32(windowHeight)
	bufferLight = graphics.GetFramebuffer(
		backbufferWidth, backbufferHeight, 1,
		map[string]int32 {
			"direct": gl.RGBA32F,
			"ambient": gl.RGBA32F,
		}, true)
	bufferLightMS = graphics.GetFramebuffer(
		backbufferWidth, backbufferHeight, 4,
		map[string]int32 {
			"direct": gl.RGBA32F,
			"ambient": gl.RGBA32F,
		}, true)
	bufferGeometry = graphics.GetFramebuffer(
		backbufferWidth, backbufferHeight, 1,
		map[string]int32 {
			"position": gl.RGBA32F,
			"normal": gl.RGBA32F,
		}, true)
	bufferSSAO = graphics.GetFramebuffer(
		backbufferWidth, backbufferHeight, 1,
		map[string]int32 {"occlusion": gl.R32F}, false)
	bufferBlur = graphics.GetFramebuffer(
		backbufferWidth, backbufferHeight, 1,
		map[string]int32 {"occlusion": gl.R32F}, false)
	bufferShading = graphics.GetFramebuffer(
		backbufferWidth, backbufferHeight, 1,
		map[string]int32 {"color": gl.RGBA8}, false)
	bufferEffect = graphics.GetFramebuffer(
		backbufferWidth, backbufferHeight, 1,
		map[string]int32 {"color": gl.RGBA8}, false)
	bufferSceneUI = graphics.GetFramebuffer(
		backbufferWidth, backbufferHeight, 1,
		map[string]int32 {"color": gl.RGBA8}, false)
	
	// Set up SSAO-related data.
	ssaoKernels = getSSAOKernels()
	noiseTexDataVec3 := getSSAONoiseTex()
	// Cast from Vec3 to float32.
	noiseTexDataFloat := *(*[noiseTexTexelCount]float32)(unsafe.Pointer(&noiseTexDataVec3[0]))
	ssaoNoiseTexture = graphics.GetTextureFloat32(ssaoNoiseTextureSize, ssaoNoiseTextureSize, 3,
												  noiseTexDataFloat[:], false)
	
	// Set up blitting quad mesh.
	screenQuad = graphics.GetMesh(screenQuadVertices[:], screenQuadIndices[:], []int{4,2})

	// Set up buffers for instanced rendering.
	instanceColorBuffer = graphics.GetInstanceBuffer(4)
	instanceModelBuffer = graphics.GetInstanceBuffer(16)

	// Initialize slices which will store data for draw calls.
	meshEntities 		  = make([]meshData, 0, 100)
	meshEntitiesInstanced = make([]meshDataInstanced, 0, 10)
	meshEntitiesSceneUI   = make([]meshData, 0, 100)
}

// RenderScene sends commands to draw meshes gathered from DrawMeshXXX calls.
func RenderScene(viewMatrix, projectionMatrix mgl32.Mat4, settings *RenderingSettings) graphics.Framebuffer {
	// Set up 3D rendering settings - no blending and depth test.
	graphics.DisableBlending()
	graphics.EnableDepthTest()
	
	// First we render the direct and indirect lighting multi-sampled.
	graphics.SetFramebuffer(bufferLightMS)
	graphics.SetFramebufferViewport(bufferLightMS)
	graphics.ClearScreen(bgColor, bgColor, bgColor, 1.0)

	// Normal, per object rendering pass.
	pipelinePBR.Start()
	pipelinePBR.SetUniform("projection_matrix", projectionMatrix)
	pipelinePBR.SetUniform("view_matrix", viewMatrix)
	pipelinePBR.SetUniform("roughness", float32(settings.Roughness))
	pipelinePBR.SetUniform("reflectivity", float32(settings.Reflectivity))
	pipelinePBR.SetUniform("direct_light_power", float32(settings.DirectLight))
	pipelinePBR.SetUniform("ambient_light_power", float32(settings.AmbientLight))
	
	for _, meshEntity := range meshEntities {
		drawMesh(pipelinePBR, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}

	// Instanced rendering pass.
	pipelinePBRInstanced.Start()
	pipelinePBRInstanced.SetUniform("projection_matrix", projectionMatrix)
	pipelinePBRInstanced.SetUniform("view_matrix", viewMatrix)
	pipelinePBRInstanced.SetUniform("roughness", float32(settings.Roughness))
	pipelinePBRInstanced.SetUniform("reflectivity", float32(settings.Reflectivity))
	pipelinePBRInstanced.SetUniform("direct_light_power", float32(settings.DirectLight))
	pipelinePBRInstanced.SetUniform("ambient_light_power", float32(settings.AmbientLight))

	for _, meshEntity := range meshEntitiesInstanced {
		drawMeshInstanced(meshEntity.mesh, meshEntity.modelMatrix,
						  meshEntity.color, meshEntity.count)
	}

	// Since color rendering was multisampled, we need to resolve into non-MS framebuffer
	// for it to be used later as a texture.
	graphics.BlitFramebufferAttachment(bufferLightMS, bufferLight, "direct", "direct")
	graphics.BlitFramebufferAttachment(bufferLightMS, bufferLight, "ambient", "ambient")

	// Next we render into geometry buffer (position + normal)
	graphics.SetFramebuffer(bufferGeometry)
	graphics.SetFramebufferViewport(bufferGeometry)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)

	// Normal, per-object pass.
	pipelineGeometry.Start()
	pipelineGeometry.SetUniform("projection_matrix", projectionMatrix)
	pipelineGeometry.SetUniform("view_matrix", viewMatrix)
	
	for _, meshEntity := range meshEntities {
		drawMesh(pipelineGeometry, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}

	// Instanced rendering pass.
	pipelineGeometryInstanced.Start()
	pipelineGeometryInstanced.SetUniform("projection_matrix", projectionMatrix)
	pipelineGeometryInstanced.SetUniform("view_matrix", viewMatrix)

	for _, meshEntity := range meshEntitiesInstanced {
		drawMeshInstanced(meshEntity.mesh, meshEntity.modelMatrix,
						  meshEntity.color, meshEntity.count)
	}

	// SSAO computation. 
	graphics.SetFramebuffer(bufferSSAO)
	graphics.SetFramebufferViewport(bufferSSAO)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(bufferGeometry, "position", 0)
	graphics.SetFramebufferTexture(bufferGeometry, "normal", 1)
	graphics.SetTexture(ssaoNoiseTexture, 2)

	width, height := graphics.GetFramebufferSize(bufferSSAO)
	pipelineSSAO.Start()
	pipelineSSAO.SetUniform("screen_size", mgl32.Vec2{float32(width), float32(height)})
	pipelineSSAO.SetUniform("projection_matrix", projectionMatrix)
	pipelineSSAO.SetUniform("kernels", ssaoKernels[:])
	pipelineSSAO.SetUniform("ssao_radius", float32(settings.SSAORadius))
	pipelineSSAO.SetUniform("ssao_range", float32(settings.SSAORange))
	pipelineSSAO.SetUniform("ssao_range_boundary", float32(settings.SSAOBoundary))
	
	graphics.DrawMesh(screenQuad)

	// Blur SSAO computed occlusion.
	graphics.SetFramebuffer(bufferBlur)
	graphics.SetFramebufferViewport(bufferBlur)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(bufferSSAO, "occlusion", 0)
	
	width, height = graphics.GetFramebufferSize(bufferBlur)
	pipelineBlur.Start()
	pipelineBlur.SetUniform("screen_size", mgl32.Vec2{float32(width), float32(height)})

	graphics.DrawMesh(screenQuad)

	// Deffered shading pass.
	graphics.SetFramebuffer(bufferShading)
	graphics.SetFramebufferViewport(bufferShading)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(bufferLight, "direct", 0)
	graphics.SetFramebufferTexture(bufferLight, "ambient", 1)
	graphics.SetFramebufferTexture(bufferBlur, "occlusion", 2)
	
	pipelineShading.Start()
	pipelineShading.SetUniform("minWhite", float32(settings.MinWhite))
	
	graphics.DrawMesh(screenQuad)
	
	// Post processing effect pass.
	graphics.SetFramebuffer(bufferEffect)
	graphics.SetFramebufferViewport(bufferEffect)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(bufferShading, "color", 0)
	
	pipelineEffect.Start()
	
	graphics.DrawMesh(screenQuad)

	// Blit scene into scene UI texture.
	graphics.BlitFramebufferAttachment(bufferEffect, bufferSceneUI, "color", "color")

	// Draw in-scene UI on top.
	graphics.EnableBlending()
	graphics.DisableDepthTest()

	pipelineSceneUI.Start()
	pipelineSceneUI.SetUniform("projection_matrix", projectionMatrix)
	pipelineSceneUI.SetUniform("view_matrix", viewMatrix)

	for _, meshEntity := range meshEntitiesSceneUI {
		drawMesh(pipelineSceneUI, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}

	// Revert settings.
	graphics.DisableBlending()
	graphics.EnableDepthTest()

	return bufferSceneUI
}

// ResetScene clears lists of meshes to draw.
// Should be called right after RenderScene().
func ResetScene() {
	meshEntities 		  = meshEntities[:0]
	meshEntitiesInstanced = meshEntitiesInstanced[:0]
	meshEntitiesSceneUI   = meshEntitiesSceneUI[:0]
}

// DrawMesh sets mesh to be drawn in scene next frame.
func DrawMesh(mesh graphics.Mesh, modelMatrix mgl32.Mat4, color mgl32.Vec4) {
	meshEntities = append(meshEntities, meshData{mesh, modelMatrix, color})
}

// DrawMeshInstanced sets mesh to be drawn multiple times in scene next frame.
func DrawMeshInstanced(mesh graphics.Mesh, modelMatrix []mgl32.Mat4, color []mgl32.Vec4, count int) {
	meshEntitiesInstanced = append(meshEntitiesInstanced, meshDataInstanced{mesh, modelMatrix, color, int32(count)})
}

// DrawMeshSceneUI sets mesh to be drawn as in-scene UI next frame.
func DrawMeshSceneUI(mesh graphics.Mesh, modelMatrix mgl32.Mat4, color mgl32.Vec4) {
	meshEntitiesSceneUI = append(meshEntitiesSceneUI, meshData{mesh, modelMatrix, color})
}

// TODO: pipeline - should be passed as param? There's already strong assumption on uniforms it has.
func drawMesh(pipeline graphics.Pipeline, mesh graphics.Mesh, modelMatrix mgl32.Mat4, color mgl32.Vec4) {
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	graphics.DrawMesh(mesh)
}

// TODO: instance buffers should be passed as args?
func drawMeshInstanced(mesh graphics.Mesh, modelMatrix []mgl32.Mat4,
					   color []mgl32.Vec4, count int32) {
	graphics.UpdateInstanceBuffer(instanceModelBuffer, int(count), modelMatrix)
	graphics.UpdateInstanceBuffer(instanceColorBuffer, int(count), color)
	graphics.DrawMeshInstanced(mesh, count,
							   []graphics.InstanceBuffer{instanceModelBuffer, instanceColorBuffer},
							   []uint32{2, 6})
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

// GetSceneBuffer retrieves bytes of the buffer which holds the final
// scene rendering, along with its dimensions.
func GetSceneBuffer() ([]byte, int32, int32) {
	buffer := graphics.GetFramebufferPixels(bufferEffect, "color")
	width, height :=  graphics.GetFramebufferSize(bufferEffect)
	return buffer, width, height
}
