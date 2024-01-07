package c

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/interlinked/b"
)

func Fn() {
	b.Fn()
	log.Println("C")
}
