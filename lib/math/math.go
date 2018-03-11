package math

import "math"

type Vec4 [4]float32;
type Vec3 [3]float32;

type Matrix4x4 [4][4]float32;

func GetPerspectiveProjectionGLRH(fov float64, aspectRatio float64, near float64, far float64) Matrix4x4 {
	var result Matrix4x4;
	
	cotFov := 1.0 / (math.Tan(fov / 2.0));
	result[0][0] = float32(cotFov / aspectRatio);
	result[1][1] = float32(cotFov);
	result[2][2] = float32((-near - far) / (far - near));
	result[2][3] = float32(-1.0);
	result[3][2] = float32(-2 * near * far / (far - near));
	
	return result;
}

func GetTranslation(x float64, y float64, z float64) Matrix4x4 {
	var result Matrix4x4;
	
	result[0][0] = float32(1.0);
	result[1][1] = float32(1.0);
	result[2][2] = float32(1.0);
	result[3][0] = float32(x);
	result[3][1] = float32(y);
	result[3][2] = float32(z);
	result[3][3] = float32(1.0);
	
	return result;
}

func length(v Vec3) float32 {
	result := math.Sqrt(float64(v[0] * v[0] + v[1] * v[1] + v[2] * v [2]))
	return float32(result)
}

func (v Vec3) muls(s float32) Vec3 {
	return Vec3{v[0] * s, v[1] * s, v[2] * s}
}

func (v0 Vec3) sub(v1 Vec3) Vec3 {
	return Vec3{v0[0] - v1[0], v0[1] - v1[1], v0[2] - v1[2]}
}

func normalize(v Vec3) Vec3 {
	length := length(v);
	if length < 0.001 {
		return Vec3{}
	}
	return v.muls(1.0 / length)
}

func cross(a Vec3, b Vec3) Vec3 {
	result := Vec3{
		a[1] * b[2] - a[2] * b[1],
		a[2] * b[0] - a[0] * b[2],
		a[0] * b[1] - a[1] * b[0],
	}
	return result;
}

func dot(a Vec3, b Vec3) float32 {
	return a[0] * b[0] + a[1] * b[1] + a[2] * b[2]
}

func GetLookAt(eye Vec3, target Vec3, up Vec3) Matrix4x4 {
	var x, y, z Vec3

	z = eye.sub(target)
	z = normalize(z);

	x = cross(up, z);
	x = normalize(x);

	y = cross(z, x);
	y = normalize(y);

	result := Matrix4x4{
		{x[0], y[0], z[0], 0},
		{x[1], y[1], z[1], 0},
		{x[2], y[2], z[2], 0},
		{-dot(x, eye), -dot(y, eye), -dot(z, eye), 1.0},
	}
	return result
}