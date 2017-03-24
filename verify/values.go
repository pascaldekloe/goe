package verify

import (
	"bytes"
	"encoding"
	"fmt"
	"reflect"
	"strings"
)

// Errorer defines error reporting conform testing.T.
type Errorer interface {
	Error(args ...interface{})
}

// Values verifies that got has all the content, and only the content, defined by want.
func Values(r Errorer, name string, got, want interface{}) (ok bool) {
	t := travel{}
	t.values(reflect.ValueOf(got), reflect.ValueOf(want), nil)

	fail := t.report(name)
	if fail != "" {
		r.Error(fail)
		return false
	}

	return true
}

func (t *travel) values(got, want reflect.Value, path []*segment) {
	if !want.IsValid() {
		if got.IsValid() {
			t.differ(path, "Unwanted %s", got.Type())
		}
		return
	}
	if !got.IsValid() {
		t.differ(path, "Missing %s", want.Type())
		return
	}

	if got.Type() != want.Type() {
		t.differ(path, "Got type %s, want %s", got.Type(), want.Type())
		return
	}

	switch got.Kind() {

	case reflect.Struct:
		seg := &segment{format: "/%s"}
		path = append(path, seg)
		for i, n := 0, got.NumField(); i < n; i++ {
			field := got.Type().Field(i)
			seg.x = field.Name
			if field.PkgPath != "" {
				if !t.valuesUnexp(got.Interface(), want.Interface(), path) && !t.valuesUnexp(got.Addr().Interface(), want.Addr().Interface(), path) {
					t.differ(path, "Can't read private fields")
				}
				break
			}
			t.values(got.Field(i), want.Field(i), path)
		}
		path = path[:len(path)-1]

	case reflect.Slice, reflect.Array:
		n := got.Len()
		if n != want.Len() {
			t.differ(path, "Got %d elements, want %d", n, want.Len())
			return
		}

		seg := &segment{format: "[%d]"}
		path = append(path, seg)
		for i := 0; i < n; i++ {
			seg.x = i
			t.values(got.Index(i), want.Index(i), path)
		}
		path = path[:len(path)-1]

	case reflect.Ptr:
		if got.Pointer() != want.Pointer() {
			t.values(got.Elem(), want.Elem(), path)
		}

	case reflect.Interface:
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
		if !(got.IsNil() && want.IsNil()) {
			t.differ(path, "Can't compare functions")
		}

	case reflect.Float32, reflect.Float64:
		a, b := got.Float(), want.Float()
		if a != b && !(a != a && b != b) {
			t.differ(path, fmt.Sprintf("Got %f (%e), want %f (%e)", a, a, b, b))
		}

	case reflect.Complex128:
		a, b := got.Complex(), want.Complex()
		if a != b && !(a != a && b != b) {
			t.differ(path, fmt.Sprintf("Got %f (%e), want %f (%e)", a, a, b, b))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a, b := got.Int(), want.Int(); a != b {
			if a < 0xA && a > -0xA && b < 0xA && b > -0xA {
				t.differ(path, fmt.Sprintf("Got %d, want %d", a, b))
			} else {
				t.differ(path, fmt.Sprintf("Got %d (0x%x), want %d (0x%x)", a, a, b, b))
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if a, b := got.Uint(), want.Uint(); a != b {
			if a < 0xA && b < 0xA {
				t.differ(path, fmt.Sprintf("Got %d, want %d", a, b))
			} else {
				t.differ(path, fmt.Sprintf("Got %d (0x%x), want %d (0x%x)", a, a, b, b))
			}
		}

	case reflect.String:
		if a, b := got.String(), want.String(); a != b {
			t.differ(path, differMsg(a, b))
		}

	default:
		if a, b := got.Interface(), want.Interface(); a != b {
			t.differ(path, fmt.Sprintf("Got %v, want %v", a, b))
		}
	}
}

func (t *travel) valuesUnexp(got, want interface{}, path []*segment) bool {
	if m, ok := got.(encoding.TextMarshaler); ok {
		path[len(path)-1].x = ".MarshalText()"

		gotBytes, gotBytesErr := m.MarshalText()
		wantBytes, wantBytesErr := want.(encoding.TextMarshaler).MarshalText()

		if gotBytesErr == nil && wantBytesErr == nil {
			if !bytes.Equal(gotBytes, wantBytes) {
				t.differ(path, differMsg(string(gotBytes), string(wantBytes)))
			}
			return true
		}
	}

	if m, ok := got.(encoding.BinaryMarshaler); ok {
		path[len(path)-1].x = ".MarshalBinary()"

		gotBytes, gotBytesErr := m.MarshalBinary()
		wantBytes, wantBytesErr := want.(encoding.BinaryMarshaler).MarshalBinary()

		if gotBytesErr == nil && wantBytesErr == nil {
			if !bytes.Equal(gotBytes, wantBytes) {
				t.differ(path, fmt.Sprintf("Got serial 0x%x, want 0x%x", gotBytes, wantBytes))
			}
			return true
		}
	}

	return false
}

func applyKeySeg(dst *segment, key reflect.Value) {
	if key.Kind() == reflect.String {
		dst.format = "[%q]"
	} else {
		dst.format = "[%v]"
	}
	dst.x = key.Interface()
}

func differMsg(got, want string) string {
	if len(got) < 9 || len(want) < 9 {
		return fmt.Sprintf("Got %q, want %q", got, want)
	}

	got, want = fmt.Sprintf("%q", got), fmt.Sprintf("%q", want)

	// find first character which differs
	var i int
	a, b := []rune(got), []rune(want)
	for i = 0; i < len(a); i++ {
		if i >= len(b) || a[i] != b[i] {
			break
		}
	}
	return fmt.Sprintf("Got %s, want %s\n    %s^", got, want, strings.Repeat(" ", i))
}
