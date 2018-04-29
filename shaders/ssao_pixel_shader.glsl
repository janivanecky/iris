#version 420 core

layout (binding = 0) uniform sampler2D position_tex;
layout (binding = 1) uniform sampler2D normal_tex;
layout (binding = 2) uniform sampler2D noise_tex;

in vec2 texcoord;

uniform mat4 projection_matrix;

uniform vec3 kernels[16];

out float occlusion;
uniform float ssao_radius;
uniform float ssao_range;
uniform float ssao_range_boundary;

uniform vec2 screen_size;

void main() {
    vec3 viewspace_position = texture(position_tex, texcoord).xyz;
    vec3 normal = texture(normal_tex, texcoord).xyz;

    vec3 tangent = texture(noise_tex, texcoord * screen_size / 4.0).xyz;
    tangent = normalize(tangent - normal * dot(tangent, normal));

    vec3 bitangent = normalize(cross(tangent, normal));
    mat3 rot = mat3(tangent, normal, bitangent);

    occlusion = 1.0;
    for (int i = 0; i < 16; ++i) {
        vec3 sample_dir = rot * kernels[i];
        vec3 sample_point = viewspace_position + sample_dir * ssao_radius;
        vec4 sample_point_projection = projection_matrix * vec4(sample_point, 1.0f);
        sample_point_projection.xyz /= sample_point_projection.w;

        vec2 sample_point_texcoord = sample_point_projection.xy * 0.5 + 0.5;
        vec3 sampled_position = texture(position_tex, sample_point_texcoord).xyz;
        
        float range_check = 1.0 - smoothstep(
            ssao_range - ssao_range_boundary,
            ssao_range + ssao_range_boundary,
            abs(viewspace_position.z - sampled_position.z)
        );
        occlusion -= sampled_position.z >= sample_point.z ? (1.0 / 16.0 * range_check) : 0;
    }
}