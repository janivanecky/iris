package main

//import "fmt"

import (
	"github.com/go-gl/glfw/v3.2/glfw";
	"github.com/go-gl/gl/v4.6-core/gl"
)

func main() {
	// Initialize GLFW context
    err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// Create our new fancy window
	window, err := glfw.CreateWindow(640, 480, "New fancy window", nil, nil)
	if err != nil {
		panic(err)
	}
	
	// God knows why this is necessary
	window.MakeContextCurrent()

	// Create an OpenGL context 
	err = gl.Init()
	if  err != nil {
		panic(err)
	}
	
	// Start our fancy-shmancy loop
	for !window.ShouldClose() {
		// Let GLFW interface with the OS - not our job, right?
		glfw.PollEvents()

		// Let's quit if user presses Esc, that cannot mean anything else.
		esc_state := window.GetKey(glfw.KeyEscape)
		if (esc_state == glfw.Press) {
			break
		}

		// We got the cleaning done bitchez.
		gl.ClearColor(0,0,1,0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Do OpenGL stuff.
		window.SwapBuffers()
	}
}