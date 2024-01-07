package c

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/transitive/b"
)

func Fn() {
	b.Fn()
	log.Println("C")
}
