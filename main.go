package main

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"io/ioutil"
	"runtime"
	"strings"
)

func init() {
	// ensure that the main loop always runs on the primary thread
	runtime.LockOSThread()
}

func main() {
	// initialize GLFW
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// set opengl core profile 3.3
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(640, 480, "GOPenGL", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// initialise OpenGL library
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// link program from shaders
	program, err := newProgram("vertex.glsl", "fragment.glsl")
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	for !window.ShouldClose() {
		// clear to black
		gl.Clear(gl.COLOR_BUFFER_BIT)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func newProgram(vertexShaderFile, fragmentShaderFile string) (uint32, error) {
	// create shaders
	vertexShader, err := compileShader(vertexShaderFile, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderFile, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	// link shaders into program
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// error handling
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	// clean up
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(sourceFile string, shaderType uint32) (uint32, error) {
	// read shader source from file
	sourceBytes, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return 0, err
	}
	// allow use as a C string
	csource := gl.Str(string(sourceBytes) + "\x00")

	// load into OpenGL
	shader := gl.CreateShader(shaderType)
	gl.ShaderSource(shader, 1, &csource, nil)
	gl.CompileShader(shader)

	// error handling
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", sourceFile, log)
	}

	return shader, nil
}
