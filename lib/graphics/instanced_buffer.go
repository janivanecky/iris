package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type InstanceBuffer struct {
	buffer uint32
	numElements int32
}

func GetInstanceBufferFloat32(numElements int32) InstanceBuffer {
	var buffer uint32
	gl.GenBuffers(1, &buffer)

	return InstanceBuffer{buffer, numElements}
}

func UpdateInstanceBuffer(buffer InstanceBuffer, count int, data interface{}) {
	dataPointer := gl.Ptr(&count)
	switch data.(type) {
	case []mgl32.Vec2:
		array := data.([]mgl32.Vec2)
		dataPointer = gl.Ptr(&array[0][0])
	case []mgl32.Vec3:
		array := data.([]mgl32.Vec3)
		dataPointer = gl.Ptr(&array[0][0])
	case []mgl32.Vec4:
		array := data.([]mgl32.Vec4)
		dataPointer = gl.Ptr(&array[0][0])
	case []mgl32.Mat4:
		array := data.([]mgl32.Mat4)
		dataPointer = gl.Ptr(&array[0][0])
	default:
		panic("INCORRECT DATA TYPE")
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer.buffer);
	gl.BufferData(gl.ARRAY_BUFFER, count * 4 * int(buffer.numElements), dataPointer, gl.DYNAMIC_DRAW);
}

func SetInstanceBuffer(buffer InstanceBuffer, location uint32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer.buffer);
	offset := 0
	currentLocation := location
	for elementsLeft := buffer.numElements; elementsLeft > 0; {
		gl.EnableVertexAttribArray(currentLocation)
		numElements := elementsLeft
		if numElements > 4 {
			numElements = 4
		}
		gl.VertexAttribPointer(currentLocation, numElements, gl.FLOAT, false, buffer.numElements * 4, gl.PtrOffset(offset))
		gl.VertexAttribDivisor(currentLocation, 1)
		currentLocation++
		elementsLeft -= numElements
		offset += int(numElements) * 4
	}
}
