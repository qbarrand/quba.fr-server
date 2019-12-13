package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/gographics/imagick.v2/imagick"

	img "git.quba.fr/qbarrand/quba.fr-server/pkg/image"
	"git.quba.fr/qbarrand/quba.fr-server/pkg/image/cache"
)

func mimeToIMFormat(mimeType string) (string, error) {
	switch mimeType {
	case "image/jpeg":
		return "jpg", nil
	case "image/webp":
		return "webp", nil
	case "image/vnd.ms-photo", "image/jxr":
		return "jxr", nil
	default:
		return "", fmt.Errorf("%q: unhandled MIME type", mimeType)
	}
}

func imFormatToMIME(imFormat string) (string, error) {
	switch imFormat {
	case "jpg":
		return "image/jpeg", nil
	case "jxr":
		return "image/jxr", nil
	case "webp":
		return "image/webp", nil
	default:
		return "", fmt.Errorf("%q: unhandled ImageMagick format", imFormat)
	}
}

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

type image struct {
	baseDir       string
	cache         cache.Cache
	processorCtor func(string) (img.Processor, error)
	quality       uint
}

func Image(baseDir string, cache cache.Cache, quality uint) http.Handler {
	processorCtor := func(path string) (img.Processor, error) {
		p, err := img.NewImagickProcessor(path)
		if err != nil {
			return nil, err
		}

		return img.Processor(p), nil
	}

	return &image{
		baseDir:       baseDir,
		cache:         cache,
		processorCtor: processorCtor,
		quality:       quality,
	}
}

func (i image) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the requested dimensions
	height, width, err := parseDimensions(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the requested format
	accept := r.Header.Get("Accept")

	if accept == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Print("Accept: " + accept)

	imFormat := ""

	if accept != "" {
		for _, MIMEType := range strings.Split(accept, ",") {
			MIMEType = strings.Trim(MIMEType, " ")

			var err error

			if imFormat, err = mimeToIMFormat(MIMEType); err == nil {
				break
			}
		}
	}

	filePath := r.URL.Path
	ifNoneMatchHeader := r.Header.Get("If-None-Match")

	// Try to get the image from the cache
	key := cache.NewImageFileKey(filePath, int(width), int(height), i.quality, imFormat)

	if i.cache == nil {
		log.Print("Cache not used")
	} else {
		cachedImgReader, metadata, err := i.cache.Get(key)

		_ = cachedImgReader
		_ = metadata

		if err == nil {
			// TODO
		}
	}

	// Otherwise, serve the image normally
	log.Printf("Could not get the image from the cache: %v", err)

	imagePath := filepath.Join(i.baseDir, filePath)

	p, err := i.processorGetter(imagePath)
	defer p.Destroy()

	if err := mw.ReadImage(filepath.Join(i.baseDir, filePath)); err != nil {
		log.Print(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("ImageMagick format: %q", imFormat)

	if err := image.Resize(mw, height, width, i.quality, imFormat); err != nil {
		log.Printf("Could not resize the image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cr, cg, cb, err := image.GetMainColor(mw)
	if err != nil {
		log.Printf("Could not get the main color: %v", err)
	}

	imageBytes := mw.GetImageBlob()

	hash, err := cache.HashBytes(imageBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not hash the file: %v", err)
		return
	}

	if hash == ifNoneMatchHeader {
		log.Printf("cached")
	}

	mainColorHexRGB := fmt.Sprintf("#%02X%02X%02X", cr, cg, cb)

	rd := bytes.NewReader(imageBytes)

	if err := writeFromStream(w, rd, hash, mainColorHexRGB); err != nil {
		log.Print(err)
	}

	if _, err := rd.Seek(0, 0); err != nil {
		log.Printf("Could not Seek() to the beginning of the file: %v", err)
		return
	}

	if err := i.cache.Add(key, rd, mainColorHexRGB, hash); err != nil {
		log.Printf("Could not add the image to the cache: %v", err)
	}
}

func writeFromStream(w http.ResponseWriter, r io.WriterTo, ETag, mainColor string) error {
	headers := w.Header()

	headers.Set("ETag", ETag)
	headers.Set("X-Main-Color", mainColor)

	n, err := r.WriteTo(w)
	if err != nil {
		return fmt.Errorf("could not write the reply: %v", err)
	}

	log.Printf("Wrote %d bytes", n)

	return nil
}

//func serveFromCache(w http.ResponseWriter, cachedImg io.Reader, metadata cache.Metadata) error {
//	cacheEtag := metadata.Hash()
//
//	if ifNoneMatchHeader == cacheEtag {
//		w.WriteHeader(http.StatusNotModified)
//	} else {
//		log.Printf("Rendering %s from the cache", filePath)
//
//		if err := writeFromStream(w, bufio.NewReader(cachedImgReader), cacheEtag, metadata.MainColor()); err != nil {
//			log.Print(err)
//			return
//		}
//
//		if err := cachedImgReader.Close(); err != nil {
//			log.Printf("Could not close the cached file: %v", err)
//		}
//	}
//
//	return
//}
