package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/gographics/imagick.v2/imagick"
)

func loggerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func imageHandler(baseDir string, quality uint, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".jpg") {
			// Not an image
			next.ServeHTTP(w, r)
			return
		}

		const (
			base = 10
			bits = 64
		)

		var (
			height uint
			width  uint
		)

		writeBadRequest := func(name, value string) {
			log.Printf("%s: invalid value %q\n", name, value)
			w.WriteHeader(http.StatusBadRequest)
		}

		if heightStr := r.FormValue("height"); heightStr != "" {
			if height64, err := strconv.ParseUint(heightStr, base, bits); err != nil {
				writeBadRequest("height", heightStr)
				return
			} else {
				height = uint(height64)
			}
		}

		if widthStr := r.FormValue("width"); widthStr != "" {
			if width64, err := strconv.ParseUint(widthStr, base, bits); err != nil {
				writeBadRequest("width", widthStr)
				return
			} else {
				width = uint(width64)
			}
		}

		if height != 0 && width != 0 {
			log.Printf("height %d, width %d", height, width)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mw := imagick.NewMagickWand()

		path := filepath.Join(baseDir, r.URL.Path)

		log.Println("Reading " + path)

		if err := mw.ReadImage(path); err != nil {
			log.Printf("Could not open %s: %v", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//
		// Sampling factor
		//

		// if err := mw.SetSamplingFactors([]float64{4, 2, 0}); err != nil {
		// 	log.Printf("Could not set the sampling factors: %v", err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		//
		// Resizing
		//

		if height != 0 || width != 0 {
			oHeight := mw.GetImageHeight()
			oWidth := mw.GetImageWidth()

			if width != 0 {
				ratio := float64(width) / float64(oWidth)
				height = uint(float64(oHeight) * ratio)
				goto resize
			}

			if height != 0 {
				ratio := float64(height) / float64(oHeight)
				width = uint(float64(oWidth) * ratio)
				goto resize
			}

		resize:
			log.Printf("Resizing to %dx%d", width, height)

			if err := mw.AdaptiveResizeImage(width, height); err != nil {
				log.Printf("Could not resize the image to %dx%d: %v", width, height, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		//
		// Quality
		//

		currentQuality := mw.GetImageCompressionQuality()

		if quality < currentQuality {
			log.Printf("Lowering the quality from %d to %d", currentQuality, quality)

			if err := mw.SetImageCompressionQuality(quality); err != nil {
				log.Printf("Could not set the quality to %d: %v", quality, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		//
		// Strip EXIF data
		//

		if err := mw.StripImage(); err != nil {
			log.Printf("Could not strip metadata: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//
		// Interlace
		//

		if err := mw.SetInterlaceScheme(imagick.INTERLACE_JPEG); err != nil {
			log.Printf("Could not set the interlace method: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//
		// Color space
		//

		if err := mw.SetColorspace(imagick.COLORSPACE_SRGB); err != nil {
			log.Printf("Could not set the color space: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if r.FormValue("format") == "webp" {
			if err := mw.SetFormat("webp"); err != nil {
				log.Printf("Could not set the format to webp: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if n, err := w.Write(mw.GetImageBlob()); err != nil {
			log.Printf("Could not write the bytes: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			log.Printf("Wrote %d bytes", n)
		}
	})
}

func startServer(addr, dir string, quality uint) error {
	imagick.Initialize()
	defer imagick.Terminate()

	http.Handle("/",
		loggerHandler(
			imageHandler(
				dir,
				quality,
				http.FileServer(http.Dir(dir)),
			),
		),
	)

	return http.ListenAndServe(addr, nil)
}
