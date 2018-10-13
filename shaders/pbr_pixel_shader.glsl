#version 420 core

uniform vec3 camera_pos;

in vec4 position;
in vec4 normal;
in vec4 in_color;

uniform float roughness;
uniform float reflectivity;

uniform float direct_light_power;
uniform float ambient_light_power;

uniform mat4 view_matrix;

layout(location = 0) out vec4 out_diffuse;
layout(location = 1) out vec4 out_ambient;

const float PI = 3.14159265359;

vec4 Shlick(vec4 F0, vec3 l, vec3 h)
{
	return F0 + (vec4(1,1,1,1) - F0) * pow(1 - clamp(dot(l,h), 0, 1),5); 
}

float TR(float alpha, vec3 n, vec3 h)
{
	float alpha2 = alpha * alpha;
	float nominator = alpha2;
	float denominator = clamp(dot(n,h), 0, 1);
	denominator = denominator * denominator * (alpha2 - 1) + 1;
	denominator = denominator * denominator * PI;
	return nominator / denominator;
}

float GGXSmith1(vec3 v, vec3 n, float alpha)
{
	alpha = alpha / 2.0f;
	float NV = abs(dot(n, v)) + 1e-5;
	float nominator = (NV);
	float denominator = NV * (1 - alpha) + alpha;
	return nominator / denominator;
}

float GGXSmith(vec3 l, vec3 v, vec3 n, float alpha)
{
	return GGXSmith1(l, n, alpha) * GGXSmith1(v, n, alpha);
}

vec4 BRDF(vec3 n, vec3 l, vec3 v, vec4 specularColor, vec4 diffuseColor, float roughness)
{
	float nl = clamp(dot(n,l), 0, 1);
	float nv = abs(dot(n,v)) + 1e-5;
	vec3 h = normalize(l + v);

	vec4 F = Shlick(specularColor, l, h);
	float D = TR(roughness, n, h);
	float G = GGXSmith(l,v,n,roughness);
	vec4 specBRDF = F * G * D / (4 * max(nl * nv, 1e-5));

	vec4 diffuseCoef = 1 - F;
	vec4 diffBRDF = diffuseColor * diffuseCoef;

	return diffBRDF + specBRDF;
}

void main()
{
	vec3 worldPos = position.xyz;
	vec4 color = in_color;
	float in_roughness = roughness;
	float in_reflectivity = reflectivity;
	
	vec3 lightPos = vec3(0.0f,50.0f,20.0f);//float3(4,4,4);
	lightPos = vec3(0,50,0.0);
	vec3 lightDir = normalize(lightPos - worldPos);
	//lightDir = normalize(vec3(2,1,3));
	//lightDir = normalize(lightPos);
	vec3 normal_ = normalize(normal.xyz);
	lightDir = normalize((transpose(inverse(view_matrix)) * vec4(0, 1, 0, 0.0)).xyz);
	float lightPower = length(lightPos - worldPos.xyz);
	vec3 camDir = normalize(-worldPos.xyz);
	//if (dot(lightDir, normal) > 0.6f)
	//	lightDir = 2 * normal * dot(normal, camDir) - camDir;

	float lightNormalDot = clamp(dot(normal_, lightDir), 0.0f, 1.0f);

	vec4 lightColor = vec4(1,1,1,1) * direct_light_power;
	vec4 specularColor = vec4(1.0f, 1.0f, 1.0f, 1.0f) * in_reflectivity;
	vec4 diffuseColor = color;
	vec4 ambientColor = color;
	float roughness = sqrt(in_roughness);
	roughness *= roughness;
	
	vec4 col = lightNormalDot * lightColor * PI * BRDF(normal_, lightDir, camDir, specularColor, diffuseColor, roughness);

	col.a = color.a;
	out_diffuse = col;

	out_ambient = ambientColor * ambient_light_power;
}