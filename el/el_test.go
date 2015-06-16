package el

import (
	"reflect"
	"testing"

	"verify"
)

type strType string

// strptr returns a pointer to s.
// Go does not allow pointers to literals.
func strptr(s string) *string {
	return &s
}

type vals struct {
	b bool
	i int64
	u uint64
	f float64
	c complex128
	s string
}

type ptrs struct {
	bp *bool
	ip *int64
	up *uint64
	fp *float64
	cp *complex128
	sp *string
}

type node struct {
	Name  *string
	Child *node
	child *node
	X     interface{}
	a     [2]interface{}
	s     []interface{}
}

var testV = vals{
	b: true,
	i: -2,
	u: 4,
	f: 8,
	c: 16i,
	s: "32",
}

var testPV = ptrs{
	bp: &testV.b,
	ip: &testV.i,
	up: &testV.u,
	fp: &testV.f,
	cp: &testV.c,
	sp: &testV.s,
}

type goldenCase struct {
	expr string
	root interface{}
	want interface{}
}

var goldenPaths = []goldenCase{
	{"/b", testV, testV.b},
	{"/ip", &testPV, testV.i},
	{"/X/X/u", node{X: node{X: testV}}, testV.u},
	{"/X/../X/fp", node{X: &testPV}, testV.f},
	{"/X/./X/c", &node{X: &node{X: &testV}}, testV.c},
	{"/", &testPV.sp, testV.s},
	{"/.[0]", "hello", uint64('h')},
	{"/s/.[0]", &node{s: []interface{}{testV.i}}, testV.i},
	{"/a[1]", node{a: [2]interface{}{testV.f, testV.s}}, testV.s},
	{"/.[true]", map[bool]string{true: "y"}, "y"},
	{`/.["I \x2f O"]`, map[strType]float64{"I / O": 99.8}, 99.8},
	{"/.[1]/.[2]", map[int]map[uint]string{1: {2: "1.2"}}, "1.2"},
	{"/.[*]/.[*]", map[int]map[uint]string{3: {4: "3.4"}}, "3.4"},
}

var nilPointer *node

var goldenPathFails = []goldenCase{
	{"/Child/Name", nilPointer, nil},
	{"malformed", node{}, nil},
	{"/mis", node{}, nil},
	{"/.[broken]", [2]bool{}, nil},
	{"/.[yes]", map[bool]bool{}, nil},
	{"/X", node{X: testV}, nil},
	{"/.[3]", testV, nil},
	{"/s[4]", node{}, nil},
	{"/a[5]", node{}, nil},
	{"/.[6.66]", map[float64]bool{}, nil},
}

func TestPaths(t *testing.T) {
	for _, gold := range append(goldenPaths, goldenPathFails...) {
		testGoldenCase(t, reflect.ValueOf(Bool), gold)
		testGoldenCase(t, reflect.ValueOf(Int), gold)
		testGoldenCase(t, reflect.ValueOf(Uint), gold)
		testGoldenCase(t, reflect.ValueOf(Float), gold)
		testGoldenCase(t, reflect.ValueOf(Complex), gold)
		testGoldenCase(t, reflect.ValueOf(String), gold)
	}
}

func testGoldenCase(t *testing.T, f reflect.Value, gold goldenCase) {
	args := []reflect.Value{
		reflect.ValueOf(gold.expr),
		reflect.ValueOf(gold.root),
	}
	result := f.Call(args)

	typ := result[0].Type()
	wantMatch := gold.want != nil && typ == reflect.TypeOf(gold.want)

	if got := result[1].Bool(); got != wantMatch {
		t.Errorf("Got %s OK %t, want %t for %q", typ, got, wantMatch, gold.expr)
		return
	}

	if got := result[0].Interface(); wantMatch && got != gold.want {
		t.Errorf("Got %s %#v, want %#v for %q", typ, got, gold.want, gold.expr)
	}
}

func BenchmarkPaths(b *testing.B) {
	todo := b.N
	for {
		for _, g := range goldenPaths {
			String(g.expr, g.root)
			todo--
			if todo == 0 {
				return
			}
		}
	}
}

func TestWildCards(t *testing.T) {
	data := &node{
		a: [2]interface{}{99, 100},
		s: []interface{}{"a", "b", 3},
	}
	valueMix := []interface{}{testV.b, testV.i, testV.u, testV.f, testV.c, testV.s, testV}

	tests := []struct {
		got, want interface{}
	}{
		{Bools("/*", testV), []bool{testV.b}},
		{Ints("/*", testV), []int64{testV.i}},
		{Uints("/*", testV), []uint64{testV.u}},
		{Floats("/*", testV), []float64{testV.f}},
		{Complexes("/*", testV), []complex128{testV.c}},
		{Strings("/*", testV), []string{testV.s}},

		{Ints("/a[*]", data), []int64{99, 100}},
		{Strings("/*[*]", data), []string{"a", "b"}},

		{Any("/.[*]", valueMix), valueMix},
		{Any("/", valueMix), []interface{}{valueMix}},
		{Any("/MisMatch", valueMix), []interface{}(nil)},
	}

	for _, test := range tests {
		verify.Values(t, "wildcard match", test.got, test.want)
	}

}

