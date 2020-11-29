package middleware

// The original work was derived from go-chi's middleware, source:
// https://github.com/go-chi/chi/tree/master/middleware/wrap_writer.go

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

const ResponseTimeHeader = "X-Response-Time"

// ResponseTime adds a "X-Response-Time" header once the handler
// writes the header of the response.
func ResponeTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := newWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
	})
}

//

type timedWriter struct {
	http.ResponseWriter
	start       time.Time
	wroteHeader bool
}

func newWriter(w http.ResponseWriter, protoMajor int) http.ResponseWriter {
	_, fl := w.(http.Flusher)

	tw := timedWriter{ResponseWriter: w, start: time.Now()}

	if protoMajor == 2 {
		_, ps := w.(http.Pusher)
		if fl && ps {
			return &http2Writer{tw}
		}
	} else {
		_, hj := w.(http.Hijacker)
		_, rf := w.(io.ReaderFrom)
		if fl && hj && rf {
			return &httpWriter{tw}
		}
	}
	if fl {
		return &flushWriter{tw}
	}

	return &tw
}

func (t *timedWriter) WriteHeader(code int) {
	if !t.wroteHeader {
		t.wroteHeader = true

		// write the response time header
		dur := int(time.Since(t.start).Milliseconds())
		t.Header().Add(ResponseTimeHeader, strconv.Itoa(dur))

		t.ResponseWriter.WriteHeader(code)
	}
}

func (t *timedWriter) Write(buf []byte) (int, error) {
	if !t.wroteHeader {
		t.WriteHeader(http.StatusOK)
	}
	return t.ResponseWriter.Write(buf)
}

type flushWriter struct {
	timedWriter
}

func (f *flushWriter) Flush() {
	f.wroteHeader = true
	fl := f.timedWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}

type httpWriter struct {
	timedWriter
}

func (h1 *httpWriter) Flush() {
	h1.wroteHeader = true
	fl := h1.timedWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}

func (h1 *httpWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := h1.timedWriter.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

func (h1 *httpWriter) ReadFrom(r io.Reader) (int64, error) {
	rf := h1.timedWriter.ResponseWriter.(io.ReaderFrom)
	if !h1.wroteHeader {
		h1.WriteHeader(http.StatusOK)
	}
	return rf.ReadFrom(r)
}

type http2Writer struct {
	timedWriter
}

func (h2 *http2Writer) Push(target string, opts *http.PushOptions) error {
	return h2.timedWriter.ResponseWriter.(http.Pusher).Push(target, opts)
}

func (h2 *http2Writer) Flush() {
	h2.wroteHeader = true
	fl := h2.timedWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}
