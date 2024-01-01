package ast

import (
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/samlitowitz/goimportcycle/internal"
)

type PrimitiveBuilder struct {
	modulePath          string
	moduleRootDirectory string

	packagesByDirName map[string]*internal.Package
}

func NewPrimitiveBuilder(modulePath, moduleRootDirectory string) *PrimitiveBuilder {
	return &PrimitiveBuilder{
		modulePath:          modulePath,
		moduleRootDirectory: moduleRootDirectory,

		packagesByDirName: make(map[string]*internal.Package),
	}
}

func (builder *PrimitiveBuilder) AddNode(node ast.Node) error {
	switch node := node.(type) {
	case *Package:
		pkg := &internal.Package{
			DirName:    node.DirName,
			ImportPath: "",
			Name:       node.Name,
			Files:      make(map[string]*internal.File, len(node.Files)),
		}
		if pkg.Name != "main" {
			pkg.ImportPath = strings.TrimPrefix(
				node.DirName,
				builder.moduleRootDirectory+string(filepath.Separator),
			)
		}
		builder.packagesByDirName[node.DirName] = pkg
	}

	return nil
}

func (builder PrimitiveBuilder) Packages() []*internal.Package {
	pkgs := make([]*internal.Package, 0, len(builder.packagesByDirName))
	for _, pkg := range builder.packagesByDirName {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}
