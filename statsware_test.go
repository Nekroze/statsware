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
	data         map[string]interface{}
	TransformURI URItransformer
}

func (b TestBackend) WriteRequest(r *http.Request, httpStatus int, t time.Duration) error {
	b.data["status"] = httpStatus
	b.data["uri"] = b.TransformURI(r.URL.RequestURI())
	b.data["time"] = t
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

	b := TestBackend{TransformURI: testURItrasformer, data: map[string]interface{}{}}
	middleware := Middleware{Handler: testhandler, Backend: &b}

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, rr.Code, b.data["status"].(int))
	assert.Equal(t, "/test", b.data["uri"].(string))
	assert.True(t, b.data["time"].(time.Duration) > 0)
}
