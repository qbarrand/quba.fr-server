//go:generate mockgen -package mock_handlers -source image_controller.go -destination mock_handlers/mock_image_controller.go imageController
//go:generate mockgen -package mock_handlers -source ../image/cache/interfaces.go -destination mock_handlers/mock_cache.go Cache

package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"git.quba.fr/qbarrand/quba.fr-server/pkg/handlers/mock_handlers"
	"git.quba.fr/qbarrand/quba.fr-server/pkg/image/cache"
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
	})

	//t.Run("Resize to width=1920 and Accept: image/webp", func(t *testing.T) {
	//	req := httptest.NewRequest(
	//		http.MethodGet,
	//		"/gopher_biplane.jpg?width=1920",
	//		nil)
	//	req.Header.Set("Accept", "image/webp")
	//
	//	w := httptest.NewRecorder()
	//
	//	controller := gomock.NewController(t)
	//	mockIC := mock_handlers.NewMockimageController(controller)
	//	mockC := mock_handlers.NewMockCache(controller)
	//
	//	const quality = 50
	//
	//	tempDir, err := ioutil.TempDir("", "")
	//	if err != nil {
	//		t.Fatalf("Could not create a temporary cache directory: %v", err)
	//	}
	//	defer os.RemoveAll(tempDir)
	//
	//	i := NewImage("testdata", cache.New(tempDir), quality)
	//	i.imageControllerCtor = func(string) (imageController, error) {
	//		return mockIC, nil
	//	}
	//
	//	key := cache.NewImageFileKey("gopher_biplane.jpg", 1920, 0, quality, "webp")
	//
	//	gomock.InOrder(
	//		mockC.EXPECT().Add(key, gomock.Any(), "000000", "lulz"),
	//		mockIC.EXPECT().Resize(uint(0), uint(1920)),
	//		mockIC.EXPECT().SetQuality(uint(quality)),
	//		mockIC.EXPECT().Convert("webp"),
	//		mockIC.EXPECT().MainColor(),
	//		mockIC.EXPECT().Bytes(),
	//		mockC.EXPECT().Add(key, gomock.Any(), "000000", "lulz"),
	//		mockIC.EXPECT().Destroy(),
	//	)
	//
	//	i.ServeHTTP(w, req)
	//
	//	res := w.Result()
	//
	//	if res.StatusCode != http.StatusOK {
	//		t.Fatalf("Got HTTP %d", res.StatusCode)
	//	}
	//})

	t.Run("Try getting from the cache", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/gopher_biplane.jpg?width=1920",
			nil)
		req.Header.Set("Accept", "image/webp")

		w := httptest.NewRecorder()

		const quality = 50

		tempDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("Could not create a temporary cache directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		controller := gomock.NewController(t)
		mockIC := mock_handlers.NewMockimageController(controller)
		mockC := mock_handlers.NewMockCache(controller)

		i := NewImage("testdata", cache.New(tempDir), quality)
		i.imageControllerCtor = func(string) (imageController, error) {
			return mockIC, nil
		}

		key := cache.NewImageFileKey("gopher_biplane.jpg", 1920, 0, quality, "webp")

		gomock.InOrder(
			mockC.EXPECT().Get(key),
			mockIC.EXPECT().Resize(uint(0), uint(1920)),
			mockIC.EXPECT().SetQuality(uint(quality)),
			mockIC.EXPECT().Convert("webp"),
			mockIC.EXPECT().MainColor(),
			mockIC.EXPECT().Bytes(),
			mockC.EXPECT().Add(key, gomock.Any(), "000000", "lulz"),
			mockIC.EXPECT().Destroy(),
		)

		i.ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}
	})
}
