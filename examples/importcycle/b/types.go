package b

import (
	"github.com/samlitowitz/goimportcycle/examples/importcycle/a"
	"github.com/samlitowitz/goimportcycle/examples/importcycle/c"
	"github.com/samlitowitz/goimportcycle/examples/importcycle/d"
	"github.com/samlitowitz/goimportcycle/examples/importcycle/f"
)

type IDer interface {
	ID() a.ID
	C() c.ID
	D() d.ID
	F() f.ID
}
