{"Name":"ui_transparent","EnableDebug":false,"Vertex":"content/renderer/src/ui.vert","VertexFlags":"","Fragment":"content/renderer/src/ui_nine.frag","FragmentFlags":"-DOIT","Geometry":"","GeometryFlags":"","TessellationControl":"","TessellationControlFlags":"","TessellationEvaluation":"","TessellationEvaluationFlags":"","LayoutGroups":[{"Type":"Vertex","Layouts":[{"Location":0,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"Position","Source":"in","Fields":null},{"Location":1,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"Normal","Source":"in","Fields":null},{"Location":2,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"Tangent","Source":"in","Fields":null},{"Location":3,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"UV0","Source":"in","Fields":null},{"Location":4,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"Color","Source":"in","Fields":null},{"Location":5,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"ivec4","Name":"JointIds","Source":"in","Fields":null},{"Location":6,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"JointWeights","Source":"in","Fields":null},{"Location":7,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec3","Name":"MorphTarget","Source":"in","Fields":null},{"Location":-1,"Binding":0,"Set":0,"InputAttachment":-1,"Type":"UniformBufferObject","Name":"","Source":"uniform","Fields":[{"Type":"mat4","Name":"view"},{"Type":"mat4","Name":"projection"},{"Type":"mat4","Name":"uiView"},{"Type":"mat4","Name":"uiProjection"},{"Type":"vec4","Name":"cameraPosition"},{"Type":"vec3","Name":"uiCameraPosition"},{"Type":"vec2","Name":"screenSize"},{"Type":"float","Name":"time"}]},{"Location":8,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"mat4","Name":"model","Source":"in","Fields":null},{"Location":12,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"uvs","Source":"in","Fields":null},{"Location":13,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fgColor","Source":"in","Fields":null},{"Location":14,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"bgColor","Source":"in","Fields":null},{"Location":15,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"scissor","Source":"in","Fields":null},{"Location":16,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"size2D","Source":"in","Fields":null},{"Location":17,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"borderRadius","Source":"in","Fields":null},{"Location":18,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"borderSize","Source":"in","Fields":null},{"Location":19,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"mat4","Name":"borderColor","Source":"in","Fields":null},{"Location":23,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"borderLen","Source":"in","Fields":null},{"Location":0,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragColor","Source":"out","Fields":null},{"Location":1,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragBGColor","Source":"out","Fields":null},{"Location":2,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragSize2D","Source":"out","Fields":null},{"Location":3,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragBorderRadius","Source":"out","Fields":null},{"Location":4,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragBorderSize","Source":"out","Fields":null},{"Location":5,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"mat4","Name":"fragBorderColor","Source":"out","Fields":null},{"Location":9,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"fragTexCoord","Source":"out","Fields":null},{"Location":10,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"fragBorderLen","Source":"out","Fields":null}]},{"Type":"Fragment","Layouts":[{"Location":0,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragColor","Source":"in","Fields":null},{"Location":1,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragBGColor","Source":"in","Fields":null},{"Location":2,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragSize2D","Source":"in","Fields":null},{"Location":3,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragBorderRadius","Source":"in","Fields":null},{"Location":4,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"fragBorderSize","Source":"in","Fields":null},{"Location":5,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"mat4","Name":"fragBorderColor","Source":"in","Fields":null},{"Location":9,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"fragTexCoord","Source":"in","Fields":null},{"Location":10,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec2","Name":"fragBorderLen","Source":"in","Fields":null},{"Location":-1,"Binding":1,"Set":-1,"InputAttachment":-1,"Type":"sampler2D","Name":"texSampler","Source":"uniform","Fields":null},{"Location":0,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"vec4","Name":"outColor","Source":"out","Fields":null},{"Location":1,"Binding":-1,"Set":-1,"InputAttachment":-1,"Type":"float","Name":"reveal","Source":"out","Fields":null}]}]}