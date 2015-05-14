# Go Enterprise

Tooling for the enterprise beauty.


## Expression Language [API](http://godoc.org/github.com/pascaldekloe/goe/el)

Goel expressions provide error free access to object content.

```
% go test -bench=GoldenCases -benchmem
PASS
BenchmarkGoldenCases	 3000000	       587 ns/op	      81 B/op	       3 allocs/op
ok  	_/Users/pkloe/Code/goe/src/el	2.367s
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
