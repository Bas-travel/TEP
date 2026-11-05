package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

func (s *Server) adminRoutes(m *http.ServeMux) {
	p := s.portal(RoleAdmin)
	m.HandleFunc("GET /admin/v1/dashboard", p(s.adminDashboard))
	m.HandleFunc("GET /admin/v1/users", p(s.adminListUsers))
	m.HandleFunc("GET /admin/v1/users/{userId}", p(s.adminGetUser))
	m.HandleFunc("PATCH /admin/v1/users/{userId}", p(s.adminPatchUser))
	m.HandleFunc("GET /admin/v1/partners", p(s.adminListPartners))
	m.HandleFunc("GET /admin/v1/partners/{partnerId}", p(s.adminGetPartner))
	m.HandleFunc("PATCH /admin/v1/partners/{partnerId}", p(s.adminPatchPartner))
	m.HandleFunc("GET /admin/v1/listings", p(s.adminListListings))
	m.HandleFunc("PATCH /admin/v1/listings/{listingId}", p(s.adminPatchListing))
	m.HandleFunc("GET /admin/v1/audit-logs", p(s.adminAuditLogs))
	m.HandleFunc("GET /admin/v1/system", p(s.adminSystem))
}

func startOfDay(t time.Time) time.Time {
	t = t.In(ictZone)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, ictZone)
}

func (s *Server) adminDashboard(w http.ResponseWriter, r *http.Request, a actor) {
	var st PlatformStats
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { st, e = tx.Stats(startOfDay(s.now())); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{
		"users": st.Users, "usersSuspended": st.UsersSuspended, "usersByRole": st.ByRole, "partners": st.Partners,
		"listingsPublished": st.ListingsPublished, "bookingsToday": st.BookingsSince, "bookingsTotal": st.BookingsTotal,
		"revenue": round2(st.Revenue),
	})
}

// ---- users ----

func adminUserView(u *userRec) AdminUserView {
	return AdminUserView{ID: u.ID, Handle: u.Handle, Email: u.Profile.Email, Name: strings.TrimSpace(u.Profile.FirstName + " " + u.Profile.LastName),
		Phone: u.Profile.Phone, Role: u.Role, Status: u.Status, CreatedAt: u.CreatedAt}
}

