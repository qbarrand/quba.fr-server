package handlers

import "testing"

func TestSitemap(t *testing.T) {
	s, err := Sitemap("/random/path")
	if err != nil {
		t.Fatal(err)
	}

	if s == nil {
		t.Fatal("Should not be nil")
	}
}
