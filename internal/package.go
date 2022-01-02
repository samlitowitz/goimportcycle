package internal

import "github.com/samlitowitz/goimportcycleviz/internal/ast/pkg"

type Package struct {
	ImportPath pkg.ImportPath
	Files      map[pkg.FilePath]*File
}
