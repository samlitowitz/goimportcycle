package a

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/transitive/c"
)

func Fn() {
	log.Println("A")
	c.Fn()
}
