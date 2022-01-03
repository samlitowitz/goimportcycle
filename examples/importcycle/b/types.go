package b

import "github.com/samlitowitz/goimportcycle/examples/importcycle/a"

type IDer interface {
	ID() a.ID
}
