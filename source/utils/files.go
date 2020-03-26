package utils

import (
	"os"
	"path/filepath"
	"regexp"
)

type directoryWalker struct {
	files   []string
	pattern string
}

// GetMatchingFiles recursively steps through all files in the current working directory and finds
// files matching the specified regex pattern.
func GetMatchingFiles(pattern string, source string) ([]string, error) {
	walker := &directoryWalker{
		files:   make([]string, 0),
		pattern: pattern,
	}
	err := filepath.Walk(source, walker.dirWalk)
	if err != nil {
		return nil, err
	}
	return walker.files, nil
}

func (walker *directoryWalker) dirWalk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	regex := regexp.MustCompile(walker.pattern)
	if regex.MatchString(path) {
		walker.files = append(walker.files, path)
	}
	return nil
}
