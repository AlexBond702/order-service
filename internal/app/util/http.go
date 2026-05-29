package util

import (
	"net/http"
	"strings"
)

func IsFilteredHttpRoute(r *http.Request) bool {
	path := r.RequestURI
	switch {
	case strings.Contains(path, "health"):
		return true
	case strings.Contains(path, "metrics"):
		return true
	case strings.Contains(path, "debug"):
		return true
	default:
		return false
	}
}
