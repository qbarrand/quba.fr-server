package pkg

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/gographics/imagick.v2/imagick"

	"git.quba.fr/qbarrand/quba.fr-server/pkg/handlers"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL.String())

		next.ServeHTTP(w, req)
	})
}

func StartServer(addr, dir string, quality uint) error {
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

	imageHandler := handlers.NewImage(dir, quality)

	r.PathPrefix("/").
		HeadersRegexp("Accept", "image/(ico|jpeg|jxr|png|webp)").
		Handler(imageHandler)

	return http.ListenAndServe(addr, r)
}
