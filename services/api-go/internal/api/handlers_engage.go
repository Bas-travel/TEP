package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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
	m.HandleFunc("GET /v1/community/feed", s.authed(s.communityFeed))
	m.HandleFunc("POST /v1/community/posts/{postId}/like", s.authed(s.likePost))
	m.HandleFunc("POST /v1/ai/itineraries", s.authed(s.createItinerary))
	m.HandleFunc("GET /v1/ai/itineraries/jobs/{jobId}", s.authed(s.getItineraryJob))
}

func fe(field, msg string) httpx.FieldError { return httpx.FieldError{Field: field, Message: msg} }

// ---- settings ----

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	v := s.extra(uid).Settings
	s.st.mu.Unlock()
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
	if in.Theme != nil && *in.Theme != "light" && *in.Theme != "dark" && *in.Theme != "system" {
		httpx.Invalid(w, r, fe("theme", "must be light, dark or system"))
		return
	}
	s.st.mu.Lock()
	x := s.extra(uid)
	if in.Language != nil {
		x.Settings.Language = *in.Language
	}
	if in.Currency != nil {
		x.Settings.Currency = *in.Currency
	}
	if in.Theme != nil {
		x.Settings.Theme = *in.Theme
	}
	if n := in.Notifications; n != nil {
		set := func(dst *bool, v *bool) {
			if v != nil {
				*dst = *v
			}
		}
		set(&x.Settings.Notifications.Push, n.Push)
		set(&x.Settings.Notifications.Email, n.Email)
		set(&x.Settings.Notifications.Trip, n.Trip)
		set(&x.Settings.Notifications.Promo, n.Promo)
	}
	v := x.Settings
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 200, v)
}

// ---- coupons ----

func (s *Server) listCoupons(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	items := append([]UserCoupon{}, s.extra(uid).Coupons...)
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) applyCoupon(w http.ResponseWriter, r *http.Request, uid int64) {
	code := strings.ToUpper(r.PathValue("code"))
	var in struct {
		Total float64 `json:"bookingDraftTotal"`
	}
	_ = httpx.Decode(r, &in)
	s.st.mu.Lock()
	var c *UserCoupon
	x := s.extra(uid)
	for i := range x.Coupons {
		if x.Coupons[i].Code == code {
			cp := x.Coupons[i]
			c = &cp
		}
	}
	s.st.mu.Unlock()
	if c == nil {
		httpx.NotFound(w, r)
		return
	}
	if c.Status != "available" {
		httpx.Invalid(w, r, fe("code", "coupon is not available"))
		return
	}
	if in.Total < c.MinSpend {
		httpx.Invalid(w, r, fe("bookingDraftTotal", "below minimum spend"))
		return
	}
	d := in.Total * c.pct
	if c.pct >= 1 || d > c.MaxDiscount {
		d = c.MaxDiscount
	}
	httpx.WriteJSON(w, 200, map[string]any{"valid": true, "discount": round2(d), "message": "ใช้คูปองสำเร็จ"})
}

// ---- notifications ----

func (s *Server) listNotifications(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	items := append([]Notification{}, s.extra(uid).Notifs...)
	s.st.mu.Unlock()
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) readAllNotifications(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	x := s.extra(uid)
	for i := range x.Notifs {
		x.Notifs[i].Read = true
	}
	s.st.mu.Unlock()
	w.WriteHeader(204)
}

func (s *Server) readNotification(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "notificationId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	found := false
	s.st.mu.Lock()
	x := s.extra(uid)
	for i := range x.Notifs {
		if x.Notifs[i].ID == id {
			x.Notifs[i].Read, found = true, true
		}
	}
	s.st.mu.Unlock()
	if !found {
		httpx.NotFound(w, r)
		return
	}
	w.WriteHeader(204)
}

// ---- trips ----

