package http

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// V1HealthHandler expone un healthcheck público sin autenticación.
type V1HealthHandler struct {
	version string
	limiter *healthLimiter
	nowFunc func() time.Time
}

func NewV1HealthHandler(version string, ratePerMin int) *V1HealthHandler {
	h := &V1HealthHandler{
		version: version,
		nowFunc: time.Now,
	}
	if ratePerMin > 0 {
		h.limiter = newHealthLimiter(ratePerMin)
	}
	return h
}

func (h *V1HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
		return
	}

	if h.limiter != nil && !h.limiter.allow(clientIP(r), h.nowFunc) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"error":   "RATE_LIMITED",
			"message": "Demasiados requests",
		}); err != nil {
			log.Printf("error writing rate-limit response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"ok":        true,
		"service":   "wsapi",
		"version":   h.version,
		"timestamp": h.nowFunc().UTC().Format(time.RFC3339),
	}); err != nil {
		log.Printf("error writing health response: %v", err)
	}
}

// clientIP devuelve la IP del cliente respetando X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		xff = strings.TrimSpace(xff)
		if xff == "" || xff[0] == ',' {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				return r.RemoteAddr
			}
			return host
		}
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				ip := strings.TrimSpace(xff[:i])
				if host, _, err := net.SplitHostPort(ip); err == nil {
					return host
				}
				return ip
			}
		}
		if host, _, err := net.SplitHostPort(xff); err == nil {
			return host
		}
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// healthLimiter implementa ventana fija de 60s por IP sin librerías externas.
type healthLimiter struct {
	mu         sync.Mutex
	ratePerMin int
	buckets    map[string]*ipBucket
	callCount  int // para limpieza periódica
}

type ipBucket struct {
	count       int
	windowStart time.Time
}

func newHealthLimiter(ratePerMin int) *healthLimiter {
	return &healthLimiter{
		ratePerMin: ratePerMin,
		buckets:    make(map[string]*ipBucket),
	}
}

func (l *healthLimiter) allow(ip string, now func() time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	t := now()
	b, ok := l.buckets[ip]
	if !ok || t.Sub(b.windowStart) >= time.Minute {
		l.buckets[ip] = &ipBucket{count: 1, windowStart: t}
		l.callCount++
		l.maybeClean(t)
		return true
	}
	if b.count >= l.ratePerMin {
		return false
	}
	b.count++
	l.callCount++
	l.maybeClean(t)
	return true
}

// maybeClean elimina buckets inactivos cada 256 llamadas para evitar leaks.
func (l *healthLimiter) maybeClean(now time.Time) {
	if l.callCount%256 != 0 {
		return
	}
	cutoff := now.Add(-5 * time.Minute)
	for ip, b := range l.buckets {
		if b.windowStart.Before(cutoff) {
			delete(l.buckets, ip)
		}
	}
}
