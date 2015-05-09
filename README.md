# Go Enterprise

Tooling for the enterprise beauty.


## Verification

Test assertions on big objects can be cumbersome with ```reflect.DeepEquals``` and ```"Got %#v, want %#v"```.
Package `verify` offers convenience with reporting. For example `verify.Values(t, "employee", got, want)` might print:

```
--- FAIL: TestToCasAd (0.00s)
	values.go:15: values at demo_test.go:72: verification for employee:
		/Title: "Agent" != "Commander"
		/Tooling/Expired["ppk"]: missing time.Time
		/UserAttributes[2]/Label: "Car" != "Vehicle"
FAIL
```
