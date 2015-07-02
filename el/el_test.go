package el

import (
	"reflect"
	"testing"

	"github.com/pascaldekloe/goe/verify"
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

var goldenPathFails = []goldenCase{
	{"/Name", (*node)(nil), nil},
	{"/Child", node{}, nil},
	{"/Child/Name", node{}, nil},
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

func BenchmarkLookups(b *testing.B) {
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

		{Any("/*", ptrs{}), []interface{}(nil)},
		{Bools("/*", ptrs{}), []bool(nil)},
		{Ints("/*", ptrs{}), []int64(nil)},
		{Uints("/*", ptrs{}), []uint64(nil)},
		{Floats("/*", ptrs{}), []float64(nil)},
		{Complexes("/*", ptrs{}), []complex128(nil)},
		{Strings("/*", ptrs{}), []string(nil)},

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

		{"/I", &struct{ I interface{} }{}, "in", 1, []string{"in"}},
		{"/U", &struct{ U interface{} }{U: true}, "up", 1, []string{"up"}},

		{"/X/S", &struct{ X *struct{ S string } }{}, "hell", 1, []string{"hell"}},
		{"/X/P", &struct{ X **struct{ P *string } }{}, "poin", 1, []string{"poin"}},
		{"/X/PP", &struct{ X **struct{ PP **string } }{}, "doub", 1, []string{"doub"}},

		{"/Child/Child/Child/Name", &node{}, "Grand Grand", 1, []string{"Grand Grand"}},

		{"/.[1]", &[3]*string{}, "up", 1, []string{"up"}},
		{"/.[2]", &[]string{"1", "2", "3"}, "up", 1, []string{"up"}},
		{"/.[3]", &[]*string{}, "in", 1, []string{"in"}},
		{"/.['p']", &map[byte]*string{}, "in", 1, []string{"in"}},
		{"/.['q']", &map[int16]*string{'q': strptr("orig")}, "up", 1, []string{"up"}},
		{"/.['r']", &map[uint]string{}, "in", 1, []string{"in"}},
		{"/.['s']", &map[int64]string{'s': "orig"}, "up", 1, []string{"up"}},
		{"/.[*]", &map[byte]*string{'x': strptr("orig"), 'y': nil}, "up", 2, []string{"up", "up"}},

		{"/.[11]/.[12]", &map[int32]map[int64]string{}, "11.12", 1, []string{"11.12"}},
		{"/.[13]/.[14]", &map[int8]**map[int16]string{}, "13.14", 1, []string{"13.14"}},
		{"/.['w']/X/Y", &map[byte]struct{ X struct{ Y ***string } }{}, "z", 1, []string{"z"}},
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

		// Wrong type
		{"/Sp", &struct{ Sp *string }{}, 9.98, 0, []string{""}},

		// String modification
		{"/.[6]", strptr("immutable"), '-', 0, nil},

		// Out of bounds
		{"/.[8]", &[2]string{}, "fail", 0, nil},

		// Malformed map keys
		{"/Sk[''']", &struct{ Sk map[string]string }{}, "fail", 0, nil},
		{"/Ik[''']", &struct{ Ik map[int]string }{}, "fail", 0, nil},
		{"/Ik[z]", &struct{ Ik map[int]string }{}, "fail", 0, nil},
		{"/Uk[''']", &struct{ Uk map[uint]string }{}, "fail", 0, nil},
		{"/Uk[z]", &struct{ Uk map[uint]string }{}, "fail", 0, nil},
		{"/Fk[z]", &struct{ Fk map[float32]string }{}, "fail", 0, nil},
		{"/Ck[z]", &struct{ Ck map[complex128]string }{}, "fail", 0, nil},
		{"/Ck[]", &struct{ Ck map[complex128]string }{}, "fail", 0, nil},

		// Non-exported
		{`/child/Name`, &node{}, "fail", 0, nil},
		{`/ns`, &struct{ ns *string }{}, "fail", 0, nil},

		// Non-exported array
		{`/na[0]`, &struct{ na [2]string }{}, "fail", 0, []string{""}},
		{`/na[1]`, &struct{ na [2]*string }{}, "fail", 0, nil},
		{`/na[*]`, &struct{ na [2]string }{}, "fail", 0, []string{"", ""}},

		// Non-exported slice
		{`/ns[0]`, &struct{ ns []string }{}, "fail", 0, nil},
		{`/ns[1]`, &struct{ ns []*string }{ns: []*string{nil, strptr("b")}}, "fail", 0, []string{"b"}},
		{`/ns[*]`, &struct{ ns []string }{ns: []string{"a"}}, "fail", 0, []string{"a"}},

		// Non-exported map
		{`/nm[0]`, &struct{ nm map[int]string }{}, "fail", 0, nil},
		{`/nm[1]`, &struct{ nm map[int]*string }{nm: map[int]*string{1: strptr("b")}}, "fail", 0, []string{"b"}},
		{`/nm[*]`, &struct{ nm map[int]string }{nm: map[int]string{2: "c"}}, "fail", 0, []string{"c"}},
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

func BenchmarkModifies(b *testing.B) {
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
