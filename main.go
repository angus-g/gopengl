package main

import (
    "github.com/go-gl/glfw/v3.1/glfw"
    "github.com/go-gl/gl/v3.3-core/gl"
    "runtime"
)

func init() {
    runtime.LockOSThread()
}

func main() {
    if err := glfw.Init(); err != nil {
        panic(err)
    }
    defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(640, 480, "GOPenGL", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

    for !window.ShouldClose() {
        window.SwapBuffers()
        glfw.PollEvents()
    }
}
