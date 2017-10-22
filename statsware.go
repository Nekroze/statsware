package main

import (
	"fmt"
	"net/http"
	"time"

	statsd "gopkg.in/alexcesaro/statsd.v2"
)

type Backend interface {
	WriteRequest(r *http.Request, httpstatus int, t time.Duration) err
}

type URItransformer func(string) string

func argInMemory(arg string, memory map[string]string) bool {
	for k, _ := range memory {
		if k == arg {
			return true
		}
	}
	return false
}

func cullMemory(memory map[string]string, limit int) {
	if limit < 1 {
		return
	}
	badkeys := []string{}
	i := 0
	for k, _ := range memory {
		if i > limit {
			badkeys = append(badkeys)
		}
		i++
	}
	for _, k := range badkeys {
		delete(memory, k)
	}
}

// A limit of 0 is considered no limit
func Memoize(f URItransformer, limit int) URItransformer {
	memory := map[string]string{}
	return func(arg string) string {
		if argInMemory(arg, memory) {
			return memory[arg]
		}
		cullMemory(memory, limit)
		memory[arg] = f(arg)
		return memory[arg]
	}
}

type StatsdBackend struct {
	statsd.Client
	statsd.Timing
	TransformURI URItransformer
}

func (b *StatsdBackend) WriteRequest(r *http.Request, httpStatus int, t time.Duration) error {
	b.Client.Increment(fmt.Sprintf("http%d", httpStatus))
	b.Timing(b.TransformURI(r.URL.RequestURI()), t)
	return nil
}

type responseCaptureWriter struct {
	http.ResponseWriter
	statusCode int
	startTime  time.Time
}

func (rw *responseCaptureWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.startTime = time.Now()
	rw.ResponseWriter.WriteHeader(code)
}

type Middleware struct {
	http.Handler
	Backend
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := responseCaptureWriter{w, http.StatusOk}
	m.Handler.ServeHTTP(rw, r)
	m.Backend.CompleteRequest(r, rw.statusCode, time.Since(rw.startTime))
}
