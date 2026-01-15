package scanner

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var defaultIgnoredDirs = map[string]struct{}{
	".git":         {},
	".next":        {},
	".turbo":       {},
	"build":        {},
	"dist":         {},
	"node_modules": {},
	"vendor":       {},
}

func Scan(scanRoot string) ([]Project, error) {
	info, err := os.Stat(scanRoot)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("scan root is not a directory")
	}

	projectRoots, err := findProjectRoots(scanRoot)
	if err != nil {
		return nil, err
	}

	sort.Strings(projectRoots)

	projects := make([]Project, 0, len(projectRoots))
	for _, root := range projectRoots {
		envFiles, err := scanProjectEnvFiles(root)
		if err != nil {
			return nil, err
		}
		projects = append(projects, Project{RootPath: root, EnvFiles: envFiles})
	}

	return projects, nil
}

func findProjectRoots(scanRoot string) ([]string, error) {
	var roots []string

	var walk func(dir string) error
	walk = func(dir string) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		// If this directory is a git repo root, record it and stop.
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() == ".git" {
				roots = append(roots, dir)
				return nil
			}
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if isIgnoredDirName(name) {
				continue
			}

			next := filepath.Join(dir, name)
			if err := walk(next); err != nil {
				return err
			}
		}

		return nil
	}

	if err := walk(scanRoot); err != nil {
		return nil, err
	}

	return roots, nil
}

func scanProjectEnvFiles(projectRoot string) ([]string, error) {
	var envFiles []string
	var ignoreStack []gitIgnoreFile

	var walk func(dir string) error
	walk = func(dir string) error {
		// load .gitignore for this directory
		ignoreFile, ok, err := loadGitIgnoreFile(dir)
		if err != nil {
			return err
		}
		if ok {
			ignoreStack = append(ignoreStack, ignoreFile)
			defer func() { ignoreStack = ignoreStack[:len(ignoreStack)-1] }()
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			name := entry.Name()
			fullPath := filepath.Join(dir, name)

			relFromProject, err := filepath.Rel(projectRoot, fullPath)
			if err != nil {
				return err
			}
			relFromProject = filepath.ToSlash(relFromProject)

			isDir := entry.IsDir()
			if isDir {
				if isIgnoredDirName(name) {
					continue
				}
				if isIgnoredByGitignore(ignoreStack, fullPath, relFromProject, true) {
					continue
				}
				if err := walk(fullPath); err != nil {
					return err
				}
				continue
			}

			if isIgnoredByGitignore(ignoreStack, fullPath, relFromProject, false) {
				continue
			}
			if isEnvFileName(name) {
				envFiles = append(envFiles, relFromProject)
			}
		}

		return nil
	}

	if err := walk(projectRoot); err != nil {
		return nil, err
	}

	sort.Strings(envFiles)
	return envFiles, nil
}

func isIgnoredDirName(name string) bool {
	_, ok := defaultIgnoredDirs[name]
	return ok
}

func isEnvFileName(name string) bool {
	if name == ".env" {
		return true
	}
	return strings.HasPrefix(name, ".env")
}

func isIgnoredByGitignore(stack []gitIgnoreFile, fullPath string, relFromProject string, isDir bool) bool {
	// Never ignore the repo's .git directory via patterns here; it's handled by defaultIgnoredDirs.
	if isDir && filepath.Base(fullPath) == ".git" {
		return true
	}

	ignored := false
	for _, ignoreFile := range stack {
		baseRel, err := filepath.Rel(ignoreFile.dir, fullPath)
		if err != nil {
			continue
		}
		baseRel = filepath.ToSlash(baseRel)

		// If the file is outside the ignore file's directory scope (shouldn't happen), skip.
		if strings.HasPrefix(baseRel, "../") {
			continue
		}

		for _, p := range ignoreFile.patterns {
			if !p.matches(baseRel, isDir) {
				continue
			}
			if p.negate {
				ignored = false
			} else {
				ignored = true
			}
		}
	}

	// Also ignore anything under .git even if somehow reached.
	if strings.HasPrefix(relFromProject, ".git/") {
		return true
	}

	return ignored
}
