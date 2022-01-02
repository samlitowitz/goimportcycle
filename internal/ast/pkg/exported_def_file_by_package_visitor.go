package pkg

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
	"unicode"
)

type FilePath string
type ImportPath string
type PackageName string
type ExportName string

type ExportedDefFileByPackageVisitor struct {
	*PackageImportsByFileVisitor
	fileByExportByPackage map[ImportPath]map[ExportName]FilePath
}

func NewExportedDefFileByPackageVisitor(modulePath string) *ExportedDefFileByPackageVisitor {
	return &ExportedDefFileByPackageVisitor{
		PackageImportsByFileVisitor: NewPackageImportsByFileVisitor(modulePath),
		fileByExportByPackage:       make(map[ImportPath]map[ExportName]FilePath),
	}
}

func (v *ExportedDefFileByPackageVisitor) FileByExportByPackage() map[ImportPath]map[ExportName]FilePath {
	return v.fileByExportByPackage
}

func (v *ExportedDefFileByPackageVisitor) Visit(n ast.Node) ast.Visitor {
	next := v.PackageImportsByFileVisitor.Visit(n)

	switch n := n.(type) {
	case *ast.GenDecl:
		if n.Tok == token.TYPE {
			return v
		}

	case *ast.FuncDecl:
		// functions with receivers are already caught under *ast.TypeSpec
		if n.Recv != nil {
			break
		}
		if !unicode.IsUpper([]rune(n.Name.String())[0]) {
			break
		}
		pkg, _ := filepath.Split(string(v.curFilePath))
		pkg = strings.TrimRight(pkg, "/")
		if _, ok := v.fileByExportByPackage[ImportPath(pkg)]; !ok {
			v.fileByExportByPackage[ImportPath(pkg)] = make(map[ExportName]FilePath)
		}
		funcName := ExportName(n.Name.String())
		if _, ok := v.fileByExportByPackage[ImportPath(pkg)][ExportName(funcName)]; !ok {
			v.fileByExportByPackage[ImportPath(pkg)][ExportName(funcName)] = FilePath(v.curFilePath)
		}

	case *ast.TypeSpec:
		pkg, _ := filepath.Split(string(v.curFilePath))
		pkg = strings.TrimRight(pkg, "/")
		if _, ok := v.fileByExportByPackage[ImportPath(pkg)]; !ok {
			v.fileByExportByPackage[ImportPath(pkg)] = make(map[ExportName]FilePath)
		}
		typeName := ExportName(n.Name.String())
		if _, ok := v.fileByExportByPackage[ImportPath(pkg)][ExportName(typeName)]; !ok {
			v.fileByExportByPackage[ImportPath(pkg)][ExportName(typeName)] = FilePath(v.curFilePath)
		}
	}
	if next != nil {
		return v
	}
	return nil
}
