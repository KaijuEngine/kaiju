package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/stages"
)

func TestStageGIProbeIsAProjectReference(t *testing.T) {
	pfs, err := project_file_system.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pfs.Close() })
	for _, folder := range []string{
		project_file_system.ContentStageFolder,
		project_file_system.ContentTemplateFolder,
		project_file_system.ContentMaterialFolder,
		project_file_system.ContentShaderFolder,
		project_file_system.ContentTableOfContentsFolder,
		project_file_system.ContentHtmlFolder,
		project_file_system.ContentCssFolder,
	} {
		if err := pfs.MkdirAll(filepath.Join(project_file_system.ContentFolder, folder), os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}
	for _, folder := range []string{project_file_system.KaijuSrcFolder, project_file_system.ProjectCodeFolder} {
		if err := pfs.MkdirAll(folder, os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}
	project := Project{fileSystem: pfs}
	const stageID = "stage-id"
	const probeID = "probe-id.kjgi"
	stage := stages.Stage{Id: stageID, GlobalIllumination: stages.StageGlobalIllumination{ProbeAsset: probeID}}
	data, err := json.Marshal(stage.ToMinimized())
	if err != nil {
		t.Fatal(err)
	}
	stagePath := filepath.Join(project_file_system.ContentFolder, project_file_system.ContentStageFolder, stageID)
	if err := pfs.WriteFile(stagePath, data, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	references, err := project.FindReferences(probeID)
	if err != nil {
		t.Fatal(err)
	}
	if len(references) != 1 || len(references[0].SubReference) != 1 {
		t.Fatalf("GI probe references = %+v", references)
	}
}
