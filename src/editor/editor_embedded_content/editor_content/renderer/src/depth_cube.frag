#version 460

layout(location = 0) in vec4 psFragPos;
layout(location = 1) in vec3 psFragSourcePos;
layout(location = 2) in float psFragShadowFar;

void main() {
    // get distance between fragment and light source
    float lightDistance = length(psFragPos.xyz - psFragSourcePos);
    
    // map to [0;1] range by dividing by shadowFar
    lightDistance = lightDistance / psFragShadowFar;
    
    // write this as modified depth
    gl_FragDepth = lightDistance;
}
