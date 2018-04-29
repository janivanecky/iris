#version 330 core
in vec4 in_position;
in vec2 in_texcoord;

out vec2 texcoord;

void main()
{
	gl_Position = in_position;
	texcoord = in_texcoord;
}