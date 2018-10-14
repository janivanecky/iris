#version 420 core
layout (location = 0) in vec4 in_position;
layout (location = 1) in vec4 in_normal;
layout (location = 2) in mat4 model_matrix;
in vec4 color;

out vec4 position;
out vec4 normal;
out vec4 in_color;

uniform mat4 projection_matrix;
uniform mat4 view_matrix;

void main()
{
    position = view_matrix * model_matrix * in_position;
	normal = transpose(inverse(view_matrix * model_matrix)) * in_normal;
	gl_Position = projection_matrix * position;
	in_color = color;
}