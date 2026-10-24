package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/auth"
	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

func (s *Server) seedDemo() {
	req := RegisterRequest{Email: "demo@tep.dev", Password: "Passw0rd!", FirstName: "สวัสดี", LastName: "อนันต์",
		Language: "th", BudgetLevel: "mid", TravelStyles: []string{"ทะเล", "คาเฟ่"}, FavoriteDestination: "ภูเก็ต"}
	u, _ := s.createUser(req)
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	u.Wallet = 1500
	now := s.now()
	u.Txs = append(u.Txs, WalletTx{ID: s.st.nextTx, Direction: "in", Label: "เติมเงิน PromptPay", Amount: 1500, CreatedAt: now.Add(-48 * time.Hour)})
	s.st.nextTx++
}

func (s *Server) createUser(req RegisterRequest) (*userRec, error) {
	h, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	id := s.st.nextUser
	s.st.nextUser++
	p := UserProfile{ID: id, Level: "Explorer", Email: req.Email, Phone: req.Phone, FirstName: req.FirstName, LastName: req.LastName,
		Gender: req.Gender, DateOfBirth: req.DateOfBirth, Nationality: req.Nationality, Language: req.Language,
		BudgetLevel: req.BudgetLevel, TravelStyles: req.TravelStyles, FavoriteDestination: req.FavoriteDestination, TravelWith: req.TravelWith}
	u := &userRec{ID: id, Role: "traveler", PassHash: h, Profile: p}
	s.st.users[id] = u
	s.st.byEmail[strings.ToLower(req.Email)] = id
	return u, nil
}

func (s *Server) issue(uid int64, role string) (AuthTokens, error) {
	at, err := auth.IssueJWT(s.cfg.JWTSecret, uid, role, s.now())
	if err != nil {
		return AuthTokens{}, err
	}
	rt := auth.NewOpaqueToken()
	s.st.mu.Lock()
	s.st.refresh[rt] = uid
	s.st.mu.Unlock()
	return AuthTokens{AccessToken: at, RefreshToken: rt, ExpiresIn: int(auth.AccessTTL.Seconds())}, nil
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Invalid(w, r, httpx.FieldError{Field: "body", Message: err.Error()})
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	var errs []httpx.FieldError
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email[strings.Index(req.Email, "@"):], ".") {
		errs = append(errs, httpx.FieldError{Field: "email", Message: "must be a valid email"})
	}
	if len(req.Password) < 8 {
		errs = append(errs, httpx.FieldError{Field: "password", Message: "must be at least 8 characters"})
	}
	if strings.TrimSpace(req.FirstName) == "" {
		errs = append(errs, httpx.FieldError{Field: "firstName", Message: "required"})
	}
	if strings.TrimSpace(req.LastName) == "" {
		errs = append(errs, httpx.FieldError{Field: "lastName", Message: "required"})
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	s.st.mu.Lock()
	_, dup := s.st.byEmail[strings.ToLower(req.Email)]
	s.st.mu.Unlock()
	if dup {
		httpx.WriteProblem(w, r, 409, httpx.CodeConflict, "Conflict", "email already registered")
		return
	}
	u, err := s.createUser(req)
	if err != nil {
		httpx.WriteProblem(w, r, 500, httpx.CodeInternal, "Internal Server Error", "")
		return
	}
	t, err := s.issue(u.ID, u.Role)
	if err != nil {
		httpx.WriteProblem(w, r, 500, httpx.CodeInternal, "Internal Server Error", "")
		return
	}
	httpx.WriteJSON(w, 201, t)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct{ Email, Password string }
	if err := httpx.Decode(r, &req); err != nil || req.Email == "" || req.Password == "" {
		httpx.Invalid(w, r, httpx.FieldError{Field: "email/password", Message: "required"})
		return
	}
	s.st.mu.Lock()
	id, ok := s.st.byEmail[strings.ToLower(strings.TrimSpace(req.Email))]
	var u *userRec
	if ok {
		u = s.st.users[id]
	}
	s.st.mu.Unlock()
	// Same error for unknown email and wrong password (no account enumeration).
	if u == nil || !auth.CheckPassword(req.Password, u.PassHash) {
		httpx.Unauthorized(w, r, "invalid email or password")
		return
	}
	t, err := s.issue(u.ID, u.Role)
	if err != nil {
		httpx.WriteProblem(w, r, 500, httpx.CodeInternal, "Internal Server Error", "")
		return
	}
	httpx.WriteJSON(w, 200, t)
}

// refreshTokens rotates the refresh token. Presenting an already-rotated token revokes every session of that user.
func (s *Server) refreshTokens(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.RefreshToken == "" {
		httpx.Invalid(w, r, httpx.FieldError{Field: "refreshToken", Message: "required"})
		return
	}
	s.st.mu.Lock()
	uid, ok := s.st.refresh[req.RefreshToken]
	if ok {
		delete(s.st.refresh, req.RefreshToken)
		s.st.used[req.RefreshToken] = uid
	} else if victim, reused := s.st.used[req.RefreshToken]; reused {
		for tok, u := range s.st.refresh {
			if u == victim {
				delete(s.st.refresh, tok)
			}
		}
	}
	u := s.st.users[uid]
	s.st.mu.Unlock()
	if !ok || u == nil {
		httpx.Unauthorized(w, r, "invalid refresh token")
		return
	}
	t, err := s.issue(uid, u.Role)
	if err != nil {
		httpx.WriteProblem(w, r, 500, httpx.CodeInternal, "Internal Server Error", "")
		return
	}
	httpx.WriteJSON(w, 200, t)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request, uid int64) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.RefreshToken == "" {
		httpx.Invalid(w, r, httpx.FieldError{Field: "refreshToken", Message: "required"})
		return
	}
	s.st.mu.Lock()
	if s.st.refresh[req.RefreshToken] == uid {
		delete(s.st.refresh, req.RefreshToken)
	}
	s.st.mu.Unlock()
	w.WriteHeader(204)
}

func (s *Server) getMe(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	p := s.st.users[uid].Profile
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 200, p)
}
