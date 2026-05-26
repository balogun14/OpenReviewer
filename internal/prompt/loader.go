package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Loader struct {
	baseDir string
}

func NewLoader(baseDir string) Loader {
	return Loader{baseDir: filepath.Clean(baseDir)}
}

func (l Loader) Load(def Definition) (string, error) {
	path, err := l.safePath(def.Path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read prompt %q: %w", def.ID, err)
	}

	resolved, err := l.processIncludes(string(content), filepath.Dir(path))
	if err != nil {
		return "", fmt.Errorf("process prompt %q includes: %w", def.ID, err)
	}

	return resolved, nil
}

func (l Loader) processIncludes(content string, currentDir string) (string, error) {
	includePattern := regexp.MustCompile(`@include\(([^)]+)\)`)
	matches := includePattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) != 2 {
			continue
		}

		includePath, err := l.safePathFrom(currentDir, strings.TrimSpace(match[1]))
		if err != nil {
			return "", err
		}

		includeContent, err := os.ReadFile(includePath)
		if err != nil {
			return "", fmt.Errorf("read include %q: %w", match[1], err)
		}

		content = strings.ReplaceAll(content, match[0], string(includeContent))
	}

	return content, nil
}

func (l Loader) safePath(relative string) (string, error) {
	return l.safePathFrom(l.baseDir, relative)
}

func (l Loader) safePathFrom(baseDir string, relative string) (string, error) {
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("absolute prompt paths are not allowed: %s", relative)
	}

	baseAbs, err := filepath.Abs(l.baseDir)
	if err != nil {
		return "", err
	}

	candidate, err := filepath.Abs(filepath.Join(baseDir, filepath.FromSlash(relative)))
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(baseAbs, candidate)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("prompt path escapes base directory: %s", relative)
	}

	return candidate, nil
}
