package modfile

import (
	"errors"
	"os"
	"path/filepath"
)

func FindGoModFile(path string) (string, error) {
	_, err := os.Stat(path + "/go.mod")
	if err == nil {
		return path + "/go.mod", nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return path + "/go.mod", nil
	}
	parent := filepath.Dir(path)
	if parent == path {
		return "", &FileNotFoundError{Name: parent + "/go.mod"}
	}
	return FindGoModFile(parent)
}
