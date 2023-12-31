package ast

import "go/ast"

type tokenType uint

const (
	TokenError tokenType = iota

	TokenPackage tokenType = iota
	TokenFile    tokenType = iota
)

type Token struct {
	typ tokenType
	val string
}

func (t Token) Type() uint {
	return uint(t.typ)
}

func (t Token) Val() string {
	return t.val
}

func NewToken(typ tokenType, val string) *Token {
	return &Token{
		typ: typ,
		val: val,
	}
}

type DependencyVisitor struct {
	out chan<- Token
}

func NewDependencyVisitor() (*DependencyVisitor, <-chan Token) {
	out := make(chan Token)
	v := &DependencyVisitor{
		out: out,
	}

	return v, out
}

func (v DependencyVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.Package:
		v.out <- Token{
			typ: TokenPackage,
			val: node.Name,
		}
	}
	return v
}

func (v *DependencyVisitor) Close() {
	close(v.out)
}
