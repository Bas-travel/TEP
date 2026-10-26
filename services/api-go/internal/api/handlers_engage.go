package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

func (s *Server) engageRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /v1/me/settings", s.authed(s.getSettings))
	m.HandleFunc("PATCH /v1/me/settings", s.authed(s.patchSettings))
	m.HandleFunc("GET /v1/coupons", s.authed(s.listCoupons))
	m.HandleFunc("POST /v1/coupons/{code}/apply", s.authed(s.applyCoupon))
	m.HandleFunc("GET /v1/notifications", s.authed(s.listNotifications))
	m.HandleFunc("POST /v1/notifications/read-all", s.authed(s.readAllNotifications))
	m.HandleFunc("POST /v1/notifications/{notificationId}/read", s.authed(s.readNotification))
	m.HandleFunc("GET /v1/trips/saved", s.authed(s.listTrips))
	m.HandleFunc("POST /v1/trips/saved", s.authed(s.createTrip))
	m.HandleFunc("GET /v1/trips/saved/{tripId}", s.authed(s.getTrip))
	m.HandleFunc("GET /v1/reviews", s.authed(s.listReviews))
	m.HandleFunc("POST /v1/reviews", s.authed(s.createReview))
	m.HandleFunc("GET /v1/support/tickets", s.authed(s.listTickets))
	m.HandleFunc("POST /v1/support/tickets", s.authed(s.createTicket))
	m.HandleFunc("POST /v1/ai/itineraries", s.authed(s.createItinerary))
	m.HandleFunc("GET /v1/ai/itineraries/jobs/{jobId}", s.authed(s.getItineraryJob))
}

func fe(field, msg string) httpx.FieldError { return httpx.FieldError{Field: field, Message: msg} }

func (s *Server) today() string { return s.now().Format("2006-01-02") }

func randID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

// ---- settings ----

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request, uid int64) {
	var v Settings
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { v, e = tx.Settings(uid); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, v)
}

