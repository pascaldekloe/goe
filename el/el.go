// Package el offers expression language "Goel".
//
// The API is error-free by design. Malformed expressions simply have no result.
//
// Struct fields can be selected by name as in "/Catalog/Title".
// Square brackets are used for index selection.
// For example, el.Int("/Movies[3]/Year", c) gets the fourth movie's year and
// el.Strings("/Movies[*]/Title", c) will get all movie titles.
package el

import (
	"reflect"
)

// Bool returns the result if, and only if, the expression has one value and the value is a boolean type.
func Bool(expr string, root interface{}) (result bool, ok bool) {
	a := eval(expr, root)
	if len(a) != 1 {
		return
	}

	v := a[0]
	if v.Kind() != reflect.Bool {
		return
	}

	return v.Bool(), true
}

// Int returns the result if, and only if, the expression has one value and the value is an integer type.
func Int(expr string, root interface{}) (result int64, ok bool) {
	a := eval(expr, root)
	if len(a) != 1 {
		return
	}

	v := a[0]
	if k := v.Kind(); k != reflect.Int64 && k != reflect.Int32 && k != reflect.Int && k != reflect.Int8 && k != reflect.Int16 {
		return
	}

	return v.Int(), true
}

// Uint returns the result if, and only if, the expression has one value and the value is an unsigned integer type.
func Uint(expr string, root interface{}) (result uint64, ok bool) {
	a := eval(expr, root)
	if len(a) != 1 {
		return
	}

	v := a[0]
	if k := v.Kind(); k != reflect.Uint64 && k != reflect.Uint32 && k != reflect.Uint && k != reflect.Uint8 && k != reflect.Uint16 {
		return
	}

	return v.Uint(), true
}

// Float returns the result if, and only if, the expression has one value and the value is a floating point type.
func Float(expr string, root interface{}) (result float64, ok bool) {
	a := eval(expr, root)
	if len(a) != 1 {
		return
	}

	v := a[0]
	if k := v.Kind(); k != reflect.Float64 && k != reflect.Float32 {
		return
	}

	return v.Float(), true
}

// Complex returns the result if, and only if, the expression has one value and the value is a complex type.
func Complex(expr string, root interface{}) (result complex128, ok bool) {
	a := eval(expr, root)
	if len(a) != 1 {
		return
	}

	v := a[0]
	if k := v.Kind(); k != reflect.Complex128 && k != reflect.Complex64 {
		return
	}

	return v.Complex(), true
}

// String returns the result if, and only if, the expression has one value and the value is a string type.
func String(expr string, root interface{}) (result string, ok bool) {
	a := eval(expr, root)
	if len(a) != 1 {
		return
	}

	v := a[0]
	if v.Kind() != reflect.String {
		return
	}

	return v.String(), true
}

// Bools returns all values with a boolean type.
func Bools(expr string, root interface{}) []bool {
	a := eval(expr, root)
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
	a := eval(expr, root)
	if len(a) == 0 {
		return nil
	}

	b := make([]int64, 0, len(a))
	for _, v := range a {
		if k := v.Kind(); k == reflect.Int64 || k == reflect.Int32 || k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 {
			b = append(b, v.Int())
		}
	}
	return b
}

// Uints returns all values with an unsigned integer type.
func Uints(expr string, root interface{}) []uint64 {
	a := eval(expr, root)
	if len(a) == 0 {
		return nil
	}

	b := make([]uint64, 0, len(a))
	for _, v := range a {
		if k := v.Kind(); k == reflect.Uint64 || k == reflect.Uint32 || k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 {
			b = append(b, v.Uint())
		}
	}
	return b
}

// Floats returns all values with a floating point type.
func Floats(expr string, root interface{}) []float64 {
	a := eval(expr, root)
	if len(a) == 0 {
		return nil
	}

	b := make([]float64, 0, len(a))
	for _, v := range a {
		if k := v.Kind(); k == reflect.Float64 || k == reflect.Float32 {
			b = append(b, v.Float())
		}
	}
	return b
}

// Complexes returns all values with a complex type.
func Complexes(expr string, root interface{}) []complex128 {
	a := eval(expr, root)
	if len(a) == 0 {
		return nil
	}

	b := make([]complex128, 0, len(a))
	for _, v := range a {
		if k := v.Kind(); k == reflect.Complex128 || k == reflect.Complex64 {
			b = append(b, v.Complex())
		}
	}
	return b
}

// Strings returns all values with a string type.
func Strings(expr string, root interface{}) []string {
	a := eval(expr, root)
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
