#version 330 core
in vec4 in_position;
in vec2 in_texcoord;

out vec2 texcoord;

uniform mat4 projection_matrix;
uniform mat4 model_matrix;
uniform vec4 source_rect;

void main()
{
	gl_Position = projection_matrix * model_matrix * in_position;
	texcoord = in_texcoord * source_rect.zw + source_rect.xy;
}