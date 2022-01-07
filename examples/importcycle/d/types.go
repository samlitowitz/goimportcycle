package d

import "github.com/samlitowitz/goimportcycle/examples/importcycle/b"

type ID string

type D struct {
	id b.IDer
}
