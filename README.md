# Go Enterprise

Common enterprise features for the Go programming language.


## Expression Language [API](http://godoc.org/github.com/pascaldekloe/goe/el)

Goel expressions provide error free access to object content.
It serves as a lightweigth alternative to [unified EL](https://docs.oracle.com/javaee/5/tutorial/doc/bnahq.html), [SpEL](http://docs.spring.io/spring/docs/current/spring-framework-reference/html/expressions.html) or even [XPath](http://www.w3.org/TR/xpath), [CSS selectors](http://www.w3.org/TR/css3-selectors) and friends.

``` Go
func FancyOneLiners() {
	// Single field selection:
	upper, applicable := el.Bool(`CharSet[0x1F]/isUpperCase`, x)

	// Escape path separator slash:
	warnings := el.Strings(`/Report/Stats["I\x2fO"]/warn[*]`, x)
```

#### Paths

Slash-separated [paths](http://golang.org/pkg/path) are used to select data. Both public and private `struct` fields can be selected by name.

Elements in indexed types `array`, `slice` and `string` are denoted with a zero based integer literal inbetween square brackets. Keys from `map` types also use the square bracket notation. Asterisk can be used as a wildcard as in `[*]` to match all entries.

``` BNF
path            ::= relative-path | "/" relative-path
relative-path   ::= segment | segment "/" segment
segment         ::= ".." | field | field key || key
field           ::= "." | go-field-name
key             ::= "[" key-selection "]"
key-selection   ::= "*" | go-literal
```

#### Performance

The implementation is highly optimized for performance. No need to precompile expressions.

``` Shell
# go test -bench=GoldenCases -benchmem
PASS
BenchmarkGoldenCases	 2000000	       732 ns/op	      97 B/op	       4 allocs/op
ok  	el	2.212s
```


## Verification [API](http://godoc.org/github.com/pascaldekloe/goe/verify)

Test assertions on big objects can be cumbersome with ```reflect.DeepEquals``` and ```"Got %#v, want %#v"```.
Package `verify` offers convenience with reporting. For example `verify.Values(t, "employee", got, want)` might print:

```
--- FAIL: TestValuesDemo (0.00s)
	values.go:15: verification for employee at demo_test.go:72:
		/Title: "Agent" != "Commander"
		/Tooling/Expired["ppk"]: missing time.Time
		/UserAttributes[2]/Label: "Car" != "Vehicle"
```
