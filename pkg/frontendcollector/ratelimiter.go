package frontendcollector

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type getTimeFn func() time.Time

func rateLimit(rps, burst int, getTime getTimeFn, next http.Handler) http.Handler {
	l := rate.NewLimiter(rate.Limit(rps), burst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.AllowN(getTime(), 1) {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, "rate limit exceeded. try again soon")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
