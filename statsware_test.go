package statsware

import (
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
	b.Status = httpStatus
	b.Uri = b.TransformURI(r.URL.RequestURI())
	b.Uri = b.TransformURI(r.RequestURI)
	b.Duration = t
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

	backend := TestBackend{TransformURI: testURItrasformer}
	middleware := Middleware{testhandler, backend}

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, rr.Code, backend.Status)
	assert.Equal(t, "/test", backend.Uri)
}
