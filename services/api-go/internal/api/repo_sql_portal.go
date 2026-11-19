package api

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// MySQL implementation of the portal primitives (see repo.go, migration 0005).

type whereBuilder struct {
	conds []string
	args  []any
}

func (w *whereBuilder) add(cond string, args ...any) {
	w.conds = append(w.conds, cond)
	w.args = append(w.args, args...)
}

func (w *whereBuilder) sql() string {
	if len(w.conds) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(w.conds, " AND ")
}

// likeArg turns user input into a safe `%…%` LIKE pattern (backslash is MySQL's default escape).
func likeArg(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(strings.TrimSpace(s)) + "%"
}

func pagingSQL(offset, limit int) string {
	if limit <= 0 {
		return ""
	}
	if offset < 0 {
		offset = 0
	}
	return " LIMIT " + sqlInt(limit) + " OFFSET " + sqlInt(offset)
}

func sqlInt(n int) string { return strconv.Itoa(n) }

func (t *sqlTx) count(from string, w *whereBuilder) (int, error) {
	var n int
	err := t.row(`SELECT COUNT(*) `+from+w.sql(), w.args...).Scan(&n)
	return n, err
}

// ---- catalog ----

func (t *sqlTx) Destinations() ([]Destination, error) {
	rows, err := t.query(`SELECT id, slug, name_th, COALESCE(subtitle,''), COALESCE(gradient_css,'') FROM destinations ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Destination{}
	for rows.Next() {
		var d Destination
		if err := rows.Scan(&d.ID, &d.Slug, &d.Name, &d.Subtitle, &d.GradientCSS); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

const listingSelect = `SELECT l.id, l.destination_id, COALESCE(l.partner_id,0), l.type, l.name, l.location_text, COALESCE(l.description,''),
  l.rating_avg, l.review_count, l.price_from, l.amenities, COALESCE(l.map_query,''), l.status FROM listings l`

const listingPublicCond = `l.status='published' AND (l.partner_id IS NULL OR EXISTS (SELECT 1 FROM partners pp WHERE pp.id=l.partner_id AND pp.status='approved'))`

func scanListing(row interface{ Scan(...any) error }) (*listingRow, error) {
	var l listingRow
	var am []byte
	if err := row.Scan(&l.ID, &l.DestinationID, &l.PartnerID, &l.Type, &l.Name, &l.Location, &l.Description,
		&l.Rating, &l.ReviewCount, &l.PriceFrom, &am, &l.MapQuery, &l.Status); err != nil {
		return nil, notFound(err)
	}
	l.Amenities = []string{}
	if len(am) > 0 {
		_ = json.Unmarshal(am, &l.Amenities)
	}
	return &l, nil
}

func (t *sqlTx) Listings(f ListingFilter) ([]listingRow, error) {
	var w whereBuilder
	if f.PublicOnly {
		w.add(listingPublicCond)
	}
	if f.Type != "" {
		w.add(`l.type=?`, f.Type)
	}
	if f.DestinationID != 0 {
		w.add(`l.destination_id=?`, f.DestinationID)
	}
	if f.PartnerID != 0 {
		w.add(`l.partner_id=?`, f.PartnerID)
	}
	if f.Status != "" {
		w.add(`l.status=?`, f.Status)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		w.add(`(l.name LIKE ? OR l.location_text LIKE ?)`, likeArg(q), likeArg(q))
	}
	rows, err := t.query(listingSelect+w.sql()+` ORDER BY l.id`, w.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []listingRow
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (t *sqlTx) ListingByID(id int64, publicOnly bool) (*listingRow, error) {
	q := listingSelect + ` WHERE l.id=?`
	if publicOnly {
		q += ` AND ` + listingPublicCond
	}
	return scanListing(t.row(q, id))
}

func (t *sqlTx) SaveListing(l *listingRow) error {
	if l.Status == "" {
		l.Status = "published"
	}
	am, err := json.Marshal(l.Amenities)
	if err != nil {
		return err
	}
	var pid any
	if l.PartnerID != 0 {
		pid = l.PartnerID
	}
	if l.ID == 0 {
		res, err := t.exec(`INSERT INTO listings (destination_id, partner_id, type, name, location_text, description, rating_avg, review_count, price_from, amenities, map_query, status)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, l.DestinationID, pid, l.Type, l.Name, l.Location, nullStr(l.Description), l.Rating, l.ReviewCount, l.PriceFrom, string(am), nullStr(l.MapQuery), l.Status)
		if err != nil {
			return err
		}
		l.ID, _ = res.LastInsertId()
		return nil
	}
	_, err = t.exec(`INSERT INTO listings (id, destination_id, partner_id, type, name, location_text, description, rating_avg, review_count, price_from, amenities, map_query, status)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE destination_id=VALUES(destination_id), partner_id=VALUES(partner_id), type=VALUES(type),
		name=VALUES(name), location_text=VALUES(location_text), description=VALUES(description), rating_avg=VALUES(rating_avg), review_count=VALUES(review_count),
		price_from=VALUES(price_from), amenities=VALUES(amenities), map_query=VALUES(map_query), status=VALUES(status)`,
		l.ID, l.DestinationID, pid, l.Type, l.Name, l.Location, nullStr(l.Description), l.Rating, l.ReviewCount, l.PriceFrom, string(am), nullStr(l.MapQuery), l.Status)
	return err
}

