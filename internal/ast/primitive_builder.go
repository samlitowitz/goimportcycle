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
		builder.packagesByDirName[node.DirName] = buildPackage(
			builder.moduleRootDirectory,
			node.DirName,
			node.Name,
			len(node.Files),
		)

		//case *File:
		//
		//	file := &internal.File{
		//		Package:      nil,
		//		FileName:     "",
		//		AbsPath:      "",
		//		Imports:      nil,
		//		TypesDefined: nil,
		//	}
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

func buildPackage(
	moduleRootDir,
	dirName, name string,
	fileCount int,
) *internal.Package {
	pkg := &internal.Package{
		DirName:    dirName,
		ImportPath: "",
		Name:       name,
		Files:      make(map[string]*internal.File, fileCount),
	}
	if pkg.Name != "main" {
		pkg.ImportPath = strings.TrimPrefix(
			pkg.DirName,
			moduleRootDir+string(filepath.Separator),
		)
	}
	return pkg
}