func (s *Server) listTrips(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	items := append([]SavedTrip{}, s.extra(uid).Trips...)
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) createTrip(w http.ResponseWriter, r *http.Request, uid int64) {
	var in SavedTripInput
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
	if in.Source == "" {
		in.Source = "manual"
	}
	if in.Source != "ai" && in.Source != "community" && in.Source != "manual" {
		errs = append(errs, fe("source", "must be ai, community or manual"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	if in.Travelers < 1 {
		in.Travelers = 1
	}
	s.st.mu.Lock()
	t := SavedTrip{ID: s.nextID(), SavedTripInput: in}
	x := s.extra(uid)
	x.Trips = append(x.Trips, t)
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 201, t)
}

func (s *Server) getTrip(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "tripId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	for _, t := range s.extra(uid).Trips {
		if t.ID == id {
			httpx.WriteJSON(w, 200, t)
			return
		}
	}
	httpx.NotFound(w, r)
}

// ---- reviews ----

func (s *Server) listReviews(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	items := append([]Review{}, s.extra(uid).Reviews...)
	s.st.mu.Unlock()
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
	} else if len(*in.Comment) > 2000 {
		errs = append(errs, fe("comment", "max 2000 characters"))
	}
	var lst *listingRow
	for i := range seedListings {
		if seedListings[i].ID == in.ListingID {
			lst = &seedListings[i]
		}
	}
	if lst == nil {
		errs = append(errs, fe("listingId", "unknown listing"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	s.st.mu.Lock()
	rv := Review{ID: s.nextID(), ItemName: lst.Name, ItemType: lst.Type, Rating: in.Rating, Comment: *in.Comment, CreatedAt: s.now()}
	x := s.extra(uid)
	x.Reviews = append(x.Reviews, rv)
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 201, rv)
}

// ---- support ----

func (s *Server) listTickets(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	items := append([]SupportTicket{}, s.extra(uid).Tickets...)
	s.st.mu.Unlock()
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
	s.st.mu.Lock()
	t := SupportTicket{ID: s.nextID(), Subject: in.Subject, Message: in.Message, CreatedAt: s.now()}
	x := s.extra(uid)
	x.Tickets = append(x.Tickets, t)
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 201, t)
}

// ---- community ----

func (s *Server) communityFeed(w http.ResponseWriter, r *http.Request, uid int64) {
	mode, dest := r.URL.Query().Get("mode"), r.URL.Query().Get("destination")
	if mode != "" && mode != "all" && mode != "following" {
		httpx.Invalid(w, r, fe("mode", "must be all or following"))
		return
	}
	now := s.now()
	s.st.mu.Lock()
	x := s.extra(uid)
	items := []Post{}
	for i, p := range seedPosts {
		if mode == "following" && !followedAuthors[p.AuthorID] {
			continue
		}
		if dest != "" && p.Destination != dest {
			continue
		}
		p.CreatedAt = now.Add(-time.Duration(i+1) * 5 * time.Hour)
		if x.Liked[p.ID] {
			p.LikeCount++
		}
		items = append(items, p)
	}
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) likePost(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	var base *Post
	for i := range seedPosts {
		if ok && seedPosts[i].ID == id {
			base = &seedPosts[i]
		}
	}
	if base == nil {
		httpx.NotFound(w, r)
		return
	}
	s.st.mu.Lock()
	x := s.extra(uid)
	x.Liked[id] = !x.Liked[id]
	liked, n := x.Liked[id], base.LikeCount
	if liked {
		n++
	}
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 200, map[string]any{"liked": liked, "likeCount": n})
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
	s.st.mu.Lock()
	job := ItineraryJob{JobID: "job_" + strconv.FormatInt(s.nextID(), 10), Status: "completed", Degraded: true,
		Result: buildItinerary(in.Destination, in.Days, in.Budget, in.Travelers)}
	s.extra(uid).Jobs[job.JobID] = job
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 202, job)
}

func (s *Server) getItineraryJob(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	job, ok := s.extra(uid).Jobs[r.PathValue("jobId")]
	s.st.mu.Unlock()
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	httpx.WriteJSON(w, 200, job)
}
