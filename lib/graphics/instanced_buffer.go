package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

// InstanceBuffer is an object used to pass per-instance data for instanced rendering.
// For each instance there's only a single item with N elements in the buffer.
// Example - vec3 is an item with 3 elements, mat4 is an item with 16 elements etc.
type InstanceBuffer struct {
	buffer uint32
	numElements int32
}

// Byte size of a single element in InstanceBuffer.
const elementByteSize = 4

// GetInstanceBuffer returns an initialized InstanceBuffer.
// numElements specified number of scalars per-item, not number of items in the buffer.
func GetInstanceBuffer(numElements int32) InstanceBuffer {
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	return InstanceBuffer{buffer, numElements}
}

// UpdateInstanceBuffer updates data inside InstanceBuffer.
func UpdateInstanceBuffer(buffer InstanceBuffer, count int, data interface{}) {
	totalBufferSize := count * elementByteSize * int(buffer.numElements)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer.buffer)
	gl.BufferData(gl.ARRAY_BUFFER, totalBufferSize, gl.Ptr(data), gl.DYNAMIC_DRAW)
}

// SetInstanceBuffer sets elements in InstanceBuffer as vertex attributes for instanced rendering.
func SetInstanceBuffer(buffer InstanceBuffer, location uint32) {
	// Bind GL buffer.
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer.buffer);
	
	// Since we cannot bind items with more than 4 elements to a single attribute,
	// we need to split larger items into more attributes. We'll do that in a loop
	// in which we'll bind elements in chunks of max size 4 and stop if there are 
	// nor more elements to bind as attributes.
	offset, currentLocation := 0, location
	stride := buffer.numElements * int32(elementByteSize)
	for elementsLeft := buffer.numElements; elementsLeft > 0; {
		// Clamp the size of attribute to 4.
		numElements := elementsLeft
		if numElements > 4 {
			numElements = 4
		}
		// Bind the current element as vertex attribute.
		gl.EnableVertexAttribArray(currentLocation)
		gl.VertexAttribPointer(currentLocation, numElements, gl.FLOAT, false, stride, gl.PtrOffset(offset))
		gl.VertexAttribDivisor(currentLocation, 1)

		// Update pointers and no. elements left.
		currentLocation++
		offset += int(numElements) * elementByteSize
		elementsLeft -= numElements
	}
}
