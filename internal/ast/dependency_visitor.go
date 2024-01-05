package ast

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

type Package struct {
	*ast.Package

	DirName string
}

type File struct {
	*ast.File

	AbsPath string
	DirName string
}

type ImportSpec struct {
	*ast.ImportSpec

	IsAliased bool
	Alias     string
}

type FuncDecl struct {
	*ast.FuncDecl

	ReceiverName  string
	QualifiedName string
}

type SelectorExpr struct {
	*ast.SelectorExpr

	ImportName string
}

func (decl FuncDecl) IsReceiver() bool {
	return decl.Recv != nil
}

type DependencyVisitor struct {
	out chan<- ast.Node

	fileImports map[string]struct{}
}

func NewDependencyVisitor() (*DependencyVisitor, <-chan ast.Node) {
	out := make(chan ast.Node)
	v := &DependencyVisitor{
		out: out,
	}

	return v, out
}

func (v *DependencyVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.Package:
		v.emitPackageAndFiles(node)

	case *ast.File:
		v.fileImports = make(map[string]struct{})

	case *ast.ImportSpec:
		v.emitImportSpec(node)

	case *ast.FuncDecl:
		v.emitFuncDecl(node)

	case *ast.GenDecl:
		switch node.Tok {
		case token.CONST:
			fallthrough
		case token.TYPE:
			fallthrough
		case token.VAR:
			v.out <- node
		}

	case *ast.SelectorExpr:
		// only references to external packages
		if node.X == nil {
			return v
		}

		impName := ""
		switch x := node.X.(type) {
		case *ast.Ident:
			impName = x.String()
		}

		// if the "import name" is actually a variable and not a package, skip it
		if _, ok := v.fileImports[impName]; !ok {
			return v
		}

		v.out <- &SelectorExpr{
			SelectorExpr: node,
			ImportName:   impName,
		}
	}
	return v
}

func (v *DependencyVisitor) emitPackageAndFiles(node *ast.Package) {
	var setImportPathAndEmitPackage bool
	var dirName string
	for filename, astFile := range node.Files {
		absPath, err := filepath.Abs(filename)
		if err != nil {
			continue
		}
		if !setImportPathAndEmitPackage {
			dirName, _ = filepath.Split(absPath)
			dirName = strings.TrimRight(dirName, "/")
			v.out <- &Package{
				Package: node,
				DirName: dirName,
			}
			setImportPathAndEmitPackage = true
		}

		v.out <- &File{
			File:    astFile,
			AbsPath: absPath,
			DirName: dirName,
		}
	}
}

func (v *DependencyVisitor) emitImportSpec(node *ast.ImportSpec) {
	node.Path.Value = strings.Trim(node.Path.Value, "\"")
	pieces := strings.Split(node.Path.Value, "/")
	name := pieces[len(pieces)-1]

	isAliased := node.Name != nil
	alias := ""

	if isAliased {
		alias = node.Name.String()
		node.Name.Name = name
		v.fileImports[alias] = struct{}{}
	}

	if !isAliased {
		node.Name = &ast.Ident{
			Name: name,
		}
		v.fileImports[name] = struct{}{}
	}

	v.out <- &ImportSpec{
		ImportSpec: node,
		IsAliased:  isAliased,
		Alias:      alias,
	}
}

func (v *DependencyVisitor) emitFuncDecl(node *ast.FuncDecl) {
	receiverName := ""
	qualifiedName := node.Name.String()

	if node.Recv != nil {
		var typName string
		switch expr := node.Recv.List[0].Type.(type) {
		case *ast.Ident:
			typName = expr.String()
		case *ast.StarExpr:
			if expr.X == nil {
				// panic error, invalid receiver method
			}
			ident, ok := expr.X.(*ast.Ident)
			if !ok {
				// panic error, invalid receiver method
			}
			typName = ident.String()
		default:
			// panic error, invalid receiver method
		}
		receiverName = typName
		qualifiedName = typName + "." + node.Name.String()
	}

	v.out <- &FuncDecl{
		FuncDecl:      node,
		ReceiverName:  receiverName,
		QualifiedName: qualifiedName,
	}
}

func (v *DependencyVisitor) Close() {
	close(v.out)
}
