package project

import (
	"fmt"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
)

const nameSetCodeTitleFileContentFormat = `package build

const Title = GameTitle("%s")
const ArchiveEncryptionKey = "%s"
`

func (p *Project) writeProjectTitle() {
	defer tracing.NewRegion("Project.writeProjectTitle").End()
	err := p.fileSystem.WriteFile(project_file_system.ProjectCodeGameTitle,
		[]byte(fmt.Sprintf(nameSetCodeTitleFileContentFormat,
			p.config.Name, p.config.ArchiveEncryptionKey)), os.ModePerm)
	if err != nil {
		slog.Error("could not set the title in source, please update or create it",
			"file", project_file_system.ProjectCodeGameTitle)
	}
}
