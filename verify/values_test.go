package verify

import (
	"math"
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
	m map[int]interface{}
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
	'b',
	byte(1),
	2,
	4.8,
	math.NaN(),
	16i,
	o{},
	&o{},
	o{
		b: true,
		i: -9,
		u: 9,
		f: .8,
		c: 7i,
		a: [2]int{6, 5},
		x: "inner",
		m: map[int]interface{}{4: 'q', 3: nil},
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
	{true, false, "Got true, want false"},
	{"a", "b", `Got "a", want "b"`},
	{o{u: 10}, o{u: 11}, "/u: Got 10 (0xa), want 11 (0xb)"},

	{int(2), uint(2), "Got type int, want uint"},
	{o{x: o{}}, o{x: &o{}},
		"/x: Got type verify.o, want *verify.o"},

	{&o{}, nil, "Unwanted *verify.o"},
	{nil, &o{}, "Missing *verify.o"},
	{o{}, o{x: o{}}, "/x: Missing verify.o"},
	{o{x: o{}}, o{}, "/x: Unwanted verify.o"},

	{[]int{1, -2}, []int{1, -3},
		"[1]: Got -2, want -3"},
	{map[int]bool{0: false, 1: false, 2: true}, map[int]bool{0: false, 1: true, 2: true},
		"[1]: Got false, want true"},
	{map[rune]bool{'f': false}, map[rune]bool{'f': false, 't': true},
		"[116]: Missing bool"},
	{map[string]int{"false": 0, "true": 1, "?": 2}, map[string]int{"false": 0, "true": 1},
		`["?"]: Unwanted int`},

	{o{x: func() int { return 9 }}, o{x: func() int { return 0 }},
		"/x: Can't compare functions"},
	{
		o{
			p: &o{s: []byte{3, 10}},
		},
		o{
			p: &o{s: []byte{5, 11}},
		},
		"/p/s[0]: Got 3, want 5\n/p/s[1]: Got 10 (0xa), want 11 (0xb)"},

	{o{t: "abcdefghijklmnoprstuvwxyz"}, o{t: "abcdefghijklmnopqrstuvwxyz"},
		"/t: Got \"abcdefghijklmnoprstuvwxyz\", want \"abcdefghijklmnopqrstuvwxyz\"\n                         ^"},
	{o{o: p{Name: "Jo"}}, o{o: p{Name: "Joe"}},
		`/o/Name: Got "Jo", want "Joe"`},
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
	var sn, se []byte
	var mn, me map[int]string
	se = make([]byte, 0, 1)
	me = make(map[int]string, 1)

	type containers struct {
		s []byte
		m map[int]string
	}

	Values(t, "nil vs empty slice", sn, se)
	Values(t, "nil vs empty map", mn, me)
	Values(t, "empty vs nil slice", se, sn)
	Values(t, "empty vs nil map", me, mn)
	Values(t, "nil vs empty embedded", containers{}, containers{s: se, m: me})
	Values(t, "empty vs nil embedded", containers{s: se, m: me}, containers{})
}
