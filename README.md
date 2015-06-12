# Go Enterprise [![GoDoc](https://godoc.org/github.com/pascaldekloe/goe?status.svg)](https://godoc.org/github.com/pascaldekloe/goe)

Common enterprise features for the Go programming language (golang).


## Expression Language [API](http://godoc.org/github.com/pascaldekloe/goe/el)

GoEL expressions provide error free access to Go types.
It serves as a lightweigth alternative to [unified EL](https://docs.oracle.com/javaee/5/tutorial/doc/bnahq.html), [SpEL](http://docs.spring.io/spring/docs/current/spring-framework-reference/html/expressions.html) or even [XPath](http://www.w3.org/TR/xpath), [CSS selectors](http://www.w3.org/TR/css3-selectors) and friends.

``` Go
func FancyOneLiners() {
	// Single field selection:
	upper, applicable := el.Bool(`/CharSet[0x1F]/isUpperCase`, x)

	// Escape path separator slash:
	warnings := el.Strings(`/Report/Stats["I\x2fO"]/warn[*]`, x)

	// Data modification:
	updateCount := el.Have(`/Nodes[*]/Cache/TTL`, x, 3600)
```

#### Selection

Slash-separated paths are used to select data. All paths are subjected to [normalization rules](http://golang.org/pkg/path#Clean).

Both exported and non-exported `struct` fields can be selected by name.

Elements in indexed types `array`, `slice` and `string` are denoted with a zero based number inbetween square brackets. Key selections from `map` types also use the square bracket notation. Asterisk can be used as a wildcard as in `[*]` to match all entries.

``` BNF
path            ::= path-component | path path-component
path-component  ::= "/" segment
segment         ::= "" | ".." | selection | selection key
selection       ::= "." | go-field-name
key             ::= "[" key-selection "]"
key-selection   ::= "*" | go-literal
```

#### Performance

The implementation is optimized for performance. No need to precompile expressions.

```
# go test -bench=. -benchmem
PASS
BenchmarkPaths-8	 1000000	      1002 ns/op	     326 B/op	      12 allocs/op
BenchmarkHaves-8	 1000000	      1423 ns/op	     478 B/op	      16 allocs/op
ok  	el	3.397s
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
