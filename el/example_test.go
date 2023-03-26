package el_test

import (
	"fmt"
	"image/gif"
	"strings"

	"github.com/pascaldekloe/goe/el"
)

func ExampleAssign_pathAllocation() {
	type node struct {
		Child *node
		Label string
	}

	x := new(node)
	el.Assign(x, "/Child/Child/Label", "Hello")

	fmt.Printf("%v", x.Child.Child)
	// Output: &{<nil> Hello}
}

func ExampleAssign_typeFlexibility() {
	type numbers struct {
		B byte
		I uint32
		F float64
	}

	x := new(numbers)
	el.Assign(x, "/*", 42)

	fmt.Printf("%+v", x)
	// Output: &{B:42 I:42 F:42}
}

func ExampleUints() {
	data := "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00\x2c\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3b"
	img, err := gif.Decode(strings.NewReader(data))
	if err != nil {
		panic(err)
	}

	fmt.Printf("RGBA: %v", el.Uints("/Palette[0]/*", img))
	// Output: RGBA: [255 255 255 255]
}
