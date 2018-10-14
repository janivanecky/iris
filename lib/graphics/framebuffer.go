package graphics

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Attachment struct {
	buffer uint32
	position int32
	format uint32
	dataType uint32
}

type Framebuffer struct {
	framebuffer uint32
	attachments map[string] Attachment
	width int32
	height int32
}

var attachmentPosToConst = map[int32] uint32 {
	0: gl.COLOR_ATTACHMENT0,
	1: gl.COLOR_ATTACHMENT1,
	2: gl.COLOR_ATTACHMENT2,
	3: gl.COLOR_ATTACHMENT3,
	4: gl.COLOR_ATTACHMENT4,
	5: gl.COLOR_ATTACHMENT5,
	6: gl.COLOR_ATTACHMENT6,
}

var internalFormatToFormat = map[int32] uint32 {
	gl.RGBA8: gl.RGBA,
	gl.RGBA32F: gl.RGBA,
	gl.R32F: gl.RED,
	gl.R16F: gl.RED,
}

var internalFormatToType = map[int32] uint32 {
	gl.RGBA8: gl.UNSIGNED_BYTE,
	gl.RGBA32F: gl.UNSIGNED_BYTE,
	gl.R32F: gl.FLOAT,
	gl.R16F: gl.FLOAT,
}

var typeToSize = map[uint32] int32 {
	gl.UNSIGNED_BYTE: 1,
	gl.FLOAT: 4,
}

var formatToChannels = map[uint32] int32 {
	gl.RGBA: 4,
	gl.RED: 1,
}

func GetFramebufferDefault() Framebuffer {
	var framebuffer Framebuffer
	framebuffer.framebuffer = 0
	framebuffer.width = backbufferWidth
	framebuffer.height = backbufferHeight
	framebuffer.attachments = make(map[string]Attachment, 0)
	return framebuffer
}

func GetFramebufferSize(framebuffer Framebuffer) (int32, int32) {
	return framebuffer.width, framebuffer.height
}

func getFramebuffer(width, height int32) Framebuffer {
	var framebuffer Framebuffer
	gl.GenFramebuffers(1, &framebuffer.framebuffer)

	framebuffer.attachments = make(map[string]Attachment, 1)
	framebuffer.width = width
	framebuffer.height = height
	return framebuffer
}

func GetFramebuffer(width, height int32, sampleCount int32, attachments map[string][]int32, depthBuffer bool) Framebuffer {
	framebuffer := getFramebuffer(width, height)
	for attachmentName, attachmentData := range attachments {
		attachmentFormat := attachmentData[0]
		attachmentIndex := attachmentData[1]
		framebuffer.addColorAttachment(attachmentName, sampleCount, attachmentFormat, attachmentIndex) 
	}

	if depthBuffer {
		framebuffer.addDepthAttachment(sampleCount)
	}

	if !framebuffer.isComplete() {
		panic("FB incomplete")
	}
	return framebuffer
}

func (framebuffer *Framebuffer) addColorAttachment(name string, sampleCount int32, format int32, index int32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.framebuffer)	

	position := attachmentPosToConst[index]

	var texColorBuffer uint32
	gl.GenTextures(1, &texColorBuffer);
	internalFormat, ok := internalFormatToFormat[format]
	if !ok {
		panic("Cannot map internal format to texture format.")
	}
	byteType, ok := internalFormatToType[format]
	if !ok {
		panic("Cannot map internal format to texture pixel type")
	}
	if sampleCount > 1 {
		gl.BindTexture(gl.TEXTURE_2D_MULTISAMPLE, texColorBuffer);
		gl.TexImage2DMultisample(gl.TEXTURE_2D_MULTISAMPLE, sampleCount, uint32(format), framebuffer.width, framebuffer.height, true)
		gl.TexParameteri(gl.TEXTURE_2D_MULTISAMPLE, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D_MULTISAMPLE, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.BindTexture(gl.TEXTURE_2D_MULTISAMPLE, 0)	

		gl.FramebufferTexture2D(gl.FRAMEBUFFER, position, gl.TEXTURE_2D_MULTISAMPLE, texColorBuffer, 0)
	} else {

		gl.BindTexture(gl.TEXTURE_2D, texColorBuffer);
		gl.TexImage2D(gl.TEXTURE_2D, 0, format, framebuffer.width, framebuffer.height, 0, internalFormat, byteType, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.BindTexture(gl.TEXTURE_2D, 0)

		gl.FramebufferTexture2D(gl.FRAMEBUFFER, position, gl.TEXTURE_2D, texColorBuffer, 0)
	}
	framebuffer.attachments[name] = Attachment{texColorBuffer, index, internalFormat, byteType}
}

func (framebuffer *Framebuffer) addDepthAttachment(sampleCount int32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.framebuffer)	

	var depthBuffer uint32
	gl.GenRenderbuffers(1, &depthBuffer)
	gl.BindRenderbuffer(gl.RENDERBUFFER, depthBuffer)
	if sampleCount > 1 {
		gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, sampleCount, gl.DEPTH_COMPONENT32F, framebuffer.width, framebuffer.height) 
	} else {
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT32F, framebuffer.width, framebuffer.height) 
	}
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)

	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, depthBuffer);
}

