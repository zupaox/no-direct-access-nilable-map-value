package test

import "fmt"

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

	if handlers["ccc"]() {
		fmt.Printf("ccc")
	}

	aaa := ABC{}
	pointers := map[string]*ABC{
		"aaa": &aaa,
	}
	fmt.Printf("value: %v", pointers["bbb"].BBB)
	pointers["bbb"].CCC()

	assignment, ok := pointers["bbb"]
	if ok {
		assignment.CCC()
	}

	structs := map[string]ABC{}
	s, ok := structs["bbb"]
	if ok {
		(&s).CCC()
	}
}

type ABC struct {
	BBB string
}

func (a *ABC) CCC() {
	fmt.Printf("CCC")
}
