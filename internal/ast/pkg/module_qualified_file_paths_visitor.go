package pkg

import (
	"go/ast"
	"path/filepath"
	"strings"
)

type ModuleQualifiedFilePathsVisitor struct {
	modulePath string
}

func NewModuleQualifiedFilePathsVisitor(modulePath string) *ModuleQualifiedFilePathsVisitor {
	return &ModuleQualifiedFilePathsVisitor{
		modulePath: modulePath,
	}
}

func (v *ModuleQualifiedFilePathsVisitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.Package:
		toAdd := make(map[string]*ast.File, len(n.Files))
		toDelete := make([]string, 0, len(n.Files))

		for filePath, astFile := range n.Files {
			absPath, err := filepath.Abs(filePath)
			if err != nil {
				continue
			}

			i := strings.LastIndex(absPath, v.modulePath)
			if i == -1 {
				continue
			}

			toAdd[absPath[i:]] = astFile
			toDelete = append(toDelete, filePath)
		}
		for filePath, astFile := range toAdd {
			n.Files[filePath] = astFile
		}
		for _, filePath := range toDelete {
			delete(n.Files, filePath)
		}
		return v
	}
	return nil
}
