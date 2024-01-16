package b

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/simple-with-forbidden-characters/a_a_a"
)

func Fn() {
	a_a_a.Fn()
	log.Println("B")
}
