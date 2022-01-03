package internal

import "github.com/samlitowitz/goimportcycle/internal/ast/pkg"

type File struct {
	Name    pkg.FilePath
	Package pkg.ImportPath
	Imports map[pkg.ImportPath]*File
}
