{"Name":"pbr","EnableDebug":false,"Vertex":"pbr.vert","VertexFlags":"-DSKINNING","Fragment":"pbr.frag","FragmentFlags":"","Geometry":"","GeometryFlags":"","TessellationControl":"","TessellationControlFlags":"","TessellationEvaluation":"","TessellationEvaluationFlags":"",

"VertexSpv":"pbr_skinned.vert.spv",
"FragmentSpv":"pbr.frag.spv",
"GeometrySpv":"",
"TessellationControlSpv":"",
"TessellationEvaluationSpv":"",

"LayoutGroups":[{"Type":"Vertex","Layouts":[{"Location":-1,"Binding":0,"Count":1,"Set":0,"InputAttachment":-1,"Type":"UniformBufferObject","Name":"","Source":"uniform","Fields":[{"Type":"mat4","Name":"view"},{"Type":"mat4","Name":"projection"},{"Type":"mat4","Name":"uiView"},{"Type":"mat4","Name":"uiProjection"},{"Type":"vec4","Name":"cameraPosition"},{"Type":"vec3","Name":"uiCameraPosition"},{"Type":"vec2","Name":"screenSize"},{"Type":"float","Name":"time"},{"Type":"Light","Name":"vertLights[20]"},{"Type":"LightInfo","Name":"lightInfos[20]"}]},
{"Location":0,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"Position","Source":"in","Fields":null},
{"Location":1,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"Normal","Source":"in","Fields":null},
{"Location":2,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"Tangent","Source":"in","Fields":null},
{"Location":3,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"UV0","Source":"in","Fields":null},
{"Location":4,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"Color","Source":"in","Fields":null},
{"Location":5,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"ivec4","Name":"JointIds","Source":"in","Fields":null},
{"Location":6,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"JointWeights","Source":"in","Fields":null},
{"Location":7,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"MorphTarget","Source":"in","Fields":null},
{"Location":8,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"mat4","Name":"model","Source":"in","Fields":null},
{"Location":-1,"Binding":2,"Set":0,"InputAttachment":-1,"Type":"SkinnedUBO","Name":"","Source":"buffer","Fields":[{"Type":"mat4","Name":"jointTransforms[][50]"}]}
{"Location":12,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"vertColors","Source":"in","Fields":null},
{"Location":13,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"metallic","Source":"in","Fields":null},
{"Location":14,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"roughness","Source":"in","Fields":null},
{"Location":15,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"emissive","Source":"in","Fields":null},
{"Location":16,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"int","Name":"lightIds","Source":"in","Fields":null},
{"Location":20,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"uint","Name":"flags","Source":"in","Fields":null},

{"Location":0,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragColor","Source":"out","Fields":null},
{"Location":1,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"fragTexCoords","Source":"out","Fields":null},{"Location":2,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragTangentViewPos","Source":"out","Fields":null},{"Location":3,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragTangentFragPos","Source":"out","Fields":null},{"Location":4,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragLightTPos","Source":"out","Fields":null},{"Location":8,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragLightTDir","Source":"out","Fields":null},{"Location":12,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragPosLightSpace","Source":"out","Fields":null},{"Location":16,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragPos","Source":"out","Fields":null},{"Location":17,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragNormal","Source":"out","Fields":null},{"Location":18,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"fragMetallic","Source":"out","Fields":null},{"Location":19,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"fragRoughness","Source":"out","Fields":null},{"Location":20,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"fragEmissive","Source":"out","Fields":null},{"Location":21,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"flat","Name":"int","Source":"out","Fields":null},{"Location":22,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"flat","Name":"int","Source":"out","Fields":null}]},{"Type":"Fragment","Layouts":[{"Location":-1,"Binding":0,"Count":1,"Set":0,"InputAttachment":-1,"Type":"UniformBufferObject","Name":"","Source":"uniform","Fields":[{"Type":"mat4","Name":"view"},{"Type":"mat4","Name":"projection"},{"Type":"mat4","Name":"uiView"},{"Type":"mat4","Name":"uiProjection"},{"Type":"vec4","Name":"cameraPosition"},{"Type":"vec3","Name":"uiCameraPosition"},{"Type":"vec2","Name":"screenSize"},{"Type":"float","Name":"time"},{"Type":"Light","Name":"vertLights[20]"},{"Type":"LightInfo","Name":"lightInfos[20]"}]},

{"Location":0,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragColor","Source":"in","Fields":null},
{"Location":1,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"fragTexCoords","Source":"in","Fields":null},
{"Location":2,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragTangentViewPos","Source":"in","Fields":null},
{"Location":3,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragTangentFragPos","Source":"in","Fields":null},
{"Location":4,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragLightTPos","Source":"in","Fields":null},
{"Location":8,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragLightTDir","Source":"in","Fields":null},
{"Location":12,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragPosLightSpace","Source":"in","Fields":null},
{"Location":16,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragPos","Source":"in","Fields":null},
{"Location":17,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"fragNormal","Source":"in","Fields":null},
{"Location":18,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"fragMetallic","Source":"in","Fields":null},
{"Location":19,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"fragRoughness","Source":"in","Fields":null},
{"Location":20,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"fragEmissive","Source":"in","Fields":null},
{"Location":21,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"flat int","Name":"lightCount","Source":"in","Fields":null},
{"Location":22,"Binding":-1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"flat int","Name":"lightIndexes","Source":"in","Fields":null},
{"Location":26,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"flat uint","Name":"fragFlags","Source":"in","Fields":null},

{"Location":0,"Binding":-1,"Count":1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"outColor","Source":"out","Fields":null},
{"Location":0,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"outPosition","Source":"out","Fields":null},
{"Location":0,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"outNormal","Source":"out","Fields":null},

{"Location":-1,"Binding":1,"Count":4,"Set":-1,"InputAttachment":-1,"Type":"sampler2D","Name":"textures","Source":"uniform","Fields":null},
{"Location":-1,"Binding":2,"Count":20,"Set":-1,"InputAttachment":-1,"Type":"sampler2D","Name":"shadowMap","Source":"uniform","Fields":null},{"Location":-1,"Binding":3,"Count":20,"Set":-1,"InputAttachment":-1,"Type":"samplerCube","Name":"shadowCubeMap","Source":"uniform","Fields":null}]}],

"SamplerLabels":["Diffuse","Normal","Metallic Roughness","Emissive"]

}