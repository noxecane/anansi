package requests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/random-guys/go-siber/json"
	"github.com/rs/zerolog"
	"syreclabs.com/go/faker"
)

func TestTimeout(t *testing.T) {
	router := chi.NewRouter()
	success := []byte("success")
	failed := []byte("failed")

	router.Use(Timeout(time.Second))
	router.Get("/timeout", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			_, _ = w.Write(success)
			return
		case <-time.After(2 * time.Second):
		}
		_, _ = w.Write(failed)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/timeout", nil)
	router.ServeHTTP(res, req)

	if res.Body.String() != string(success) {
		t.Errorf("Expected http to return %s string got %s", string(success), res.Body.String())
	}
}

func TestAttachLog(t *testing.T) {
	router := chi.NewRouter()
	success := []byte("success")
	failed := []byte("failed")
	logOut := &bytes.Buffer{}

	logger := zerolog.
		New(logOut).
		With().
		Timestamp().
		Logger()

	router.Use(AttachLogger(logger))
	router.Get("/attach", func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())
		if log.GetLevel() == zerolog.Disabled {
			_, _ = w.Write(failed)
			return
		}

		_, _ = w.Write(success)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/attach", nil)
	router.ServeHTTP(res, req)

	if res.Body.String() != string(success) {
		t.Error("Expected logger to be attached to request")
	}
}

func TestLog(t *testing.T) {
	type requestLog struct {
		URL     string                 `json:"url"`
		Method  string                 `json:"method"`
		Address string                 `json:"remote_address"`
		Body    map[string]interface{} `json:"request"`
	}

	router := chi.NewRouter()
	logOut := &bytes.Buffer{}

	logger := zerolog.
		New(logOut).
		With().
		Timestamp().
		Logger()

	router.Use(AttachLogger(logger))
	router.Use(Log)
	router.Post("/attach", func(w http.ResponseWriter, r *http.Request) {
		// write to buffer
		log := zerolog.Ctx(r.Context())
		log.Info().Msg("")

		_, _ = w.Write([]byte(""))
	})

	name := faker.Name().Name()
	b, _ := json.Marshal(map[string]string{"name": name})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/attach", bytes.NewBuffer(b))
	router.ServeHTTP(res, req)

	logs := requestLog{}
	_ = json.Unmarshal(logOut.Bytes(), &logs)

	if logs.URL == "" {
		t.Error("Expected URL to be logged")
	}

	if logs.Method == "" {
		t.Error("Expected request method to be logged")
	}

	if logs.Address == "" {
		t.Error("Expected request address to be logged")
	}

	nameInBody, ok := logs.Body["name"].(string)
	if !ok {
		t.Errorf("Expected request body to have name as string, got %T", logs.Body["name"])
	}

	if nameInBody != name {
		t.Errorf("Expected request body to have name as %s, got %s", name, nameInBody)
	}
}