type goldenHave struct {
	path  string
	root  interface{}
	value interface{}

	// updates is the wanted number of updates.
	updates int
	// result is the wanted content at path.
	result []string
}

func newGoldenHaves() []goldenHave {
	return []goldenHave{
		{"/", strptr("hello"), "hell", 1, []string{"hell"}},
		{"/.", strptr("hello"), "hell", 1, []string{"hell"}},
		{"/", strptr("hello"), strptr("poin"), 1, []string{"poin"}},

		{"/S", &struct{ S string }{}, "hell", 1, []string{"hell"}},
		{"/SC", &struct{ SC string }{}, strType("hell"), 1, []string{"hell"}},
		{"/CC", &struct{ CC strType }{}, strType("hell"), 1, []string{"hell"}},
		{"/CS", &struct{ CS strType }{}, "hell", 1, []string{"hell"}},

		{"/P", &struct{ P *string }{P: new(string)}, "poin", 1, []string{"poin"}},
		{"/PP", &struct{ PP **string }{PP: new(*string)}, "doub", 1, []string{"doub"}},
		{"/PPP", &struct{ PPP ***string }{PPP: new(**string)}, "trip", 1, []string{"trip"}},

		{"/X/S", &struct{ X *struct{ S string } }{}, "hell", 1, []string{"hell"}},
		{"/X/P", &struct{ X **struct{ P *string } }{}, "poin", 1, []string{"poin"}},
		{"/X/PP", &struct{ X **struct{ PP **string } }{}, "doub", 1, []string{"doub"}},

		{"/Child/Child/Child/Name", &node{}, "Grand Grand", 1, []string{"Grand Grand"}},

		{"/.[1]", &[3]*string{}, "up", 1, []string{"up"}},
		{"/.[2]", &[]string{"1", "2", "3"}, "up", 1, []string{"up"}},
		{"/.['p']", &map[byte]*string{}, "in", 1, []string{"in"}},
		{"/.['q']", &map[byte]*string{'q': strptr("orig")}, "up", 1, []string{"up"}},
		{"/.['r']", &map[byte]string{}, "in", 1, []string{"in"}},
		{"/.['s']", &map[byte]string{'s': "orig"}, "up", 1, []string{"up"}},
		{"/.[*]", &map[byte]*string{'x': strptr("orig"), 'y': nil}, "up", 2, []string{"up", "up"}},

		{"/.[11]/.[12]", &map[int32]map[int64]string{}, "11.12", 1, []string{"11.12"}},
		{"/.[13]/.[14]", &map[int8]**map[int16]string{}, "13.14", 1, []string{"13.14"}},
		{"/.['w']/X/Y", &map[byte]struct{X struct{Y ***string}}{}, "z", 1, []string{"z"}},
	}
}

func newGoldenHaveFails() []goldenHave {
	return []goldenHave{
		// No expression
		{"", strptr("hello"), "fail", 0, nil},

		// Nil root
		{"/", nil, "fail", 0, nil},

		// Nil value
		{"/", strptr("hello"), nil, 0, []string{"hello"}},

		// Not addresable
		{"/", "hello", "fail", 0, []string{"hello"}},

		// Too abstract
		{"/X/anyField", &node{}, "fail", 0, nil},

		// Initialize with zero value on type mismatch
		{"/WrongType", &struct{ WrongType *string }{}, 9.98, 0, []string{""}},

		// String modification
		{"/.[6]", strptr("immutable"), '-', 0, nil},

		// Out of bounds
		{"/.[7]", &[]*string{}, "fail", 0, nil},
		{"/.[8]", &[2]string{}, "fail", 0, nil},

		// Non-exported
		{`/s`, &struct{ s *string }{}, "can't use", 0, nil},
		{`/child/Name`, &node{}, "can't use", 0, nil},
		{`/a[1]`, &struct{ a [2]string }{}, "can't use", 0, []string{""}},
		{`/m[3]`, &struct{ m map[int]string }{m: map[int]string{3: "three"}}, "can't use", 0, []string{"three"}},
		{`/a[*]`, &struct{ a [2]string }{}, "can't use", 0, []string{"", ""}},
		{`/m[*]`, &struct{ m map[int]string }{m: map[int]string{1: "four"}}, "can't use", 0, []string{"four"}},
		{`/m[4]`, &struct{ m map[int]string }{}, "can't use", 0, nil},
	}
}

func TestHaves(t *testing.T) {
	for _, gold := range append(newGoldenHaves(), newGoldenHaveFails()...) {
		n := Have(gold.root, gold.path, gold.value)
		if n != gold.updates {
			t.Errorf("Got n=%d, want %d for %s", n, gold.updates, gold.path)
		}

		got := Strings(gold.path, gold.root)
		verify.Values(t, gold.path, got, gold.result)
	}
}

func BenchmarkHaves(b *testing.B) {
	b.StopTimer()
	todo := b.N
	for {
		cases := newGoldenHaves()
		b.StartTimer()
		for _, g := range cases {
			Have(g.root, g.path, g.value)
			todo--
			if todo == 0 {
				return
			}
		}
		b.StopTimer()
	}
}
