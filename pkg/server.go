package pkg

import (
	"net/http"

	"gopkg.in/gographics/imagick.v2/imagick"

	"git.quba.fr/qbarrand/quba.fr-server/pkg/handlers"
)

func StartServer(addr, dir string, quality uint) error {
	imagick.Initialize()
	defer imagick.Terminate()

	s, err := handlers.Sitemap(dir)
	if err != nil {
		return err
	}

	http.Handle("/sitemap.xml", handlers.Logger(s))

	http.Handle("/",
		handlers.Logger(
			handlers.Image(
				dir,
				quality,
				http.FileServer(http.Dir(dir)),
			),
		),
	)

	http.Handle("/health",
		handlers.Logger(
			handlers.Health(),
		),
	)

	return http.ListenAndServe(addr, nil)
}
