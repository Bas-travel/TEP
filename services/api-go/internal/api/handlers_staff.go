package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

// Staff routes accept staff and admin (admin can step in for support).
func (s *Server) staffRoutes(m *http.ServeMux) {
	p := s.portal(RoleStaff, RoleAdmin)
	m.HandleFunc("GET /staff/v1/dashboard", p(s.staffDashboard))
	m.HandleFunc("GET /staff/v1/bookings", p(s.staffListBookings))
	m.HandleFunc("GET /staff/v1/bookings/{bookingId}", p(s.staffGetBooking))
	m.HandleFunc("POST /staff/v1/bookings/{bookingId}/check", p(s.staffCheckBooking))
	m.HandleFunc("GET /staff/v1/tickets", p(s.staffListTickets))
	m.HandleFunc("GET /staff/v1/tickets/{ticketId}", p(s.staffGetTicket))
	m.HandleFunc("PATCH /staff/v1/tickets/{ticketId}", p(s.staffPatchTicket))
	m.HandleFunc("POST /staff/v1/tickets/{ticketId}/replies", p(s.staffReplyTicket))
	m.HandleFunc("GET /staff/v1/reports/daily", p(s.staffDailyReport))
}

func (s *Server) staffBookingView(tx Tx, b *bookingRec) (StaffBookingView, error) {
	v := StaffBookingView{BookingDetail: b.BookingDetail, Checks: []BookingCheckView{}, Verification: "none"}
	if u, e := tx.UserAny(b.UserID, false); e == nil {
		v.TravelerName, v.TravelerEmail = memberName(u.Profile), u.Profile.Email
	}
	cs, err := tx.BookingChecks(b.ID)
	if err != nil {
		return v, err
	}
	for i, c := range cs { // newest first
		if i == 0 {
			v.Verification = c.Result
		}
		v.Checks = append(v.Checks, BookingCheckView{ID: c.ID, StaffName: c.StaffName, Result: c.Result, Note: c.Note, CreatedAt: c.CreatedAt})
	}
	return v, nil
}

