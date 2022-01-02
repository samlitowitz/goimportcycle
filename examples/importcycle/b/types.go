package b

import "github.com/samlitowitz/goimportcycleviz/examples/importcycle/a"

type IDer interface {
	ID() a.ID
}
