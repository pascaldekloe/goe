package prototype

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"verify"
)

func TestFromFileURL(t *testing.T) {
	Fatalf = t.Fatalf

	path, err := filepath.Abs("./sample.json")
	if err != nil {
		t.Fatal("Absolute path for sample:", err)
	}

	got := FromURL(new(TestItem), "file://"+path).Build()

	want := &TestItem{Title: "JSON Sample"}
	verify.Values(t, "sample", got, want)
}

func TestFromHTTPURL(t *testing.T) {
	Fatalf = t.Fatalf

	path := "/sample"
	srvJSON := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Errorf("Want path %s, got %s", path, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write([]byte(`{"Title": "HTTP Sample"}`)); err != nil {
			t.Fatal("Can't serve JSON:", err)
		}
	}

	srv := httptest.NewServer(http.HandlerFunc(srvJSON))

	got := FromURL(new(TestItem), srv.URL+path).Build()

	want := &TestItem{Title: "HTTP Sample"}
	verify.Values(t, "sample", got, want)
}
