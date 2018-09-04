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
var pipelineGeometry Pipeline
var pipelineSSAO Pipeline
var pipelineBlur Pipeline
var pipelineBlit Pipeline
var pipelineShading Pipeline
var pipelineEffect Pipeline
var pipelineUI Pipeline

// Framebuffers used for 3D scene rendering.
var bufferGeometry graphics.Framebuffer
var bufferScene graphics.Framebuffer
var bufferSceneMS graphics.Framebuffer
var bufferSSAO graphics.Framebuffer
var bufferBlur graphics.Framebuffer
var bufferShading graphics.Framebuffer
var bufferEffect graphics.Framebuffer

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

// Adjustable parameters for rendering.
var Roughness = 1.0
var Reflectivity = 0.05
var SSAORadius = 0.5
var SSAORange = 3.0
var SSAOBoundary = 1.0
var DirectLight = 0.5
var AmbientLight = 0.75
var MinWhite = 8.0

// Projection related info.
var screenWidth, screenHeight float64
var near, far float32 = 0.01, 500.0

func initSceneRendering(windowWidth, windowHeight float64) {
	// Initialize 3D scene rendering pipelines.
	pipelinePBR = InitPipeline("shaders/geometry_vertex_shader.glsl", "shaders/pbr_pixel_shader.glsl")
	pipelineGeometry = InitPipeline("shaders/geometry_vertex_shader.glsl", "shaders/geometry_pixel_shader.glsl")
	pipelineBlur = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/blur_pixel_shader.glsl")
	pipelineBlit = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/blit_pixel_shader.glsl")
	pipelineEffect = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/effect_pixel_shader.glsl")
	pipelineShading = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/shading_pixel_shader.glsl")
	pipelineSSAO = InitPipeline("shaders/blit_vertex_shader.glsl", "shaders/ssao_pixel_shader.glsl")
	pipelineUI = InitPipeline("shaders/geometry_vertex_shader.glsl", "shaders/flat_pixel_shader.glsl")

	// Set up framebuffers for rendering.
	backbufferWidth, backbufferHeight := windowWidth, windowHeight
	bufferScene = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {
			"direct": []int32{gl.RGBA32F, 0},
			"ambient": []int32{gl.RGBA32F, 1},
		}, true)
	bufferSceneMS = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 4,
		map[string][]int32 {
			"direct": []int32{gl.RGBA32F, 0},
			"ambient": []int32{gl.RGBA32F, 1},
		}, true)
	bufferGeometry = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {
			"position": []int32{gl.RGBA32F, 0},
			"normal": []int32{gl.RGBA32F, 1},
		}, true)
	bufferSSAO = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {"occlusion": []int32{gl.R32F, 0}}, false)
	bufferBlur = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {"occlusion": []int32{gl.R32F, 0}}, false)
	bufferShading = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {"color": []int32{gl.RGBA8, 0}}, false)
	bufferEffect = graphics.GetFramebuffer(
		int32(backbufferWidth), int32(backbufferHeight), 1,
		map[string][]int32 {"color": []int32{gl.RGBA8, 0}}, false)
	

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
}

func renderScene(targetBuffer graphics.Framebuffer, meshEntities []meshData, meshEntitiesUI []meshData) {
	bgColor := float32(0.9)
	// Set up 3D rendering settings - no blending and depth test.
	graphics.DisableBlending()
	graphics.EnableDepthTest()
	
	// First we render the direct and indirect lighting multi-sampled.
	graphics.SetFramebuffer(bufferSceneMS)
	graphics.SetFramebufferViewport(bufferSceneMS)
	graphics.ClearScreen(bgColor, bgColor, bgColor, 1.0)

	pipelinePBR.Start()
	pipelinePBR.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelinePBR.SetUniform("view_matrix", sceneViewMatrix)
	pipelinePBR.SetUniform("roughness", float32(Roughness))
	pipelinePBR.SetUniform("reflectivity", float32(Reflectivity))
	pipelinePBR.SetUniform("direct_light_power", float32(DirectLight))
	pipelinePBR.SetUniform("ambient_light_power", float32(AmbientLight))
	
	for _, meshEntity := range meshEntities {
		drawMesh(pipelinePBR, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}

	// Since color rendering was multisampled, we need to resolve into non-MS framebuffer for it to be used
	// later as a texture.
	graphics.BlitFramebufferAttachment(bufferSceneMS, bufferScene, "direct", "direct")
	graphics.BlitFramebufferAttachment(bufferSceneMS, bufferScene, "ambient", "ambient")

	// Next we render into geometry buffer (position + normal)
	graphics.SetFramebuffer(bufferGeometry)
	graphics.SetFramebufferViewport(bufferGeometry)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)

	pipelineGeometry.Start()
	pipelineGeometry.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelineGeometry.SetUniform("view_matrix", sceneViewMatrix)
	
	for _, meshEntity := range meshEntities {
		drawMesh(pipelineGeometry, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
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
	pipelineSSAO.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelineSSAO.SetUniform("kernels", ssaoKernels[:])
	pipelineSSAO.SetUniform("ssao_radius", float32(SSAORadius))
	pipelineSSAO.SetUniform("ssao_range", float32(SSAORange))
	pipelineSSAO.SetUniform("ssao_range_boundary", float32(SSAOBoundary))
	
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
	graphics.SetFramebufferTexture(bufferScene, "direct", 0)
	graphics.SetFramebufferTexture(bufferScene, "ambient", 1)
	graphics.SetFramebufferTexture(bufferBlur, "occlusion", 2)
	
	pipelineShading.Start()
	pipelineShading.SetUniform("minWhite", float32(MinWhite))
	
	graphics.DrawMesh(screenQuad)
	
	// Post processing effect pass.
	graphics.SetFramebuffer(bufferEffect)
	graphics.SetFramebufferViewport(bufferEffect)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(bufferShading, "color", 0)
	
	pipelineEffect.Start()
	
	graphics.DrawMesh(screenQuad)

	// Blit to screen
	graphics.SetFramebuffer(targetBuffer)
	graphics.SetFramebufferViewport(targetBuffer)
	graphics.ClearScreen(0.0, 0.0, 0.0, 0.0)
	graphics.SetFramebufferTexture(bufferEffect, "color", 0)
	
	pipelineBlit.Start()
	
	graphics.DrawMesh(screenQuad)

	// Draw in-scene UI on top
	graphics.EnableBlending()
	graphics.DisableDepthTest()
	pipelineUI.Start()
	pipelineUI.SetUniform("projection_matrix", sceneProjectionMatrix)
	pipelineUI.SetUniform("view_matrix", sceneViewMatrix)
	for _, meshEntity := range meshEntitiesUI {
		drawMesh(pipelineUI, meshEntity.mesh, meshEntity.modelMatrix, meshEntity.color)
	}
}

func drawMesh(pipeline Pipeline, mesh graphics.Mesh, modelMatrix mgl32.Mat4, color mgl32.Vec4) {
	pipeline.SetUniform("model_matrix", modelMatrix)
	pipeline.SetUniform("color", color)
	graphics.DrawMesh(mesh)
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
	buffer := graphics.GetFramebufferPixels(bufferEffect, "color")
	width, height :=  graphics.GetFramebufferSize(bufferEffect)
	return buffer, width, height
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