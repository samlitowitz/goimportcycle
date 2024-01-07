package a

import (
	"log"

	"github.com/samlitowitz/goimportcycle/examples/interlinked/b"
)

func Fn() {
	log.Println("A")
	b.Fn()
}