func (s *Server) staffListBookings(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	off, lim, errs := pageParams(r)
	if st := q.Get("status"); st != "" && !oneOf(st, "pending_payment", "confirmed", "completed", "cancelled") {
		errs = append(errs, fe("status", "invalid status"))
	}
	f := BookingFilter{Status: q.Get("status"), Q: q.Get("q"), Offset: off, Limit: lim}
	if v := q.Get("checked"); v != "" {
		if v != "true" && v != "false" {
			errs = append(errs, fe("checked", "must be true or false"))
		}
		c := v == "true"
		f.Checked = &c
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var items []StaffBookingView
	var total int
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		bs, t, e := tx.AdminBookings(f)
		if e != nil {
			return e
		}
		total = t
		for _, b := range bs {
			v, e := s.staffBookingView(tx, b)
			if e != nil {
				return e
			}
			items = append(items, v)
		}
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	if items == nil {
		items = []StaffBookingView{}
	}
	pageJSON(w, items, total, off, lim)
}

func (s *Server) staffGetBooking(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "bookingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var out StaffBookingView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		b, e := tx.GetBooking(id, false)
		if e != nil {
			return e
		}
		out, e = s.staffBookingView(tx, b)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) staffCheckBooking(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "bookingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Result string `json:"result"`
		Note   string `json:"note"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	in.Note = strings.TrimSpace(in.Note)
	var errs []httpx.FieldError
	if !oneOf(in.Result, "verified", "flagged") {
		errs = append(errs, fe("result", "must be verified or flagged"))
	}
	if in.Result == "flagged" && in.Note == "" {
		errs = append(errs, fe("note", "required when flagging a booking"))
	}
	if len([]rune(in.Note)) > 500 {
		errs = append(errs, fe("note", "max 500 characters"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out StaffBookingView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		b, e := tx.GetBooking(id, true)
		if e != nil {
			return e
		}
		if b.Status != "confirmed" && b.Status != "completed" {
			return conflict("only paid bookings (confirmed or completed) are verified; this one is " + b.Status)
		}
		if e := tx.AddBookingCheck(&bookingCheck{BookingID: b.ID, StaffID: a.ID, Result: in.Result, Note: in.Note, CreatedAt: s.now()}); e != nil {
			return e
		}
		if e := s.audit(tx, a, "staff.booking.check", "booking", b.ID, map[string]any{"result": in.Result}); e != nil {
			return e
		}
		out, e = s.staffBookingView(tx, b)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// ---- tickets ----

func (s *Server) staffTicketView(tx Tx, t *ticketRec) StaffTicketView {
	v := StaffTicketView{SupportTicket: t.SupportTicket, ReplyCount: t.ReplyCount}
	if v.Replies == nil {
		v.Replies = []TicketReply{}
	}
	if u, e := tx.UserAny(t.UserID, false); e == nil {
		v.TravelerName, v.TravelerEmail = memberName(u.Profile), u.Profile.Email
	}
	return v
}

func (s *Server) staffListTickets(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	off, lim, errs := pageParams(r)
	if st := q.Get("status"); st != "" && !oneOf(st, "open", "answered", "resolved") {
		errs = append(errs, fe("status", "invalid status"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var items []StaffTicketView
	var total int
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		ts, t, e := tx.AdminTickets(TicketFilter{Status: q.Get("status"), Q: q.Get("q"), Offset: off, Limit: lim})
		if e != nil {
			return e
		}
		total = t
		for _, tk := range ts {
			items = append(items, s.staffTicketView(tx, tk))
		}
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	if items == nil {
		items = []StaffTicketView{}
	}
	pageJSON(w, items, total, off, lim)
}

func (s *Server) staffGetTicket(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "ticketId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var out StaffTicketView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		t, e := tx.TicketByID(id, false)
		if e == nil {
			out = s.staffTicketView(tx, t)
		}
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func validReplyText(msg string) (string, []httpx.FieldError) {
	msg = strings.TrimSpace(msg)
	if msg == "" || len([]rune(msg)) > 2000 {
		return msg, []httpx.FieldError{fe("message", "required, max 2000 characters")}
	}
	return msg, nil
}

func (s *Server) staffReplyTicket(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "ticketId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Message string `json:"message"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	msg, errs := validReplyText(in.Message)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out StaffTicketView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		t, e := tx.TicketByID(id, true)
		if e != nil {
			return e
		}
		if t.Status == "resolved" {
			return conflict("ticket is resolved; reopen it before replying")
		}
		role := RoleStaff
		if a.Role == RoleAdmin {
			role = RoleAdmin
		}
		if _, e := tx.AddTicketReply(t.ID, a.ID, role, msg, s.now()); e != nil {
			return e
		}
		if _, e := tx.AddNotification(t.UserID, Notification{Type: "system", Title: "ทีมงานตอบกลับคำร้องของคุณ", Message: t.Subject, CreatedAt: s.now()}); e != nil {
			return e
		}
		if e := s.audit(tx, a, "staff.ticket.reply", "ticket", t.ID, nil); e != nil {
			return e
		}
		if t, e = tx.TicketByID(id, false); e != nil {
			return e
		}
		out = s.staffTicketView(tx, t)
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, out)
}

func (s *Server) staffPatchTicket(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "ticketId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := httpx.Decode(r, &in); err != nil || !oneOf(in.Status, "open", "resolved") {
		httpx.Invalid(w, r, fe("status", "must be open or resolved"))
		return
	}
	var out StaffTicketView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		t, e := tx.TicketByID(id, true)
		if e != nil {
			return e
		}
		if t.Status != in.Status {
			if e := tx.SetTicketStatus(t.ID, in.Status, s.now()); e != nil {
				return e
			}
			if e := s.audit(tx, a, "staff.ticket.status", "ticket", t.ID, map[string]any{"status": []string{t.Status, in.Status}}); e != nil {
				return e
			}
		}
		if t, e = tx.TicketByID(id, false); e != nil {
			return e
		}
		out = s.staffTicketView(tx, t)
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// ---- dashboard & daily report ----

