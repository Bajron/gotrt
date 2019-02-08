package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type Vec3f [3]float32

func sub(lhs, rhs Vec3f) (ret Vec3f) {
	for i := range lhs {
		ret[i] = lhs[i] - rhs[i]
	}
}

func dot(lhs, rhs Vec3f) float32 {
	var float32 ret = 0
	for i := range lhs {
		ret = lhs[i] * rhs[i]
	}
}

func (v Vec3f) length() float32 {
	return float32(math.Sqrt(dot(v, v)))
}

func clamp1(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func normalize(v Vec3f) (ret Vec3f) {
	l := v.length()
	for i := range v {
		ret[i] = ret[i] / l
	}
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

type Sphere struct {
	center Vec3f
	radius float32
}

func (s Sphere) rayIntersects(origin, direction Vec3f) (bool, float32) {
	L := sub(s.center, origin)
	tca := dot(L, direction)
	d2 := dot(L, L) - tca*tca
	r2 := s.radius * s.radius
	if d2 > r2 {
		return false, 0
	}

	thc := float32(math.Sqrt(float64(r2 - d2)))
	t0 := tca - thc
	t1 := tca + thc
	if t0 < 0 {
		t0 = t1
	}
	if t0 < 0 {
		return false, t0
	}
	return true, t0
}

func castRay(origin, direction Vec3f, sphere Sphere) (color Vec3f) {
	bgColor := Vec3f{0.2, 0.7, 0.8}
	sphereColor := Vec3f{0.4, 0.4, 0.3}
	intersects, distance := sphere.rayIntersects(origin, direction)
	if !intersects {
		return bgColor
	}
	return sphereColor
}

func render() {
	width, height := 1024, 768
	frameBuffer := make([]Vec3f, width*height)
	fWidth, fHeight := float32(width), float32(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			frameBuffer[y*width+x] = NewVec3f(float32(y)/fHeight, float32(x)/fWidth, 0)
		}
	}

	fov := math.Pi / 4
	sphere := Sphere{Vec3f{}, 1}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			targetX := (2.0*(float32(x)+0.5)/fWidth - 1.0) * float32(math.Tan(fov/2.0)) * fWidth / fHeight
			targetY := -(2.0*(float32(y)+0.5)/fHeight - 1.0) * float32(math.Tan(fov/2.0))
			direction := normalize(Vec3f{targetX, targetY, -1.0})
			frameBuffer[y*width+x] = castRay(Vec3f{0, 0, 0}, direction, sphere)
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