func (s *Server) adminListUsers(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	off, lim, errs := pageParams(r)
	if v := q.Get("role"); v != "" && !oneOf(v, RoleTraveler, RolePartner, RoleStaff, RoleModerator, RoleAdmin) {
		errs = append(errs, fe("role", "invalid role"))
	}
	if v := q.Get("status"); v != "" && !oneOf(v, "active", "suspended", "deleted") {
		errs = append(errs, fe("status", "invalid status"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var rows []*userRec
	var total int
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) {
		rows, total, e = tx.AdminUsers(UserFilter{Q: q.Get("q"), Role: q.Get("role"), Status: q.Get("status"), Offset: off, Limit: lim})
		return
	})
	if s.fail(w, r, err) {
		return
	}
	items := make([]AdminUserView, 0, len(rows))
	for _, u := range rows {
		items = append(items, adminUserView(u))
	}
	pageJSON(w, items, total, off, lim)
}

func (s *Server) adminGetUser(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "userId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var u *userRec
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { u, e = tx.UserAny(id, false); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, adminUserView(u))
}

func (s *Server) adminPatchUser(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "userId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Role   *string `json:"role"`
		Status *string `json:"status"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if in.Role == nil && in.Status == nil {
		errs = append(errs, fe("body", "role or status required"))
	}
	if in.Role != nil && !oneOf(*in.Role, RoleTraveler, RoleStaff, RoleModerator, RoleAdmin) {
		errs = append(errs, fe("role", "must be traveler, staff, moderator or admin (partners are managed under /admin/v1/partners)"))
	}
	if in.Status != nil && !oneOf(*in.Status, "active", "suspended") {
		errs = append(errs, fe("status", "must be active or suspended"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out AdminUserView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		u, e := tx.UserAny(id, true)
		if e != nil {
			return e
		}
		if u.Status == "deleted" {
			return conflict("deleted accounts cannot be changed")
		}
		if u.ID == a.ID {
			return conflict("administrators cannot change their own role or status")
		}
		revoke := false
		meta := map[string]any{}
		if in.Role != nil && *in.Role != u.Role {
			if u.Role == RolePartner {
				return conflict("partner accounts are managed under /admin/v1/partners")
			}
			meta["role"] = []string{u.Role, *in.Role}
			u.Role, revoke = *in.Role, true
		}
		if in.Status != nil && *in.Status != u.Status {
			meta["status"] = []string{u.Status, *in.Status}
			u.Status = *in.Status
			revoke = revoke || u.Status == "suspended"
		}
		if len(meta) == 0 {
			out = adminUserView(u)
			return nil
		}
		if e := tx.SaveUser(u); e != nil {
			return e
		}
		if revoke {
			if e := tx.RevokeAllRefresh(u.ID); e != nil {
				return e
			}
		}
		if e := s.audit(tx, a, "admin.user.update", "user", u.ID, meta); e != nil {
			return e
		}
		out = adminUserView(u)
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// ---- partners ----

func (s *Server) adminListPartners(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	off, lim, errs := pageParams(r)
	if v := q.Get("status"); v != "" && !oneOf(v, PartnerPending, PartnerApproved, PartnerRejected, PartnerSuspended) {
		errs = append(errs, fe("status", "invalid status"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var items []PartnerView
	var total int
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		ps, t, e := tx.Partners(PartnerFilter{Q: q.Get("q"), Status: q.Get("status"), Offset: off, Limit: lim})
		if e != nil {
			return e
		}
		total = t
		for _, p := range ps {
			v, e := s.partnerView(tx, p, true)
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
		items = []PartnerView{}
	}
	pageJSON(w, items, total, off, lim)
}

func (s *Server) adminGetPartner(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "partnerId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var out PartnerView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		p, e := tx.PartnerByID(id, false)
		if e != nil {
			return e
		}
		out, e = s.partnerView(tx, p, true)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

var partnerTransitions = map[string][]string{
	PartnerPending:   {PartnerApproved, PartnerRejected},
	PartnerApproved:  {PartnerSuspended},
	PartnerSuspended: {PartnerApproved},
	PartnerRejected:  {PartnerApproved},
}

func (s *Server) adminPatchPartner(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "partnerId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Status         *string  `json:"status"`
		CommissionRate *float64 `json:"commissionRate"`
		DecisionNote   *string  `json:"decisionNote"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	if in.Status == nil && in.CommissionRate == nil && in.DecisionNote == nil {
		errs = append(errs, fe("body", "status, commissionRate or decisionNote required"))
	}
	if in.Status != nil && !oneOf(*in.Status, PartnerApproved, PartnerRejected, PartnerSuspended) {
		errs = append(errs, fe("status", "must be approved, rejected or suspended"))
	}
	if in.CommissionRate != nil && (*in.CommissionRate < 0 || *in.CommissionRate > 50) {
		errs = append(errs, fe("commissionRate", "must be between 0 and 50"))
	}
	if in.DecisionNote != nil && len([]rune(*in.DecisionNote)) > 500 {
		errs = append(errs, fe("decisionNote", "max 500 characters"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out PartnerView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		p, e := tx.PartnerByID(id, true)
		if e != nil {
			return e
		}
		meta := map[string]any{}
		if in.DecisionNote != nil {
			p.DecisionNote = strings.TrimSpace(*in.DecisionNote)
		}
		if in.CommissionRate != nil && round2(*in.CommissionRate) != p.CommissionRate {
			meta["commissionRate"] = []float64{p.CommissionRate, round2(*in.CommissionRate)}
			p.CommissionRate = round2(*in.CommissionRate)
		}
		var notify string
		if in.Status != nil && *in.Status != p.Status {
			if !oneOf(*in.Status, partnerTransitions[p.Status]...) {
				return conflict("cannot change partner from " + p.Status + " to " + *in.Status)
			}
			if *in.Status != PartnerApproved && p.DecisionNote == "" {
				return invalid(fe("decisionNote", "a note is required to reject or suspend"))
			}
			u, e := tx.UserAny(p.UserID, true)
			if e != nil {
				return e
			}
			if *in.Status == PartnerApproved {
				if u.Role != RoleTraveler && u.Role != RolePartner {
					return conflict("the owner account has role " + u.Role + " and cannot become a partner")
				}
				if u.Role != RolePartner {
					u.Role = RolePartner
					if e := tx.SaveUser(u); e != nil {
						return e
					}
				}
			}
			meta["status"] = []string{p.Status, *in.Status}
			p.Status, p.DecidedBy = *in.Status, a.ID
			now := s.now()
			p.DecidedAt = &now
			notify = map[string]string{
				PartnerApproved:  "ใบสมัครพาร์ทเนอร์ของคุณได้รับการอนุมัติแล้ว",
				PartnerRejected:  "ใบสมัครพาร์ทเนอร์ไม่ผ่านการพิจารณา: " + p.DecisionNote,
				PartnerSuspended: "บัญชีพาร์ทเนอร์ถูกระงับชั่วคราว: " + p.DecisionNote,
			}[p.Status]
			if p.Status == PartnerApproved {
				p.DecisionNote = ""
			}
			if in.DecisionNote != nil && p.Status == PartnerApproved {
				p.DecisionNote = strings.TrimSpace(*in.DecisionNote)
			}
		}
		if e := tx.SavePartner(p); e != nil {
			return e
		}
		if notify != "" {
			if _, e := tx.AddNotification(p.UserID, Notification{Type: "system", Title: "สถานะพาร์ทเนอร์", Message: notify, CreatedAt: s.now()}); e != nil {
				return e
			}
		}
		if e := s.audit(tx, a, "admin.partner.update", "partner", p.ID, meta); e != nil {
			return e
		}
		out, e = s.partnerView(tx, p, true)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// ---- listings (moderation) ----

func (s *Server) adminListListings(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	if st := q.Get("status"); st != "" && !oneOf(st, "draft", "published", "archived") {
		httpx.Invalid(w, r, fe("status", "invalid status"))
		return
	}
	f := ListingFilter{Q: q.Get("q"), Type: q.Get("type"), Status: q.Get("status")}
	if v := q.Get("partnerId"); v != "" {
		n, ok := atoi64(v)
		if !ok {
			httpx.Invalid(w, r, fe("partnerId", "must be a number"))
			return
		}
		f.PartnerID = n
	}
	var rows []listingRow
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { rows, e = tx.Listings(f); return })
	if s.fail(w, r, err) {
		return
	}
	items := make([]ListingAdminView, 0, len(rows))
	for _, l := range rows {
		items = append(items, listingAdminView(l))
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) adminPatchListing(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "listingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Status *string `json:"status"`
	}
	if err := httpx.Decode(r, &in); err != nil || in.Status == nil || !oneOf(*in.Status, "draft", "published", "archived") {
		httpx.Invalid(w, r, fe("status", "must be draft, published or archived"))
		return
	}
	var out ListingAdminView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		l, e := tx.ListingByID(id, false)
		if e != nil {
			return e
		}
		old := l.Status
		l.Status = *in.Status
		if e := tx.SaveListing(l); e != nil {
			return e
		}
		if e := s.audit(tx, a, "admin.listing.status", "listing", l.ID, map[string]any{"status": []string{old, l.Status}}); e != nil {
			return e
		}
		out = listingAdminView(*l)
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// ---- audit & system ----

func (s *Server) adminAuditLogs(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	off, lim, errs := pageParams(r)
	f := AuditFilter{Action: q.Get("action"), Offset: off, Limit: lim}
	if v := q.Get("actorId"); v != "" {
		n, ok := atoi64(v)
		if !ok {
			errs = append(errs, fe("actorId", "must be a number"))
		}
		f.ActorID = n
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var rows []auditRec
	var total int
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { rows, total, e = tx.Audits(f); return })
	if s.fail(w, r, err) {
		return
	}
	items := make([]AuditView, 0, len(rows))
	for _, x := range rows {
		items = append(items, AuditView{ID: x.ID, ActorID: x.ActorID, ActorName: x.ActorName, Action: x.Action, Entity: x.Entity, EntityID: x.EntityID, Meta: x.Meta, CreatedAt: x.CreatedAt})
	}
	pageJSON(w, items, total, off, lim)
}

func (s *Server) adminSystem(w http.ResponseWriter, r *http.Request, a actor) {
	deps := map[string]string{}
	for _, name := range []string{"mysql", "redis"} {
		switch chk := s.checks[name]; {
		case chk == nil:
			deps[name] = "disabled"
		case chk(r.Context()) != nil:
			deps[name] = "down"
		default:
			deps[name] = "ok"
		}
	}
	httpx.WriteJSON(w, 200, map[string]any{"version": s.cfg.Version, "env": s.cfg.AppEnv, "mode": s.cfg.APIMode, "serverTime": s.now().UTC(), "dependencies": deps})
}
