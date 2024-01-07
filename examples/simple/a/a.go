package a

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/simple/b"
)

func Fn() {
	log.Println("A")
	b.Fn()
}
