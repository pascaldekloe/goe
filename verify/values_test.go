package verify

import (
	"reflect"
	"testing"
)

type o struct {
	i int
	a [2]int
	s []int
	m map[int]bool
}

type p struct {
	x interface{}
	o *o
}

var goldenIdentities = []interface{}{
	nil,
	false,
	true,
	"",
	"a",
	int(2),
	o{},
	p{},
	&p{},
	p{
		x: "",
		o: &o{
			i: 8,
			a: [...]int{1, 2},
			s: []int{3},
			m: map[int]bool{8: true},
		},
	},
}

func TestGoldenIdentities(t *testing.T) {
	for _, x := range goldenIdentities {
		if !Values(t, "same", x, x) {
			t.Errorf("Rejected %#v", x)
		}
	}
}

var goldenDiffers = []struct {
	a, b interface{}
	msg  string
}{
	{true, false,
		"\ntrue != false"},
	{1, 2,
		"\n1 != 2"},
	{&o{}, nil,
		"\nunwanted <*verify.o Value>"},
	{[]byte{1, 2}, []byte{1, 4},
		"\n[1]: 2 != 4"},
	{map[int]bool{1: false, 2: true, 3: false}, map[int]bool{2: false, 3: false, 4: false},
		"\n[1]: not wanted\n[2]: true != false\n[4]: not available"},
	{o{i: 1}, o{},
		"\n/i: 1 != 0"},
}

func TestGoldenDiffers(t *testing.T) {
	for _, gold := range goldenDiffers {
		s := state{}
		s.values(reflect.ValueOf(gold.a), reflect.ValueOf(gold.b), "")
		if s.String() != gold.msg {
			t.Errorf("Got %q, want %q for %#v and %#v", s.String(), gold.msg, gold.a, gold.b)
		}
	}
}
