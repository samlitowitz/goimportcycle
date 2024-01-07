package b

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/transitive/a"
)

func Fn() {
	a.Fn()
	log.Println("B")
}
