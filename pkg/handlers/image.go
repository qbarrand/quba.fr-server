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

func getPreferredIMFormat(accept string) (string, string) {
	for _, mimeType := range strings.Split(accept, ",") {
		mimeType = strings.Trim(mimeType, " ")

		switch mimeType {
		case "image/jpeg":
			return mimeType, "jpg"
		case "image/webp":
			return mimeType, "webp"
		case "image/vnd.ms-photo", "image/jxr":
			return mimeType, "jxr"
		}
	}

	return "", ""
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

	mimeType, imFormat := getPreferredIMFormat(accept)
	if imFormat == "" {
		log.Printf("No accepted format among %q", accept)
		w.WriteHeader(http.StatusNotAcceptable)
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
	headers.Set("Content-Length", strconv.Itoa(len(imageBytes)))
	headers.Set("Content-Type", mimeType)
	headers.Set("X-Date", p.ExifField("comment"))
	headers.Set("X-Location", p.ExifField("Iptc4xmpCore:Location"))
	headers.Set("X-Main-Color", fmt.Sprintf("#%02X%02X%02X", cr, cg, cb))

	if n, err := w.Write(imageBytes); err != nil {
		log.Printf("could not write the reply: %v", err)
	} else {
		log.Printf("Wrote %d bytes", n)
	}
}
