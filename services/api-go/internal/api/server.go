// Package api implements the TEP HTTP API (Sprint 1: contract-first mock mode over net/http).
package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	repo        Repo
	log         *slog.Logger
	now         func() time.Time
	checks      map[string]Check
	settleAfter time.Duration // mock PSP (MOCK_PAYMENT=true): how long a pending intent stays pending
}

// Option customises New.
type Option func(*Server)

// WithRepo selects the storage backend (default: in-memory, for API_MODE=mock and unit tests).
func WithRepo(r Repo) Option { return func(s *Server) { s.repo = r } }

func New(cfg config.Config, log *slog.Logger, checks map[string]Check, opts ...Option) *Server {
	s := &Server{cfg: cfg, repo: newMemRepo(), log: log, now: time.Now, checks: checks, settleAfter: 3 * time.Second}
	for _, o := range opts {
		o(s)
	}
	if cfg.AppEnv == "dev" {
		if err := s.seedDemo(context.Background()); err != nil {
			log.Error("demo seed failed", "err", err)
		}
	}
	return s
}

func round2(x float64) float64 { return math.Round(x*100) / 100 }

// apiErr is a business error: returning it from a Tx function undoes the transaction and is rendered as RFC 7807.
type apiErr struct {
	Status      int
	Code, Title string
	Detail      string
	Errors      []httpx.FieldError
}

func (e *apiErr) Error() string { return e.Code + ": " + e.Detail }

func invalid(errs ...httpx.FieldError) *apiErr {
	return &apiErr{Status: 422, Code: httpx.CodeValidation, Title: "Validation failed", Detail: "one or more fields are invalid", Errors: errs}
}
func conflict(detail string) *apiErr {
	return &apiErr{Status: 409, Code: httpx.CodeConflict, Title: "Conflict", Detail: detail}
}
func forbidden(detail string) *apiErr {
	return &apiErr{Status: 403, Code: httpx.CodeForbidden, Title: "Forbidden", Detail: detail}
}

// fail renders err (if any) and reports whether it did. Unknown errors are logged and returned as 500 without detail.
func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	var ae *apiErr
	switch {
	case errors.As(err, &ae):
		httpx.WriteProblem(w, r, ae.Status, ae.Code, ae.Title, ae.Detail, ae.Errors...)
	case errors.Is(err, ErrNotFound):
		httpx.NotFound(w, r)
	case errors.Is(err, ErrForbidden):
		httpx.WriteProblem(w, r, 403, httpx.CodeForbidden, "Forbidden", "not allowed")
	case errors.Is(err, ErrInsufficient):
		httpx.WriteProblem(w, r, 402, httpx.CodeWalletShort, "Payment Required", "wallet balance is insufficient")
	default:
		s.log.Error("request failed", "request_id", httpx.RequestID(r), "path", r.URL.Path, "err", err)
		httpx.WriteProblem(w, r, 500, httpx.CodeInternal, "Internal Server Error", "")
	}
	return true
}

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
		err = s.repo.Tx(r.Context(), func(tx Tx) error {
			_, e := tx.UserByID(c.Sub, false)
			return e
		})
		if errors.Is(err, ErrNotFound) {
			httpx.Unauthorized(w, r, "unknown user")
			return
		}
		if s.fail(w, r, err) {
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
	m.HandleFunc("PUT /v1/auth/password", s.authed(s.changePassword))
	m.HandleFunc("GET /v1/me", s.authed(s.getMe))
	m.HandleFunc("DELETE /v1/me", s.authed(s.deleteMe))

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

	m.HandleFunc("POST /v1/webhooks/payments/{provider}", s.paymentWebhook)
	m.HandleFunc("POST /v1/devices/push-tokens", s.authed(s.registerPushToken))
	m.HandleFunc("POST /v1/media/presign", s.authed(s.presignUpload))

	s.engageRoutes(m)
	s.communityRoutes(m)

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

// hashKey maps an opaque client value (idempotency key, refresh token) to a fixed-width hex digest for storage.
func hashKey(k string) string {
	if k == "" {
		return ""
	}
	h := sha256.Sum256([]byte(k))
	return hex.EncodeToString(h[:])
}

// idempotent runs fn in a transaction. With an Idempotency-Key header the response is stored in the same transaction,
// so a retry (or a concurrent duplicate) replays it instead of repeating the side effect.
// fn returns the HTTP status and body; returning an error (e.g. *apiErr) rolls back and stores nothing.
func (s *Server) idempotent(w http.ResponseWriter, r *http.Request, uid int64, endpoint string, fn func(Tx) (int, any, error)) {
	key := hashKey(r.Header.Get("Idempotency-Key"))
	if len(r.Header.Get("Idempotency-Key")) > 200 {
		httpx.Invalid(w, r, fe("Idempotency-Key", "max 200 characters"))
		return
	}
	for attempt := 0; attempt < 2; attempt++ {
		var replay *idemRec
		var status int
		var body any
		err := s.repo.Tx(r.Context(), func(tx Tx) error {
			replay, status, body = nil, 0, nil
			if key != "" {
				rec, e := tx.GetIdem(uid, endpoint, key)
				if e == nil {
					replay = rec
					return nil
				}
				if !errors.Is(e, ErrNotFound) {
					return e
				}
			}
			st, b, e := fn(tx)
			if e != nil {
				return e
			}
			status, body = st, b
			if key == "" {
				return nil
			}
			raw, e := json.Marshal(b)
			if e != nil {
				return e
			}
			return tx.PutIdem(uid, endpoint, key, idemRec{Status: st, Body: append(raw, '\n')}, s.now())
		})
		if key != "" && errors.Is(err, ErrDuplicate) {
			continue // a concurrent request with the same key won: replay its response on the next attempt
		}
		if s.fail(w, r, err) {
			return
		}
		if replay != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Idempotent-Replay", "true")
			w.WriteHeader(replay.Status)
			_, _ = w.Write(replay.Body)
			return
		}
		httpx.WriteJSON(w, status, body)
		return
	}
	httpx.WriteProblem(w, r, 409, httpx.CodeConflict, "Conflict", "request with this Idempotency-Key is still in progress")
}
