package ast

import (
	"go/ast"
	"go/token"
)

type File struct {
	*ast.File

	Filename string
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
		v.out <- node

		for filename, astFile := range node.Files {
			v.out <- &File{
				File:     astFile,
				Filename: filename,
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
	}
	return v
}

func (v *DependencyVisitor) Close() {
	close(v.out)
}
