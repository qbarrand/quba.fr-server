package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestImage(t *testing.T) {
	if Image("", nil, 0) == nil {
		t.Fatal("Should not return nil")
	}
}

func TestImage_ServeHTTP(t *testing.T) {
	t.Run("no accept: HTTP 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existent-file.jpg", nil)
		w := httptest.NewRecorder()

		Image("testdata", nil, 80).ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("non-existing file: HTTP 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existent-file.jpg", nil)
		req.Header.Set("Accept", "image/jpeg")

		w := httptest.NewRecorder()

		Image("testdata", nil, 80).ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("Accept: image/webp", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/gopher_biplane.jpg", nil)
		req.Header.Set("Accept", "image/webp")

		w := httptest.NewRecorder()

		Image("testdata", nil, 80).ServeHTTP(w, req)

		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("wd: %s", wd)

		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})
}
