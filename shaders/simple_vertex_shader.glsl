#version 330 core
in vec4 in_position;
in vec4 in_normal;

out vec4 normal;

uniform mat4 projection_matrix;
uniform mat4 view_matrix;

void main()
{
	gl_Position = projection_matrix * view_matrix * in_position;
	normal = in_normal;
}