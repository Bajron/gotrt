package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

type Vec3f [3]float32

func clamp1(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (v Vec3f) color(i int) uint8 {
	return uint8(clamp1(v[i]) * 255)
}

func (v Vec3f) R() uint8 {
	return v.color(0)
}
func (v Vec3f) G() uint8 {
	return v.color(1)
}
func (v Vec3f) B() uint8 {
	return v.color(2)
}

func (v Vec3f) ToNRGBA() color.NRGBA {
	return color.NRGBA{v.R(), v.G(), v.B(), 255}
}

func NewVec3f(x, y, z float32) Vec3f {
	return [3]float32{x, y, z}
}

func render() {
	width, height := 1024, 768
	frameBuffer := make([]Vec3f, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			frameBuffer[y*width+x] = NewVec3f(float32(y)/float32(height), float32(x)/float32(width), 0)
		}
	}

	image := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			image.SetNRGBA(x, y, frameBuffer[y*width+x].ToNRGBA())
		}
	}
	file, err := os.Create("output.png")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	writer := bufio.NewWriter(file)
	err = png.Encode(writer, image)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	writer.Flush()
	fmt.Fprintln(os.Stdout, "rendering done")
}

func main() {
	render()
}
