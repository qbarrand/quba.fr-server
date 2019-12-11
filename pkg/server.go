package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/gographics/imagick.v2/imagick"

	"git.quba.fr/qbarrand/quba.fr-server/pkg/handlers"
	"git.quba.fr/qbarrand/quba.fr-server/pkg/img/cache"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL.String())

		next.ServeHTTP(w, req)
	})
}

func StartServer(addr, dir string, quality uint, cacheDir string) error {
	if cacheDir == "" {
		log.Printf("cacheDir not specified; using a random temporary directory")

		tempDir, err := ioutil.TempDir("", "quba.fr-")
		if err != nil {
			return fmt.Errorf("could not get a temporary directory: %v", err)
		}

		cacheDir = tempDir

		log.Printf("Using %q as the cache directory", cacheDir)
	}

	imagick.Initialize()
	defer imagick.Terminate()

	r := mux.NewRouter().Methods(http.MethodGet).Subrouter()

	r.Use(Logger)

	r.Handle("/health", handlers.Health())

	sitemapHandler, err := handlers.Sitemap(dir)
	if err != nil {
		return err
	}

	r.Handle("/sitemap.xml", sitemapHandler)

	imageHandler := handlers.Image(dir, cache.FsCache(cacheDir), quality)

	r.PathPrefix("/").
		HeadersRegexp("Accept", "image/(ico|jpeg|jxr|png|webp)").
		Handler(imageHandler)

	r.PathPrefix("/").
		Handler(http.FileServer(http.Dir(dir)))

	http.Handle("/", r)

	return http.ListenAndServe(addr, nil)
}
