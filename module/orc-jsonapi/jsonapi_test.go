package jsonapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

type fakeResponseWriter struct {
	header http.Header
	data   []byte
	code   int
}

func newFakeResponseWriter() *fakeResponseWriter {
	return &fakeResponseWriter{
		header: map[string][]string{},
	}
}

func (f *fakeResponseWriter) Header() http.Header {
	return f.header
}

func (f *fakeResponseWriter) Write(writedata []byte) (int, error) {
	if f.code == 0 {
		f.code = 200
	}
	f.data = append(f.data, writedata...)
	return len(writedata), nil
}
func (f *fakeResponseWriter) WriteHeader(statusCode int) {
	f.code = statusCode
}

func TestJSONAPI(t *testing.T) {
	m := &Module{apimux: mux.NewRouter()}

	m.Handle("/foo/{fooID}/bar/{barID}/", Methods{
		Post: DiscardBody(func(req *http.Request) (interface{}, error) {
			return map[string]string{
				"echo": req.URL.Path,
			}, nil
		},
			ValidateVar("fooID", Numeric),
			ValidateVar("barID", Numeric),
		),
	},
	)

	fake := newFakeResponseWriter()

	req, _ := http.NewRequest("POST", "https://example.com/foo/123/bar/456/", &bytes.Buffer{})
	m.apimux.ServeHTTP(fake, req)

	var want interface{} = map[string]interface{}{
		"echo": "/foo/123/bar/456/",
	}

	if wantCode := 200; fake.code != wantCode {
		t.Errorf("got code %v want %v", fake.code, wantCode)
	}
	var got interface{}
	if err := json.Unmarshal(fake.data, &got); err != nil {
		t.Fatalf("unable to unmarshal %q: %v", string(fake.data), err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestJSONAPIError(t *testing.T) {
	m := &Module{apimux: mux.NewRouter()}

	m.Handle("/foo/{fooID}/bar/{barID}/", Methods{
		Post: DiscardBody(func(req *http.Request) (interface{}, error) {
			return map[string]string{
				"echo": req.URL.Path,
			}, nil
		},
			ValidateVar("fooID", Numeric),
			ValidateVar("barID", Numeric),
		),
	})

	fake := newFakeResponseWriter()

	req, _ := http.NewRequest("POST", "https://example.com/foo/123/bar/45a6/", &bytes.Buffer{})
	m.apimux.ServeHTTP(fake, req)

	var want interface{} = map[string]interface{}{
		"ok":    false,
		"error": `Bad request: Bad URL component "barID"="45a6": not numeric`,
	}

	if wantCode := 400; fake.code != wantCode {
		t.Errorf("got code %v want %v", fake.code, wantCode)
	}
	var got interface{}
	if err := json.Unmarshal(fake.data, &got); err != nil {
		t.Fatalf("unable to unmarshal %q: %v", string(fake.data), err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
