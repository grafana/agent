package utils

import (
	"github.com/minio/pkg/wildcard"
)

// URLMatchesOrigins returns true if URL matches at least one of origin prefix. Wildcard '*' and '?' supported
func URLMatchesOrigins(URL string, origins []string) bool {
	for _, origin := range origins {
		if origin == "*" || wildcard.Match(origin+"*", URL) {
			return true
		}
	}
	return false
}
