package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mgeovany/sentra/internal/scanner"
)

func Execute(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "scan":
		if len(args) > 1 {
			return errors.New("sentra scan does not accept flags/args yet")
		}
		return runScan()
	default:
		return usageError()
	}
}

func usageError() error {
	return errors.New("usage: sentra scan")
}

func runScan() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	scanRoot := filepath.Join(homeDir, "dev")

	projects, err := scanner.Scan(scanRoot)
	if err != nil {
		return err
	}

	envCount := 0
	for _, p := range projects {
		envCount += len(p.EnvFiles)
	}

	fmt.Printf("✔ %d projects found\n", len(projects))
	fmt.Printf("✔ %d env files detected\n\n", envCount)

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].RootPath < projects[j].RootPath
	})

	var lines []string
	for _, project := range projects {
		relProjectRoot, err := filepath.Rel(scanRoot, project.RootPath)
		if err != nil {
			return err
		}

		for _, envRel := range project.EnvFiles {
			fullRel := filepath.ToSlash(filepath.Join(relProjectRoot, envRel))
			lines = append(lines, fullRel)
		}
	}

	sort.Strings(lines)
	for _, line := range lines {
		fmt.Println(strings.TrimPrefix(line, "./"))
	}

	return nil
}
