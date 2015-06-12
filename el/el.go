// Package el offers expression language "GoEL".
//
// The API is error-free by design. Malformed expressions simply have no result.
//
// Slash-separated paths are used to select data. All paths are subjected to
// normalization rules. See http://golang.org/pkg/path#Clean.
//
// Both exported and non-exported struct fields can be selected by name.
//
// Elements in indexed types array, slice and string are denoted with a zero
// based integer literal inbetween square brackets. Key selections from map
// types also use the square bracket notation. Asterisk can be used as a
// wildcard as in `[*]` to match all entries.
package el

import (
	"reflect"
)

func eval(expr string, root interface{}, doBuild bool) []value {
	if expr == "" {
		return nil
	}

	switch expr[0] {
	case '/':
		return resolve(expr, root, doBuild)
	default:
		return nil
	}
}

// Have applies want to the path on root and returns the number of successes.
//
// All content in the path is instantiated the fly with the zero value where
// possible. This implies automatic construction of structs, pointers and maps.
//
// For the operation to succeed the targets must be settable conform to the
// third law of reflection. See http://blog.golang.org/laws-of-reflection#TOC_8.
// In short, root should be a pointer and the destination should be exported.
// See https://golang.org/ref/spec#Exported_identifiers.
func Have(root interface{}, path string, want interface{}) (n int) {
	values := eval(path, root, true)

	var w reflect.Value
	if wp := follow(reflect.ValueOf(want), false); wp == nil {
		return
	} else {
		w = (*wp).(reflect.Value)
	}
	wt := w.Type()

	for _, v := range values {
		if !v.CanSet() {
			continue
		}

		switch vt := v.Type(); {
		case wt.AssignableTo(vt):
			v.Set(w)
			n++
		case wt.ConvertibleTo(vt):
			v.Set(w.Convert(vt))
			n++
		}
	}

	return n
}

// Bool returns the result if, and only if, the expression has one value and the
// value is a boolean type.
func Bool(expr string, root interface{}) (result bool, ok bool) {
	a := eval(expr, root, false)
	if len(a) != 1 {
		return
	}

	v := a[0]
	if v.Kind() == reflect.Bool {
		return v.Bool(), true
	}
	return
}

// Int returns the result if, and only if, the expression has one value and the
// value is an integer type.
func Int(expr string, root interface{}) (result int64, ok bool) {
	a := eval(expr, root, false)
	if len(a) != 1 {
		return
	}

	v := a[0]
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	}
	return
}

// Uint returns the result if, and only if, the expression has one value and the
// value is an unsigned integer type.
func Uint(expr string, root interface{}) (result uint64, ok bool) {
	a := eval(expr, root, false)
	if len(a) != 1 {
		return
	}

	v := a[0]
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), true
	}
	return
}

// Float returns the result if, and only if, the expression has one value and
// the value is a floating point type.
func Float(expr string, root interface{}) (result float64, ok bool) {
	a := eval(expr, root, false)
	if len(a) != 1 {
		return
	}

	v := a[0]
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	}
	return
}

// Complex returns the result if, and only if, the expression has one value and
// the value is a complex type.
func Complex(expr string, root interface{}) (result complex128, ok bool) {
	a := eval(expr, root, false)
	if len(a) != 1 {
		return
	}

	v := a[0]
	switch v.Kind() {
	case reflect.Complex64, reflect.Complex128:
		return v.Complex(), true
	}
	return
}

// String returns the result if, and only if, the expression has one value and
// the value is a string type.
func String(expr string, root interface{}) (result string, ok bool) {
	a := eval(expr, root, false)
	if len(a) != 1 {
		return
	}

	v := a[0]
	if v.Kind() == reflect.String {
		return v.String(), true
	}
	return
}

// Bools returns all values with a boolean type.
func Bools(expr string, root interface{}) []bool {
	a := eval(expr, root, false)
	if len(a) == 0 {
		return nil
	}

	b := make([]bool, 0, len(a))
	for _, v := range a {
		if v.Kind() == reflect.Bool {
			b = append(b, v.Bool())
		}
	}
	return b
}

// Ints returns all values with an integer type.
func Ints(expr string, root interface{}) []int64 {
	a := eval(expr, root, false)
	if len(a) == 0 {
		return nil
	}

	b := make([]int64, 0, len(a))
	for _, v := range a {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			b = append(b, v.Int())
		}
	}
	return b
}

// Uints returns all values with an unsigned integer type.
func Uints(expr string, root interface{}) []uint64 {
	a := eval(expr, root, false)
	if len(a) == 0 {
		return nil
	}

	b := make([]uint64, 0, len(a))
	for _, v := range a {
		switch v.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			b = append(b, v.Uint())
		}
	}
	return b
}

// Floats returns all values with a floating point type.
func Floats(expr string, root interface{}) []float64 {
	a := eval(expr, root, false)
	if len(a) == 0 {
		return nil
	}

	b := make([]float64, 0, len(a))
	for _, v := range a {
		switch v.Kind() {
		case reflect.Float32, reflect.Float64:
			b = append(b, v.Float())
		}
	}
	return b
}

// Complexes returns all values with a complex type.
func Complexes(expr string, root interface{}) []complex128 {
	a := eval(expr, root, false)
	if len(a) == 0 {
		return nil
	}

	b := make([]complex128, 0, len(a))
	for _, v := range a {
		switch v.Kind() {
		case reflect.Complex64, reflect.Complex128:
			b = append(b, v.Complex())
		}
	}
	return b
}

// Strings returns all values with a string type.
func Strings(expr string, root interface{}) []string {
	a := eval(expr, root, false)
	if len(a) == 0 {
		return nil
	}

	b := make([]string, 0, len(a))
	for _, v := range a {
		if v.Kind() == reflect.String {
			b = append(b, v.String())
		}
	}
	return b
}