// ---- partners ----

const partnerSelect = `SELECT p.id, p.user_id, p.business_name, COALESCE(p.contact_phone,''), COALESCE(p.description,''), p.status, p.commission_rate,
  COALESCE(p.decision_note,''), COALESCE(p.decided_by,0), p.decided_at, p.created_at FROM partners p`

func scanPartner(row interface{ Scan(...any) error }) (*partnerRec, error) {
	var p partnerRec
	var dec sql.NullTime
	if err := row.Scan(&p.ID, &p.UserID, &p.BusinessName, &p.ContactPhone, &p.Description, &p.Status, &p.CommissionRate, &p.DecisionNote, &p.DecidedBy, &dec, &p.CreatedAt); err != nil {
		return nil, notFound(err)
	}
	if dec.Valid {
		tt := dec.Time
		p.DecidedAt = &tt
	}
	return &p, nil
}

func (t *sqlTx) CreatePartner(p *partnerRec) error {
	res, err := t.exec(`INSERT INTO partners (user_id, business_name, contact_phone, description, status, commission_rate, created_at) VALUES (?,?,?,?,?,?,?)`,
		p.UserID, p.BusinessName, nullStr(p.ContactPhone), nullStr(p.Description), p.Status, p.CommissionRate, utc(p.CreatedAt))
	if err != nil {
		if isDup(err) {
			return ErrDuplicate
		}
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (t *sqlTx) PartnerByID(id int64, lock bool) (*partnerRec, error) {
	return scanPartner(t.row(partnerSelect+` WHERE p.id=?`+lockClause(lock), id))
}

func (t *sqlTx) PartnerByUser(uid int64, lock bool) (*partnerRec, error) {
	return scanPartner(t.row(partnerSelect+` WHERE p.user_id=?`+lockClause(lock), uid))
}

func (t *sqlTx) SavePartner(p *partnerRec) error {
	var by, at any
	if p.DecidedBy != 0 {
		by = p.DecidedBy
	}
	if p.DecidedAt != nil {
		at = utc(*p.DecidedAt)
	}
	_, err := t.exec(`UPDATE partners SET business_name=?, contact_phone=?, description=?, status=?, commission_rate=?, decision_note=?, decided_by=?, decided_at=? WHERE id=?`,
		p.BusinessName, nullStr(p.ContactPhone), nullStr(p.Description), p.Status, p.CommissionRate, nullStr(p.DecisionNote), by, at, p.ID)
	return err
}

func (t *sqlTx) Partners(f PartnerFilter) ([]*partnerRec, int, error) {
	var w whereBuilder
	if f.Status != "" {
		w.add(`p.status=?`, f.Status)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		w.add(`(p.business_name LIKE ? OR u.email LIKE ? OR CONCAT(pr.first_name,' ',pr.last_name) LIKE ?)`, likeArg(q), likeArg(q), likeArg(q))
	}
	const join = ` FROM partners p JOIN users u ON u.id=p.user_id JOIN user_profiles pr ON pr.user_id=u.id`
	total, err := t.count(join, &w)
	if err != nil {
		return nil, 0, err
	}
	rows, err := t.query(partnerSelect+` JOIN users u ON u.id=p.user_id JOIN user_profiles pr ON pr.user_id=u.id`+w.sql()+` ORDER BY p.id DESC`+pagingSQL(f.Offset, f.Limit), w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*partnerRec{}
	for rows.Next() {
		p, err := scanPartner(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

// ---- users (admin) ----

func (t *sqlTx) UserAny(id int64, lock bool) (*userRec, error) {
	return scanUser(t.row(userSelect+` WHERE u.id=?`+lockClause(lock), id))
}

func (t *sqlTx) AdminUsers(f UserFilter) ([]*userRec, int, error) {
	w := whereBuilder{}
	w.add(`u.status<>'deleted'`)
	if f.Role != "" {
		w.add(`u.role=?`, f.Role)
	}
	if f.Status != "" {
		w.add(`u.status=?`, f.Status)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		w.add(`(u.email LIKE ? OR u.handle LIKE ? OR CONCAT(p.first_name,' ',p.last_name) LIKE ?)`, likeArg(q), likeArg(q), likeArg(q))
	}
	total, err := t.count(`FROM users u JOIN user_profiles p ON p.user_id=u.id`, &w)
	if err != nil {
		return nil, 0, err
	}
	rows, err := t.query(userSelect+w.sql()+` ORDER BY u.id DESC`+pagingSQL(f.Offset, f.Limit), w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*userRec{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

// ---- bookings ----

func (t *sqlTx) AdminBookings(f BookingFilter) ([]*bookingRec, int, error) {
	var w whereBuilder
	if f.Status != "" {
		w.add(`status=?`, f.Status)
	}
	if len(f.Statuses) > 0 {
		ph := make([]string, len(f.Statuses))
		args := make([]any, len(f.Statuses))
		for i, st := range f.Statuses {
			ph[i], args[i] = "?", st
		}
		w.add(`status IN (`+strings.Join(ph, ",")+`)`, args...)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		w.add(`(booking_number LIKE ? OR title LIKE ?)`, likeArg(q), likeArg(q))
	}
	if !f.Since.IsZero() {
		w.add(`booked_at >= ?`, utc(f.Since))
	}
	if f.Checked != nil {
		if *f.Checked {
			w.add(`EXISTS (SELECT 1 FROM booking_checks c WHERE c.booking_id=bookings.id)`)
		} else {
			w.add(`NOT EXISTS (SELECT 1 FROM booking_checks c WHERE c.booking_id=bookings.id)`)
		}
	}
	if f.PartnerID != 0 {
		w.add(`EXISTS (SELECT 1 FROM booking_line_items li JOIN listings l ON l.id=li.listing_id WHERE li.booking_id=bookings.id AND l.partner_id=?)`, f.PartnerID)
	}
	total, err := t.count(`FROM bookings`, &w)
	if err != nil {
		return nil, 0, err
	}
	rows, err := t.query(bookingSelect+w.sql()+` ORDER BY booked_at DESC, id DESC`+pagingSQL(f.Offset, f.Limit), w.args...)
	if err != nil {
		return nil, 0, err
	}
	var out []*bookingRec
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			rows.Close()
			return nil, 0, err
		}
		out = append(out, b)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := t.loadLines(out); err != nil {
		return nil, 0, err
	}
	if out == nil {
		out = []*bookingRec{}
	}
	return out, total, nil
}

func (t *sqlTx) AddBookingCheck(c *bookingCheck) error {
	if err := t.exists(`SELECT 1 FROM bookings WHERE id=?`, c.BookingID); err != nil {
		return err
	}
	res, err := t.exec(`INSERT INTO booking_checks (booking_id, staff_id, result, note, created_at) VALUES (?,?,?,?,?)`,
		c.BookingID, c.StaffID, c.Result, nullStr(c.Note), utc(c.CreatedAt))
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	var first, last string
	if err := t.row(`SELECT first_name, last_name FROM user_profiles WHERE user_id=?`, c.StaffID).Scan(&first, &last); err == nil {
		c.StaffName = memberName(UserProfile{FirstName: first, LastName: last})
	}
	return nil
}

func (t *sqlTx) BookingChecks(bookingID int64) ([]bookingCheck, error) {
	rows, err := t.query(`SELECT c.id, c.booking_id, c.staff_id, pr.first_name, pr.last_name, c.result, COALESCE(c.note,''), c.created_at
		FROM booking_checks c JOIN user_profiles pr ON pr.user_id=c.staff_id WHERE c.booking_id=? ORDER BY c.id DESC`, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bookingCheck{}
	for rows.Next() {
		var c bookingCheck
		var first, last string
		if err := rows.Scan(&c.ID, &c.BookingID, &c.StaffID, &first, &last, &c.Result, &c.Note, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.StaffName = memberName(UserProfile{FirstName: first, LastName: last})
		out = append(out, c)
	}
	return out, rows.Err()
}

func (t *sqlTx) CheckStats(staffID int64, from, to time.Time) (verified, flagged int, err error) {
	q := `SELECT COALESCE(SUM(result='verified'),0), COALESCE(SUM(result='flagged'),0) FROM booking_checks WHERE created_at >= ? AND created_at < ?`
	args := []any{utc(from), utc(to)}
	if staffID != 0 {
		q += ` AND staff_id=?`
		args = append(args, staffID)
	}
	err = t.row(q, args...).Scan(&verified, &flagged)
	return
}

// ---- support desk ----

func (t *sqlTx) loadReplies(ids []int64) (map[int64][]TicketReply, error) {
	out := map[int64][]TicketReply{}
	if len(ids) == 0 {
		return out, nil
	}
	ph := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		ph[i], args[i] = "?", id
	}
	rows, err := t.query(`SELECT r.ticket_id, r.id, r.author_role, pr.first_name, pr.last_name, r.message, r.created_at
		FROM support_ticket_replies r JOIN user_profiles pr ON pr.user_id=r.author_id WHERE r.ticket_id IN (`+strings.Join(ph, ",")+`) ORDER BY r.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tid int64
		var r TicketReply
		var first, last string
		if err := rows.Scan(&tid, &r.ID, &r.AuthorRole, &first, &last, &r.Message, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.AuthorName = memberName(UserProfile{FirstName: first, LastName: last})
		out[tid] = append(out[tid], r)
	}
	return out, rows.Err()
}

const ticketSelect = `SELECT t.id, t.user_id, t.subject, t.message, t.status, t.created_at, t.first_replied_at, t.resolved_at,
  (SELECT COUNT(*) FROM support_ticket_replies r WHERE r.ticket_id=t.id) FROM support_tickets t`

func scanTicket(row interface{ Scan(...any) error }) (*ticketRec, error) {
	var r ticketRec
	var fr, rs sql.NullTime
	if err := row.Scan(&r.ID, &r.UserID, &r.Subject, &r.Message, &r.Status, &r.CreatedAt, &fr, &rs, &r.ReplyCount); err != nil {
		return nil, notFound(err)
	}
	if fr.Valid {
		tt := fr.Time
		r.FirstRepliedAt = &tt
	}
	if rs.Valid {
		tt := rs.Time
		r.ResolvedAt = &tt
	}
	r.Replies = []TicketReply{}
	return &r, nil
}

func (t *sqlTx) AdminTickets(f TicketFilter) ([]*ticketRec, int, error) {
	var w whereBuilder
	if f.Status != "" {
		w.add(`t.status=?`, f.Status)
	}
	if f.UserID != 0 {
		w.add(`t.user_id=?`, f.UserID)
	}
	if !f.Since.IsZero() {
		s := utc(f.Since)
		w.add(`(t.created_at >= ? OR t.first_replied_at >= ? OR t.resolved_at >= ?)`, s, s, s)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		w.add(`(t.subject LIKE ? OR t.message LIKE ? OR u.email LIKE ? OR CONCAT(pr.first_name,' ',pr.last_name) LIKE ?)`, likeArg(q), likeArg(q), likeArg(q), likeArg(q))
	}
	const join = ` JOIN users u ON u.id=t.user_id JOIN user_profiles pr ON pr.user_id=u.id`
	total, err := t.count(`FROM support_tickets t`+join, &w)
	if err != nil {
		return nil, 0, err
	}
	rows, err := t.query(ticketSelect+join+w.sql()+` ORDER BY t.created_at DESC, t.id DESC`+pagingSQL(f.Offset, f.Limit), w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*ticketRec{}
	for rows.Next() {
		r, err := scanTicket(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func (t *sqlTx) TicketByID(id int64, lock bool) (*ticketRec, error) {
	r, err := scanTicket(t.row(ticketSelect+` WHERE t.id=?`+lockClause(lock), id))
	if err != nil {
		return nil, err
	}
	rep, err := t.loadReplies([]int64{id})
	if err != nil {
		return nil, err
	}
	if x := rep[id]; x != nil {
		r.Replies = x
	}
	return r, nil
}

func (t *sqlTx) AddTicketReply(ticketID, authorID int64, authorRole, message string, now time.Time) (TicketReply, error) {
	var owner int64
	if err := t.row(`SELECT user_id FROM support_tickets WHERE id=? FOR UPDATE`, ticketID).Scan(&owner); err != nil {
		return TicketReply{}, notFound(err)
	}
	res, err := t.exec(`INSERT INTO support_ticket_replies (ticket_id, author_id, author_role, message, created_at) VALUES (?,?,?,?,?)`,
		ticketID, authorID, authorRole, message, utc(now))
	if err != nil {
		return TicketReply{}, err
	}
	if authorRole == "traveler" {
		_, err = t.exec(`UPDATE support_tickets SET status='open', resolved_at=NULL WHERE id=?`, ticketID)
	} else {
		_, err = t.exec(`UPDATE support_tickets SET status='answered', resolved_at=NULL, first_replied_at=COALESCE(first_replied_at, ?) WHERE id=?`, utc(now), ticketID)
	}
	if err != nil {
		return TicketReply{}, err
	}
	r := TicketReply{AuthorRole: authorRole, Message: message, CreatedAt: now}
	r.ID, _ = res.LastInsertId()
	var first, last string
	if err := t.row(`SELECT first_name, last_name FROM user_profiles WHERE user_id=?`, authorID).Scan(&first, &last); err == nil {
		r.AuthorName = memberName(UserProfile{FirstName: first, LastName: last})
	}
	return r, nil
}

func (t *sqlTx) SetTicketStatus(id int64, status string, now time.Time) error {
	if err := t.exists(`SELECT 1 FROM support_tickets WHERE id=? FOR UPDATE`, id); err != nil {
		return err
	}
	var resolved any
	if status == "resolved" {
		resolved = utc(now)
	}
	_, err := t.exec(`UPDATE support_tickets SET status=?, resolved_at=? WHERE id=?`, status, resolved, id)
	return err
}

// ---- audit & stats ----

func (t *sqlTx) AddAudit(a auditRec) error {
	var meta, actor any
	if len(a.Meta) > 0 {
		b, err := json.Marshal(a.Meta)
		if err != nil {
			return err
		}
		meta = string(b)
	}
	if a.ActorID != 0 {
		actor = a.ActorID
	}
	_, err := t.exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, meta, created_at) VALUES (?,?,?,?,?,?)`,
		actor, a.Action, nullStr(a.Entity), nullStr(a.EntityID), meta, utc(a.CreatedAt))
	return err
}

func (t *sqlTx) Audits(f AuditFilter) ([]auditRec, int, error) {
	var w whereBuilder
	if f.Action != "" {
		w.add(`a.action LIKE ?`, strings.TrimSpace(strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(f.Action))+"%")
	}
	if f.ActorID != 0 {
		w.add(`a.user_id=?`, f.ActorID)
	}
	total, err := t.count(`FROM audit_logs a`, &w)
	if err != nil {
		return nil, 0, err
	}
	rows, err := t.query(`SELECT a.id, COALESCE(a.user_id,0), COALESCE(pr.first_name,''), COALESCE(pr.last_name,''), a.action, COALESCE(a.entity,''), COALESCE(a.entity_id,''), a.meta, a.created_at
		FROM audit_logs a LEFT JOIN user_profiles pr ON pr.user_id=a.user_id`+w.sql()+` ORDER BY a.id DESC`+pagingSQL(f.Offset, f.Limit), w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []auditRec{}
	for rows.Next() {
		var a auditRec
		var first, last string
		var meta []byte
		if err := rows.Scan(&a.ID, &a.ActorID, &first, &last, &a.Action, &a.Entity, &a.EntityID, &meta, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		if first != "" {
			a.ActorName = memberName(UserProfile{FirstName: first, LastName: last})
		}
		if len(meta) > 0 {
			_ = json.Unmarshal(meta, &a.Meta)
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

func (t *sqlTx) Stats(since time.Time) (PlatformStats, error) {
	st := PlatformStats{ByRole: map[string]int{}, Partners: map[string]int{}}
	rows, err := t.query(`SELECT role, status, COUNT(*) FROM users WHERE status IN ('active','suspended') GROUP BY role, status`)
	if err != nil {
		return st, err
	}
	for rows.Next() {
		var role, status string
		var n int
		if err := rows.Scan(&role, &status, &n); err != nil {
			rows.Close()
			return st, err
		}
		st.Users += n
		if status == "suspended" {
			st.UsersSuspended += n
		} else {
			st.ByRole[role] += n
		}
	}
	rows.Close()
	prows, err := t.query(`SELECT status, COUNT(*) FROM partners GROUP BY status`)
	if err != nil {
		return st, err
	}
	for prows.Next() {
		var status string
		var n int
		if err := prows.Scan(&status, &n); err != nil {
			prows.Close()
			return st, err
		}
		st.Partners[status] = n
	}
	prows.Close()
	if err := t.row(`SELECT COUNT(*) FROM listings l WHERE ` + listingPublicCond).Scan(&st.ListingsPublished); err != nil {
		return st, err
	}
	if err := t.row(`SELECT COUNT(*), COALESCE(SUM(booked_at >= ?),0), COALESCE(SUM(CASE WHEN status IN ('confirmed','completed') THEN total ELSE 0 END),0) FROM bookings`, utc(since)).
		Scan(&st.BookingsTotal, &st.BookingsSince, &st.Revenue); err != nil {
		return st, err
	}
	st.Revenue = round2(st.Revenue)
	return st, nil
}