func (s *Server) patchSettings(w http.ResponseWriter, r *http.Request, uid int64) {
	var in struct {
		Language      *string `json:"language"`
		Currency      *string `json:"currency"`
		Theme         *string `json:"theme"`
		Notifications *struct {
			Push  *bool `json:"push"`
			Email *bool `json:"email"`
			Trip  *bool `json:"trip"`
			Promo *bool `json:"promo"`
		} `json:"notifications"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if in.Theme != nil && *in.Theme != "light" && *in.Theme != "dark" && *in.Theme != "system" {
		errs = append(errs, fe("theme", "must be light, dark or system"))
	}
	if in.Language != nil && (len(*in.Language) < 2 || len(*in.Language) > 10) {
		errs = append(errs, fe("language", "must be a language code (2..10 chars)"))
	}
	if in.Currency != nil && len(*in.Currency) != 3 {
		errs = append(errs, fe("currency", "must be a 3-letter code"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var v Settings
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		cur, e := tx.Settings(uid)
		if e != nil {
			return e
		}
		if in.Language != nil {
			cur.Language = *in.Language
		}
		if in.Currency != nil {
			cur.Currency = strings.ToUpper(*in.Currency)
		}
		if in.Theme != nil {
			cur.Theme = *in.Theme
		}
		if n := in.Notifications; n != nil {
			set := func(dst *bool, v *bool) {
				if v != nil {
					*dst = *v
				}
			}
			set(&cur.Notifications.Push, n.Push)
			set(&cur.Notifications.Email, n.Email)
			set(&cur.Notifications.Trip, n.Trip)
			set(&cur.Notifications.Promo, n.Promo)
		}
		v = cur
		return tx.SaveSettings(uid, cur)
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, v)
}

// ---- coupons ----

func (s *Server) listCoupons(w http.ResponseWriter, r *http.Request, uid int64) {
	var items []UserCoupon
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.UserCoupons(uid, s.today()); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) applyCoupon(w http.ResponseWriter, r *http.Request, uid int64) {
	code := strings.ToUpper(r.PathValue("code"))
	var in struct {
		Total float64 `json:"bookingDraftTotal"`
	}
	_ = httpx.Decode(r, &in)
	var disc float64
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		list, e := tx.UserCoupons(uid, s.today())
		if e != nil {
			return e
		}
		var c *UserCoupon
		for i := range list {
			if list[i].Code == code {
				c = &list[i]
			}
		}
		if c == nil {
			return ErrNotFound
		}
		if c.Status != "available" {
			return invalid(fe("code", "coupon is not available"))
		}
		if in.Total < c.MinSpend {
			return invalid(fe("bookingDraftTotal", "below minimum spend"))
		}
		disc = in.Total * c.pct
		if c.pct >= 1 || disc > c.MaxDiscount {
			disc = c.MaxDiscount
		}
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"valid": true, "discount": round2(disc), "message": "ใช้คูปองสำเร็จ"})
}

// ---- notifications ----

func (s *Server) listNotifications(w http.ResponseWriter, r *http.Request, uid int64) {
	var items []Notification
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.Notifications(uid); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) readAllNotifications(w http.ResponseWriter, r *http.Request, uid int64) {
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) error { return tx.MarkAllNotificationsRead(uid) })) {
		return
	}
	w.WriteHeader(204)
}

func (s *Server) readNotification(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "notificationId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) error { return tx.MarkNotificationRead(uid, id) })) {
		return
	}
	w.WriteHeader(204)
}

// ---- trips ----

func (s *Server) listTrips(w http.ResponseWriter, r *http.Request, uid int64) {
	var items []SavedTrip
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.Trips(uid); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) createTrip(w http.ResponseWriter, r *http.Request, uid int64) {
	var in SavedTripInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if strings.TrimSpace(in.Destination) == "" || len([]rune(in.Destination)) > 120 {
		errs = append(errs, fe("destination", "required, max 120 characters"))
	}
	if in.Days < 1 || in.Days > 14 {
		errs = append(errs, fe("days", "must be 1..14"))
	}
	if in.Source == "" {
		in.Source = "manual"
	}
	if in.Source != "ai" && in.Source != "community" && in.Source != "manual" {
		errs = append(errs, fe("source", "must be ai, community or manual"))
	}
	if in.Budget < 0 {
		errs = append(errs, fe("budget", "must not be negative"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	if in.Travelers < 1 {
		in.Travelers = 1
	}
	var t SavedTrip
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { t, e = tx.AddTrip(uid, in, s.now()); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, t)
}

func (s *Server) getTrip(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "tripId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var t *SavedTrip
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { t, e = tx.Trip(uid, id); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, t)
}

// ---- reviews ----

func reviewOut(rv reviewRec) Review {
	l, _ := findListing(rv.ListingID)
	return Review{ID: rv.ID, ItemName: l.Name, ItemType: l.Type, Rating: rv.Rating, Comment: rv.Comment, CreatedAt: rv.CreatedAt}
}

func (s *Server) listReviews(w http.ResponseWriter, r *http.Request, uid int64) {
	items := []Review{}
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		rs, e := tx.Reviews(uid)
		for _, rv := range rs {
			items = append(items, reviewOut(rv))
		}
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) createReview(w http.ResponseWriter, r *http.Request, uid int64) {
	var in struct {
		ListingID int64   `json:"listingId"`
		Rating    int     `json:"rating"`
		Comment   *string `json:"comment"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if in.Rating < 1 || in.Rating > 5 {
		errs = append(errs, fe("rating", "must be 1..5"))
	}
	if in.Comment == nil {
		errs = append(errs, fe("comment", "required"))
	} else if len([]rune(*in.Comment)) > 2000 {
		errs = append(errs, fe("comment", "max 2000 characters"))
	}
	if _, ok := findListing(in.ListingID); !ok {
		errs = append(errs, fe("listingId", "unknown listing"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out Review
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		rv, e := tx.AddReview(uid, in.ListingID, in.Rating, *in.Comment, s.now())
		out = reviewOut(rv)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, out)
}

// ---- support ----

func (s *Server) listTickets(w http.ResponseWriter, r *http.Request, uid int64) {
	var items []SupportTicket
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.Tickets(uid); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) createTicket(w http.ResponseWriter, r *http.Request, uid int64) {
	var in struct{ Subject, Message string }
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if strings.TrimSpace(in.Subject) == "" || len([]rune(in.Subject)) > 120 {
		errs = append(errs, fe("subject", "required, max 120 characters"))
	}
	if strings.TrimSpace(in.Message) == "" || len([]rune(in.Message)) > 2000 {
		errs = append(errs, fe("message", "required, max 2000 characters"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var t SupportTicket
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { t, e = tx.AddTicket(uid, in.Subject, in.Message, s.now()); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, t)
}

// ---- AI planner (template fallback, synchronous) ----

func (s *Server) createItinerary(w http.ResponseWriter, r *http.Request, uid int64) {
	var in struct {
		Destination string   `json:"destination"`
		Days        int      `json:"days"`
		Budget      float64  `json:"budget"`
		Styles      []string `json:"styles"`
		Travelers   int      `json:"travelers"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if strings.TrimSpace(in.Destination) == "" {
		errs = append(errs, fe("destination", "required"))
	}
	if in.Days < 1 || in.Days > 14 {
		errs = append(errs, fe("days", "must be 1..14"))
	}
	if in.Budget <= 0 {
		errs = append(errs, fe("budget", "must be > 0"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	job := ItineraryJob{JobID: randID("job_"), Status: "completed", Degraded: true,
		Result: buildItinerary(in.Destination, in.Days, in.Budget, in.Travelers)}
	err := s.repo.Tx(r.Context(), func(tx Tx) error { return tx.AddJob(uid, jobRec{ItineraryJob: job, Input: in}, s.now()) })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 202, job)
}

func (s *Server) getItineraryJob(w http.ResponseWriter, r *http.Request, uid int64) {
	var job *ItineraryJob
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { job, e = tx.Job(uid, r.PathValue("jobId")); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, job)
}
