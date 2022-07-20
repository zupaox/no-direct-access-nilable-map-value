package test

import (
	"fmt"
)

func main() {
	handlers := map[string]func() bool{
		"aaa": func() bool {
			fmt.Printf("aaa")
			return true
		},
		"bbb": func() bool {
			fmt.Printf("bbb")
			return true
		},
	}

	// this is bad
	if handlers["ccc"]() {
		fmt.Printf("ccc")
	}

	aaa := ABC{}
	pointers := map[string]*ABC{
		"aaa": &aaa,
	}

	// this is bad
	fmt.Printf("value: %v", pointers["bbb"].BBB)
	// this is bad
	pointers["bbb"].CCC()

	// this is good
	assignment, ok := pointers["bbb"]
	if ok {
		assignment.CCC()
	}

	// this is bad
	p, _ := pointers["DDD"]
	p.CCC()

	// this is bad
	p = pointers["bbb"]
	p.CCC()

	structs := map[string]ABC{}
	// this is ok
	s, _ := structs["bbb"]
	(&s).CCC()

	// this is ok
	s = structs["ccc"]
	(&s).CCC()
}

type ABC struct {
	BBB string
}

func (a *ABC) CCC() {
	fmt.Printf("CCC")
}
