package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

// Mesh is a structure encapsulating OpenGL VAO.
type Mesh struct {
	Vao uint32
	indexCount int32	
}

// GetMesh returns a Mesh object from vertices, indices and specified vertex attributes.
// VertexAttribs are specified as a list of byte sizes of individual vertex attributes,
// in order in which they appear in the vertex shader.
func GetMesh(vertices []float32, indices []uint32, vertexAttribs []int) Mesh {
	// Create VAO.
	var cubeVao uint32
	gl.GenVertexArrays(1, &cubeVao)
	gl.BindVertexArray(cubeVao)

	// Create vertex buffer.
	var cubeVbo uint32
	gl.GenBuffers(1, &cubeVbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices) * 4, gl.Ptr(&vertices[0]), gl.STATIC_DRAW)
	
	// Create index buffer.
	var cubeIbo uint32
	gl.GenBuffers(1, &cubeIbo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, cubeIbo);
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices) * 4, gl.Ptr(&indices[0]), gl.STATIC_DRAW);

	// Compute stride for single vertex.
	var stride int32
	for _, vertexAttrib := range vertexAttribs {
		stride += int32(vertexAttrib * 4)
	}

	// Set vertex attribute pointers for the VAO.
	offset := 0
	for i, vertexAttrib := range vertexAttribs {
		index := uint32(i)
		gl.VertexAttribPointer(index, int32(vertexAttrib), gl.FLOAT, false, stride, gl.PtrOffset(offset));
		gl.EnableVertexAttribArray(index);  
		offset += vertexAttrib * 4
	}
	
	return Mesh{cubeVao, int32(len(indices))}
}

// DrawMesh sends a command to draw Mesh to current framebuffer.
func DrawMesh(mesh Mesh) {
	gl.BindVertexArray(mesh.Vao)
	gl.DrawElements(gl.TRIANGLES, mesh.indexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
}

// DrawMeshInstanced sends a command to draw instanceCount instances of Mesh to current framebuffer.
// It also binds InstanceBuffers specified in instanceBuffers slice to locations specified by instanceBufferLocations.
func DrawMeshInstanced(mesh Mesh, instanceCount int32, instanceBuffers []InstanceBuffer, instanceBufferLocations []uint32) {
	gl.BindVertexArray(mesh.Vao)
	for i := range instanceBuffers {
		SetInstanceBuffer(instanceBuffers[i], instanceBufferLocations[i])
	}
	gl.DrawElementsInstanced(gl.TRIANGLES, mesh.indexCount, gl.UNSIGNED_INT, gl.PtrOffset(0), instanceCount)
}
