package api

import (
	"sort"
	"strings"
	"time"
)

// In-memory implementation of the portal primitives (see repo.go). Runs under the memRepo mutex.

func cloneListing(l *listingRow) listingRow {
	cp := *l
	cp.Amenities = append([]string{}, l.Amenities...)
	return cp
}

func (m *memRepo) Destinations() ([]Destination, error) {
	return append([]Destination{}, seedDestinations...), nil
}

func (m *memRepo) listingPublic(l *listingRow) bool {
	if l.Status != "published" {
		return false
	}
	if l.PartnerID == 0 {
		return true
	}
	p := m.partners[l.PartnerID]
	return p != nil && p.Status == PartnerApproved
}

func (m *memRepo) Listings(f ListingFilter) ([]listingRow, error) {
	q := strings.ToLower(strings.TrimSpace(f.Q))
	var out []listingRow
	for _, l := range m.listings {
		if f.PublicOnly && !m.listingPublic(l) {
			continue
		}
		if f.Type != "" && l.Type != f.Type {
			continue
		}
		if f.DestinationID != 0 && l.DestinationID != f.DestinationID {
			continue
		}
		if f.PartnerID != 0 && l.PartnerID != f.PartnerID {
			continue
		}
		if f.Status != "" && l.Status != f.Status {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(l.Name+" "+l.Location), q) {
			continue
		}
		out = append(out, cloneListing(l))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *memRepo) ListingByID(id int64, publicOnly bool) (*listingRow, error) {
	l, ok := m.listings[id]
	if !ok || (publicOnly && !m.listingPublic(l)) {
		return nil, ErrNotFound
	}
	cp := cloneListing(l)
	return &cp, nil
}

func (m *memRepo) SaveListing(l *listingRow) error {
	if l.Status == "" {
		l.Status = "published"
	}
	if l.ID == 0 {
		l.ID = m.nextListing
	}
	if l.ID >= m.nextListing {
		m.nextListing = l.ID + 1
	}
	cp := cloneListing(l)
	m.listings[l.ID] = &cp
	return nil
}

// ---- partners ----

func (m *memRepo) CreatePartner(p *partnerRec) error {
	for _, x := range m.partners {
		if x.UserID == p.UserID {
			return ErrDuplicate
		}
	}
	p.ID = m.nextPartner
	m.nextPartner++
	cp := *p
	m.partners[p.ID] = &cp
	return nil
}

func (m *memRepo) PartnerByID(id int64, _ bool) (*partnerRec, error) {
	p, ok := m.partners[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *memRepo) PartnerByUser(uid int64, _ bool) (*partnerRec, error) {
	for _, p := range m.partners {
		if p.UserID == uid {
			cp := *p
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *memRepo) SavePartner(p *partnerRec) error {
	if _, ok := m.partners[p.ID]; !ok {
		return ErrNotFound
	}
	cp := *p
	m.partners[p.ID] = &cp
	return nil
}

func page[T any](all []T, offset, limit int) []T {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(all) {
		return []T{}
	}
	end := len(all)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return all[offset:end]
}

func (m *memRepo) Partners(f PartnerFilter) ([]*partnerRec, int, error) {
	q := strings.ToLower(strings.TrimSpace(f.Q))
	var all []*partnerRec
	for _, p := range m.partners {
		if f.Status != "" && p.Status != f.Status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.BusinessName)
			if u := m.users[p.UserID]; u != nil {
				hay += " " + strings.ToLower(u.Profile.Email+" "+u.Profile.FirstName+" "+u.Profile.LastName)
			}
			if !strings.Contains(hay, q) {
				continue
			}
		}
		cp := *p
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID > all[j].ID })
	return page(all, f.Offset, f.Limit), len(all), nil
}

// ---- users (admin) ----

func (m *memRepo) UserAny(id int64, _ bool) (*userRec, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *memRepo) AdminUsers(f UserFilter) ([]*userRec, int, error) {
	q := strings.ToLower(strings.TrimSpace(f.Q))
	var all []*userRec
	for _, u := range m.users {
		if u.Status == "deleted" {
			continue
		}
		if f.Role != "" && u.Role != f.Role {
			continue
		}
		if f.Status != "" && u.Status != f.Status {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(u.Profile.Email+" "+u.Handle+" "+u.Profile.FirstName+" "+u.Profile.LastName), q) {
			continue
		}
		cp := *u
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID > all[j].ID })
	return page(all, f.Offset, f.Limit), len(all), nil
}

// ---- bookings (admin / staff / partner) ----

func (m *memRepo) hasCheck(bookingID int64) bool {
	for _, c := range m.checks {
		if c.BookingID == bookingID {
			return true
		}
	}
	return false
}

func (m *memRepo) AdminBookings(f BookingFilter) ([]*bookingRec, int, error) {
	q := strings.ToLower(strings.TrimSpace(f.Q))
	var all []*bookingRec
	for _, b := range m.bookings {
		if f.Status != "" && b.Status != f.Status {
			continue
		}
		if len(f.Statuses) > 0 {
			ok := false
			for _, st := range f.Statuses {
				ok = ok || st == b.Status
			}
			if !ok {
				continue
			}
		}
		if q != "" && !strings.Contains(strings.ToLower(b.BookingNumber+" "+b.Title), q) {
			continue
		}
		if !f.Since.IsZero() && b.Date.Before(f.Since) {
			continue
		}
		if f.Checked != nil && m.hasCheck(b.ID) != *f.Checked {
			continue
		}
		if f.PartnerID != 0 {
			mine := false
			for _, ln := range b.Lines {
				if l := m.listings[ln.ListingID]; l != nil && l.PartnerID == f.PartnerID {
					mine = true
				}
			}
			if !mine {
				continue
			}
		}
		cp := *b
		cp.Lines = append([]bookingLine{}, b.Lines...)
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Date.Equal(all[j].Date) {
			return all[i].ID > all[j].ID
		}
		return all[i].Date.After(all[j].Date)
	})
	return page(all, f.Offset, f.Limit), len(all), nil
}

