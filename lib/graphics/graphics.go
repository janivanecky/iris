package graphics

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

var backbufferWidth, backbufferHeight int32
func getBackbufferSize() (int32, int32) {
	var data [4]int32
	dataPtr := (*int32)(unsafe.Pointer(&data[0]))
	gl.GetIntegerv(gl.VIEWPORT, dataPtr)
	return data[2], data[3]
}

// Init initializes OpenGL context.
func Init() {
	err := gl.Init()
	if err != nil {
		panic(err)
	}

	backbufferWidth, backbufferHeight = getBackbufferSize()
}

// ClearScreen clears current framebuffer to specific color.
func ClearScreen(r float32, g float32, b float32, a float32) {
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

// EnableBlending enables alpha blending.
// col = src_col * src_alpha + dst_col * (1 - src_alpha)
func EnableBlending() {
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

// DisableBlending disables alpha blending.
func DisableBlending() {
	gl.Disable(gl.BLEND)
}

// EnableDepthTest enables depth test.
func EnableDepthTest() {
	gl.Enable(gl.DEPTH_TEST)
}

// DisableDepthTest disables depth test.
func DisableDepthTest() {
	gl.Disable(gl.DEPTH_TEST)
}

// EnableSRGBRendering sets current framebuffer to be in SRGB color space.
func EnableSRGBRendering() {
	gl.Enable(gl.FRAMEBUFFER_SRGB)
}

// DisableSRGBRendering sets current framebuffer to be in linear color space.
func DisableSRGBRendering() {
	gl.Disable(gl.FRAMEBUFFER_SRGB)
}
