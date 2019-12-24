package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	img "git.quba.fr/qbarrand/quba.fr-server/pkg/image"
)

func getPreferredIMFormat(accept string) string {
	for _, MIMEType := range strings.Split(accept, ",") {
		MIMEType = strings.Trim(MIMEType, " ")

		switch MIMEType {
		case "image/jpeg":
			return "jpg"
		case "image/webp":
			return "webp"
		case "image/vnd.ms-photo", "image/jxr":
			return "jxr"
		}
	}

	return ""
}

// func mimeToIMFormat(mimeType string) (string, error) {
// 	switch mimeType {
// 	case "image/jpeg":
// 		return "jpg", nil
// 	case "image/webp":
// 		return "webp", nil
// 	case "image/vnd.ms-photo", "image/jxr":
// 		return "jxr", nil
// 	default:
// 		return "", fmt.Errorf("%q: unhandled MIME type", mimeType)
// 	}
// }

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

type Image struct {
	baseDir             string
	bytesHasher         func([]byte) (string, error)
	imageControllerCtor func(string) (imageController, error)
	quality             uint
}

func NewImage(baseDir string, quality uint) *Image {
	imageProcessorCtor := func(path string) (imageController, error) {
		p, err := img.NewImagickProcessor(path)
		if err != nil {
			return nil, err
		}

		return imageController(p), nil
	}

	return &Image{
		baseDir:             baseDir,
		bytesHasher:         hashBytes,
		imageControllerCtor: imageProcessorCtor,
		quality:             quality,
	}
}

func (i Image) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the requested dimensions
	height, width, err := parseDimensions(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Find the Image
	accept := r.Header.Get("Accept")

	log.Print("Accept: " + accept)

	imFormat := getPreferredIMFormat(accept)
	if imFormat == "" {
		log.Printf("No accepted format among %q", accept)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filePath := r.URL.Path
	imagePath := filepath.Join(i.baseDir, filePath)

	p, err := i.imageControllerCtor(imagePath)
	if err != nil {
		log.Printf("could not create the image controller: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer p.Destroy()

	log.Printf("ImageMagick format: %q", imFormat)

	if height != 0 || width != 0 {
		if err := p.Resize(height, width); err != nil {
			log.Printf("Could not resize the Image: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if err := p.SetQuality(i.quality); err != nil {
		log.Printf("Could not set the quality to %d: %v", i.quality, err)
	}

	if err := p.Convert(imFormat); err != nil {
		log.Printf("Could not convert to %q: %v", imFormat, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cr, cg, cb, err := p.MainColor()
	if err != nil {
		log.Printf("Could not get the main color: %v", err)
	}

	imageBytes := p.Bytes()

	hash, err := i.bytesHasher(imageBytes)
	if err != nil {
		log.Printf("Could not hash the reponse bytes: %v", err)
	} else if r.Header.Get("If-None-Match") == hash {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	headers := w.Header()
	headers.Set("ETag", hash)
	headers.Set("X-Main-Color", fmt.Sprintf("#%02X%02X%02X", cr, cg, cb))
	headers.Set("Content-Length", strconv.Itoa(len(imageBytes)))

	if n, err := w.Write(imageBytes); err != nil {
		log.Printf("could not write the reply: %v", err)
	} else {
		log.Printf("Wrote %d bytes", n)
	}
}
