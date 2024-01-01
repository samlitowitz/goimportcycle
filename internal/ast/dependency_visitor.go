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

func (v DependencyVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.Package:
		var setImportPathAndEmitPackage bool
		for filename, astFile := range node.Files {
			absPath, err := filepath.Abs(filename)
			if err != nil {
				continue
			}
			if !setImportPathAndEmitPackage {
				dirName, _ := filepath.Split(absPath)
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
			}
		}
	case *ast.ImportSpec:
		v.out <- node

	case *ast.GenDecl:
		switch node.Tok {
		case token.CONST:
			fallthrough
		case token.TYPE:
			fallthrough
		case token.VAR:
			v.out <- node
		}

	case *ast.FuncDecl:
		v.out <- node

	case *ast.SelectorExpr:
		v.out <- node
	}
	return v
}

func (v *DependencyVisitor) Close() {
	close(v.out)
}
