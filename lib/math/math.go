package math

import "math"

type Vec4 [4]float32;

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