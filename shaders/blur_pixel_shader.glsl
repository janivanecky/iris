#version 330 core

in vec2 texcoord;
out vec4 out_color;

uniform sampler2D tex;
uniform vec2 screen_size;

void main()
{
    int kernel_size = 5;
    out_color = vec4(0.0);
    for (int y = -2; y <= 2; ++y) {
        for (int x = -2; x <=2; ++x) {
            vec2 offset = vec2(x + 0.5, y + 0.5) / screen_size ;
            out_color += texture(tex, texcoord + offset);
        }
    }

	out_color /= 25.0;
}