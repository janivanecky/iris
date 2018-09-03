#version 420 core

in vec2 texcoord;
out vec4 out_color;

uniform sampler2D tex;
uniform vec2 screen_size;

void main()
{
    out_color += texture(tex, texcoord);
    float r = texture(tex, texcoord + vec2(.001f, 0)).r;
    float b = texture(tex, texcoord + vec2(.001f, 0)).b;
    out_color.r = r;
    out_color.b = b;

    vec2 pos = texcoord * 2.0 - 1.0;
	//float d = length(pos * pos * pos);
    float d = length(pos);// * pos);

    out_color = clamp(out_color, 0, 1);
	out_color = pow(out_color, vec4(1/2.2f));
	out_color -= d * 0.1f;
    out_color.a = 1.0;
}