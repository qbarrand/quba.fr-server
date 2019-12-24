package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	if Health() == nil {
		t.Fatal("Should not return nil")
	}
}

func TestHealth_ServeHTTP(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)

	t.Run("empty DNS records: HTTP 500", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler := &health{
			cache: &healthCache{},
			dnsQueryier: func(_ string) ([]string, error) {
				return nil, nil
			},
		}

		handler.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("too many DNS records: HTTP 500", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler := &health{
			cache: &healthCache{},
			dnsQueryier: func(_ string) ([]string, error) {
				return []string{"a", "b"}, nil
			},
		}

		handler.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("DNS query error: HTTP 500", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler := &health{
			cache: &healthCache{},
			dnsQueryier: func(_ string) ([]string, error) {
				return nil, errors.New("whatever")
			},
		}

		handler.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("should query ping.quba.fr", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler := &health{
			cache: &healthCache{},
			dnsQueryier: func(q string) ([]string, error) {
				if q != "ping.quba.fr" {
					t.Fatalf("Queried %s", q)
				}

				return nil, nil
			},
		}

		handler.ServeHTTP(w, req)
	})

	t.Run("unexpected TXT contents: HTTP 500", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler := &health{
			cache: &healthCache{},
			dnsQueryier: func(_ string) ([]string, error) {
				return []string{"test"}, nil
			},
		}

		handler.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("expected TXT contents: HTTP 200", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler := &health{
			cache: &healthCache{},
			dnsQueryier: func(_ string) ([]string, error) {
				return []string{"quentin@quba.fr"}, nil
			},
		}

		handler.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})
}
