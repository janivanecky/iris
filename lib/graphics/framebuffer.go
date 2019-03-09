package graphics

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Attachment specifies single buffer/texture in framebuffer.
type Attachment struct {
	buffer uint32
	index int32
	format uint32
	dataType uint32
}

// Framebuffer represents OpenGL framebuffer.
// Has multiple attachments.
type Framebuffer struct {
	framebuffer uint32
	attachments map[string] Attachment
	width int32
	height int32
}

var attachmentIdxToConst = map[int32] uint32 {
	0: gl.COLOR_ATTACHMENT0,
	1: gl.COLOR_ATTACHMENT1,
	2: gl.COLOR_ATTACHMENT2,
	3: gl.COLOR_ATTACHMENT3,
	4: gl.COLOR_ATTACHMENT4,
	5: gl.COLOR_ATTACHMENT5,
	6: gl.COLOR_ATTACHMENT6,
}

// Helper mappings for various GL formats.
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

// GetFramebufferDefault returns Framebuffer representing backbuffer/screen.
func GetFramebufferDefault() Framebuffer {
	var framebuffer Framebuffer
	framebuffer.framebuffer = 0  // OpenGL considers framebuffer 0 to be backbuffer.
	framebuffer.width = backbufferWidth
	framebuffer.height = backbufferHeight
	framebuffer.attachments = make(map[string]Attachment, 0)
	return framebuffer
}

// GetFramebufferSize returns framebuffer's size in pixels.
func GetFramebufferSize(framebuffer Framebuffer) (int32, int32) {
	return framebuffer.width, framebuffer.height
}

// Use to create empty framebuffer.
func getFramebuffer(width, height int32) Framebuffer {
	var framebuffer Framebuffer
	gl.GenFramebuffers(1, &framebuffer.framebuffer)

	framebuffer.attachments = make(map[string]Attachment, 0)
	framebuffer.width = width
	framebuffer.height = height
	return framebuffer
}

// GetFramebuffer returns initialized Framebuffer object with multiple attachments.
// attachment arguments is a map from attachment name to attachment format (e.g. gl.RGBA8).
func GetFramebuffer(width, height int32, sampleCount int32, attachments map[string]int32, depthBuffer bool) Framebuffer {
	// Get empty framebuffer.
	framebuffer := getFramebuffer(width, height)
	
	// Add individual attachments.
	attachmentIndex := int32(0)
	for attachmentName, attachmentFormat := range attachments {
		framebuffer.addColorAttachment(attachmentName, sampleCount, attachmentFormat, attachmentIndex) 
		attachmentIndex++
	}

	// Optionally add depth attachment.
	if depthBuffer {
		framebuffer.addDepthAttachment(sampleCount)
	}

	// Check that Framebuffer was created successfully.
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.framebuffer)	
	if  gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Framebuffer incomplete, something went wrong during initialization.")
	}
	return framebuffer
}

// Add color attachment to Framebuffer.
func (framebuffer *Framebuffer) addColorAttachment(name string, sampleCount int32, format int32, index int32) {
	// Get color attachment position enum - e.g. COLOR_ATTACHMENT0.
	position := attachmentIdxToConst[index]

	// Get texture format enum and byte type enum.
	textureFormat, ok := internalFormatToFormat[format]
	if !ok {
		panic("Cannot map internal format to texture format.")
	}
	byteType, ok := internalFormatToType[format]
	if !ok {
		panic("Cannot map internal format to texture pixel type")
	}

	// Bind the framebuffer.
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.framebuffer)	
	
	// Create a texture for the attachement.
	var texColorBuffer uint32
	gl.GenTextures(1, &texColorBuffer);

	// We need separate path for multisampled attachments.
	if sampleCount > 1 {
		// Bind texture and create storage for the attachment.
		gl.BindTexture(gl.TEXTURE_2D_MULTISAMPLE, texColorBuffer);
		gl.TexImage2DMultisample(gl.TEXTURE_2D_MULTISAMPLE, sampleCount, uint32(format), framebuffer.width, framebuffer.height, true)

		// Set texture parameters.
		gl.TexParameteri(gl.TEXTURE_2D_MULTISAMPLE, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D_MULTISAMPLE, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		// Bind the attachment to the framebuffer.
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, position, gl.TEXTURE_2D_MULTISAMPLE, texColorBuffer, 0)
	} else {
		// Bind texture and create storage for the attachment.
		gl.BindTexture(gl.TEXTURE_2D, texColorBuffer);
		gl.TexImage2D(gl.TEXTURE_2D, 0, format, framebuffer.width, framebuffer.height, 0, textureFormat, byteType, nil)

		// Set texture parameters.
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		// Bind the attachment to the framebuffer.
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, position, gl.TEXTURE_2D, texColorBuffer, 0)
	}

	framebuffer.attachments[name] = Attachment{texColorBuffer, index, textureFormat, byteType}
}

