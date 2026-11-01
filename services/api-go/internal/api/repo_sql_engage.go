package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// ---- settings ----

func (t *sqlTx) Settings(uid int64) (Settings, error) {
	s := DefaultSettings()
	var push, email, trip, promo bool
	err := t.row(`SELECT language, currency, theme, notify_push, notify_email, notify_trip, notify_promo FROM user_settings WHERE user_id=?`, uid).
		Scan(&s.Language, &s.Currency, &s.Theme, &push, &email, &trip, &promo)
	if errors.Is(err, sql.ErrNoRows) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	s.Notifications = NotifySetting{Push: push, Email: email, Trip: trip, Promo: promo}
	return s, nil
}

func (t *sqlTx) SaveSettings(uid int64, s Settings) error {
	_, err := t.exec(`INSERT INTO user_settings (user_id, language, currency, theme, notify_push, notify_email, notify_trip, notify_promo)
		VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE language=VALUES(language), currency=VALUES(currency), theme=VALUES(theme),
		notify_push=VALUES(notify_push), notify_email=VALUES(notify_email), notify_trip=VALUES(notify_trip), notify_promo=VALUES(notify_promo)`,
		uid, s.Language, s.Currency, s.Theme, s.Notifications.Push, s.Notifications.Email, s.Notifications.Trip, s.Notifications.Promo)
	return err
}

// ---- coupons ----

func (t *sqlTx) UpsertCoupon(c couponDef) error {
	var pct, maxd, exp any
	if c.PercentOff > 0 {
		pct = c.PercentOff
	}
	if c.MaxDiscount > 0 {
		maxd = c.MaxDiscount
	}
	exp = nullStr(c.ExpiresOn)
	_, err := t.exec(`INSERT INTO coupons (code, title, detail, percent_off, min_spend, max_discount, expires_on) VALUES (?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE title=VALUES(title), detail=VALUES(detail), percent_off=VALUES(percent_off), min_spend=VALUES(min_spend),
		max_discount=VALUES(max_discount), expires_on=VALUES(expires_on)`, c.Code, c.Title, nullStr(c.Detail), pct, c.MinSpend, maxd, exp)
	return err
}

func (t *sqlTx) GrantCoupon(uid int64, code string) error {
	var cid int64
	if err := t.row(`SELECT id FROM coupons WHERE code=?`, code).Scan(&cid); err != nil {
		return notFound(err)
	}
	_, err := t.exec(`INSERT IGNORE INTO user_coupons (user_id, coupon_id) VALUES (?,?)`, uid, cid)
	return err
}

