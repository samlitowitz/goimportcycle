package c

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/none/b"
)

func Fn() {
	b.Fn()
	log.Println("C")
}
