package responses

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestResponseTime(t *testing.T) {
	router := chi.NewRouter()

	router.Use(ResponseTime)
	router.Get("/sleep", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(1 * time.Second)
		_, _ = w.Write([]byte(""))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sleep", nil)
	router.ServeHTTP(res, req)

	d, err := time.ParseDuration(res.Header().Get(ResponseTimeHeader))

	if err != nil {
		t.Fatal(err)
	}

	if d < time.Second {
		t.Errorf("Expected %s header to be %s or more, got %s", ResponseTimeHeader, time.Second, d)
	}
}

func TestResponseDuration(t *testing.T) {
	router := chi.NewRouter()
	registry := prometheus.NewRegistry()
	fqdn := "http_request_duration_seconds"

	router.Use(ResponseTime)
	router.Use(RequestDuration(registry))
	router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(""))
	})
	router.Handle("/metrics", promhttp.InstrumentMetricHandler(
		registry,
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	))

	req := httptest.NewRequest("GET", "/", nil)

	for i := 0; i < 10; i += 1 {
		router.ServeHTTP(httptest.NewRecorder(), req)
	}

	res := httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("Expected metrics endpoint to response with 200 OK, got %d %s", res.Code, http.StatusText(res.Code))
	}

	metrics := res.Body.String()

	if !strings.Contains(metrics, fqdn) {
		t.Errorf("Expected metrics endpoint to return %s metric", fqdn)
	}
}
