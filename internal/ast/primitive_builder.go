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
	filesByAbsPath    map[string]*internal.File

	curPkg  *internal.Package
	curFile *internal.File
}

func NewPrimitiveBuilder(modulePath, moduleRootDirectory string) *PrimitiveBuilder {
	return &PrimitiveBuilder{
		modulePath:          modulePath,
		moduleRootDirectory: moduleRootDirectory,

		packagesByDirName: make(map[string]*internal.Package),
		filesByAbsPath:    make(map[string]*internal.File),
	}
}

func (builder *PrimitiveBuilder) AddNode(node ast.Node) error {
	switch node := node.(type) {
	case *Package:
		if _, ok := builder.packagesByDirName[node.DirName]; !ok {
			builder.packagesByDirName[node.DirName] = buildPackage(
				builder.moduleRootDirectory,
				node.DirName,
				node.Name,
				len(node.Files),
			)
		}
		builder.curPkg = builder.packagesByDirName[node.DirName]

	case *File:
		if builder.curPkg == nil {
			// return custom error, undefined package
		}
		if _, ok := builder.filesByAbsPath[node.AbsPath]; !ok {
			builder.filesByAbsPath[node.AbsPath] = &internal.File{
				Package:      builder.packagesByDirName[node.DirName],
				FileName:     filepath.Base(node.AbsPath),
				AbsPath:      node.AbsPath,
				Imports:      make(map[string]*internal.Import),
				TypesDefined: make(map[string]*internal.Type),
			}
		}
		builder.packagesByDirName[node.DirName].Files[node.AbsPath] = builder.filesByAbsPath[node.AbsPath]
		builder.curFile = builder.filesByAbsPath[node.AbsPath]
	}

	return nil
}

func (builder *PrimitiveBuilder) Files() []*internal.File {
	files := make([]*internal.File, 0, len(builder.filesByAbsPath))
	for _, file := range builder.filesByAbsPath {
		files = append(files, file)
	}
	return files
}

func (builder *PrimitiveBuilder) Packages() []*internal.Package {
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
