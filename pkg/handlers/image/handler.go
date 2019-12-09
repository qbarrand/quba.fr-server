package image

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"gopkg.in/gographics/imagick.v2/imagick"

	"git.quba.fr/qbarrand/quba.fr-server/pkg/img"
	"git.quba.fr/qbarrand/quba.fr-server/pkg/img/cache"
)

func parseDimensions(r *http.Request) (uint, uint, error) {
	var (
		height uint
		width  uint
	)

	const (
		base = 10
		bits = 64
	)

	if heightStr := r.FormValue("height"); heightStr != "" {
		if height64, err := strconv.ParseUint(heightStr, base, bits); err != nil {
			return 0, 0, fmt.Errorf("%q: invalid height: %v", heightStr, err)
		} else {
			height = uint(height64)
		}
	}

	if widthStr := r.FormValue("width"); widthStr != "" {
		if width64, err := strconv.ParseUint(widthStr, base, bits); err != nil {
			return 0, 0, fmt.Errorf("%q: invalid width: %v", widthStr, err)
		} else {
			width = uint(width64)
		}
	}

	if height != 0 && width != 0 {
		return 0, 0, fmt.Errorf("height and width both set (%dx%d)", height, width)
	}

	return height, width, nil
}

type Handler struct {
	baseDir string
	cache   cache.Cache
	quality uint
}

func New(baseDir string, cache cache.Cache, quality uint) *Handler {
	return &Handler{
		baseDir: baseDir,
		cache:   cache,
		quality: quality,
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const mainColorHeaderName = "X-Main-Color"

	height, width, err := parseDimensions(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filename := r.URL.Path

	// Try to get the image from the cache
	key := cache.NewImageFileKey(filename, int(width), int(height), h.quality)
	cachedImgReader, metadata, err := h.cache.Get(key)
	if err == nil {
		w.Header().Set(mainColorHeaderName, metadata.MainColor())
		w.Header().Set("Etag", metadata.Hash())

		n, err := bufio.NewReader(cachedImgReader).WriteTo(w)
		if err != nil {
			log.Printf("Could not write from the cached image reader to the response: %v", err)
		}

		log.Printf("Wrote %d bytes from the cache", n)

		cachedImgReader.Close()
		return
	}

	// Otherwise, serve the image normally
	log.Printf("Could not get the image from the cache: %v", err)

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImage(filepath.Join(h.baseDir, filename)); err != nil {
		log.Print(err)
		http.NotFound(w, r)
		return
	}

	if err := img.Resize(mw, height, width, h.quality, r.FormValue("format")); err != nil {
		log.Printf("Could not resize the image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cr, cg, cb, err := img.GetMainColor(mw)
	if err != nil {
		log.Printf("Could not get the main color: %v", err)
	}

	mainColorHexRGB := fmt.Sprintf("#%02X%02X%02X", cr, cg, cb)
	w.Header().Set(mainColorHeaderName, mainColorHexRGB)

	imageBytes := mw.GetImageBlob()

	if n, err := w.Write(imageBytes); err != nil {
		log.Printf("Could not write the bytes: %v", err)
		// w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Printf("Wrote %d bytes", n)
	}

	if err := h.cache.Add(key, bytes.NewReader(imageBytes), mainColorHexRGB); err != nil {
		log.Printf("Could not add the image to the cache: %v", err)
	}
}
