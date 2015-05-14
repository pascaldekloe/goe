// Package el offers expression language "Goel".
// Goel is error-free by design. Malformed expressions have no result.
package el

import (
	"reflect"
)

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
