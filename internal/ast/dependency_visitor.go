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

type FuncDecl struct {
	*ast.FuncDecl

	ReceiverName  string
	QualifiedName string
}

func (decl FuncDecl) IsReceiver() bool {
	return decl.Recv != nil
}

type DependencyVisitor struct {
	out chan<- ast.Node
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
	case *ast.ImportSpec:
		node.Path.Value = strings.Trim(node.Path.Value, "\"")
		if node.Name == nil {
			pieces := strings.Split(node.Path.Value, "/")
			node.Name = &ast.Ident{
				Name: pieces[len(pieces)-1],
			}
		}

		v.out <- node

	case *ast.FuncDecl:
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
		v.out <- node
	}
	return v
}

func (v *DependencyVisitor) Close() {
	close(v.out)
}
