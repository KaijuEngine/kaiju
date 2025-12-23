package content_database

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
)

func init() { addCategory(ParticleSystem{}) }

// ParticleSystem is a [ContentCategory] represented by a file with a
// ".particlesystem" extension. It is a list of emitters that make up a particle
// system.
type ParticleSystem struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (ParticleSystem) Path() string       { return project_file_system.ContentParticlesFolder }
func (ParticleSystem) TypeName() string   { return "ParticleSystem" }
func (ParticleSystem) ExtNames() []string { return []string{".particles"} }

func (ParticleSystem) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("ParticleSystem.Import").End()
	return pathToTextData(src)
}

func (c ParticleSystem) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("ParticleSystem.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (ParticleSystem) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
