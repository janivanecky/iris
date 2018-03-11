#version 330 core
out vec4 out_color;
in vec4 normal;

uniform vec3 light_position;

void main()
{
	vec3 light_direction = normalize(light_position);
	float d = clamp(dot(light_direction, normal.xyz), 0, 1);
	out_color.xyz = vec3(0,1,0) * d;
	out_color.a = 1.0f;
}