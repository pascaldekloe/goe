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
	{"/", testV.s, testV.s},
	{"/field", testV.s, nil},
	{"/mis", node{}, nil},
	{"/sub/x", node{}, nil},
	{"/s[1]", node{}, nil},
	{"/a[4]", node{}, nil},
	{"/s[0]", &node{s: []interface{}{testV.i}}, testV.i},
	{"/a[1]", node{a: [2]interface{}{testV.f, testV.s}}, testV.s},
	{"/[1]", "hello", testV.u + 97},
	{"/[1", "hello", nil},
	{"/[1]", testV, nil},
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
	r := f.Call(args)
	got, ok := r[0].Interface(), r[1].Bool()

	typ := r[0].Type()

	if !ok {
		if gold.want != nil && reflect.TypeOf(gold.want) == typ {
			t.Errorf("Got %s not OK for %q on %#v", typ, gold.expr, gold.data)
		}
		return
	}

	if gold.want == nil {
		t.Errorf("Got %s OK with %#v for %q on %#v", typ, got, gold.expr, gold.data)
		return
	}

	want := reflect.ValueOf(gold.want).Interface()
	if got != want {
		t.Errorf("Got %v, want %v for %s", got, want, gold.expr)
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
