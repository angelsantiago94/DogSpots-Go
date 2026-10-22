package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/angellsan94/dogspots-api/internal/auth"
	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
)

type rateLimitEntry struct {
	windowStart time.Time
	requests    int
}

type rateLimiter struct {
	mu         sync.Mutex
	entries    map[string]rateLimitEntry
	limit      int
	interval   time.Duration
	trustProxy bool
	lastSweep  time.Time
}

func NewRateLimiter(limit int, interval time.Duration, trustProxy bool) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		entries:    make(map[string]rateLimitEntry),
		limit:      limit,
		interval:   interval,
		trustProxy: trustProxy,
	}
	return limiter.Middleware
}

func (l *rateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l.limit <= 0 || l.interval <= 0 {
			next.ServeHTTP(w, r)
			return
		}

		now := time.Now()
		key := rateLimitKey(r, l.trustProxy)
		l.mu.Lock()
		if l.lastSweep.IsZero() || now.Sub(l.lastSweep) >= l.interval {
			for entryKey, entry := range l.entries {
				if now.Sub(entry.windowStart) >= l.interval {
					delete(l.entries, entryKey)
				}
			}
			l.lastSweep = now
		}
		entry := l.entries[key]
		if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= l.interval {
			entry = rateLimitEntry{windowStart: now}
		}
		if entry.requests >= l.limit {
			retryAfter := int(math.Ceil(l.interval.Seconds() - now.Sub(entry.windowStart).Seconds()))
			l.mu.Unlock()
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			apierrors.NewProblem(http.StatusTooManyRequests, "Too Many Requests", "Rate limit exceeded", "rate_limit_exceeded").Write(w)
			return
		}
		entry.requests++
		l.entries[key] = entry
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func rateLimitKey(r *http.Request, trustProxy bool) string {
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		return "user:" + userID.String()
	}
	return "ip:" + clientIP(r, trustProxy)
}

// Los headers de proxy solo son fiables con TRUST_PROXY activo (la API vive detrás de nginx).
func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			return strings.TrimSpace(realIP)
		}
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			first, _, _ := strings.Cut(forwarded, ",")
			return strings.TrimSpace(first)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
