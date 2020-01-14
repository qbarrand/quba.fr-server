//go:generate mockgen -package mock_handlers -source image_controller.go -destination mock_handlers/mock_image_controller.go imageController

package handlers

import (
	"net/http"
	"net/http/httptest"
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
	checkContentLength := func(t *testing.T, res *http.Response) {
		if res.Header.Get("Content-Length") == "" {
			t.Fatal("Content-Length undefined")
		}
	}

	checkContentType := func(t *testing.T, res *http.Response, expected string) {
		got := res.Header.Get("Content-Type")

		if got != expected {
			t.Fatalf("Unexpected Content-Type: expected %q, got %q", expected, got)
		}
	}

	t.Run("no accept: HTTP 406", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existent-file.jpg", nil)
		w := httptest.NewRecorder()

		NewImage("testdata", 80).ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusNotAcceptable {
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
			m.EXPECT().SetQuality(uint(80)),
			m.EXPECT().Convert("webp"),
			m.EXPECT().MainColor(),
			m.EXPECT().Bytes(),
			m.EXPECT().ExifField("comment"),
			m.EXPECT().ExifField("Iptc4xmpCore:Location"),
			m.EXPECT().Destroy(),
		)

		i.ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}

		checkContentLength(t, res)
		checkContentType(t, res, "image/webp")
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
			mockIC.EXPECT().ExifField("comment"),
			mockIC.EXPECT().ExifField("Iptc4xmpCore:Location"),
			mockIC.EXPECT().Destroy(),
		)

		i.ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}

		checkContentLength(t, res)
		checkContentType(t, res, "image/webp")
	})

	t.Run("Resize to 1920w and Accept: image/jpeg", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/gopher_biplane.jpg?width=1920",
			nil)
		req.Header.Set("Accept", "image/jpeg")

		w := httptest.NewRecorder()

		const (
			quality = 50
			width   = 1920
		)

		controller := gomock.NewController(t)
		mockIC := mock_handlers.NewMockimageController(controller)

		i := NewImage("testdata", quality)
		i.imageControllerCtor = func(string) (imageController, error) {
			return mockIC, nil
		}

		gomock.InOrder(
			mockIC.EXPECT().Resize(uint(0), uint(width)),
			mockIC.EXPECT().SetQuality(uint(quality)),
			mockIC.EXPECT().Convert("jpg"),
			mockIC.EXPECT().MainColor().Return(uint(0), uint(0), uint(0), nil),
			mockIC.EXPECT().Bytes(),
			mockIC.EXPECT().ExifField("comment"),
			mockIC.EXPECT().ExifField("Iptc4xmpCore:Location"),
			mockIC.EXPECT().Destroy(),
		)

		i.ServeHTTP(w, req)

		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Got HTTP %d", res.StatusCode)
		}

		checkContentLength(t, res)
		checkContentType(t, res, "image/jpeg")
	})
}

func Test_getPreferredHeader(t *testing.T) {
	cases := []struct {
		input        string
		expectedMIME string
		expectedIM   string
	}{
		{
			input:        "image/jpeg,image/webp",
			expectedMIME: "image/jpeg",
			expectedIM:   "jpg",
		},
		{
			input:        "image/webp,image/jpeg",
			expectedMIME: "image/webp",
			expectedIM:   "webp",
		},
		{
			input:        "a,image/jpeg",
			expectedMIME: "image/jpeg",
			expectedIM:   "jpg",
		},
		{
			input:        "a,b,text/plain,image/jpeg",
			expectedMIME: "image/jpeg",
			expectedIM:   "jpg",
		},
		{
			input:        "a,b,c",
			expectedMIME: "",
			expectedIM:   "",
		},
		{
			input:        "",
			expectedMIME: "",
			expectedIM:   "",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			mimeType, imFormat := getPreferredIMFormat(c.input)

			if mimeType != c.expectedMIME {
				t.Fatalf("Unexpected MIME type: expected %q, got %q", c.expectedMIME, mimeType)
			}

			if imFormat != c.expectedIM {
				t.Fatalf("Unexpected IM format: expected %q, got %q", c.expectedIM, imFormat)
			}
		})
	}
}
