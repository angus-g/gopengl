package main

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/draw"
	_ "image/png"
	"io/ioutil"
	"os"
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

	// vertex attribute object holds links between attributes and vbo
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// vertex buffer with per-vertex data
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// set up position attribute with layout of vertices
	posAttrib := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	gl.VertexAttribPointer(posAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(posAttrib)

	// vertex colour attribute
	colAttrib := uint32(gl.GetAttribLocation(program, gl.Str("color\x00")))
	gl.VertexAttribPointer(colAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(colAttrib)

	// vertex texture coordinate attribute
	texAttrib := uint32(gl.GetAttribLocation(program, gl.Str("texCoord\x00")))
	gl.VertexAttribPointer(texAttrib, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(texAttrib)

	if _, err := newTexture("kitten.png", gl.TEXTURE0); err != nil {
		panic(err)
	}

	uniModel := gl.GetUniformLocation(program, gl.Str("model\x00"))
	uniView := gl.GetUniformLocation(program, gl.Str("view\x00"))
	uniProj := gl.GetUniformLocation(program, gl.Str("proj\x00"))

	matView := mgl32.LookAt(2.0, 2.0, 2.0,
		0.0, 0.0, 0.0,
		0.0, 0.0, 1.0)
	gl.UniformMatrix4fv(uniView, 1, false, &matView[0])

	matProj := mgl32.Perspective(mgl32.DegToRad(45.0), 640.0/480.0, 1.0, 10.0)
	gl.UniformMatrix4fv(uniProj, 1, false, &matProj[0])

	startTime := glfw.GetTime()
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	for !window.ShouldClose() {
		// clear buffer
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		matRot := mgl32.HomogRotate3DZ(float32(glfw.GetTime() - startTime))
		gl.UniformMatrix4fv(uniModel, 1, false, &matRot[0])

		gl.DrawArrays(gl.TRIANGLES, 0, 36)

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

func newTexture(file string, texNum uint32) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	img, _, err := image.Decode(imgFile)

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.ZP, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(texNum)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	return texture, nil
}

var vertices = []float32{
	-0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 0.0,
	0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 1.0, 0.0,
	0.5, 0.5, -0.5, 1.0, 1.0, 1.0, 1.0, 1.0,
	0.5, 0.5, -0.5, 1.0, 1.0, 1.0, 1.0, 1.0,
	-0.5, 0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,
	-0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 0.0,

	-0.5, -0.5, 0.5, 1.0, 1.0, 1.0, 0.0, 0.0,
	0.5, -0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,
	0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 1.0,
	0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 1.0,
	-0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 0.0, 1.0,
	-0.5, -0.5, 0.5, 1.0, 1.0, 1.0, 0.0, 0.0,

	-0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,
	-0.5, 0.5, -0.5, 1.0, 1.0, 1.0, 1.0, 1.0,
	-0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,
	-0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,
	-0.5, -0.5, 0.5, 1.0, 1.0, 1.0, 0.0, 0.0,
	-0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,

	0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,
	0.5, 0.5, -0.5, 1.0, 1.0, 1.0, 1.0, 1.0,
	0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,
	0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,
	0.5, -0.5, 0.5, 1.0, 1.0, 1.0, 0.0, 0.0,
	0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,

	-0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,
	0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 1.0, 1.0,
	0.5, -0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,
	0.5, -0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,
	-0.5, -0.5, 0.5, 1.0, 1.0, 1.0, 0.0, 0.0,
	-0.5, -0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,

	-0.5, 0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,
	0.5, 0.5, -0.5, 1.0, 1.0, 1.0, 1.0, 1.0,
	0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,
	0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 1.0, 0.0,
	-0.5, 0.5, 0.5, 1.0, 1.0, 1.0, 0.0, 0.0,
	-0.5, 0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 1.0,

	-1.0, -1.0, -0.5, 0.0, 0.0, 0.0, 0.0, 0.0,
	1.0, -1.0, -0.5, 0.0, 0.0, 0.0, 1.0, 0.0,
	1.0, 1.0, -0.5, 0.0, 0.0, 0.0, 1.0, 1.0,
	1.0, 1.0, -0.5, 0.0, 0.0, 0.0, 1.0, 1.0,
	-1.0, 1.0, -0.5, 0.0, 0.0, 0.0, 0.0, 1.0,
	-1.0, -1.0, -0.5, 0.0, 0.0, 0.0, 0.0, 0.0,
}
