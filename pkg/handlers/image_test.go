//go:generate mockgen -package mock_handlers -source image_controller.go -destination mock_handlers/mock_image_controller.go imageController

package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"git.quba.fr/qbarrand/quba.fr-server/pkg/handlers/mock_handlers"
)

func TestImage(t *testing.T) {
	if NewImage("", 0) == nil {
		t.Fatal("Should not return nil")
	}
}

func TestImage_ServeHTTP(t *testing.T) {
	t.Run("no accept: HTTP 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existent-file.jpg", nil)
		w := httptest.NewRecorder()

		NewImage("testdata", 80).ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})

	t.Run("non-existing file: HTTP 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existent-file.jpg", nil)
		req.Header.Set("Accept", "image/jpeg")

		w := httptest.NewRecorder()

		NewImage("testdata", 80).ServeHTTP(w, req)

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

		i := NewImage("testdata", 80)
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
	})

	t.Run("Resize to 1920w and Accept: image/webp", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/gopher_biplane.jpg?width=1920",
			nil)
		req.Header.Set("Accept", "image/webp")

		w := httptest.NewRecorder()

		const (
			quality = 50
			width   = 1920
		)

		tempDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("Could not create a temporary cache directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		controller := gomock.NewController(t)
		mockIC := mock_handlers.NewMockimageController(controller)

		i := NewImage("testdata", quality)
		i.imageControllerCtor = func(string) (imageController, error) {
			return mockIC, nil
		}

		gomock.InOrder(
			mockIC.EXPECT().Resize(uint(0), uint(width)),
			mockIC.EXPECT().SetQuality(uint(quality)),
			mockIC.EXPECT().Convert("webp"),
			mockIC.EXPECT().MainColor().Return(uint(0), uint(0), uint(0), nil),
			mockIC.EXPECT().Bytes(),
			mockIC.EXPECT().Destroy(),
		)

		i.ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})
}
