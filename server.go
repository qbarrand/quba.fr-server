package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type bgData struct {
	File     string
	Location string
	Date     string
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

	if err := resize(i.mw, height, width, i.quality, r.FormValue("format")); err != nil {
		log.Printf("Could not resize the image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if n, err := w.Write(i.mw.GetImageBlob()); err != nil {
		log.Printf("Could not write the bytes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Printf("Wrote %d bytes", n)
	}
}

func loggerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
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

func imageHandler(baseDir string, quality uint, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handledExtensions := map[string]bool{
			".jpg": true,
			".png": true,
		}

		if !handledExtensions[filepath.Ext(r.URL.Path)] {
			// Not an image
			next.ServeHTTP(w, r)
			return
		}

		mw := imagick.NewMagickWand()

		if err := mw.ReadImage(filepath.Join(baseDir, r.URL.Path)); err != nil {
			log.Print(err)
			http.NotFound(w, r)
			return
		}

		ih{mw: mw, quality: quality}.ServeHTTP(w, r)
	})
}

func randomHandler(dir string, quality uint) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fd, err := os.Open(filepath.Join(dir, "db.json"))
		if err != nil {
			log.Printf("Could not open the background database: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		d := make([]bgData, 0)

		if err := json.NewDecoder(fd).Decode(&d); err != nil {
			log.Printf("Could not decode bg.json: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		selected := d[rand.Intn(len(d))]

		mw := imagick.NewMagickWand()

		if err := mw.ReadImage(filepath.Join(dir, selected.File)); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cr, cg, cb, err := getMainColor(mw)
		if err != nil {
			log.Printf("Could not get the image's main color: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		hexRGB := fmt.Sprintf("#%02X%02X%02X", cr, cg, cb)

		w.Header().Set("X-Quba-Date", selected.Date)
		w.Header().Set("X-Quba-Location", selected.Location)
		w.Header().Set("X-Quba-MainColor", hexRGB)

		ih{mw: mw, quality: quality}.ServeHTTP(w, r)
	})
}

func sitemapHandler(dir string) (http.Handler, error) {
	const sitemapTemplateStr = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>https://quba.fr/</loc>
		<lastmod>{{ .LastMod }}</lastmod>
		<changefreq>monthly</changefreq>
		<priority>1.0</priority>
	</url>
</urlset>`

	sitemapTempate, err := template.New("sitemap").Parse(sitemapTemplateStr)
	if err != nil {
		return nil, fmt.Errorf("Could not parse the sitemap template: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command("git", "log", "-1", "--format=%ad", "--date=iso-strict")
		cmd.Dir = dir

		out, err := cmd.Output()
		if err != nil {
			log.Printf("Error while running git: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		date := strings.TrimSuffix(string(out), "\n")

		sitemapTempate.Execute(w, struct{ LastMod string }{date})
	})

	return handler, nil
}

func startServer(addr, dir string, quality uint) error {
	imagick.Initialize()
	defer imagick.Terminate()

	s, err := sitemapHandler(dir)
	if err != nil {
		return err
	}

	http.Handle("/sitemap.xml", loggerHandler(s))

	http.Handle("/",
		loggerHandler(
			imageHandler(
				dir,
				quality,
				http.FileServer(http.Dir(dir)),
			),
		),
	)

	imgDir := filepath.Join(dir, "images", "bg")

	http.Handle("/images/bg/random", loggerHandler(randomHandler(imgDir, quality)))

	return http.ListenAndServe(addr, nil)
}
