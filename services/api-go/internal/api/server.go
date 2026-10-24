// Package api implements the TEP HTTP API (Sprint 1: contract-first mock mode over net/http).
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/auth"
	"github.com/Bas-travel/tep/services/api-go/internal/config"
	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

// Check reports whether a dependency is reachable. Nil Check map entry => "disabled".
type Check func(ctx context.Context) error

type Server struct {
	cfg         config.Config
	st          *store
	log         *slog.Logger
	now         func() time.Time
	checks      map[string]Check
	settleAfter time.Duration // mock PSP: how long a pending intent stays pending
}

func New(cfg config.Config, log *slog.Logger, checks map[string]Check) *Server {
	s := &Server{cfg: cfg, st: newStore(), log: log, now: time.Now, checks: checks, settleAfter: 3 * time.Second}
	if cfg.AppEnv == "dev" {
		s.seedDemo()
	}
	return s
}

func round2(x float64) float64 { return math.Round(x*100) / 100 }

type authedHandler func(w http.ResponseWriter, r *http.Request, uid int64)

func (s *Server) authed(h authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get("Authorization")
		if !strings.HasPrefix(hdr, "Bearer ") {
			httpx.Unauthorized(w, r, "missing bearer token")
			return
		}
		c, err := auth.ParseJWT(s.cfg.JWTSecret, strings.TrimPrefix(hdr, "Bearer "), s.now())
		if err != nil {
			httpx.Unauthorized(w, r, "invalid or expired access token")
			return
		}
		s.st.mu.Lock()
		_, ok := s.st.users[c.Sub]
		s.st.mu.Unlock()
		if !ok {
			httpx.Unauthorized(w, r, "unknown user")
			return
		}
		h(w, r, c.Sub)
	}
}

// Handler builds the route table (see docs/api/openapi.yaml).
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health/live", s.healthLive)
	m.HandleFunc("GET /health/ready", s.healthReady)

	m.HandleFunc("POST /v1/auth/register", s.register)
	m.HandleFunc("POST /v1/auth/login", s.login)
	m.HandleFunc("POST /v1/auth/refresh", s.refreshTokens)
	m.HandleFunc("POST /v1/auth/logout", s.authed(s.logout))
	m.HandleFunc("GET /v1/me", s.authed(s.getMe))

	m.HandleFunc("GET /v1/catalog/destinations", s.listDestinations)
	m.HandleFunc("GET /v1/catalog/listings", s.searchListings)
	m.HandleFunc("GET /v1/catalog/listings/{listingId}", s.getListing)
	m.HandleFunc("GET /v1/support/faqs", s.listFaqs)

	m.HandleFunc("GET /v1/bookings", s.authed(s.listBookings))
	m.HandleFunc("POST /v1/bookings", s.authed(s.createBooking))
	m.HandleFunc("GET /v1/bookings/{bookingId}", s.authed(s.getBooking))
	m.HandleFunc("POST /v1/bookings/{bookingId}/cancel", s.authed(s.cancelBooking))

	m.HandleFunc("POST /v1/payments/intents", s.authed(s.createIntent))
	m.HandleFunc("GET /v1/payments/intents/{intentId}", s.authed(s.getIntent))
	m.HandleFunc("GET /v1/wallet", s.authed(s.getWallet))
	m.HandleFunc("POST /v1/wallet/top-ups", s.authed(s.topUp))

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { httpx.NotFound(w, r) })
	return m
}

// ---- health ----

func (s *Server) healthLive(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, 200, map[string]any{"status": "ok", "version": s.cfg.Version})
}

func (s *Server) healthReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	res := map[string]string{}
	status, code := "ok", 200
	for _, name := range []string{"mysql", "redis"} {
		chk := s.checks[name]
		switch {
		case chk == nil:
			res[name] = "disabled"
		case chk(ctx) != nil:
			res[name] = "down"
			status, code = "degraded", 503
		default:
			res[name] = "ok"
		}
	}
	httpx.WriteJSON(w, code, map[string]any{"status": status, "version": s.cfg.Version, "checks": res})
}

// ---- helpers ----

func pathID(r *http.Request, name string) (int64, bool) {
	v, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	return v, err == nil && v > 0
}

func (s *Server) idemKey(r *http.Request, uid int64, endpoint string) string {
	k := r.Header.Get("Idempotency-Key")
	if k == "" {
		return ""
	}
	return strconv.FormatInt(uid, 10) + "|" + endpoint + "|" + k
}

// replay writes a stored idempotent response if present.
func (s *Server) replay(w http.ResponseWriter, key string) bool {
	if key == "" {
		return false
	}
	s.st.mu.Lock()
	rec, ok := s.st.idem[key]
	s.st.mu.Unlock()
	if !ok {
		return false
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Idempotent-Replay", "true")
	w.WriteHeader(rec.Status)
	_, _ = w.Write(rec.Body)
	return true
}

// respondIdem writes v as JSON and records it under key (if any). Caller must not hold st.mu.
func (s *Server) respondIdem(w http.ResponseWriter, key string, status int, v any) {
	if key != "" {
		b, _ := json.Marshal(v)
		s.st.mu.Lock()
		s.st.idem[key] = idemRec{Status: status, Body: append(b, '\n')}
		s.st.mu.Unlock()
	}
	httpx.WriteJSON(w, status, v)
}
