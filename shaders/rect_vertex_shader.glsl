#version 330 core
in vec4 in_position;

uniform mat4 projection_matrix;
uniform mat4 model_matrix;

void main()
{
	gl_Position = projection_matrix * model_matrix * in_position;
}