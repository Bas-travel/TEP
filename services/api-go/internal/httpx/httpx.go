// Package httpx holds framework-agnostic HTTP helpers: RFC 7807 problems, JSON IO, middleware.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type ctxKey int

const requestIDKey ctxKey = 1

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Problem struct {
	Type      string       `json:"type,omitempty"`
	Title     string       `json:"title"`
	Status    int          `json:"status"`
	Detail    string       `json:"detail,omitempty"`
	Code      string       `json:"code"`
	RequestID string       `json:"requestId,omitempty"`
	Errors    []FieldError `json:"errors,omitempty"`
}

// Stable error codes (docs/api/conventions.md).
const (
	CodeUnauthenticated = "AUTH_UNAUTHENTICATED"
	CodeForbidden       = "AUTH_FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeValidation      = "VALIDATION_FAILED"
	CodeWalletShort     = "WALLET_INSUFFICIENT"
	CodeRateLimited     = "RATE_LIMITED"
	CodeConflict        = "CONFLICT"
	CodeInternal        = "INTERNAL"
)

func RequestID(r *http.Request) string {
	if v, ok := r.Context().Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func WriteProblem(w http.ResponseWriter, r *http.Request, status int, code, title, detail string, errs ...FieldError) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type: "about:blank", Title: title, Status: status, Detail: detail,
		Code: code, RequestID: RequestID(r), Errors: errs,
	})
}

func Unauthorized(w http.ResponseWriter, r *http.Request, detail string) {
	WriteProblem(w, r, 401, CodeUnauthenticated, "Unauthorized", detail)
}
func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteProblem(w, r, 404, CodeNotFound, "Not Found", "resource not found")
}
func Invalid(w http.ResponseWriter, r *http.Request, errs ...FieldError) {
	WriteProblem(w, r, 422, CodeValidation, "Validation failed", "one or more fields are invalid", errs...)
}

// Decode reads a JSON body (max 1 MiB). Returns an error suitable for a 422 detail.
func Decode(r *http.Request, v any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	if err := dec.Decode(v); err != nil {
		return errors.New("invalid JSON body")
	}
	return nil
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(c int) { s.status = c; s.ResponseWriter.WriteHeader(c) }

// Chain applies middleware: request id → recover → access log → CORS (dev only).
func Chain(h http.Handler, log *slog.Logger, corsDev bool) http.Handler {
	h = accessLog(h, log)
	h = recoverer(h, log)
	if corsDev {
		h = cors(h)
	}
	return requestID(h)
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

func recoverer(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic", "request_id", RequestID(r), "panic", rec, "stack", string(debug.Stack()))
				WriteProblem(w, r, 500, CodeInternal, "Internal Server Error", "unexpected error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func accessLog(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		log.Info("http", "method", r.Method, "path", r.URL.Path, "status", sw.status,
			"ms", time.Since(start).Milliseconds(), "request_id", RequestID(r))
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
