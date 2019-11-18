package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"git.quba.fr/qbarrand/quba.fr-server/pkg/img"

	"gopkg.in/gographics/imagick.v2/imagick"
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

type ih struct {
	mw      *imagick.MagickWand
	quality uint
}

func (i ih) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	height, width, err := parseDimensions(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := img.Resize(i.mw, height, width, i.quality, r.FormValue("format")); err != nil {
		log.Printf("Could not resize the image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cr, cg, cb, err := img.GetMainColor(i.mw)

	if err != nil {
		log.Printf("Could not get the main color: %v", err)
	} else {
		hexRGB := fmt.Sprintf("#%02X%02X%02X", cr, cg, cb)
		w.Header().Set("X-Quba-MainColor", hexRGB)
	}

	if n, err := w.Write(i.mw.GetImageBlob()); err != nil {
		log.Printf("Could not write the bytes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Printf("Wrote %d bytes", n)
	}
}
