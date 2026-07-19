package content_database

import (
	"fmt"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(GIProbe{}) }

// GIProbe is a validated, versioned baked global-illumination probe field.
type GIProbe struct{}

func (GIProbe) Path() string       { return project_file_system.ContentGIProbeFolder }
func (GIProbe) TypeName() string   { return "GI Probe" }
func (GIProbe) ExtNames() []string { return []string{".kjgi"} }

func (GIProbe) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("GIProbe.Import").End()
	processed, err := pathToBinaryData(src)
	if err != nil {
		return ProcessedImport{}, err
	}
	if len(processed.Variants) != 1 {
		return ProcessedImport{}, fmt.Errorf("GI probe import produced %d variants", len(processed.Variants))
	}
	if _, err := gi.UnmarshalProbeAsset(processed.Variants[0].Data); err != nil {
		return ProcessedImport{}, fmt.Errorf("invalid GI probe asset: %w", err)
	}
	return processed, nil
}

func (c GIProbe) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("GIProbe.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (GIProbe) PostImportProcessing(ProcessedImport, *ImportResult, *project_file_system.FileSystem, *Cache, string) error {
	return nil
}
