package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"

	"git.quba.fr/qbarrand/quba.fr-server/pkg/handlers/mock_handlers"
)

func TestImage(t *testing.T) {
	if NewImage("", nil, 0) == nil {
		t.Fatal("Should not return nil")
	}
}

func TestImage_ServeHTTP(t *testing.T) {
	t.Run("no accept: HTTP 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existent-file.jpg", nil)
		w := httptest.NewRecorder()

		NewImage("testdata", nil, 80).ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("non-existing file: HTTP 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existent-file.jpg", nil)
		req.Header.Set("Accept", "image/jpeg")

		w := httptest.NewRecorder()

		NewImage("testdata", nil, 80).ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("Accept: image/webp", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/gopher_biplane.jpg", nil)
		req.Header.Set("Accept", "image/webp")

		w := httptest.NewRecorder()

		c := gomock.NewController(t)
		m := mock_handlers.NewMockimageController(c)

		i := NewImage("testdata", nil, 80)
		i.imageControllerCtor = func(string) (imageController, error) {
			return m, nil
		}

		gomock.InOrder(
			m.EXPECT().Resize(uint(0), uint(0)),
			m.EXPECT().SetQuality(uint(80)),
			m.EXPECT().Convert("webp"),
			m.EXPECT().MainColor(),
			m.EXPECT().Bytes(),
			m.EXPECT().Destroy(),
		)

		i.ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}

		// resize := m.EXPECT().Resize(0, 0)
		// convert := m.EXPECT().Convert("webp").After(resize)
		// mainColor := m.EXPECT().MainColor().After(convert)
		// bytes := m.EXPECT().MainColor().After(mainColor)
		// m.EXPECT().Destroy().After(bytes)
	})
}
