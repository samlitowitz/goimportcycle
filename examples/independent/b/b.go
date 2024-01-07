package b

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/independent/a"
)

func Fn() {
	a.Fn()
	log.Println("B")
}
