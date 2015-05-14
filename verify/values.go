package verify

import (
	"reflect"
	"testing"
)

// Values verifies that got has all the content, and only the content, defined by want.
func Values(t *testing.T, name string, got, want interface{}) (ok bool) {
	tr := travel{}
	tr.values(reflect.ValueOf(got), reflect.ValueOf(want), nil)

	fail := tr.report(name)
	if fail != "" {
		t.Error(fail)
		return false
	}

	return true
}

func (t *travel) values(got, want reflect.Value, path []*segment) {
	if !want.IsValid() {
		if got.IsValid() {
			t.differ(path, "unwanted %s", got.Type())
		}
		return
	}
	if !got.IsValid() {
		t.differ(path, "missing %s", want.Type())
		return
	}

	if got.Type() != want.Type() {
		t.differ(path, "types differ, %s != %s", got.Type(), want.Type())
		return
	}

	switch got.Kind() {

	case reflect.Struct:
		seg := &segment{format: "/%s"}
		path = append(path, seg)
		for i, n := 0, got.NumField(); i < n; i++ {
			seg.x = got.Type().Field(i).Name
			t.values(got.Field(i), want.Field(i), path)
		}
		path = path[:len(path)-1]

	case reflect.Slice, reflect.Array:
		n := got.Len()
		if n != want.Len() {
			t.differ(path, "got %d elements, want %d", n, want.Len())
			return
		}

		seg := &segment{format: "[%d]"}
		path = append(path, seg)
		for i := 0; i < n; i++ {
			seg.x = i
			t.values(got.Index(i), want.Index(i), path)
		}
		path = path[:len(path)-1]

	case reflect.Ptr, reflect.Interface:
		t.values(got.Elem(), want.Elem(), path)

	case reflect.Map:
		seg := &segment{}
		path = append(path, seg)
		for _, key := range want.MapKeys() {
			applyKeySeg(seg, key)
			t.values(got.MapIndex(key), want.MapIndex(key), path)
		}

		for _, key := range got.MapKeys() {
			v := want.MapIndex(key)
			if v.IsValid() {
				continue
			}
			applyKeySeg(seg, key)
			t.values(got.MapIndex(key), v, path)
		}
		path = path[:len(path)-1]

	case reflect.Func:
		t.differ(path, "can't compare functions")

	case reflect.String:
		a, b := got.String(), want.String()
		if a != b {
			t.differ(path, "%q != %q", a, b)
		}

	default:
		a, b := asInterface(got), asInterface(want)
		if a != b {
			t.differ(path, "%v != %v", a, b)
		}

	}
}

func applyKeySeg(dst *segment, key reflect.Value) {
	if key.Kind() == reflect.String {
		dst.format = "[%q]"
	} else {
		dst.format = "[%v]"
	}
	dst.x = asInterface(key)
}

func asInterface(v reflect.Value) interface{} {
	if v.CanInterface() {
		return v.Interface()
	}

	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Complex64, reflect.Complex128:
		return v.Complex()
	}

	panic("verify: can't interface kind " + v.Kind().String())
}
