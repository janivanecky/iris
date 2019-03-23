package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

// Texture is a handle to OpenGL texture.
type Texture struct {
	texture 	  uint32
	Width, Height int
}
//type Texture uint32

var channelsToFormat = map[int]int32 {
	1: gl.RED,
	3: gl.RGB,
	4: gl.RGBA,
}

var channelsToFormat32 = map[int]int32 {
	1: gl.R32F,
	3: gl.RGB32F,
	4: gl.RGBA32F,
}

var slotToEnum = [...]uint32 {
	gl.TEXTURE0,
	gl.TEXTURE1,
	gl.TEXTURE2,
	gl.TEXTURE3,
}

// GetTextureUint8 creates an OpenGL texture with 8-bit unsigned integer per channel
// and returns a Texture object. If linearFilter is false, nearest neighbor filtering
// will be used when texture is sampled. Supported textures with 1, 3 and 4 channels.
func GetTextureUint8(width int, height int, channels int, data []uint8, linearFilter bool) Texture {
	// Create GL texture.
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	// Set texture parameters.
	filterMode := int32(gl.NEAREST)
	if linearFilter {
		filterMode = gl.LINEAR
	}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filterMode)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filterMode)

	// Upload texture data and generate mipmaps.
	format, ok := channelsToFormat[channels]
	if !ok {
		panic("Incorrect number of channels! Must be 1, 3, or 4.")
	}

	// OpenGL requires textures to be in row order bottom->top, so we need to invert it.
	stride := int(width * channels)
	invertedBytes := invertRows(data, stride, height)

	gl.TexImage2D(gl.TEXTURE_2D, 0, format, int32(width), int32(height),
				  0, uint32(format), gl.UNSIGNED_BYTE, gl.Ptr(invertedBytes))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	return Texture{textureID, width, height}
}

// GetTextureFloat32 creates an OpenGL texture with 32-bit float per channel
// and returns a Texture object. If linearFilter is false, nearest neighbor filtering
// will be used when texture is sampled. Supported textures with 1, 3 and 4 channels.
func GetTextureFloat32(width int, height int, channels int, data []float32, linearFilter bool) Texture {
	// Create GL texture.
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	// Set texture parameters.
	filterMode := int32(gl.NEAREST)
	if linearFilter {
		filterMode = gl.LINEAR
	}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filterMode)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filterMode)

	// Upload texture data.
	format, ok := channelsToFormat[channels]
	if !ok {
		panic("Incorrect number of channels!")
	}
	internalFormat, ok := channelsToFormat32[channels]
	if !ok {
		panic("Incorrect number of channels!")
	}
	// OpenGL requires textures to be in row order bottom->top, so we need to invert it.
	stride := int(width * channels)
	invertedBytes := invertRowsF32(data, stride, int(height))

	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat, int32(width), int32(height),
				  0, uint32(format), gl.FLOAT, gl.Ptr(invertedBytes))

	return Texture{textureID, width, height}
}

// DelTexture releases texture from memory.
func DelTexture(texture Texture) {
	gl.DeleteTextures(1, &texture.texture)
	texture.texture = 0
}

// SetTexture binds a texture to specific slot.
func SetTexture(texture Texture, slot int) {
	gl.ActiveTexture(slotToEnum[slot])
	gl.BindTexture(gl.TEXTURE_2D, texture.texture) 
}

func invertRows(bytes []byte, rowLength, rowCount int) []byte {
	invertedBytes := make([]byte, len(bytes))
	for i := 0; i < rowCount; i++ {
		srcRowStart := i * rowLength
		srcRowEnd := (i + 1) * rowLength
		dstRowStart := (int(rowCount) - 1 - i) * rowLength
		dstRowEnd := (int(rowCount) - i) * rowLength
		copy(invertedBytes[dstRowStart:dstRowEnd], bytes[srcRowStart:srcRowEnd])
	}
	return invertedBytes
}

func invertRowsF32(bytes []float32, rowLength, rowCount int) []float32 {
	invertedBytes := make([]float32, len(bytes))
	for i := 0; i < rowCount; i++ {
		srcRowStart := i * rowLength
		srcRowEnd := (i + 1) * rowLength
		dstRowStart := (int(rowCount) - 1 - i) * rowLength
		dstRowEnd := (int(rowCount) - i) * rowLength
		copy(invertedBytes[dstRowStart:dstRowEnd], bytes[srcRowStart:srcRowEnd])
	}
	return invertedBytes
}