func (s *Server) staffCounters(tx Tx, from, to time.Time) (map[string]any, error) {
	_, open, err := tx.AdminTickets(TicketFilter{Status: "open", Limit: 1})
	if err != nil {
		return nil, err
	}
	no := false
	_, unchecked, err := tx.AdminBookings(BookingFilter{Statuses: []string{"confirmed", "completed"}, Checked: &no, Limit: 1})
	if err != nil {
		return nil, err
	}
	verified, flagged, err := tx.CheckStats(0, from, to)
	if err != nil {
		return nil, err
	}
	all, _, err := tx.AdminTickets(TicketFilter{})
	if err != nil {
		return nil, err
	}
	var sum float64
	n := 0
	for _, t := range all {
		if t.FirstRepliedAt != nil {
			sum += t.FirstRepliedAt.Sub(t.CreatedAt).Hours()
			n++
		}
	}
	avg := 0.0
	if n > 0 {
		avg = round2(sum / float64(n))
	}
	return map[string]any{"openTickets": open, "uncheckedBookings": unchecked, "checkedToday": verified + flagged, "verifiedToday": verified,
		"flaggedToday": flagged, "avgFirstResponseHours": avg}, nil
}

func (s *Server) staffDashboard(w http.ResponseWriter, r *http.Request, a actor) {
	var out map[string]any
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) {
		from := startOfDay(s.now())
		out, e = s.staffCounters(tx, from, from.Add(24*time.Hour))
		return
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// staffDailyReport summarises one Asia/Bangkok day (default today): ?date=YYYY-MM-DD.
func (s *Server) staffDailyReport(w http.ResponseWriter, r *http.Request, a actor) {
	day := startOfDay(s.now())
	if v := r.URL.Query().Get("date"); v != "" {
		d, err := time.ParseInLocation("2006-01-02", v, ictZone)
		if err != nil {
			httpx.Invalid(w, r, fe("date", "must be YYYY-MM-DD"))
			return
		}
		day = d
	}
	end := day.Add(24 * time.Hour)
	var out map[string]any
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		verified, flagged, e := tx.CheckStats(0, day, end)
		if e != nil {
			return e
		}
		mine, mf, e := tx.CheckStats(a.ID, day, end)
		if e != nil {
			return e
		}
		bs, total, e := tx.AdminBookings(BookingFilter{Since: day})
		if e != nil {
			return e
		}
		created := 0
		for _, b := range bs {
			if b.Date.Before(end) {
				created++
			}
		}
		_ = total
		ts, _, e := tx.AdminTickets(TicketFilter{Since: day})
		if e != nil {
			return e
		}
		var opened, answered, resolved int
		for _, t := range ts {
			if !t.CreatedAt.Before(day) && t.CreatedAt.Before(end) {
				opened++
			}
			if t.FirstRepliedAt != nil && !t.FirstRepliedAt.Before(day) && t.FirstRepliedAt.Before(end) {
				answered++
			}
			if t.ResolvedAt != nil && !t.ResolvedAt.Before(day) && t.ResolvedAt.Before(end) {
				resolved++
			}
		}
		out = map[string]any{"date": day.Format("2006-01-02"), "timezone": "Asia/Bangkok", "bookingsCreated": created,
			"bookingsVerified": verified, "bookingsFlagged": flagged, "myChecks": mine + mf,
			"ticketsOpened": opened, "ticketsFirstAnswered": answered, "ticketsResolved": resolved}
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// ---- traveler: reply to own ticket ----

func (s *Server) replyOwnTicket(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "ticketId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Message string `json:"message"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	msg, errs := validReplyText(in.Message)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out SupportTicket
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		t, e := tx.TicketByID(id, true)
		if e != nil {
			return e
		}
		if t.UserID != uid {
			return ErrNotFound
		}
		if _, e := tx.AddTicketReply(t.ID, uid, RoleTraveler, msg, s.now()); e != nil {
			return e
		}
		t, e = tx.TicketByID(id, false)
		if e == nil {
			out = t.SupportTicket
		}
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, out)
}
