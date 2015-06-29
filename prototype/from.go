package prototype

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

var (
	// HTTPClient is used for URL retrieval.
	HTTPClient = http.DefaultClient

	// HTTPAccept is the Accept header value used for URL retrieval.
	HTTPAccept = "text/json, application/json;charset=UTF-8, application/json, text/xml, application/xml;charset=UTF-8, application/xml;q=0.4, text/*;q=0.2, application/*;q=0.1"

	// HTTPAcceptCharset is the Accept-Charset header value used for URL retrieval.
	HTTPAcceptCharset = "UTF-8, US-ASCII"
)

// FromJSON gets a new template with the unmarshalled content.
func FromJSON(x interface{}, data []byte) Template {
	if err := json.Unmarshal(data, &x); err != nil {
		Fatalf("prototype: can't apply JSON: %s", err)
	}
	return New(x)
}

func FromJSONStream(x interface{}, stream io.Reader) Template {
	data, err := ioutil.ReadAll(stream)
	if err != nil {
		Fatalf("prototype: JSON stream fail: %s", err)
	}
	return FromJSON(x, data)
}

func FromXML(x interface{}, data []byte) Template {
	if err := xml.Unmarshal(data, &x); err != nil {
		Fatalf("prototype: can't apply XML: %s", err)
	}
	return New(x)
}

func FromXMLStream(x interface{}, stream io.Reader) Template {
	data, err := ioutil.ReadAll(stream)
	if err != nil {
		Fatalf("prototype: XML stream fail: %s", err)
	}
	return FromXML(x, data)
}

func FromURL(x interface{}, u string) Template {
	data, typ := readURL(u)

	if typ == "" {
		typ = mime.TypeByExtension(path.Ext(u))
	}
	typ, _, _ = mime.ParseMediaType(typ)

	if strings.HasSuffix(typ, "/json") || strings.HasSuffix(typ, "+json") {
		return FromJSON(x, data)
	}
	if strings.HasSuffix(typ, "/xml") || strings.HasSuffix(typ, "+xml") {
		return FromXML(x, data)
	}

	// Detect by magic
	switch data[0] {
	case '{', '[':
		return FromJSON(x, data)
	case '<':
		return FromXML(x, data)
	default:
		Fatalf("prototype: can't recognize the content of %s", u)
		panic("continue after prototype.Fatalf call.")
	}
}

func readURL(s string) (bytes []byte, mimeType string) {
	u, err := url.Parse(s)
	if err != nil {
		Fatalf("prototype: can't use URL: %s", err)
	}

	var stream io.Reader
	switch u.Scheme {
	case "http", "https":
		req, err := http.NewRequest("GET", s, nil)
		if err != nil {
			Fatalf("prototype: can't request %s: %s", s, err)
		}
		req.Header.Set("Accept", HTTPAccept)
		req.Header.Set("Accept-Charset", HTTPAcceptCharset)

		res, err := HTTPClient.Do(req)
		if err != nil {
			Fatalf("prototype: can't get %s: %s", s, err)
		}
		defer res.Body.Close()

		mimeType = res.Header.Get("Content-Type")
		stream = res.Body

	case "file":
		f, err := os.Open(u.Path)
		if err != nil {
			Fatalf("prototype: can't open %s: %s", s, err)
		}
		defer f.Close()

		stream = f

	default:
		Fatalf("prototype: unsupported scheme in URL %s", s)

	}

	bytes, err = ioutil.ReadAll(stream)
	if err != nil {
		Fatalf("prototype: can't read %s: %s", s, err)
	}
	return
}
