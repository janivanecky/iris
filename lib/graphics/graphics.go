package graphics

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func getBackbufferSize() (int32, int32) {
	var data [4]int32
	dataPtr := (*int32)(unsafe.Pointer(&data[0]))
	gl.GetIntegerv(gl.VIEWPORT, dataPtr)
	return data[2], data[3]
}

var backbufferWidth, backbufferHeight int32

func Init() {
	err := gl.Init()
	if err != nil {
		panic(err)
	}

	backbufferWidth, backbufferHeight = getBackbufferSize()
}

func ClearScreen(r float32, g float32, b float32, a float32) {
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func EnableBlending() {
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func DisableBlending() {
	gl.Disable(gl.BLEND)
}

func EnableDepthTest() {
	gl.Enable(gl.DEPTH_TEST)
}

func DisableDepthTest() {
	gl.Disable(gl.DEPTH_TEST)
}

func EnableSRGBRendering() {
	gl.Enable(gl.FRAMEBUFFER_SRGB)
}

func DisableSRGBRendering() {
	gl.Disable(gl.FRAMEBUFFER_SRGB)
}