func (framebuffer *Framebuffer) isComplete() bool {
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.framebuffer)	
	return gl.CheckFramebufferStatus(gl.FRAMEBUFFER) == gl.FRAMEBUFFER_COMPLETE
}

func SetFramebufferDefault() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func SetFramebuffer(framebuffer Framebuffer) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.framebuffer)
	
	attachmentCount := len(framebuffer.attachments)
	if attachmentCount == 0 {
		return
	}
	attachments := make([]uint32, attachmentCount)
	for _, attachmentData := range framebuffer.attachments {
		attachmentIndex := attachmentData.position
		attachments[attachmentIndex] = attachmentPosToConst[attachmentIndex]
	}
	gl.DrawBuffers(int32(attachmentCount), (*uint32)(unsafe.Pointer(&attachments[0])))
}

func SetFramebufferTexture(framebuffer Framebuffer, attachment string, slot int) {
	gl.ActiveTexture(slotToEnum[slot])
	gl.BindTexture(gl.TEXTURE_2D, uint32(framebuffer.attachments[attachment].buffer)) 
}

func GetFramebufferTexture(framebuffer Framebuffer, attachment string) Texture {
	return Texture(framebuffer.attachments[attachment].buffer)
}

func SetFramebufferViewport(framebuffer Framebuffer) {
	gl.Viewport(0, 0, framebuffer.width, framebuffer.height)
}

func SetFramebufferCustomViewport(framebuffer Framebuffer, x, y, width, height float64) {
	x = x * float64(framebuffer.width)
	y = y * float64(framebuffer.height)
	width = width * float64(framebuffer.width)
	height = height * float64(framebuffer.height)
	gl.Viewport(int32(x), int32(y), int32(width), int32(height))
}

func BlitFramebufferAttachment(from, to Framebuffer, fromAttachment, toAttachment string) {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, from.framebuffer)
	gl.ReadBuffer(attachmentPosToConst[from.attachments[fromAttachment].position])
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, to.framebuffer)
	gl.DrawBuffer(attachmentPosToConst[to.attachments[toAttachment].position])
	gl.BlitFramebuffer(0, 0, from.width, from.height, 0, 0, to.width, to.height, gl.COLOR_BUFFER_BIT, gl.LINEAR); 
}

func BlitFramebufferToScreen(from Framebuffer, attachment string) {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, from.framebuffer)
	gl.ReadBuffer(attachmentPosToConst[from.attachments[attachment].position])

	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	gl.DrawBuffer(gl.FRONT)
	gl.BlitFramebuffer(0, 0, from.width, from.height, 0, 0, backbufferWidth, backbufferHeight, gl.COLOR_BUFFER_BIT, gl.LINEAR); 
}

func GetFramebufferPixels(framebuffer Framebuffer, attachmentName string) []byte {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer.framebuffer)
	gl.ReadBuffer(attachmentPosToConst[framebuffer.attachments[attachmentName].position])
	
	attachment := framebuffer.attachments[attachmentName]
	perPixelBytes := typeToSize[attachment.dataType] * formatToChannels[attachment.format]
	totalBytes := framebuffer.width * framebuffer.height * perPixelBytes
	
	buffer := make([]byte, totalBytes)
	// TODO: Check if necessary to pass dataType, maybe instead we can always pass gl.UNSIGNED_BYTE
	gl.ReadPixels(0, 0, framebuffer.width, framebuffer.height, attachment.format, attachment.dataType, gl.Ptr(&buffer[0]))
	return buffer
}
