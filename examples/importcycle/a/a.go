package a

import (
	"github.com/samlitowitz/goimportcycle/examples/importcycle/b"
	c "github.com/samlitowitz/goimportcycle/examples/importcycle/b"
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

func IsIDer(i c.IDer) bool {
	return true
}

func IsIDerToo(i b.IDer) bool {
	return true
}
