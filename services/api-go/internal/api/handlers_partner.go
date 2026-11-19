package api

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

// ---- traveler side: apply to become a partner ----

type partnerApplyReq struct {
	BusinessName string `json:"businessName"`
	ContactPhone string `json:"contactPhone"`
	Description  string `json:"description"`
}

func validatePartnerFields(in partnerApplyReq, requireName bool) []httpx.FieldError {
	var errs []httpx.FieldError
	n := len([]rune(strings.TrimSpace(in.BusinessName)))
	if (requireName && n == 0) || n > 120 {
		errs = append(errs, fe("businessName", "required, max 120 characters"))
	}
	if len([]rune(in.ContactPhone)) > 30 {
		errs = append(errs, fe("contactPhone", "max 30 characters"))
	}
	if len([]rune(in.Description)) > 1000 {
		errs = append(errs, fe("description", "max 1000 characters"))
	}
	return errs
}

func (s *Server) partnerView(tx Tx, p *partnerRec, withOwner bool) (PartnerView, error) {
	v := PartnerView{ID: p.ID, UserID: p.UserID, BusinessName: p.BusinessName, ContactPhone: p.ContactPhone, Description: p.Description,
		Status: p.Status, CommissionRate: p.CommissionRate, DecisionNote: p.DecisionNote, CreatedAt: p.CreatedAt, DecidedAt: p.DecidedAt}
	ls, err := tx.Listings(ListingFilter{PartnerID: p.ID})
	if err != nil {
		return v, err
	}
	v.ListingCount = len(ls)
	if withOwner {
		if u, e := tx.UserAny(p.UserID, false); e == nil {
			v.OwnerName, v.OwnerEmail = memberName(u.Profile), u.Profile.Email
		}
	}
	return v, nil
}

func (s *Server) applyPartner(w http.ResponseWriter, r *http.Request, uid int64) {
	var in partnerApplyReq
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	if errs := validatePartnerFields(in, true); len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out PartnerView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		u, e := tx.UserByID(uid, true)
		if e != nil {
			return e
		}
		if u.Role != RoleTraveler {
			return conflict("only traveler accounts can apply to become a partner")
		}
		p := &partnerRec{UserID: uid, BusinessName: strings.TrimSpace(in.BusinessName), ContactPhone: strings.TrimSpace(in.ContactPhone),
			Description: strings.TrimSpace(in.Description), Status: PartnerPending, CommissionRate: DefaultCommissionRate, CreatedAt: s.now()}
		if e := tx.CreatePartner(p); e != nil {
			if e == ErrDuplicate {
				old, e2 := tx.PartnerByUser(uid, true)
				if e2 != nil {
					return e2
				}
				if old.Status != PartnerRejected {
					return conflict("a partner application already exists (" + old.Status + ")")
				}
				// A rejected applicant may re-apply: the record goes back to pending with the new details.
				old.BusinessName, old.ContactPhone, old.Description = p.BusinessName, p.ContactPhone, p.Description
				old.Status, old.DecisionNote, old.DecidedAt, old.DecidedBy = PartnerPending, "", nil, 0
				if e := tx.SavePartner(old); e != nil {
					return e
				}
				p = old
			} else {
				return e
			}
		}
		out, e = s.partnerView(tx, p, false)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, out)
}

