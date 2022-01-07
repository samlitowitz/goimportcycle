package a

type ID string

const (
	IDOne ID = "one"
	IDTwo ID = "two"
	IDThree, IDFour ID = "three", "four"
)

var (
	V1 ID = "one"
	V2, V3 ID = "two", "three"
)
