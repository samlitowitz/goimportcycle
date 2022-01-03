package main

import (
	"fmt"

	"github.com/samlitowitz/goimportcycle/examples/importcycle/a"
)

func main() {
	a1, a2 := a.NewA("1"), a.NewA("2")
	fmt.Printf("ID: %s, %s\n", a1.ID(), a2.ID())
}
