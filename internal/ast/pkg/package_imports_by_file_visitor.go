package pkg

import (
	"go/ast"
	"go/token"
	"strings"
)

type Package struct {
	Name       PackageName
	Path       ImportPath
	ImportedBy []FilePath
}

type PackageImportsByFileVisitor struct {
	*ModuleQualifiedFilePathsVisitor
	curFilePath       FilePath
	filePathByAstFile map[*ast.File]FilePath
	importsByFilePath map[FilePath][]*Package
}

func NewPackageImportsByFileVisitor(modulePath string) *PackageImportsByFileVisitor {
	return &PackageImportsByFileVisitor{
		ModuleQualifiedFilePathsVisitor: NewModuleQualifiedFilePathsVisitor(modulePath),
		filePathByAstFile:               make(map[*ast.File]FilePath),
		importsByFilePath:               make(map[FilePath][]*Package),
	}
}

func (v *PackageImportsByFileVisitor) PackageImportsByFile() map[FilePath][]*Package {
	return v.importsByFilePath
}

func (v *PackageImportsByFileVisitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.Package:
		v.ModuleQualifiedFilePathsVisitor.Visit(n)
		v.filePathByAstFile = make(map[*ast.File]FilePath)
		for path, astFile := range n.Files {
			v.filePathByAstFile[astFile] = FilePath(path)
		}
		return v
	case *ast.File:
		path, ok := v.filePathByAstFile[n]
		if !ok {
			return nil
		}
		v.curFilePath = path
		for _, astImportSpec := range n.Imports {
			// this should probably never happen?
			if astImportSpec.Path.Kind != token.STRING {
				continue
			}
			pkg := &Package{
				Path:       ImportPath(strings.Trim(astImportSpec.Path.Value, "\"")),
				ImportedBy: []FilePath{path},
			}
			pkg.Name = PackageName(extractPackageNameFromImportPath(string(pkg.Path)))
			if astImportSpec.Name != nil {
				pkg.Name = PackageName(astImportSpec.Name.String())
			}
			if _, ok := v.importsByFilePath[path]; !ok {
				v.importsByFilePath[path] = make([]*Package, 0)
			}
			v.importsByFilePath[path] = append(v.importsByFilePath[path], pkg)
		}
		return v
	}

	return nil
}

func extractPackageNameFromImportPath(path string) string {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return path
	}
	return path[i+1:]
}
