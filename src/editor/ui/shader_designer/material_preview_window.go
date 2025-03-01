package shader_designer

import (
	"kaiju/host_container"
)

func PreviewMaterial(path string) {
	m, ok := loadMaterialData(path)
	if !ok {
		return
	}
	container := host_container.New("Material Preview", nil)
	go container.Run(480, 480, -1, -1)
	<-container.PrepLock
	container.RunFunction(func() {
		host := container.Host
		m.Compile(host.AssetDatabase(), host.Window.Renderer)
		// TODO:  Complete
	})
}
