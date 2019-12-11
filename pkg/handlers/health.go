package handlers

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const secondsBetweenChecks = 120

type healthCache struct {
	lastCheck time.Time
	value     bool

	m sync.Mutex
}

type health struct {
	cache *healthCache
}

func Health() http.Handler {
	return &health{
		cache: &healthCache{},
	}
}

func (h *health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const fqdn = "ping.quba.fr"

	log.Printf("Last DNS lookup: %v", h.cache.lastCheck.String())

	now := time.Now()

	if now.Sub(h.cache.lastCheck) <= secondsBetweenChecks*time.Second {
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

		h.cache.m.Lock()

		if got != expected {
			log.Printf("Expected %s, got %s", expected, got)
			h.cache.value = false
		} else {
			h.cache.value = true
		}

		h.cache.lastCheck = now

		h.cache.m.Unlock()
	}

	if !h.cache.value {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
