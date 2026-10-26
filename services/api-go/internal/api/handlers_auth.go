package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/auth"
	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

const refreshTTL = 30 * 24 * time.Hour

// seedDemo creates the DEV demo account with sample engagement data. Idempotent: skipped when it already exists.
func (s *Server) seedDemo(ctx context.Context) error {
	hash, err := auth.HashPassword("Passw0rd!")
	if err != nil {
		return err
	}
	return s.repo.Tx(ctx, func(tx Tx) error {
		if _, err := tx.UserByEmail("demo@tep.dev"); err == nil {
			return nil
		}
		return s.seedAll(tx, hash)
	})
}

func (s *Server) issue(tx Tx, uid int64, role string) (AuthTokens, error) {
	now := s.now()
	at, err := auth.IssueJWT(s.cfg.JWTSecret, uid, role, now)
	if err != nil {
		return AuthTokens{}, err
	}
	rt := auth.NewOpaqueToken()
	if err := tx.AddRefresh(hashKey(rt), uid, now.Add(refreshTTL), now); err != nil {
		return AuthTokens{}, err
	}
	return AuthTokens{AccessToken: at, RefreshToken: rt, ExpiresIn: int(auth.AccessTTL.Seconds())}, nil
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	var errs []httpx.FieldError
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email[strings.Index(req.Email, "@"):], ".") || len(req.Email) > 190 {
		errs = append(errs, fe("email", "must be a valid email"))
	}
	if len(req.Password) < 8 || len(req.Password) > 128 {
		errs = append(errs, fe("password", "must be 8..128 characters"))
	}
	if strings.TrimSpace(req.FirstName) == "" {
		errs = append(errs, fe("firstName", "required"))
	}
	if strings.TrimSpace(req.LastName) == "" {
		errs = append(errs, fe("lastName", "required"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var tokens AuthTokens
	err = s.repo.Tx(r.Context(), func(tx Tx) error {
		if _, e := tx.UserByEmail(req.Email); e == nil {
			return conflict("email already registered")
		}
		u := &userRec{PassHash: hash, Profile: UserProfile{Email: req.Email, Phone: req.Phone, FirstName: req.FirstName, LastName: req.LastName,
			Gender: req.Gender, DateOfBirth: req.DateOfBirth, Nationality: req.Nationality, Language: req.Language,
			BudgetLevel: req.BudgetLevel, TravelStyles: req.TravelStyles, FavoriteDestination: req.FavoriteDestination, TravelWith: req.TravelWith}}
		if e := tx.CreateUser(u, s.now()); e != nil {
			if errors.Is(e, ErrDuplicate) {
				return conflict("email already registered")
			}
			return e
		}
		if e := tx.GrantCoupon(u.ID, "WELCOME10"); e != nil && !errors.Is(e, ErrNotFound) {
			return e
		}
		t, e := s.issue(tx, u.ID, u.Role)
		tokens = t
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, tokens)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct{ Email, Password string }
	if err := httpx.Decode(r, &req); err != nil || req.Email == "" || req.Password == "" {
		httpx.Invalid(w, r, fe("email/password", "required"))
		return
	}
	var tokens AuthTokens
	bad := false
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		u, e := tx.UserByEmail(strings.TrimSpace(req.Email))
		// Same outcome for unknown email and wrong password (no account enumeration).
		if errors.Is(e, ErrNotFound) || (e == nil && !auth.CheckPassword(req.Password, u.PassHash)) {
			bad = true
			return nil
		}
		if e != nil {
			return e
		}
		tokens, e = s.issue(tx, u.ID, u.Role)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	if bad {
		httpx.Unauthorized(w, r, "invalid email or password")
		return
	}
	httpx.WriteJSON(w, 200, tokens)
}

// refreshTokens rotates the refresh token. Presenting an already-rotated token revokes every session of that user.
func (s *Server) refreshTokens(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.RefreshToken == "" {
		httpx.Invalid(w, r, fe("refreshToken", "required"))
		return
	}
	var tokens AuthTokens
	ok := false
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		uid, st, e := tx.UseRefresh(hashKey(req.RefreshToken), s.now())
		if e != nil {
			return e
		}
		switch st {
		case RefreshReused:
			return tx.RevokeAllRefresh(uid) // committed: theft response
		case RefreshOK:
			u, e := tx.UserByID(uid, false)
			if errors.Is(e, ErrNotFound) {
				return nil
			}
			if e != nil {
				return e
			}
			tokens, e = s.issue(tx, uid, u.Role)
			ok = e == nil
			return e
		}
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	if !ok {
		httpx.Unauthorized(w, r, "invalid refresh token")
		return
	}
	httpx.WriteJSON(w, 200, tokens)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request, uid int64) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.RefreshToken == "" {
		httpx.Invalid(w, r, fe("refreshToken", "required"))
		return
	}
	err := s.repo.Tx(r.Context(), func(tx Tx) error { return tx.RevokeRefresh(hashKey(req.RefreshToken), uid) })
	if s.fail(w, r, err) {
		return
	}
	w.WriteHeader(204)
}

func (s *Server) getMe(w http.ResponseWriter, r *http.Request, uid int64) {
	var p UserProfile
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		u, e := tx.UserByID(uid, false)
		if e == nil {
			p = u.Profile
		}
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, p)
}

// deleteMe erases the account: personal data and engagement content are removed, financial records (bookings, wallet)
// are kept with the identity anonymised. All sessions are revoked.
func (s *Server) deleteMe(w http.ResponseWriter, r *http.Request, uid int64) {
	err := s.repo.Tx(r.Context(), func(tx Tx) error { return tx.DeleteUser(uid, s.now()) })
	if s.fail(w, r, err) {
		return
	}
	w.WriteHeader(204)
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request, uid int64) {
	var in struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if in.CurrentPassword == "" {
		errs = append(errs, fe("currentPassword", "required"))
	}
	if len(in.NewPassword) < 8 || len(in.NewPassword) > 128 {
		errs = append(errs, fe("newPassword", "must be 8..128 characters"))
	}
	if in.NewPassword == in.CurrentPassword && in.CurrentPassword != "" {
		errs = append(errs, fe("newPassword", "must differ from the current password"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	newHash, err := auth.HashPassword(in.NewPassword)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	err = s.repo.Tx(r.Context(), func(tx Tx) error {
		u, e := tx.UserByID(uid, true)
		if e != nil {
			return e
		}
		if !auth.CheckPassword(in.CurrentPassword, u.PassHash) {
			return invalid(fe("currentPassword", "incorrect"))
		}
		u.PassHash = newHash
		if e := tx.SaveUser(u); e != nil {
			return e
		}
		return tx.RevokeAllRefresh(uid) // sign out every device
	})
	if s.fail(w, r, err) {
		return
	}
	w.WriteHeader(204)
}

func (s *Server) registerPushToken(w http.ResponseWriter, r *http.Request, uid int64) {
	var in struct{ Token, Platform string }
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if in.Token == "" || len(in.Token) > 512 {
		errs = append(errs, fe("token", "required, max 512 characters"))
	}
	if in.Platform != "ios" && in.Platform != "android" {
		errs = append(errs, fe("platform", "must be ios or android"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	err := s.repo.Tx(r.Context(), func(tx Tx) error { return tx.UpsertPushToken(uid, in.Token, in.Platform, s.now()) })
	if s.fail(w, r, err) {
		return
	}
	w.WriteHeader(204)
}
