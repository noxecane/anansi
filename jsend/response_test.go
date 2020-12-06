package jsend

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/random-guys/go-siber/json"
	"github.com/random-guys/go-siber/requests"
	"github.com/rs/zerolog"
	"syreclabs.com/go/faker"
)

type responseLog struct {
	Status  int                 `json:"status"`
	Length  int                 `json:"length"`
	Headers map[string][]string `json:"response_headers"`
	Error   string              `json:"error"`
}

func TestSuccess(t *testing.T) {
	type message struct {
		Name string `json:"name"`
	}

	router := chi.NewRouter()
	name := faker.Name().Name()
	logOut := &bytes.Buffer{}

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		Success(r, w, message{name})
	})

	router.
		With(requests.AttachLogger(zerolog.New(logOut).With().Logger())).
		Get("/logged", func(w http.ResponseWriter, r *http.Request) {
			Success(r, w, message{name})
		})

	t.Run("sends jsend response", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		router.ServeHTTP(res, req)

		resp := jsendSuccess{}
		err := json.Unmarshal(res.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err)
		}

		if resp.Code != http.StatusOK {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusOK, resp.Code)
		}

		m := resp.Data.(map[string]interface{})
		mName := m["name"].(string)
		if mName != name {
			t.Errorf("Expected the name to be %s, got %s", name, mName)
		}
	})

	t.Run("logs response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/logged", nil)
		router.ServeHTTP(httptest.NewRecorder(), req)

		log := responseLog{}
		_ = json.Unmarshal(logOut.Bytes(), &log)

		if log.Status == 0 {
			t.Error("Expected status code to be logged")
		}

		if log.Length == 0 {
			t.Error("Expected response length to be logged")
		}

		if len(log.Headers) == 0 {
			t.Error("Expected response headers to be logged")
		}
	})
}

func TestErr(t *testing.T) {
	type metadata struct {
		Name string `json:"name"`
	}

	router := chi.NewRouter()
	name := faker.Name().Name()
	logOut := &bytes.Buffer{}
	errMessage := "Cannot process request"
	internalErr := "cannot process request"

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		Error(r, w, Err{
			Code:    http.StatusUnprocessableEntity,
			Message: errMessage,
			Data:    metadata{name},
			Err:     errors.New(internalErr),
		})
	})

	router.
		With(requests.AttachLogger(zerolog.New(logOut).With().Logger())).
		Get("/logged", func(w http.ResponseWriter, r *http.Request) {
			Error(r, w, Err{
				Code:    http.StatusUnprocessableEntity,
				Message: errMessage,
				Data:    metadata{name},
				Err:     errors.New(internalErr),
			})
		})

	t.Run("sends jsend response", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		router.ServeHTTP(res, req)

		resp := Err{}
		err := json.Unmarshal(res.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err)
		}

		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusUnprocessableEntity, resp.Code)
		}

		if resp.Message != errMessage {
			t.Errorf("Expected error message to be %s, got %s", errMessage, resp.Message)
		}

		if resp.Err != nil {
			t.Errorf("Expected error to be discarded during serialization, got %s", resp.Err)
		}

		m := resp.Data.(map[string]interface{})
		mName := m["name"].(string)
		if mName != name {
			t.Errorf("Expected the name to be %s, got %s", name, mName)
		}
	})

	t.Run("logs response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/logged", nil)
		router.ServeHTTP(httptest.NewRecorder(), req)

		log := responseLog{}
		_ = json.Unmarshal(logOut.Bytes(), &log)

		if log.Status == 0 {
			t.Error("Expected status code to be logged")
		}

		if log.Length == 0 {
			t.Error("Expected response length to be logged")
		}

		if log.Error == "" {
			t.Error("Expected internal error to be logged")
		}

		if len(log.Headers) == 0 {
			t.Error("Expected response headers to be logged")
		}
	})
}
