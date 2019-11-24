package handlers

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type healthCache struct {
	lastCheck time.Time
	value     bool

	m sync.Mutex
}

func Health() http.Handler {
	const secondsBetweenChecks = 120

	c := &healthCache{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const fqdn = "ping.quba.fr"

		log.Printf("Last DNS lookup: %v", c.lastCheck.String())

		now := time.Now()

		if now.Sub(c.lastCheck) <= secondsBetweenChecks*time.Second {
			log.Print("Using cache")
		} else {
			log.Printf("Older than %d seconds; cache invalidated", secondsBetweenChecks)

			records, err := net.LookupTXT(fqdn)
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if len(records) != 1 {
				log.Printf("%s/TXT: not enough records", fqdn)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			const expected = "quentin@quba.fr"
			got := records[0]

			c.m.Lock()

			if got != expected {
				log.Printf("Expected %s, got %s", expected, got)
				c.value = false
			} else {
				c.value = true
			}

			c.lastCheck = now

			c.m.Unlock()
		}

		if !c.value {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

func Image(baseDir string, quality uint, next http.Handler) http.Handler {
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
		defer mw.Destroy()

		if err := mw.ReadImage(filepath.Join(baseDir, r.URL.Path)); err != nil {
			log.Print(err)
			http.NotFound(w, r)
			return
		}

		ih{mw: mw, quality: quality}.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func Sitemap(dir string) (http.Handler, error) {
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
