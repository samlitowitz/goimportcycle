package b

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/none/a"
)

func Fn() {
	a.Fn()
	log.Println("B")
}
