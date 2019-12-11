package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"text/template"
)

type sitemap struct {
	dir      string
	template *template.Template
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

	template, err := template.New("sitemap").Parse(sitemapTemplateStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse the sitemap template: %v", err)
	}

	h := sitemap{
		dir:      dir,
		template: template,
	}

	return h, err
}

func (s sitemap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("git", "log", "-1", "--format=%ad", "--date=iso-strict")
	cmd.Dir = s.dir

	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error while running git: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	date := strings.TrimSuffix(string(out), "\n")

	w.Header().Set("Content-Type", "application/xml")

	if err := s.template.Execute(w, struct{ LastMod string }{date}); err != nil {
		log.Printf("could not render the template to the reply: %v", err)
	}
}