func (s *Server) myPartnerApplication(w http.ResponseWriter, r *http.Request, uid int64) {
	var out PartnerView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		p, e := tx.PartnerByUser(uid, false)
		if e != nil {
			return e
		}
		out, e = s.partnerView(tx, p, false)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// ---- partner portal (/partner/v1) ----

func (s *Server) partnerRoutes(m *http.ServeMux) {
	p := s.portal(RolePartner)
	m.HandleFunc("GET /partner/v1/dashboard", p(s.partnerDashboard))
	m.HandleFunc("GET /partner/v1/profile", p(s.partnerGetProfile))
	m.HandleFunc("PATCH /partner/v1/profile", p(s.partnerPatchProfile))
	m.HandleFunc("GET /partner/v1/listings", p(s.partnerListListings))
	m.HandleFunc("POST /partner/v1/listings", p(s.partnerCreateListing))
	m.HandleFunc("GET /partner/v1/listings/{listingId}", p(s.partnerGetListing))
	m.HandleFunc("PATCH /partner/v1/listings/{listingId}", p(s.partnerPatchListing))
	m.HandleFunc("GET /partner/v1/bookings", p(s.partnerListBookings))
	m.HandleFunc("GET /partner/v1/bookings/{bookingId}", p(s.partnerGetBooking))
	m.HandleFunc("POST /partner/v1/bookings/{bookingId}/complete", p(s.partnerCompleteBooking))
	m.HandleFunc("POST /partner/v1/bookings/{bookingId}/reject", p(s.partnerRejectBooking))
	m.HandleFunc("GET /partner/v1/earnings", p(s.partnerEarnings))
}

func (s *Server) partnerGetProfile(w http.ResponseWriter, r *http.Request, a actor) {
	var out PartnerView
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { out, e = s.partnerView(tx, a.Partner, false); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) partnerPatchProfile(w http.ResponseWriter, r *http.Request, a actor) {
	var in struct {
		BusinessName *string `json:"businessName"`
		ContactPhone *string `json:"contactPhone"`
		Description  *string `json:"description"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var out PartnerView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		p, e := tx.PartnerByID(a.Partner.ID, true)
		if e != nil {
			return e
		}
		next := partnerApplyReq{BusinessName: p.BusinessName, ContactPhone: p.ContactPhone, Description: p.Description}
		if in.BusinessName != nil {
			next.BusinessName = strings.TrimSpace(*in.BusinessName)
		}
		if in.ContactPhone != nil {
			next.ContactPhone = strings.TrimSpace(*in.ContactPhone)
		}
		if in.Description != nil {
			next.Description = strings.TrimSpace(*in.Description)
		}
		if errs := validatePartnerFields(next, true); len(errs) > 0 {
			return invalid(errs...)
		}
		p.BusinessName, p.ContactPhone, p.Description = next.BusinessName, next.ContactPhone, next.Description
		if e := tx.SavePartner(p); e != nil {
			return e
		}
		if e := s.audit(tx, a, "partner.profile.update", "partner", p.ID, nil); e != nil {
			return e
		}
		out, e = s.partnerView(tx, p, false)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// ---- listings ----

type listingInput struct {
	Name          *string   `json:"name"`
	Type          *string   `json:"type"`
	Location      *string   `json:"location"`
	DestinationID *int64    `json:"destinationId"`
	PriceFrom     *float64  `json:"priceFrom"`
	Description   *string   `json:"description"`
	Amenities     *[]string `json:"amenities"`
	MapQuery      *string   `json:"mapQuery"`
	ImageURL      *string   `json:"imageUrl"`
	Status        *string   `json:"status"`
}

func listingAdminView(l listingRow) ListingAdminView {
	if l.Amenities == nil {
		l.Amenities = []string{}
	}
	return ListingAdminView{ListingDetail: l.ListingDetail, DestinationID: l.DestinationID, PartnerID: l.PartnerID, Status: l.Status}
}

// applyListing merges in into l and validates the result (create passes an empty row, patch the stored one).
func applyListing(tx Tx, l *listingRow, in listingInput, create bool) error {
	if in.Name != nil {
		l.Name = strings.TrimSpace(*in.Name)
	}
	if in.Type != nil {
		l.Type = *in.Type
	}
	if in.Location != nil {
		l.Location = strings.TrimSpace(*in.Location)
	}
	if in.DestinationID != nil {
		l.DestinationID = *in.DestinationID
	}
	if in.PriceFrom != nil {
		l.PriceFrom = round2(*in.PriceFrom)
	}
	if in.Description != nil {
		l.Description = strings.TrimSpace(*in.Description)
	}
	if in.Amenities != nil {
		l.Amenities = *in.Amenities
	}
	if in.MapQuery != nil {
		l.MapQuery = strings.TrimSpace(*in.MapQuery)
	}
	if in.ImageURL != nil {
		l.ImageURL = strings.TrimSpace(*in.ImageURL)
	}
	if in.Status != nil {
		l.Status = *in.Status
	}
	var errs []httpx.FieldError
	if n := len([]rune(l.Name)); n == 0 || n > 190 {
		errs = append(errs, fe("name", "required, max 190 characters"))
	}
	if !oneOf(l.Type, "hotel", "restaurant", "attraction") {
		errs = append(errs, fe("type", "must be hotel, restaurant or attraction"))
	}
	if n := len([]rune(l.Location)); n == 0 || n > 190 {
		errs = append(errs, fe("location", "required, max 190 characters"))
	}
	if l.PriceFrom <= 0 || l.PriceFrom > 1000000 {
		errs = append(errs, fe("priceFrom", "must be > 0 and <= 1,000,000"))
	}
	if len([]rune(l.Description)) > 2000 {
		errs = append(errs, fe("description", "max 2000 characters"))
	}
	if len(l.Amenities) > 20 {
		errs = append(errs, fe("amenities", "max 20 items"))
	}
	for i, am := range l.Amenities {
		if t := strings.TrimSpace(am); t == "" || len([]rune(t)) > 60 {
			errs = append(errs, fe("amenities", "item "+strconv.Itoa(i)+" must be 1..60 characters"))
			break
		}
	}
	if l.ImageURL != "" {
		if u, err := url.Parse(l.ImageURL); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || len(l.ImageURL) > 500 {
			errs = append(errs, fe("imageUrl", "must be an http(s) URL, max 500 characters"))
		}
	}
	if !oneOf(l.Status, "draft", "published", "archived") {
		errs = append(errs, fe("status", "must be draft, published or archived"))
	}
	ds, err := tx.Destinations()
	if err != nil {
		return err
	}
	found := false
	for _, d := range ds {
		found = found || d.ID == l.DestinationID
	}
	if !found {
		errs = append(errs, fe("destinationId", "unknown destination"))
	}
	if len(errs) > 0 {
		return invalid(errs...)
	}
	return nil
}

func (s *Server) partnerListListings(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	if st := q.Get("status"); st != "" && !oneOf(st, "draft", "published", "archived") {
		httpx.Invalid(w, r, fe("status", "invalid status"))
		return
	}
	var rows []listingRow
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) {
		rows, e = tx.Listings(ListingFilter{PartnerID: a.Partner.ID, Status: q.Get("status"), Type: q.Get("type"), Q: q.Get("q")})
		return
	})
	if s.fail(w, r, err) {
		return
	}
	items := make([]ListingAdminView, 0, len(rows))
	for _, l := range rows {
		items = append(items, listingAdminView(l))
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) ownListing(tx Tx, a actor, id int64) (*listingRow, error) {
	l, err := tx.ListingByID(id, false)
	if err != nil {
		return nil, err
	}
	if l.PartnerID != a.Partner.ID {
		return nil, ErrNotFound // other partners' listings do not exist for this partner
	}
	return l, nil
}

func (s *Server) partnerGetListing(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "listingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var out ListingAdminView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		l, e := s.ownListing(tx, a, id)
		if e == nil {
			out = listingAdminView(*l)
		}
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) partnerCreateListing(w http.ResponseWriter, r *http.Request, a actor) {
	var in listingInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var out ListingAdminView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		l := &listingRow{PartnerID: a.Partner.ID, Status: "draft"}
		if e := applyListing(tx, l, in, true); e != nil {
			return e
		}
		if e := tx.SaveListing(l); e != nil {
			return e
		}
		if e := s.audit(tx, a, "partner.listing.create", "listing", l.ID, map[string]any{"status": l.Status}); e != nil {
			return e
		}
		out = listingAdminView(*l)
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, out)
}

func (s *Server) partnerPatchListing(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "listingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in listingInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var out ListingAdminView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		l, e := s.ownListing(tx, a, id)
		if e != nil {
			return e
		}
		if e := applyListing(tx, l, in, false); e != nil {
			return e
		}
		if e := tx.SaveListing(l); e != nil {
			return e
		}
		if e := s.audit(tx, a, "partner.listing.update", "listing", l.ID, map[string]any{"status": l.Status}); e != nil {
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

// ---- bookings ----

// partnerShare sums the booking lines that belong to the partner and reports whether every line is theirs.
func partnerShare(b *bookingRec, own map[int64]bool) (gross float64, all bool) {
	all = len(b.Lines) > 0
	for _, ln := range b.Lines {
		if own[ln.ListingID] {
			gross += float64(ln.Quantity) * ln.UnitPrice
		} else {
			all = false
		}
	}
	return round2(gross), all
}

func ownListingIDs(tx Tx, partnerID int64) (map[int64]bool, error) {
	ls, err := tx.Listings(ListingFilter{PartnerID: partnerID})
	own := map[int64]bool{}
	for _, l := range ls {
		own[l.ID] = true
	}
	return own, err
}

func (s *Server) partnerBookingView(tx Tx, a actor, own map[int64]bool, b *bookingRec) PartnerBookingView {
	gross, all := partnerShare(b, own)
	comm := round2(gross * a.Partner.CommissionRate / 100)
	name := ""
	if u, e := tx.UserAny(b.UserID, false); e == nil {
		name = memberName(u.Profile)
	}
	return PartnerBookingView{ID: b.ID, BookingNumber: b.BookingNumber, Type: b.Type, Title: b.Title, Status: b.Status, Date: b.Date,
		StartDate: b.StartDate, EndDate: b.EndDate, Guests: b.Guests, TravelerName: name, Gross: gross, Commission: comm, Net: round2(gross - comm),
		CanAct: all && (b.Status == "confirmed" || b.Status == "pending_payment"), CancelReason: b.CancelReason}
}

func (s *Server) partnerListBookings(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	off, lim, errs := pageParams(r)
	if st := q.Get("status"); st != "" && !oneOf(st, "pending_payment", "confirmed", "completed", "cancelled") {
		errs = append(errs, fe("status", "invalid status"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var items []PartnerBookingView
	var total int
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		own, e := ownListingIDs(tx, a.Partner.ID)
		if e != nil {
			return e
		}
		bs, t, e := tx.AdminBookings(BookingFilter{Status: q.Get("status"), Q: q.Get("q"), PartnerID: a.Partner.ID, Offset: off, Limit: lim})
		if e != nil {
			return e
		}
		total = t
		for _, b := range bs {
			items = append(items, s.partnerBookingView(tx, a, own, b))
		}
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	if items == nil {
		items = []PartnerBookingView{}
	}
	pageJSON(w, items, total, off, lim)
}

// partnerBooking loads a booking that contains at least one of the partner's listings (otherwise 404).
func (s *Server) partnerBooking(tx Tx, a actor, id int64, lock bool) (*bookingRec, map[int64]bool, error) {
	own, err := ownListingIDs(tx, a.Partner.ID)
	if err != nil {
		return nil, nil, err
	}
	b, err := tx.GetBooking(id, lock)
	if err != nil {
		return nil, nil, err
	}
	for _, ln := range b.Lines {
		if own[ln.ListingID] {
			return b, own, nil
		}
	}
	return nil, nil, ErrNotFound
}

func (s *Server) partnerGetBooking(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "bookingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var out PartnerBookingView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		b, own, e := s.partnerBooking(tx, a, id, false)
		if e == nil {
			out = s.partnerBookingView(tx, a, own, b)
		}
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// partnerAct runs complete / reject. Both need every line of the booking to belong to the partner, because the
// action applies to the whole booking (a mixed booking is handled by staff / support).
func (s *Server) partnerAct(w http.ResponseWriter, r *http.Request, a actor, action string) {
	id, ok := pathID(r, "bookingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if action == "reject" {
		if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Reason) == "" || len([]rune(req.Reason)) > 450 {
			httpx.Invalid(w, r, fe("reason", "required, max 450 characters"))
			return
		}
	}
	var out PartnerBookingView
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		b, own, e := s.partnerBooking(tx, a, id, true)
		if e != nil {
			return e
		}
		if _, all := partnerShare(b, own); !all {
			return conflict("this booking includes items of other providers; contact support to change it")
		}
		switch action {
		case "complete":
			if b.Status != "confirmed" {
				return conflict("only a confirmed booking can be completed (now " + b.Status + ")")
			}
			b.Status = "completed"
			if e := tx.SaveBooking(b); e != nil {
				return e
			}
		default:
			if b.Status != "pending_payment" && b.Status != "confirmed" {
				return conflict("booking cannot be rejected in state " + b.Status)
			}
			if e := s.refundBooking(tx, b, "ผู้ให้บริการปฏิเสธ: "+strings.TrimSpace(req.Reason)); e != nil {
				return e
			}
		}
		if _, e := tx.AddNotification(b.UserID, Notification{Type: "booking", CreatedAt: s.now(),
			Title:   map[string]string{"complete": "การจองเสร็จสิ้น", "reject": "ผู้ให้บริการปฏิเสธการจอง"}[action],
			Message: "การจอง " + b.BookingNumber + map[string]string{"complete": " เสร็จสิ้นแล้ว ขอบคุณที่ใช้บริการ", "reject": " ถูกยกเลิกและคืนเงินเข้ากระเป๋าแล้ว (ถ้าชำระแล้ว)"}[action]}); e != nil {
			return e
		}
		if e := s.audit(tx, a, "partner.booking."+action, "booking", b.ID, nil); e != nil {
			return e
		}
		out = s.partnerBookingView(tx, a, own, b)
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) partnerCompleteBooking(w http.ResponseWriter, r *http.Request, a actor) {
	s.partnerAct(w, r, a, "complete")
}
func (s *Server) partnerRejectBooking(w http.ResponseWriter, r *http.Request, a actor) {
	s.partnerAct(w, r, a, "reject")
}

// ---- earnings & dashboard ----

func (s *Server) partnerEarningsFor(tx Tx, p *partnerRec) (Earnings, error) {
	own, err := ownListingIDs(tx, p.ID)
	if err != nil {
		return Earnings{}, err
	}
	bs, _, err := tx.AdminBookings(BookingFilter{Statuses: []string{"confirmed", "completed"}, PartnerID: p.ID})
	if err != nil {
		return Earnings{}, err
	}
	e := Earnings{CommissionRate: p.CommissionRate, Months: []EarningsMonth{}}
	idx := map[string]int{}
	for _, b := range bs { // newest first, so months come out newest first
		gross, _ := partnerShare(b, own)
		key := b.Date.In(ictZone).Format("2006-01")
		i, ok := idx[key]
		if !ok {
			i = len(e.Months)
			idx[key] = i
			e.Months = append(e.Months, EarningsMonth{Month: key})
		}
		e.Months[i].Bookings++
		e.Months[i].Gross += gross
	}
	for i := range e.Months {
		m := &e.Months[i]
		m.Gross = round2(m.Gross)
		m.Commission = round2(m.Gross * p.CommissionRate / 100)
		m.Net = round2(m.Gross - m.Commission)
		e.Gross += m.Gross
		e.Commission += m.Commission
		e.Net += m.Net
	}
	e.Gross, e.Commission, e.Net = round2(e.Gross), round2(e.Commission), round2(e.Net)
	return e, nil
}

func (s *Server) partnerEarnings(w http.ResponseWriter, r *http.Request, a actor) {
	var out Earnings
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { out, e = s.partnerEarningsFor(tx, a.Partner); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) partnerDashboard(w http.ResponseWriter, r *http.Request, a actor) {
	out := map[string]any{}
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		ls, e := tx.Listings(ListingFilter{PartnerID: a.Partner.ID})
		if e != nil {
			return e
		}
		byStatus := map[string]int{"draft": 0, "published": 0, "archived": 0}
		for _, l := range ls {
			byStatus[l.Status]++
		}
		count := func(st string) (int, error) {
			_, t, e := tx.AdminBookings(BookingFilter{Status: st, PartnerID: a.Partner.ID, Limit: 1})
			return t, e
		}
		toComplete, e := count("confirmed")
		if e != nil {
			return e
		}
		awaiting, e := count("pending_payment")
		if e != nil {
			return e
		}
		earn, e := s.partnerEarningsFor(tx, a.Partner)
		if e != nil {
			return e
		}
		cur := s.now().In(ictZone).Format("2006-01")
		month := EarningsMonth{Month: cur}
		for _, m := range earn.Months {
			if m.Month == cur {
				month = m
			}
		}
		out = map[string]any{"businessName": a.Partner.BusinessName, "commissionRate": a.Partner.CommissionRate, "listings": byStatus,
			"bookingsToComplete": toComplete, "bookingsAwaitingPayment": awaiting, "thisMonth": month, "totalNet": earn.Net}
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}
