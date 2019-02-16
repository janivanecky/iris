package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

// Texture is a handle to OpenGL texture.
type Texture uint32

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
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, int32(width), int32(height),
				  0, uint32(format), gl.UNSIGNED_BYTE, gl.Ptr(data))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	return Texture(textureID)
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
	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat, int32(width), int32(height),
				  0, uint32(format), gl.FLOAT, gl.Ptr(data))

	return Texture(textureID)
}

// DelTexture releases texture from memory.
func DelTexture(texture Texture) {
	tex := uint32(texture)
	gl.DeleteTextures(1, &tex)
}

// SetTexture binds a texture to specific slot.
func SetTexture(texture Texture, slot int) {
	gl.ActiveTexture(slotToEnum[slot])
	gl.BindTexture(gl.TEXTURE_2D, uint32(texture)) 
}
