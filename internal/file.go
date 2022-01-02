package internal

import "github.com/samlitowitz/goimportcycleviz/internal/ast/pkg"

type File struct {
	Name    pkg.FilePath
	Imports map[pkg.ImportPath]*Package
}
