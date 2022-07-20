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

	// this is ok
	pointers["ccc"] = &ABC{}

	// this is bad
	pointers["ccc"] = pointers["aaa"].Clone()

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

	// this is bad since it direct access to something possibily nil
	arrays := []*ABC{pointers["aaa"]}

	// slice index checking is not in this scope, so this is considered ok
	arrays[0].CCC()

	// this is bad as well
	(&s).FFF["ddd"].CCC()

	// this is bad as well
	(&s).Map()["ddd"].CCC()
}

type ABC struct {
	BBB string
	FFF map[string]*ABC
}

func (a *ABC) CCC() {
	fmt.Printf("CCC")
}

func (a *ABC) Clone() *ABC {
	return &ABC{}
}

func (a *ABC) Map() map[string]*ABC {
	return map[string]*ABC{
		"AAA": a.Clone(),
	}
}
