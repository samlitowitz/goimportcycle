package b

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/simple/a"
)

func Fn() {
	a.Fn()
	log.Println("B")
}
