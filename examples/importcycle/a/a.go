package a

import (
	"github.com/samlitowitz/goimportcycleviz/examples/importcycle/b"
	c "github.com/samlitowitz/goimportcycleviz/examples/importcycle/b"
)

type A struct {
	id ID
}

func NewA(id ID) *A {
	return &A{
		id: id,
	}
}

func (a *A) ID() ID {
	return a.id
}

func AAsIDer(a *A) b.IDer {
	return b.IDer(a)
}

func IsIDer(i c.IDer) bool {
	return true
}
