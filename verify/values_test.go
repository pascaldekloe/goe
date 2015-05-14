package verify

import (
	"reflect"
	"strings"
	"testing"
)

type o struct {
	b bool
	i int
	u uint
	f float64
	c complex64
	a [2]int
	x interface{}
	m map[int]bool
	p *o
	s []byte
	t string
	o p
}

type p struct {
	Name string
}

var goldenIdentities = []interface{}{
	nil,
	false,
	true,
	"",
	"a",
	int(2),
	o{},
	&o{},
	o{
		b: true,
		i: -9,
		u: 9,
		f: 8,
		c: 7,
		a: [2]int{6, 5},
		x: "inner",
		m: map[int]bool{4: false, 3: true},
		p: &o{
			t: "inner",
			p: &o{},
		},
		s: []byte{2, 1},
		t: "text",
		o: p{Name: "test"},
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
	{true, false, "true != false"},
	{"a", "b", `"a" != "b"`},
	{1, 2, "1 != 2"},
	{1, true, "types differ, int != bool"},
	{o{x: o{}}, o{x: "o"},
		"/x: types differ, verify.o != string"},
	{&o{}, nil,
		"unwanted *verify.o"},
	{[]byte{1, 2}, []byte{1, 3},
		"[1]: 2 != 3"},
	{map[int]bool{1: false, 2: false, 3: true}, map[int]bool{1: false, 2: true, 3: true},
		"[2]: false != true"},
	{map[rune]bool{'f': false, 't': true}, map[rune]bool{'f': false},
		"[116]: unwanted bool"},
	{map[string]int{"f": 0, "t": 1}, map[string]int{"f": 0, "t": 1, "?": 2},
		`["?"]: missing int`},
	{o{i: 1}, o{},
		"/i: 1 != 0"},
	{o{x: func() int { return 9 }}, o{x: func() int { return 0 }},
		"/x: can't compare functions"},
	{
		o{
			p: &o{s: []byte{3, 4}},
			s: []byte{1, 2},
		},
		o{
			p: &o{s: []byte{5, 6}},
			s: []byte{0},
		},
		"/p/s[0]: 3 != 5\n/p/s[1]: 4 != 6\n/s: got 2 elements, want 1"},
	{o{t: "abcdefghijklmnoprstuvwxyz"}, o{t: "abcdefghijklmnopqrstuvwxyz"},
		`/t: "abcdefghijklmnoprstuvwxyz" != "abcdefghijklmnopqrstuvwxyz"`},
	{o{o: p{Name: "Jo"}}, o{o: p{Name: "Joe"}},
		`/o/Name: "Jo" != "Joe"`},
}

func TestGoldenDiffers(t *testing.T) {
	for _, gold := range goldenDiffers {
		tr := &travel{}
		tr.values(reflect.ValueOf(gold.a), reflect.ValueOf(gold.b), nil)

		msg := tr.report("case")
		if len(msg) == 0 {
			t.Errorf("No report for %#v and %#v", gold.a, gold.b)
			continue
		}

		if i := strings.IndexRune(msg, '\n'); i < 0 {
			t.Fatalf("Missing report header in: %s", msg)
		} else {
			msg = msg[i+1:]
		}

		if msg != gold.msg {
			t.Errorf("Got %q, want %q for %#v and %#v", msg, gold.msg, gold.a, gold.b)
		}
	}
}

func TestNilEquivalents(t *testing.T) {
	var s1, s2 []byte
	var m1, m2 map[rune]string
	s2 = make([]byte, 0, 1)
	m2 = make(map[rune]string, 1)

	Values(t, "nil slice", s1, s2)
	Values(t, "nil map", m1, m2)
	Values(t, "empty slice", s2, s1)
	Values(t, "empty map", m2, m1)
}
