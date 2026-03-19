package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// App holds the provider clients and the config sequence.
type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
}

// ServeHTTP: GET /?count=N&offset=M → JSON array of content items.
func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	count, offset, ok := parseCountOffset(req)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userIP := userIPFromRequest(req)
	items := FetchContent(a.Config, a.ContentClients, userIP, count, offset)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(items)
}

// parseCountOffset reads and validates query params. ok false => invalid.
// Missing offset is treated as 0.
func parseCountOffset(req *http.Request) (count, offset int, ok bool) {
	q := req.URL.Query()
	count, err := strconv.Atoi(q.Get("count"))
	if err != nil || count < 1 {
		return 0, 0, false
	}
	offsetStr := q.Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0, 0, false
	}
	return count, offset, true
}

func userIPFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	if host != "" {
		return host
	}
	return req.RemoteAddr
}
