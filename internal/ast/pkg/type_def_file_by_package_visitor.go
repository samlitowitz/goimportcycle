package pkg

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

type FilePath string
type ImportPath string
type PackageName string
type TypeName string

type TypeDefFileByPackageVisitor struct {
	*PackageImportsByFileVisitor
	fileByTypebyPackage map[ImportPath]map[TypeName]FilePath
}

func NewTypeDefFileByPackageVisitor(modulePath string) *TypeDefFileByPackageVisitor {
	return &TypeDefFileByPackageVisitor{
		PackageImportsByFileVisitor: NewPackageImportsByFileVisitor(modulePath),
		fileByTypebyPackage:         make(map[ImportPath]map[TypeName]FilePath),
	}
}

func (v *TypeDefFileByPackageVisitor) FileByTypeByPackage() map[ImportPath]map[TypeName]FilePath {
	return v.fileByTypebyPackage
}

func (v *TypeDefFileByPackageVisitor) Visit(n ast.Node) ast.Visitor {
	next := v.PackageImportsByFileVisitor.Visit(n)

	switch n := n.(type) {
	case *ast.GenDecl:
		if n.Tok == token.TYPE {
			return v
		}

	case *ast.TypeSpec:
		pkg, _ := filepath.Split(string(v.curFilePath))
		pkg = strings.TrimRight(pkg, "/")
		if _, ok := v.fileByTypebyPackage[ImportPath(pkg)]; !ok {
			v.fileByTypebyPackage[ImportPath(pkg)] = make(map[TypeName]FilePath)
		}
		typeName := TypeName(n.Name.String())
		if _, ok := v.fileByTypebyPackage[ImportPath(pkg)][TypeName(typeName)]; !ok {
			v.fileByTypebyPackage[ImportPath(pkg)][TypeName(typeName)] = FilePath(v.curFilePath)
		}
	}
	if next != nil {
		return v
	}
	return nil
}
