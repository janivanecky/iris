package graphics

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Mesh struct {
	Vao uint32
	indexCount int32	
}

func GetMesh(vertices []float32, indices []uint32, vertexAttribs []int) Mesh {
	var cubeVao uint32
	gl.GenVertexArrays(1, &cubeVao)
	gl.BindVertexArray(cubeVao)

	var cubeVbo uint32
	gl.GenBuffers(1, &cubeVbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices) * 4, gl.Ptr(&vertices[0]), gl.STATIC_DRAW)
	
	var stride int32 = 0
	for _, vertexAttrib := range vertexAttribs {
		stride += int32(vertexAttrib * 4)
	}

	offset := 0
	for i, vertexAttrib := range vertexAttribs {
		index := uint32(i)
		gl.VertexAttribPointer(index, int32(vertexAttrib), gl.FLOAT, false, stride, gl.PtrOffset(offset));
		gl.EnableVertexAttribArray(index);  
		offset += vertexAttrib * 4
	}
	
	var cubeIbo uint32
	gl.GenBuffers(1, &cubeIbo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, cubeIbo);
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices) * 4, gl.Ptr(&indices[0]), gl.STATIC_DRAW);

	return Mesh{cubeVao, int32(len(indices))}
}

func DrawMesh(mesh Mesh) {
	gl.BindVertexArray(mesh.Vao)
	gl.DrawElements(gl.TRIANGLES, mesh.indexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
}

func DrawMeshInstanced(mesh Mesh, instanceCount int32, instanceBuffers []InstanceBuffer, instanceBufferLocations []uint32) {
	gl.BindVertexArray(mesh.Vao)
	for i := range instanceBuffers {
		SetInstanceBuffer(instanceBuffers[i], instanceBufferLocations[i])
	}
	gl.DrawElementsInstanced(gl.TRIANGLES, mesh.indexCount, gl.UNSIGNED_INT, gl.PtrOffset(0), instanceCount)
}


func DrawLineMesh(mesh Mesh) {
	gl.BindVertexArray(mesh.Vao)
	gl.DrawElements(gl.LINE_STRIP, mesh.indexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
}