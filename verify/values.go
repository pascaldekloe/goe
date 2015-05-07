package verify

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

// Values verifies that got has all the content, and only the content, defined by want.
func Values(t *testing.T, name string, got, want interface{}) bool {
	s := state{}
	s.values(reflect.ValueOf(got), reflect.ValueOf(want), "")
	if s.Len() != 0 {
		t.Error(name + " values not equivalent:" + s.String())
		return false
	}
	return true
}

type state struct {
	bytes.Buffer
}

func (s *state) fail(path, msg string, args ...interface{}) {
	s.WriteByte('\n')
	if path != "" {
		s.WriteString(path)
		s.WriteString(": ")
	}
	s.WriteString(fmt.Sprintf(msg, args...))
}

func (s *state) values(got, want reflect.Value, path string) {
	if !want.IsValid() {
		if got.IsValid() {
			s.fail(path, "unwanted %s", got)
		}
		return
	}
	if !got.IsValid() {
		s.fail(path, "missing %s", want)
		return
	}

	if got.Type() != want.Type() {
		s.fail(path, "types differ")
	}

	switch got.Kind() {
	case reflect.Struct:
		for i, n := 0, got.NumField(); i < n; i++ {
			p := path + "/" + got.Type().Field(i).Name
			s.values(got.Field(i), want.Field(i), p)
		}
	case reflect.Slice, reflect.Array:
		n := got.Len()
		if n != want.Len() {
			s.fail(path, "different length")
		}
		for i := 0; i < n; i++ {
			p := fmt.Sprintf("%s[%d]", path, i)
			s.values(got.Index(i), want.Index(i), p)
		}
	case reflect.Ptr, reflect.Interface:
		s.values(got.Elem(), want.Elem(), path)
	case reflect.Map:
		todo := make(map[interface{}]bool)
		for _, key := range want.MapKeys() {
			kv := asInterface(key)
			todo[kv] = true
		}

		for _, key := range got.MapKeys() {
			kv := asInterface(key)
			p := fmt.Sprintf("%s[%v]", path, kv)
			if !todo[kv] {
				s.fail(p, "not wanted")
			} else {
				todo[kv] = false
				s.values(got.MapIndex(key), want.MapIndex(key), p)
			}
		}

		for key, pending := range todo {
			if !pending {
				continue
			}
			p := fmt.Sprintf("%s[%v]", path, key)
			s.fail(p, "not available")
		}
	case reflect.Func:
		s.fail(path, "can't compare functions")
	default:
		a, b := asInterface(got), asInterface(want)
		if a != b {
			s.fail(path, "%v != %v", a, b)
		}
	}
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
	case reflect.String:
		return v.String()
	}

	panic("TODO: can't interface kind %s " + v.Kind().String())
}
