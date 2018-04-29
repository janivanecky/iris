#version 330 core
in vec4 in_position;
in vec4 in_normal;

out vec4 position;
out vec4 normal;

uniform mat4 projection_matrix;
uniform mat4 view_matrix;
uniform mat4 model_matrix;

void main()
{
    position = view_matrix * model_matrix * in_position;
	normal = transpose(inverse(view_matrix * model_matrix)) * in_normal;
	gl_Position = projection_matrix * position;
}