package modfile

import "fmt"

type FileNotFoundError struct {
	Name string
}

func (err *FileNotFoundError) Error() string {
	return fmt.Sprintf("`%s` not found.", err.Name)
}

type ModulePathNotFoundError struct{}

func (err *ModulePathNotFoundError) Error() string {
	return "go module path not found"
}
