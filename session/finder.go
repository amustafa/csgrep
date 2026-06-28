package session

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var projectsBase string

func init() {
	if env := os.Getenv("CSGREP_PROJECTS_DIR"); env != "" {
		projectsBase = env
		return
	}
	projectsBase = defaultProjectsDir()
}

func defaultProjectsDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "claude", "projects")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not determine home directory: %v\n", err)
		home = "."
	}
	return filepath.Join(home, ".claude", "projects")
}

func encodeDirPath(cwd string) string {
	cwd = filepath.Clean(cwd)
	encoded := strings.ReplaceAll(cwd, string(filepath.Separator), "-")
	if runtime.GOOS == "windows" {
		encoded = strings.ReplaceAll(encoded, ":", "-")
	}
	encoded = strings.ReplaceAll(encoded, ".", "-")
	encoded = strings.TrimPrefix(encoded, "-")
	return "-" + encoded
}

func findSessionsDir(cwd string) string {
	key := encodeDirPath(cwd)
	candidate := filepath.Join(projectsBase, key)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}
	trimmed := strings.TrimPrefix(key, "-")
	entries, err := os.ReadDir(projectsBase)
	if err != nil {
		return candidate
	}
	for _, e := range entries {
		if e.IsDir() && strings.Contains(e.Name(), trimmed) {
			return filepath.Join(projectsBase, e.Name())
		}
	}
	return candidate
}

func FindFiles(filter Filter) []string {
	if filter.Dir != "" {
		raw := filter.Dir
		if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "~") || strings.HasPrefix(raw, ".") {
			dir := resolveDir(raw)
			return findFilesForDir(dir)
		}
		return findBySubstring(raw)
	}
	return findAllFiles()
}

func isRootDir(dir string) bool {
	return dir == "/" || dir == "." || dir == filepath.VolumeName(dir)+string(filepath.Separator)
}

func findFilesForDir(dir string) []string {
	for !isRootDir(dir) {
		sessDir := findSessionsDir(dir)
		files := globJSONL(sessDir)
		if len(files) > 0 {
			return files
		}
		dir = filepath.Dir(dir)
	}
	return nil
}

func findAllFiles() []string {
	var files []string
	entries, err := os.ReadDir(projectsBase)
	if err != nil {
		return files
	}
	for _, e := range entries {
		if e.IsDir() {
			files = append(files, globJSONL(filepath.Join(projectsBase, e.Name()))...)
		}
	}
	return files
}

func findBySubstring(substr string) []string {
	var files []string
	lower := strings.ToLower(substr)
	entries, err := os.ReadDir(projectsBase)
	if err != nil {
		return files
	}
	for _, e := range entries {
		if e.IsDir() && strings.Contains(strings.ToLower(e.Name()), lower) {
			files = append(files, globJSONL(filepath.Join(projectsBase, e.Name()))...)
		}
	}
	return files
}

func FindByID(sessionID string) string {
	entries, err := os.ReadDir(projectsBase)
	if err != nil {
		return ""
	}
	lower := strings.ToLower(sessionID)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(projectsBase, e.Name())
		files := globJSONL(dir)
		for _, f := range files {
			base := strings.TrimSuffix(filepath.Base(f), ".jsonl")
			if strings.ToLower(base) == lower || strings.HasPrefix(strings.ToLower(base), lower) {
				return f
			}
		}
	}
	return ""
}

func globJSONL(dir string) []string {
	matches, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return nil
	}
	return matches
}

func resolveDir(d string) string {
	if strings.HasPrefix(d, "~") {
		home, _ := os.UserHomeDir()
		d = home + d[1:]
	}
	abs, err := filepath.Abs(d)
	if err != nil {
		return d
	}
	return abs
}
