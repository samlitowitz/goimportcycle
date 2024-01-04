package ast

import (
	"go/ast"
	"go/token"
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
		if _, ok := builder.filesByAbsPath[node.AbsPath]; ok {
			// return custom error, duplicate file
		}
		builder.filesByAbsPath[node.AbsPath] = &internal.File{
			Package:  builder.packagesByDirName[node.DirName],
			FileName: filepath.Base(node.AbsPath),
			AbsPath:  node.AbsPath,
			Imports:  make(map[string]*internal.Import),
			Decls:    make(map[string]*internal.Decl),
		}
		builder.packagesByDirName[node.DirName].Files[node.AbsPath] = builder.filesByAbsPath[node.AbsPath]
		builder.curFile = builder.filesByAbsPath[node.AbsPath]

	case *ast.ImportSpec:
		if builder.curPkg == nil {
			// return custom error, undefined package
		}
		if builder.curFile == nil {
			// return custom error, undefined file
		}
		if _, ok := builder.curFile.Imports[node.Name.String()]; ok {
			// return custom error, duplicate import
		}
		builder.curFile.Imports[node.Name.String()] = &internal.Import{
			Package:         builder.curPkg,
			File:            builder.curFile,
			Name:            node.Name.String(),
			Path:            node.Path.Value,
			ReferencedTypes: make(map[string]*internal.Decl),
		}

	case *FuncDecl:
		if builder.curPkg == nil {
			// return custom error, undefined package
		}
		if builder.curFile == nil {
			// return custom error, undefined file
		}
		if node.Name.String() == "" {
			// return custom error, invalid function name
		}
		if _, ok := builder.curFile.Decls[node.QualifiedName]; ok {
			// return custom error, duplicate decl
		}
		var receiverDecl *internal.Decl
		for _, file := range builder.curPkg.Files {
			if _, ok := file.Decls[node.ReceiverName]; ok {
				receiverDecl = file.Decls[node.ReceiverName]
				break
			}
		}
		builder.curFile.Decls[node.QualifiedName] = &internal.Decl{
			File:         builder.curFile,
			ReceiverDecl: receiverDecl,
			Name:         node.Name.String(),
		}

	case *ast.GenDecl:
		if builder.curPkg == nil {
			// return custom error, undefined package
		}
		if builder.curFile == nil {
			// return custom error, undefined file
		}
		for _, spec := range node.Specs {
			switch spec := spec.(type) {
			case *ast.TypeSpec:
				if node.Tok != token.TYPE {
					// return custom error, invalid declaration
				}
				if _, ok := builder.curFile.Decls[spec.Name.String()]; ok {
					// return custom error, duplicate decl
				}
				if spec.Name.String() == "" {
					// return custom error, invalid type name
				}
				builder.curFile.Decls[spec.Name.String()] = &internal.Decl{
					File:         builder.curFile,
					ReceiverDecl: nil,
					Name:         spec.Name.String(),
				}

			case *ast.ValueSpec:
				if node.Tok != token.CONST && node.Tok != token.VAR {
					// return custom error, invalid declaration
				}
				for _, name := range spec.Names {
					if _, ok := builder.curFile.Decls[name.String()]; ok {
						// return custom error, duplicate decl
					}
					if name.String() == "" {
						// return custom error, invalid const/var name
					}
					builder.curFile.Decls[name.String()] = &internal.Decl{
						File:         builder.curFile,
						ReceiverDecl: nil,
						Name:         name.String(),
					}
				}

			default:
				// return custom error, unhandled spec type
			}
		}
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
