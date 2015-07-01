package prototype

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pascaldekloe/goe/verify"
)

var fatals []string

func recordFatals() {
	fatals = nil
	Fatalf = func(format string, v ...interface{}) {
		fatals = append(fatals, fmt.Sprintf(format, v...))
	}
}

type TestItem struct {
	Title string
}

func TestMutations(t *testing.T) {
	Fatalf = t.Fatalf
	in := new(TestItem)

	p := New(in)
	out, ok := p.Build().(*TestItem)
	if !ok {
		t.Fatalf("Got type %s, want %s", reflect.TypeOf(in), reflect.TypeOf(out))
	}

	verify.Values(t, "identity", out, new(TestItem))

	in.Title = "Changed"
	verify.Values(t, "source changed", p.Build(), new(TestItem))

	out.Title = "Changed"
	verify.Values(t, "instance changed", p.Build(), new(TestItem))
}

func TestPathIn(t *testing.T) {
	Fatalf = t.Fatalf
	templates := New(new(TestItem)).HaveIn("/Title", "First", "Second")

	got := make([]*TestItem, len(templates))
	for i, _ := range templates {
		x, ok := templates[i].Build().(*TestItem)
		if !ok {
			t.Fatal("Got wrong type")
		}
		got[i] = x
	}

	want := []*TestItem{
		{Title: "First"},
		{Title: "Second"},
	}

	verify.Values(t, "prototype instances", got, want)
}

func TestWithToFail(t *testing.T) {
	recordFatals()
	New(new(TestItem)).Have("/DoesNotExist", 42)

	want := []string{"prototype: can't apply /DoesNotExist on *prototype.TestItem"}
	verify.Values(t, "messages", fatals, want)
}

func TestUnserializable(t *testing.T) {
	recordFatals()

	type loop *loop
	var x loop
	x = &x

	p := New(x)
	wantPrefix := "prototype: can't serialize prototype.loop: "
	if len(fatals) != 1 {
		t.Fatalf("Want 1 fatal, got %q", fatals)
	} else if !strings.HasPrefix(fatals[0], wantPrefix) {
		t.Errorf("Want message start %q, got %q", wantPrefix, fatals[0])
	}

	p.Build()
	wantPrefix = "prototype: can't deserialize prototype.loop: "
	if len(fatals) != 2 {
		t.Fatalf("Want 2 fatals, got %q", fatals)
	} else if !strings.HasPrefix(fatals[1], wantPrefix) {
		t.Errorf("Want message start %q, got %q", wantPrefix, fatals[1])
	}
}

func TestSerials(t *testing.T) {
	Fatalf = t.Fatalf
	p := New(TestItem{Title: "Serialize"})

	wantJSON := `{"Title":"Serialize"}`
	if got := string(p.BuildJSON()); got != wantJSON {
		t.Errorf("Want %s, got %s", wantJSON, got)
	}

	wantXML := `<TestItem><Title>Serialize</Title></TestItem>`
	if got := string(p.BuildXML()); got != wantXML {
		t.Errorf("Want %s, got %s", wantXML, got)
	}
}

func TestSerialFails(t *testing.T) {
	recordFatals()
	p := New(make(map[int]complex128))

	p.BuildJSON()
	wantPrefix := "prototype: can't serialize map[int]complex128 to JSON: "
	if len(fatals) != 1 {
		t.Errorf("Want 1 fatal, got %q", fatals)
	} else if !strings.HasPrefix(fatals[0], wantPrefix) {
		t.Errorf("Want message start %q, got %q", wantPrefix, fatals[0])
	}

	p.BuildXML()
	wantPrefix = "prototype: can't serialize map[int]complex128 to XML: "
	if len(fatals) != 2 {
		t.Errorf("Want 2 fatals, got %q", fatals)
	} else if !strings.HasPrefix(fatals[1], wantPrefix) {
		t.Errorf("Want message start %q, got %q", wantPrefix, fatals[1])
	}
}
