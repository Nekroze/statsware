package statsware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestBackend struct {
	Status int
	Uri    string
	time.Duration
	TransformURI URItransformer
}

func (b TestBackend) WriteRequest(r *http.Request, httpStatus int, t time.Duration) error {
	fmt.Printf("TTT stats %#v uri %#v duration %s\n", b.Status, b.Uri, b.Duration)
	b.Status = httpStatus
	b.Uri = b.TransformURI(r.URL.RequestURI())
	b.Duration = t
	fmt.Printf("WWW stats %#v uri %#v duration %s\n", b.Status, b.Uri, b.Duration)
	return nil
}

func testURItrasformer(in string) string {
	return in
}

func TestMiddleware(t *testing.T) {
	testcheck := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"alive": true}`)
	}
	testhandler := http.HandlerFunc(testcheck)

	b := TestBackend{TransformURI: testURItrasformer, Status: 1}
	middleware := Middleware{Handler: testhandler, Backend: &b}

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	fmt.Printf("XXX stats %#v uri %#v duration %s\n", b.Status, b.Uri, b.Duration)
	middleware.ServeHTTP(rr, req)
	fmt.Printf("YYY stats %#v uri %#v duration %s\n", b.Status, b.Uri, b.Duration)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, rr.Code, b.Status)
	assert.Equal(t, "/test", b.Uri)
}
