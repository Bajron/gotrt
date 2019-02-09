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
	return
}

func add(lhs, rhs Vec3f) (ret Vec3f) {
	for i := range lhs {
		ret[i] = lhs[i] + rhs[i]
	}
	return
}

func accumulate(vectors ...Vec3f) Vec3f {
	ret := Vec3f{0, 0, 0}
	for _, v := range vectors {
		for i := range v {
			ret[i] += v[i]
		}
	}
	return ret
}

func scale(v Vec3f, f float32) (ret Vec3f) {
	for i := range v {
		ret[i] = v[i] * f
	}
	return
}

func dot(lhs, rhs Vec3f) float32 {
	ret := float32(0)
	for i := range lhs {
		ret += lhs[i] * rhs[i]
	}
	return ret
}

func (v Vec3f) length() float32 {
	return float32(math.Sqrt(float64(dot(v, v))))
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
	for i, p := range v {
		ret[i] = p / l
	}
	return
}

func negate(v Vec3f) (ret Vec3f) {
	for i, p := range v {
		ret[i] = -p
	}
	return
}

func reflect(I, normal Vec3f) Vec3f {
	return sub(I, scale(normal, 2*dot(I, normal)))
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

type Material struct {
	diffuseColor     Vec3f
	albedo           [3]float32
	specularExponent float64
}

type Sphere struct {
	center   Vec3f
	radius   float32
	material Material
}

type Light struct {
	position  Vec3f
	intensity float32
}

func (s Sphere) rayIntersects(origin, direction Vec3f) (bool, float32) {
	L := sub(s.center, origin)
	tca := dot(L, direction)
	d2 := dot(L, L) - tca*tca
	r2 := s.radius * s.radius

	if d2 > r2 {
		return false, float32(math.MaxFloat32)
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

func sceneIntersect(origin, direction Vec3f, spheres []Sphere) (intersects bool, hit, N Vec3f, material Material) {
	closestDistance := float32(math.MaxFloat32)

	for _, s := range spheres {
		intersection, distance := s.rayIntersects(origin, direction)
		if intersection && distance < closestDistance {
			closestDistance = distance
			hit = add(origin, scale(direction, distance))
			N = normalize(sub(hit, s.center))
			material = s.material
		}
	}
	intersects = closestDistance < 1000
	return
}

func castRay(origin, direction Vec3f, spheres []Sphere, lights []Light, depth int) (color Vec3f) {
	bgColor := Vec3f{0.2, 0.7, 0.8}

	intersects, point, normal, material := sceneIntersect(origin, direction, spheres)
	if depth < 1 || !intersects {
		return bgColor
	}

	reflectDirection := normalize(reflect(direction, normal))
	// Not to hit the object itself with reflection check
	pointCorrection := scale(normal, 0.001)
	if dot(reflectDirection, normal) < 0 {
		pointCorrection = negate(pointCorrection)
	}
	reflectOrigin := add(point, pointCorrection)
	reflectColor := castRay(reflectOrigin, reflectDirection, spheres, lights, depth-1)

	diffuseLightIntensity, specularLightIntensity := float32(0), float32(0)
	for _, light := range lights {
		lightDirection := normalize(sub(light.position, point))
		lightDistance := sub(light.position, point).length()

		// Not to hit the object itself with shadow check
		pointCorrection := scale(normal, 0.001)
		if dot(lightDirection, normal) < 0 {
			pointCorrection = negate(pointCorrection)
		}
		shadowOrigin := add(point, pointCorrection)

		shadowIntersects, shadowPoint, _, _ := sceneIntersect(shadowOrigin, lightDirection, spheres)
		if shadowIntersects && sub(shadowPoint, shadowOrigin).length() < lightDistance {
			// We hit something before the light ray reaches the point
			continue
		}

		diffuseLightIntensity += light.intensity * float32(math.Max(0, float64(dot(lightDirection, normal))))

		viewAngleToLightReflectionAngleValue := math.Max(0, float64(-dot(reflect(negate(lightDirection), normal), direction)))
		specularLightIntensity += float32(math.Pow(viewAngleToLightReflectionAngleValue, material.specularExponent)) * light.intensity
	}

	white := Vec3f{1, 1, 1}
	return accumulate(
		scale(material.diffuseColor, diffuseLightIntensity*material.albedo[0]),
		scale(white, specularLightIntensity*material.albedo[1]),
		scale(reflectColor, material.albedo[2]))
}

func render(spheres []Sphere, lights []Light) {
	width, height := 1024, 768
	frameBuffer := make([]Vec3f, width*height)
	fWidth, fHeight := float32(width), float32(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			frameBuffer[y*width+x] = NewVec3f(float32(y)/fHeight, float32(x)/fWidth, 0)
		}
	}

	renderingDepth := 4
	fov := math.Pi / 2
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			targetX := (2.0*(float32(x)+0.5)/fWidth - 1.0) * float32(math.Tan(fov/2.0)) * fWidth / fHeight
			targetY := -(2.0*(float32(y)+0.5)/fHeight - 1.0) * float32(math.Tan(fov/2.0))
			direction := normalize(Vec3f{targetX, targetY, -1.0})
			frameBuffer[y*width+x] = castRay(Vec3f{0, 0, 0}, direction, spheres, lights, renderingDepth)
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
	ivory := Material{Vec3f{0.4, 0.4, 0.3}, [3]float32{0.6, 0.3, 0.1}, 50}
	redRubber := Material{Vec3f{0.3, 0.1, 0.1}, [3]float32{0.9, 0.1, 0}, 10}
	mirror := Material{Vec3f{1, 1, 1}, [3]float32{0.0, 10, 0.8}, 1425}

	spheres := []Sphere{}
	spheres = append(spheres, Sphere{Vec3f{-3, 0, -16}, 2, ivory})
	spheres = append(spheres, Sphere{Vec3f{-1, -1.5, -12}, 2, mirror})
	spheres = append(spheres, Sphere{Vec3f{1.5, -0.5, -18}, 3, redRubber})
	spheres = append(spheres, Sphere{Vec3f{7, 5, -18}, 4, mirror})

	lights := []Light{}
	lights = append(lights, Light{Vec3f{-20, 20, 20}, 1.5})
	lights = append(lights, Light{Vec3f{30, 50, -25}, 1.8})
	lights = append(lights, Light{Vec3f{30, 20, 30}, 1.7})

	render(spheres, lights)
}
