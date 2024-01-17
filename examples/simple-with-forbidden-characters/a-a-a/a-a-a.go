package a_a_a

import (
	"log"

	b "github.com/samlitowitz/goimportcycle/examples/simple-with-forbidden-characters/b.b.b"
)

func Fn() {
	log.Println("A")
	b.Fn()
}
