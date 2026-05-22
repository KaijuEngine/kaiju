/******************************************************************************/
/* project_title_writer.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"fmt"
	"log/slog"
	"os"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

const nameSetCodeTitleFileContentFormat = `package build

const Title = GameTitle("%s")
const ArchiveEncryptionKey = "%s"
`

func (p *Project) writeProjectTitle() {
	defer tracing.NewRegion("Project.writeProjectTitle").End()
	err := p.fileSystem.WriteFile(project_file_system.ProjectCodeGameTitle,
		[]byte(fmt.Sprintf(nameSetCodeTitleFileContentFormat,
			p.Settings.Name, p.Settings.ArchiveEncryptionKey)), os.ModePerm)
	if err != nil {
		slog.Error("could not set the title in source, please update or create it",
			"file", project_file_system.ProjectCodeGameTitle)
	}
}
