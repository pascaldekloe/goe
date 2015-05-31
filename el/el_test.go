package el

import (
	"reflect"
	"testing"
)

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
	sub interface{}
	a   [2]interface{}
	s   []interface{}
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
	data interface{}
	want interface{}
}

var golden = []goldenCase{
	{"/b", testV, testV.b},
	{"/ip", &testPV, testV.i},
	{"/sub/sub/u", node{sub: node{sub: testV}}, testV.u},
	{"/sub/../sub/fp", node{sub: &testPV}, testV.f},
	{"/sub/./sub/c", &node{sub: &node{sub: &testV}}, testV.c},
	{"sub/sp", &node{sub: &testPV}, testV.s},
	{"/", &testPV.sp, testV.s},
	{".", "hello", "hello"},
	{"[1]", "hello", uint64('e')},
	{"/s/[0]", &node{s: []interface{}{testV.i}}, testV.i},
	{"/a[1]", node{a: [2]interface{}{testV.f, testV.s}}, testV.s},
	{"/field", testV.s, nil},
	{"/mis", node{}, nil},
	{"/sub", node{sub: testV}, nil},
	{"/[1]", testV, nil},
	{"/s[0]", node{}, nil},
	{"/a[4]", node{}, nil},
}

func TestGolden(t *testing.T) {
	for _, gold := range golden {
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
		reflect.ValueOf(gold.data),
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

func TestWildCards(t *testing.T) {
	data := &node{
		a: [2]interface{}{99, 100},
		s: []interface{}{"a", "b", 3},
	}

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
	}
	for _, test := range tests {
		if !reflect.DeepEqual(test.got, test.want) {
			t.Errorf("Got %#v, want %#v", test.got, test.want)
		}
	}
}

func BenchmarkGoldenCases(b *testing.B) {
	todo := b.N
	for {
		for _, g := range golden {
			String(g.expr, g.data)
			todo--
			if todo == 0 {
				return
			}
		}
	}
}
