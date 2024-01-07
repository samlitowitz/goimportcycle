package a

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/independent/b"
	"github.com/samlitowitz/goimportcycle/examples/independent/c"
)

func Fn() {
	log.Println("A")
	b.Fn()
	c.Fn()
}