func (m *memRepo) AddBookingCheck(c *bookingCheck) error {
	if _, ok := m.bookings[c.BookingID]; !ok {
		return ErrNotFound
	}
	c.ID = m.nextCheck
	m.nextCheck++
	if u := m.users[c.StaffID]; u != nil {
		c.StaffName = memberName(u.Profile)
	}
	m.checks = append(m.checks, *c)
	return nil
}

func (m *memRepo) BookingChecks(bookingID int64) ([]bookingCheck, error) {
	out := []bookingCheck{}
	for i := len(m.checks) - 1; i >= 0; i-- {
		if m.checks[i].BookingID == bookingID {
			out = append(out, m.checks[i])
		}
	}
	return out, nil
}

func (m *memRepo) CheckStats(staffID int64, from, to time.Time) (verified, flagged int, err error) {
	for _, c := range m.checks {
		if (staffID != 0 && c.StaffID != staffID) || c.CreatedAt.Before(from) || !c.CreatedAt.Before(to) {
			continue
		}
		if c.Result == "verified" {
			verified++
		} else {
			flagged++
		}
	}
	return
}

// ---- support desk ----

func (m *memRepo) AdminTickets(f TicketFilter) ([]*ticketRec, int, error) {
	q := strings.ToLower(strings.TrimSpace(f.Q))
	var all []*ticketRec
	for _, t := range m.tickets {
		if f.Status != "" && t.Status != f.Status {
			continue
		}
		if f.UserID != 0 && t.UserID != f.UserID {
			continue
		}
		if !f.Since.IsZero() {
			act := !t.CreatedAt.Before(f.Since) ||
				(t.FirstRepliedAt != nil && !t.FirstRepliedAt.Before(f.Since)) ||
				(t.ResolvedAt != nil && !t.ResolvedAt.Before(f.Since))
			if !act {
				continue
			}
		}
		if q != "" {
			hay := strings.ToLower(t.Subject + " " + t.Message)
			if u := m.users[t.UserID]; u != nil {
				hay += " " + strings.ToLower(u.Profile.Email+" "+u.Profile.FirstName+" "+u.Profile.LastName)
			}
			if !strings.Contains(hay, q) {
				continue
			}
		}
		all = append(all, m.ticketOut(t, false))
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].CreatedAt.Equal(all[j].CreatedAt) {
			return all[i].ID > all[j].ID
		}
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	return page(all, f.Offset, f.Limit), len(all), nil
}

func (m *memRepo) TicketByID(id int64, _ bool) (*ticketRec, error) {
	t, ok := m.tickets[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m.ticketOut(t, true), nil
}

func (m *memRepo) AddTicketReply(ticketID, authorID int64, authorRole, message string, now time.Time) (TicketReply, error) {
	t, ok := m.tickets[ticketID]
	if !ok {
		return TicketReply{}, ErrNotFound
	}
	r := TicketReply{ID: m.nextReply, AuthorRole: authorRole, Message: message, CreatedAt: now}
	m.nextReply++
	if u := m.users[authorID]; u != nil {
		r.AuthorName = memberName(u.Profile)
	}
	m.replies[ticketID] = append(m.replies[ticketID], r)
	if authorRole == "traveler" {
		t.Status, t.ResolvedAt = "open", nil
	} else {
		t.Status, t.ResolvedAt = "answered", nil
		if t.FirstRepliedAt == nil {
			ts := now
			t.FirstRepliedAt = &ts
		}
	}
	return r, nil
}

func (m *memRepo) SetTicketStatus(id int64, status string, now time.Time) error {
	t, ok := m.tickets[id]
	if !ok {
		return ErrNotFound
	}
	t.Status = status
	if status == "resolved" {
		ts := now
		t.ResolvedAt = &ts
	} else {
		t.ResolvedAt = nil
	}
	return nil
}

// ---- audit & stats ----

func (m *memRepo) AddAudit(a auditRec) error {
	a.ID = m.nextAudit
	m.nextAudit++
	if u := m.users[a.ActorID]; u != nil {
		a.ActorName = memberName(u.Profile)
	}
	m.audits = append(m.audits, a)
	return nil
}

func (m *memRepo) Audits(f AuditFilter) ([]auditRec, int, error) {
	var all []auditRec
	for i := len(m.audits) - 1; i >= 0; i-- {
		a := m.audits[i]
		if f.Action != "" && !strings.HasPrefix(a.Action, f.Action) {
			continue
		}
		if f.ActorID != 0 && a.ActorID != f.ActorID {
			continue
		}
		all = append(all, a)
	}
	return page(all, f.Offset, f.Limit), len(all), nil
}

func (m *memRepo) Stats(since time.Time) (PlatformStats, error) {
	st := PlatformStats{ByRole: map[string]int{}, Partners: map[string]int{}}
	for _, u := range m.users {
		switch u.Status {
		case "active":
			st.Users++
			st.ByRole[u.Role]++
		case "suspended":
			st.Users++
			st.UsersSuspended++
		}
	}
	for _, p := range m.partners {
		st.Partners[p.Status]++
	}
	for _, l := range m.listings {
		if m.listingPublic(l) {
			st.ListingsPublished++
		}
	}
	for _, b := range m.bookings {
		st.BookingsTotal++
		if !b.Date.Before(since) {
			st.BookingsSince++
		}
		if b.Status == "confirmed" || b.Status == "completed" {
			st.Revenue += b.Amount
		}
	}
	st.Revenue = round2(st.Revenue)
	return st, nil
}