func (framebuffer *Framebuffer) addDepthAttachment(sampleCount int32) {
	// Bind the framebuffer.
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.framebuffer)	

	// Create and bind renderbuffer for depth buffer.
	var depthBuffer uint32
	gl.GenRenderbuffers(1, &depthBuffer)
	gl.BindRenderbuffer(gl.RENDERBUFFER, depthBuffer)

	// Create storage for depth buffer.
	if sampleCount > 1 {
		gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, sampleCount, gl.DEPTH_COMPONENT32F, framebuffer.width, framebuffer.height) 
	} else {
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT32F, framebuffer.width, framebuffer.height) 
	}

	// Bind depth buffer to framebuffer.
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, depthBuffer);
}

// SetFramebuffer sets specified framebuffer for rendering.
func SetFramebuffer(framebuffer Framebuffer) {
	// Bind framebuffer itself.
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.framebuffer)
	
	// In case framebuffer has no attachments, that's all we needed to do.
	attachmentCount := len(framebuffer.attachments)
	if attachmentCount == 0 {
		return
	}

	// We need to get array of GL enums (e.g. COLOR_ATTACHMENT0), specifying
	// indices of all the attachment in the framebuffer.
	attachments := make([]uint32, attachmentCount)
	for _, attachmentData := range framebuffer.attachments {
		attachmentIndex := attachmentData.index
		attachments[attachmentIndex] = attachmentIdxToConst[attachmentIndex]
	}

	// Set attachments as draw buffers.
	gl.DrawBuffers(int32(attachmentCount), (*uint32)(unsafe.Pointer(&attachments[0])))
}

// SetFramebufferTexture sets attachment in framebuffer as a texture in specific slot.
func SetFramebufferTexture(framebuffer Framebuffer, attachment string, slot int) {
	gl.ActiveTexture(slotToEnum[slot])
	gl.BindTexture(gl.TEXTURE_2D, uint32(framebuffer.attachments[attachment].buffer)) 
}

// GetFramebufferTexture returns framebuffer attachment as a texture.
func GetFramebufferTexture(framebuffer Framebuffer, attachment string) Texture {
	return Texture(framebuffer.attachments[attachment].buffer)
}

// SetFramebufferViewport sets viewport to be full size of framebuffer.
func SetFramebufferViewport(framebuffer Framebuffer) {
	gl.Viewport(0, 0, framebuffer.width, framebuffer.height)
}

// BlitFramebufferAttachment blits 'fromAttachment' of 'from' framebuffer
// into 'toAttachment' of 'to' framebuffer.
func BlitFramebufferAttachment(from, to Framebuffer, fromAttachment, toAttachment string) {
	// Bind read framebuffer and attachment.
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, from.framebuffer)
	gl.ReadBuffer(attachmentIdxToConst[from.attachments[fromAttachment].index])

	// Bind draw framebuffer and attachment.
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, to.framebuffer)
	gl.DrawBuffer(attachmentIdxToConst[to.attachments[toAttachment].index])

	gl.BlitFramebuffer(0, 0, from.width, from.height, 0, 0, to.width, to.height, gl.COLOR_BUFFER_BIT, gl.LINEAR); 
}

// BlitFramebufferToScreen blits attachment of framebuffer onto screen/backbuffer. 
func BlitFramebufferToScreen(from Framebuffer, attachment string) {
	// Bind read framebuffer and attachment.
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, from.framebuffer)
	gl.ReadBuffer(attachmentIdxToConst[from.attachments[attachment].index])

	// Bind backbuffer.
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	gl.DrawBuffer(gl.BACK)

	gl.BlitFramebuffer(0, 0, from.width, from.height, 0, 0, backbufferWidth, backbufferHeight, gl.COLOR_BUFFER_BIT, gl.LINEAR); 
}

// GetFramebufferPixels gets attachment's contents as a byte array.
func GetFramebufferPixels(framebuffer Framebuffer, attachmentName string) []byte {
	// Get total number of bytes in attachment.
	attachment := framebuffer.attachments[attachmentName]
	perPixelBytes := typeToSize[attachment.dataType] * formatToChannels[attachment.format]
	totalBytes := framebuffer.width * framebuffer.height * perPixelBytes

	// Bind read framebuffer and attachment.
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer.framebuffer)
	gl.ReadBuffer(attachmentIdxToConst[attachment.index])
	
	// Read attachment pixels into a buffer.
	buffer := make([]byte, totalBytes)
	// TODO: Check if necessary to pass dataType, maybe instead we can always pass gl.UNSIGNED_BYTE
	gl.ReadPixels(0, 0, framebuffer.width, framebuffer.height, attachment.format, attachment.dataType, gl.Ptr(&buffer[0]))
	return buffer
}