func (t *sqlTx) UserCoupons(uid int64, today string) ([]UserCoupon, error) {
	rows, err := t.query(`SELECT c.id, c.code, c.title, COALESCE(c.detail,''), COALESCE(c.percent_off,0), c.min_spend, COALESCE(c.max_discount,0),
		COALESCE(DATE_FORMAT(c.expires_on,'%Y-%m-%d'),'') FROM user_coupons uc JOIN coupons c ON c.id=uc.coupon_id WHERE uc.user_id=? ORDER BY c.id`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []UserCoupon{}
	for rows.Next() {
		var id int64
		var c couponDef
		if err := rows.Scan(&id, &c.Code, &c.Title, &c.Detail, &c.PercentOff, &c.MinSpend, &c.MaxDiscount, &c.ExpiresOn); err != nil {
			return nil, err
		}
		out = append(out, toUserCoupon(id, c, today))
	}
	return out, rows.Err()
}

// ---- notifications ----

func (t *sqlTx) AddNotification(uid int64, n Notification) (Notification, error) {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	res, err := t.exec(`INSERT INTO notifications (user_id, type, title, message, is_read, created_at) VALUES (?,?,?,?,?,?)`,
		uid, n.Type, n.Title, n.Message, n.Read, utc(n.CreatedAt))
	if err != nil {
		return n, err
	}
	n.ID, _ = res.LastInsertId()
	return n, nil
}

func (t *sqlTx) Notifications(uid int64) ([]Notification, error) {
	rows, err := t.query(`SELECT id, type, title, message, is_read, created_at FROM notifications WHERE user_id=? ORDER BY created_at DESC, id DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Notification{}
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Type, &n.Title, &n.Message, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (t *sqlTx) MarkNotificationRead(uid, id int64) error {
	res, err := t.exec(`UPDATE notifications SET is_read=1 WHERE id=? AND user_id=?`, id, uid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 { // already read, or not the caller's
		return t.exists(`SELECT 1 FROM notifications WHERE id=? AND user_id=?`, id, uid)
	}
	return nil
}

func (t *sqlTx) MarkAllNotificationsRead(uid int64) error {
	_, err := t.exec(`UPDATE notifications SET is_read=1 WHERE user_id=? AND is_read=0`, uid)
	return err
}

// ---- trips ----

func (t *sqlTx) AddTrip(uid int64, in SavedTripInput, now time.Time) (SavedTrip, error) {
	var breakdown any
	if in.Itinerary != nil {
		b, err := json.Marshal(in.Itinerary.BudgetBreakdown)
		if err != nil {
			return SavedTrip{}, err
		}
		breakdown = string(b)
	}
	src := in.Source
	if src == "" {
		src = "manual"
	}
	res, err := t.exec(`INSERT INTO saved_trips (user_id, destination_name, days, travelers, budget_total, budget_breakdown, source, created_at) VALUES (?,?,?,?,?,?,?,?)`,
		uid, in.Destination, in.Days, in.Travelers, in.Budget, breakdown, src, utc(now))
	if err != nil {
		return SavedTrip{}, err
	}
	id, _ := res.LastInsertId()
	if in.Itinerary != nil {
		for i, d := range in.Itinerary.Days {
			dr, err := t.exec(`INSERT INTO trip_days (trip_id, day_no, title) VALUES (?,?,?)`, id, i+1, d.Title)
			if err != nil {
				return SavedTrip{}, err
			}
			did, _ := dr.LastInsertId()
			for j, it := range d.Items {
				if _, err := t.exec(`INSERT INTO trip_day_items (trip_day_id, sort_no, time_label, text, place) VALUES (?,?,?,?,?)`,
					did, j, nullStr(it.Time), it.Text, nullStr(it.Place)); err != nil {
					return SavedTrip{}, err
				}
			}
		}
	}
	return SavedTrip{ID: id, SavedTripInput: in}, nil
}

func (t *sqlTx) loadTrips(where string, args ...any) ([]SavedTrip, error) {
	rows, err := t.query(`SELECT id, destination_name, days, travelers, COALESCE(budget_total,0), budget_breakdown, source FROM saved_trips WHERE `+where+` ORDER BY id`, args...)
	if err != nil {
		return nil, err
	}
	var out []SavedTrip
	hasIt := map[int64]bool{}
	crumbs := map[int64][]byte{}
	for rows.Next() {
		var st SavedTrip
		var bd []byte
		if err := rows.Scan(&st.ID, &st.Destination, &st.Days, &st.Travelers, &st.Budget, &bd, &st.Source); err != nil {
			rows.Close()
			return nil, err
		}
		if bd != nil {
			hasIt[st.ID] = true
			crumbs[st.ID] = bd
		}
		out = append(out, st)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		id := out[i].ID
		if !hasIt[id] {
			continue
		}
		it := &ItineraryResult{Days: []ItineraryDay{}, BudgetBreakdown: []BudgetLine{}}
		_ = json.Unmarshal(crumbs[id], &it.BudgetBreakdown)
		if it.BudgetBreakdown == nil {
			it.BudgetBreakdown = []BudgetLine{}
		}
		drows, err := t.query(`SELECT d.id, d.title, i.time_label, i.text, i.place FROM trip_days d LEFT JOIN trip_day_items i ON i.trip_day_id=d.id
			WHERE d.trip_id=? ORDER BY d.day_no, i.sort_no`, id)
		if err != nil {
			return nil, err
		}
		var lastDay int64 = -1
		for drows.Next() {
			var did int64
			var title string
			var tl, tx, pl sql.NullString
			if err := drows.Scan(&did, &title, &tl, &tx, &pl); err != nil {
				drows.Close()
				return nil, err
			}
			if did != lastDay {
				it.Days = append(it.Days, ItineraryDay{Title: title, Items: []ItineraryItem{}})
				lastDay = did
			}
			if tx.Valid {
				d := &it.Days[len(it.Days)-1]
				d.Items = append(d.Items, ItineraryItem{Time: tl.String, Text: tx.String, Place: pl.String})
			}
		}
		drows.Close()
		if err := drows.Err(); err != nil {
			return nil, err
		}
		out[i].Itinerary = it
	}
	return out, nil
}

func (t *sqlTx) Trips(uid int64) ([]SavedTrip, error) {
	out, err := t.loadTrips(`user_id=?`, uid)
	if out == nil {
		out = []SavedTrip{}
	}
	return out, err
}

func (t *sqlTx) Trip(uid, id int64) (*SavedTrip, error) {
	out, err := t.loadTrips(`user_id=? AND id=?`, uid, id)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, ErrNotFound
	}
	return &out[0], nil
}

// ---- reviews, tickets, jobs ----

func (t *sqlTx) AddReview(uid, listingID int64, rating int, comment string, now time.Time) (reviewRec, error) {
	if err := t.exists(`SELECT 1 FROM listings WHERE id=?`, listingID); err != nil {
		return reviewRec{}, err
	}
	res, err := t.exec(`INSERT INTO reviews (user_id, listing_id, rating, comment, created_at) VALUES (?,?,?,?,?)`, uid, listingID, rating, comment, utc(now))
	if err != nil {
		return reviewRec{}, err
	}
	id, _ := res.LastInsertId()
	return reviewRec{ID: id, ListingID: listingID, Rating: rating, Comment: comment, CreatedAt: now}, nil
}

func (t *sqlTx) Reviews(uid int64) ([]reviewRec, error) {
	rows, err := t.query(`SELECT id, listing_id, rating, comment, created_at FROM reviews WHERE user_id=? ORDER BY id`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []reviewRec{}
	for rows.Next() {
		var r reviewRec
		if err := rows.Scan(&r.ID, &r.ListingID, &r.Rating, &r.Comment, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (t *sqlTx) AddTicket(uid int64, subject, message string, now time.Time) (SupportTicket, error) {
	res, err := t.exec(`INSERT INTO support_tickets (user_id, subject, message, created_at) VALUES (?,?,?,?)`, uid, subject, message, utc(now))
	if err != nil {
		return SupportTicket{}, err
	}
	id, _ := res.LastInsertId()
	return SupportTicket{ID: id, Subject: subject, Message: message, CreatedAt: now}, nil
}

func (t *sqlTx) Tickets(uid int64) ([]SupportTicket, error) {
	rows, err := t.query(`SELECT id, subject, message, created_at FROM support_tickets WHERE user_id=? ORDER BY id`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SupportTicket{}
	for rows.Next() {
		var s SupportTicket
		if err := rows.Scan(&s.ID, &s.Subject, &s.Message, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (t *sqlTx) AddJob(uid int64, j jobRec, now time.Time) error {
	in, err := json.Marshal(j.Input)
	if err != nil {
		return err
	}
	var result, done any
	if j.Result != nil {
		b, err := json.Marshal(j.Result)
		if err != nil {
			return err
		}
		result = string(b)
	}
	if j.Status == "completed" {
		done = utc(now)
	}
	_, err = t.exec(`INSERT INTO ai_planner_jobs (id, user_id, status, degraded, input, result, completed_at) VALUES (?,?,?,?,?,?,?)`,
		j.JobID, uid, j.Status, j.Degraded, string(in), result, done)
	return err
}

func (t *sqlTx) Job(uid int64, id string) (*ItineraryJob, error) {
	var j ItineraryJob
	var res []byte
	err := t.row(`SELECT id, status, degraded, result FROM ai_planner_jobs WHERE id=? AND user_id=?`, id, uid).Scan(&j.JobID, &j.Status, &j.Degraded, &res)
	if err != nil {
		return nil, notFound(err)
	}
	if res != nil {
		j.Result = &ItineraryResult{}
		if err := json.Unmarshal(res, j.Result); err != nil {
			return nil, err
		}
	}
	return &j, nil
}

// ---- community ----

func (t *sqlTx) UpsertGroup(g GroupDetail) error {
	rules, err := jsonOrNil(g.Rules, g.Rules == nil)
	if err != nil {
		return err
	}
	_, err = t.exec(`INSERT INTO community_groups (name, description, destination, gradient_css, rules) VALUES (?,?,?,?,?)
		ON DUPLICATE KEY UPDATE description=VALUES(description), destination=VALUES(destination), gradient_css=VALUES(gradient_css), rules=VALUES(rules)`,
		g.Name, nullStr(g.Description), g.Destination, nullStr(g.GradientCSS), rules)
	return err
}

const groupSelect = `SELECT g.id, g.name, COALESCE(g.description,''), g.destination, COALESCE(g.gradient_css,''), g.rules,
  (SELECT COUNT(*) FROM community_group_members m WHERE m.group_id=g.id),
  EXISTS(SELECT 1 FROM community_group_members m WHERE m.group_id=g.id AND m.user_id=?) FROM community_groups g`

func scanGroup(row interface{ Scan(...any) error }) (*GroupDetail, error) {
	var g GroupDetail
	var dest sql.NullString
	var rules []byte
	if err := row.Scan(&g.ID, &g.Name, &g.Description, &dest, &g.GradientCSS, &rules, &g.MemberCount, &g.Joined); err != nil {
		return nil, notFound(err)
	}
	if dest.Valid {
		d := dest.String
		g.Destination = &d
	}
	g.Rules = []string{}
	if rules != nil {
		_ = json.Unmarshal(rules, &g.Rules)
	}
	return &g, nil
}

func (t *sqlTx) Groups(uid int64) ([]Group, error) {
	rows, err := t.query(groupSelect+` ORDER BY g.id`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Group{}
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g.Group)
	}
	return out, rows.Err()
}

func (t *sqlTx) GroupByID(uid, id int64) (*GroupDetail, error) {
	return scanGroup(t.row(groupSelect+` WHERE g.id=?`, uid, id))
}

func (t *sqlTx) ToggleGroup(uid, id int64) (bool, error) {
	var one int
	if err := t.row(`SELECT 1 FROM community_groups WHERE id=? FOR UPDATE`, id).Scan(&one); err != nil {
		return false, notFound(err)
	}
	res, err := t.exec(`DELETE FROM community_group_members WHERE group_id=? AND user_id=?`, id, uid)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return false, nil
	}
	_, err = t.exec(`INSERT INTO community_group_members (group_id, user_id, joined_at) VALUES (?,?,UTC_TIMESTAMP())`, id, uid)
	return err == nil, err
}

const postSelect = `SELECT p.id, u.handle, p.post_type, p.destination, p.group_id, p.body, p.like_count, p.image_key IS NOT NULL, p.shared_trip, p.created_at,
  pr.first_name, pr.last_name,
  (SELECT COUNT(*) FROM community_comments c WHERE c.post_id=p.id),
  EXISTS(SELECT 1 FROM post_likes l WHERE l.post_id=p.id AND l.user_id=?),
  EXISTS(SELECT 1 FROM post_saves s WHERE s.post_id=p.id AND s.user_id=?)
  FROM community_posts p JOIN users u ON u.id=p.author_id JOIN user_profiles pr ON pr.user_id=u.id`

func scanPost(row interface{ Scan(...any) error }) (*Post, error) {
	var p Post
	var dest sql.NullString
	var gid sql.NullInt64
	var shared []byte
	var first, last string
	if err := row.Scan(&p.ID, &p.AuthorID, &p.Type, &dest, &gid, &p.Text, &p.LikeCount, &p.HasImage, &shared, &p.CreatedAt, &first, &last,
		&p.CommentCount, &p.Liked, &p.Saved); err != nil {
		return nil, notFound(err)
	}
	p.Destination = dest.String
	if gid.Valid {
		g := gid.Int64
		p.GroupID = &g
	}
	if shared != nil {
		p.SharedTrip = &SharedTrip{}
		if err := json.Unmarshal(shared, p.SharedTrip); err != nil {
			return nil, err
		}
	}
	p.AuthorName = memberName(UserProfile{FirstName: first, LastName: last})
	return &p, nil
}

func (t *sqlTx) Feed(viewer int64, mode, destination string) ([]Post, error) {
	q, args := postSelect+` WHERE NOT EXISTS (SELECT 1 FROM post_hides h WHERE h.post_id=p.id AND h.user_id=?)`, []any{viewer, viewer, viewer}
	if mode == "following" {
		q += ` AND p.author_id IN (SELECT followee_id FROM user_follows WHERE follower_id=?)`
		args = append(args, viewer)
	}
	if destination != "" {
		q += ` AND p.destination=?`
		args = append(args, destination)
	}
	rows, err := t.query(q+` ORDER BY p.created_at DESC, p.id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Post{}
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (t *sqlTx) PostByID(viewer, id int64) (*Post, error) {
	return scanPost(t.row(postSelect+` WHERE p.id=?`, viewer, viewer, id))
}

func (t *sqlTx) Comments(postID int64) ([]Comment, error) {
	rows, err := t.query(`SELECT c.id, u.handle, pr.first_name, pr.last_name, c.body, c.like_count, c.is_best_answer, c.created_at FROM community_comments c
		JOIN users u ON u.id=c.author_id JOIN user_profiles pr ON pr.user_id=u.id WHERE c.post_id=? ORDER BY c.id`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Comment{}
	for rows.Next() {
		var c Comment
		var first, last string
		if err := rows.Scan(&c.ID, &c.AuthorID, &first, &last, &c.Text, &c.LikeCount, &c.IsBestAnswer, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.AuthorName = memberName(UserProfile{FirstName: first, LastName: last})
		out = append(out, c)
	}
	return out, rows.Err()
}

func (t *sqlTx) AddPost(authorUID int64, p Post, imageKey string, now time.Time) (*Post, error) {
	if p.GroupID != nil {
		if err := t.exists(`SELECT 1 FROM community_groups WHERE id=?`, *p.GroupID); err != nil {
			return nil, err
		}
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	shared, err := jsonOrNil(p.SharedTrip, p.SharedTrip == nil)
	if err != nil {
		return nil, err
	}
	res, err := t.exec(`INSERT INTO community_posts (author_id, group_id, post_type, destination, body, shared_trip, image_key, like_count, created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		authorUID, p.GroupID, p.Type, nullStr(p.Destination), p.Text, shared, nullStr(imageKey), p.LikeCount, utc(p.CreatedAt))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return t.PostByID(authorUID, id)
}

func (t *sqlTx) DeletePost(uid, id int64) error {
	var author int64
	if err := t.row(`SELECT author_id FROM community_posts WHERE id=? FOR UPDATE`, id).Scan(&author); err != nil {
		return notFound(err)
	}
	if author != uid {
		return ErrForbidden
	}
	_, err := t.exec(`DELETE FROM community_posts WHERE id=?`, id)
	return err
}

func (t *sqlTx) lockPost(id int64) (likes int, err error) {
	err = notFound(t.row(`SELECT like_count FROM community_posts WHERE id=? FOR UPDATE`, id).Scan(&likes))
	return
}

func (t *sqlTx) TogglePostLike(uid, id int64) (bool, int, error) {
	n, err := t.lockPost(id)
	if err != nil {
		return false, 0, err
	}
	res, err := t.exec(`DELETE FROM post_likes WHERE post_id=? AND user_id=?`, id, uid)
	if err != nil {
		return false, 0, err
	}
	liked := false
	if c, _ := res.RowsAffected(); c > 0 {
		n = max(n-1, 0)
	} else {
		if _, err := t.exec(`INSERT INTO post_likes (post_id, user_id) VALUES (?,?)`, id, uid); err != nil {
			return false, 0, err
		}
		n++
		liked = true
	}
	_, err = t.exec(`UPDATE community_posts SET like_count=? WHERE id=?`, n, id)
	return liked, n, err
}

func (t *sqlTx) TogglePostSave(uid, id int64) (bool, error) {
	if _, err := t.lockPost(id); err != nil {
		return false, err
	}
	res, err := t.exec(`DELETE FROM post_saves WHERE post_id=? AND user_id=?`, id, uid)
	if err != nil {
		return false, err
	}
	if c, _ := res.RowsAffected(); c > 0 {
		return false, nil
	}
	_, err = t.exec(`INSERT INTO post_saves (post_id, user_id) VALUES (?,?)`, id, uid)
	return err == nil, err
}

func (t *sqlTx) HidePost(uid, id int64) error {
	if err := t.exists(`SELECT 1 FROM community_posts WHERE id=?`, id); err != nil {
		return err
	}
	_, err := t.exec(`INSERT IGNORE INTO post_hides (post_id, user_id) VALUES (?,?)`, id, uid)
	return err
}

func (t *sqlTx) ReportPost(uid, id int64, reason string, now time.Time) error {
	if err := t.exists(`SELECT 1 FROM community_posts WHERE id=?`, id); err != nil {
		return err
	}
	_, err := t.exec(`INSERT INTO post_reports (post_id, user_id, reason, created_at) VALUES (?,?,?,?)`, id, uid, reason, utc(now))
	return err
}

func (t *sqlTx) AddComment(uid, postID int64, text string, now time.Time) (Comment, error) {
	if err := t.exists(`SELECT 1 FROM community_posts WHERE id=?`, postID); err != nil {
		return Comment{}, err
	}
	res, err := t.exec(`INSERT INTO community_comments (post_id, author_id, body, created_at) VALUES (?,?,?,?)`, postID, uid, text, utc(now))
	if err != nil {
		return Comment{}, err
	}
	c := Comment{Text: text, CreatedAt: now}
	c.ID, _ = res.LastInsertId()
	u, err := t.UserByID(uid, false)
	if err != nil {
		return Comment{}, err
	}
	c.AuthorID, c.AuthorName = u.Handle, memberName(u.Profile)
	return c, nil
}

func (t *sqlTx) userIDByHandle(handle string) (int64, error) {
	var id int64
	err := t.row(`SELECT id FROM users WHERE handle=? AND status='active'`, handle).Scan(&id)
	return id, notFound(err)
}

func (t *sqlTx) Member(viewer int64, handle string) (*MemberProfile, error) {
	var m MemberProfile
	var bio sql.NullString
	var first, last string
	err := t.row(`SELECT u.handle, pr.first_name, pr.last_name, pr.level, pr.bio,
		(SELECT COUNT(*) FROM user_follows f WHERE f.followee_id=u.id),
		EXISTS(SELECT 1 FROM user_follows f WHERE f.followee_id=u.id AND f.follower_id=?)
		FROM users u JOIN user_profiles pr ON pr.user_id=u.id WHERE u.handle=? AND u.status='active'`, viewer, handle).
		Scan(&m.ID, &first, &last, &m.Level, &bio, &m.FollowerCount, &m.Following)
	if err != nil {
		return nil, notFound(err)
	}
	m.Name, m.Bio = memberName(UserProfile{FirstName: first, LastName: last}), bio.String
	return &m, nil
}

func (t *sqlTx) ToggleFollow(viewer int64, handle string) (bool, error) {
	id, err := t.userIDByHandle(handle)
	if err != nil {
		return false, err
	}
	res, err := t.exec(`DELETE FROM user_follows WHERE follower_id=? AND followee_id=?`, viewer, id)
	if err != nil {
		return false, err
	}
	if c, _ := res.RowsAffected(); c > 0 {
		return false, nil
	}
	_, err = t.exec(`INSERT INTO user_follows (follower_id, followee_id, created_at) VALUES (?,?,UTC_TIMESTAMP())`, viewer, id)
	return err == nil, err
}
