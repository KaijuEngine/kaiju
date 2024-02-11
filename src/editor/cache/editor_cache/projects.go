package editor_cache

import (
	"kaiju/filesystem"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func projectCacheFolder() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cache = filepath.Join(cache, CacheFolder)
	if _, err := os.Stat(cache); os.IsNotExist(err) {
		os.Mkdir(cache, os.ModePerm)
	}
	return cache, nil
}

func AddProject(project string) error {
	cache, err := projectCacheFolder()
	if err != nil {
		return err
	}
	list, err := ListProjects()
	if err != nil {
		return err
	}
	if slices.Contains(list, project) {
		list = append(list, project)
		filesystem.WriteTextFile(cache, strings.Join(list, "\n"))
	}
	return nil
}

func ListProjects() ([]string, error) {
	cache, err := projectCacheFolder()
	if err != nil {
		return []string{}, err
	}
	projectsList := filepath.Join(cache, "projects.txt")
	list, err := filesystem.ReadTextFile(projectsList)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		} else {
			return []string{}, err
		}
	}
	lines := strings.Split(list, "\n")
	projects := make([]string, 0, len(lines))
	for _, s := range lines {
		s = strings.TrimSpace(s)
		if s != "" {
			projects = append(projects, s)
		}
	}
	return projects, nil
}
