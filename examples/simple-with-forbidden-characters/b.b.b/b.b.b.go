package b

import (
	"log"

	a "github.com/samlitowitz/goimportcycle/examples/simple-with-forbidden-characters/a_a_a"
)

func Fn() {
	a.Fn()
	log.Println("B")
}
