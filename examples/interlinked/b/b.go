package b

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/interlinked/a"
	"github.com/samlitowitz/goimportcycle/examples/interlinked/c"
)

func Fn() {
	a.Fn()
	log.Println("B")
	c.Fn()
}
