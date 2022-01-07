package c

import (
	"github.com/samlitowitz/goimportcycle/examples/importcycle/d"
	"github.com/samlitowitz/goimportcycle/examples/importcycle/e"
)

type ID string

type C struct {
	d d.ID
	e e.ID
}

func Fn1() {
	panic("not implemented")
}
