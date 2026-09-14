package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/middleware"
	"github.com/go-chi/chi/v5"
)

func TestRequestTimeoutRejectsSlowHandlers(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(10 * time.Millisecond))
	router.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/slow", nil))

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusGatewayTimeout)
	}

	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if body.Success || body.Error.Code != "GATEWAY_TIMEOUT" {
		t.Errorf("unexpected error envelope: %+v", body)
	}
}

func TestRequestTimeoutPassesFastHandlers(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(time.Second))
	router.Get("/fast", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/fast", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != "ok" {
		t.Errorf("got body %q, want %q", recorder.Body.String(), "ok")
	}
}

func TestRequestTimeoutSkipsEventStreamRequests(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(10 * time.Millisecond))
	router.Get("/notifications", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	request.Header.Set("Accept", "text/event-stream")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("SSE requests must not be cancelled by the timeout middleware: got status %d", recorder.Code)
	}
}

func TestRequestTimeoutSkipsStreamPathWithoutAcceptHeader(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(10 * time.Millisecond))
	router.Get("/notifications/stream", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/notifications/stream", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("routes ending in /stream must be exempt: got status %d", recorder.Code)
	}
}

func TestRequestTimeoutWithCustomSkip(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeoutWithSkip(10*time.Millisecond, func(r *http.Request) bool {
		return r.URL.Path == "/reports/export"
	}))
	router.Get("/reports/export", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/reports/export", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("skipped route must run to completion: got status %d", recorder.Code)
	}
}

func TestRequestTimeoutDoesNotOverrideCommittedResponse(t *testing.T) {
	committed := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})

	router := chi.NewRouter()
	// Generous deadline: the behaviour under test is what happens *after* the
	// handler already committed a response, not the timeout itself.
	router.Use(middleware.RequestTimeout(100 * time.Millisecond))
	router.Get("/streaming", func(w http.ResponseWriter, r *http.Request) {
		defer close(finished)

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("partial"))
		close(committed)

		// Hold the response open past the deadline, then try a late write that
		// the middleware must discard.
		<-release
		_, _ = w.Write([]byte("late"))
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/streaming", nil))

	// ServeHTTP has returned, so the deadline has been processed. Let the
	// abandoned handler finish before asserting, to keep the test race free.
	<-committed
	close(release)
	<-finished

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("committed status must be preserved: got %d, want %d", recorder.Code, http.StatusAccepted)
	}
	if body := recorder.Body.String(); body != "partial" {
		t.Errorf("late writes must be discarded: got body %q, want %q", body, "partial")
	}
}

func TestRequestTimeoutPropagatesPanics(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(time.Second))
	router.Get("/boom", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	defer func() {
		if recover() == nil {
			t.Error("expected the panic to propagate to the outer recoverer")
		}
	}()

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/boom", nil))
}

func TestRequestTimeoutDisabledWhenNonPositive(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(0))
	router.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/slow", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("a non-positive timeout must disable the middleware: got status %d", recorder.Code)
	}
}

func TestRequestTimeoutPropagatesDeadlineToContext(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(50 * time.Millisecond))
	router.Get("/ctx", func(w http.ResponseWriter, r *http.Request) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Error("expected the request context to carry a deadline")
			return
		}
		if until := time.Until(deadline); until <= 0 || until > 50*time.Millisecond {
			t.Errorf("deadline must fall within the configured timeout, got %s", until)
		}
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ctx", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestRequestTimeoutWriterSupportsFlushing(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(time.Second))
	router.Get("/flush", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("the timeout writer must implement http.Flusher")
			return
		}
		w.WriteHeader(http.StatusOK)
		flusher.Flush()
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/flush", nil))

	if !recorder.Flushed {
		t.Error("expected the flush to reach the underlying ResponseWriter")
	}
}

func TestRequestTimeoutRejectsWritesAfterDeadline(t *testing.T) {
	release := make(chan struct{})
	finished := make(chan struct{})
	var lateWriteErr error

	router := chi.NewRouter()
	router.Use(middleware.RequestTimeout(20 * time.Millisecond))
	router.Get("/late", func(w http.ResponseWriter, r *http.Request) {
		defer close(finished)

		// Do nothing until the deadline has passed, then try to write.
		<-release
		_, lateWriteErr = w.Write([]byte("late"))
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/late", nil))

	close(release)
	<-finished

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusGatewayTimeout)
	}
	if !errors.Is(lateWriteErr, http.ErrHandlerTimeout) {
		t.Errorf("got write error %v, want %v", lateWriteErr, http.ErrHandlerTimeout)
	}
}