package internal

import "github.com/samlitowitz/goimportcycleviz/internal/ast/pkg"

type File struct {
	Name    pkg.FilePath
	Package pkg.ImportPath
	Imports map[pkg.ImportPath]*File
}
