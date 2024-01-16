package a_a_a

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/simple-with-forbidden-characters/b"
)

func Fn() {
	log.Println("A")
	b.Fn()
}
