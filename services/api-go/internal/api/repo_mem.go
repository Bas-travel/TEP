package api

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// DefaultSettings are used until a user saves settings.
func DefaultSettings() Settings {
	return Settings{Language: "th", Currency: "THB", Theme: "light", Notifications: NotifySetting{Push: true, Email: true, Trip: true, Promo: true}}
}

type sessRec struct {
	uid     int64
	expires time.Time
	used    bool
}

type walletRec struct {
	bal float64
	txs []WalletTx
}

type payRec struct {
	intent, method, status string
	amount                 float64
}

type postRec struct {
	Post
	authorUID int64
	imageKey  string
}

type commentRec struct {
	Comment
	postID    int64
	authorUID int64
}

type groupRec struct {
	GroupDetail
	members map[int64]bool
}

type memRepo struct {
	mu sync.Mutex

	users    map[int64]*userRec
	byEmail  map[string]int64
	byHandle map[string]int64
	nextUser int64
	sess     map[string]*sessRec
	push     map[string]int64

	wallets  map[int64]*walletRec
	nextTx   int64
	bookings map[int64]*bookingRec
	nextBk   int64
	intents  map[string]*intentRec
	pays     map[int64]*payRec
	idem     map[string]idemRec
	webhooks map[string]bool

	settings   map[int64]Settings
	coupons    map[string]*couponDef
	couponIDs  map[string]int64
	nextCoupon int64
	granted    map[int64][]string
	notifs     map[int64][]Notification
	nextNotif  int64
	trips      map[int64][]SavedTrip
	nextTrip   int64
	reviews    map[int64][]reviewRec
	nextRev    int64
	tickets    map[int64][]SupportTicket
	nextTicket int64
	jobs       map[string]jobRec
	jobOwner   map[string]int64

	groups    map[int64]*groupRec
	nextGroup int64
	posts     map[int64]*postRec
	nextPost  int64
	comments  []commentRec
	nextCmt   int64
	likes     map[int64]map[int64]bool // post -> user
	saves     map[int64]map[int64]bool
	hides     map[int64]map[int64]bool // user -> post
	follows   map[int64]map[int64]bool // follower -> followee
	reports   int
}

func newMemRepo() *memRepo {
	return &memRepo{
		users: map[int64]*userRec{}, byEmail: map[string]int64{}, byHandle: map[string]int64{}, nextUser: 1,
		sess: map[string]*sessRec{}, push: map[string]int64{},
		wallets: map[int64]*walletRec{}, nextTx: 1, bookings: map[int64]*bookingRec{}, nextBk: 1,
		intents: map[string]*intentRec{}, pays: map[int64]*payRec{}, idem: map[string]idemRec{}, webhooks: map[string]bool{},
		settings: map[int64]Settings{}, coupons: map[string]*couponDef{}, couponIDs: map[string]int64{}, nextCoupon: 1,
		granted: map[int64][]string{}, notifs: map[int64][]Notification{}, nextNotif: 1,
		trips: map[int64][]SavedTrip{}, nextTrip: 1, reviews: map[int64][]reviewRec{}, nextRev: 1,
		tickets: map[int64][]SupportTicket{}, nextTicket: 1, jobs: map[string]jobRec{}, jobOwner: map[string]int64{},
		groups: map[int64]*groupRec{}, nextGroup: 1, posts: map[int64]*postRec{}, nextPost: 1, nextCmt: 1,
		likes: map[int64]map[int64]bool{}, saves: map[int64]map[int64]bool{}, hides: map[int64]map[int64]bool{}, follows: map[int64]map[int64]bool{},
	}
}

func (m *memRepo) Tx(_ context.Context, fn func(Tx) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fn(m)
}

func (m *memRepo) Close() error { return nil }

// ---- users & sessions ----

func (m *memRepo) CreateUser(u *userRec, now time.Time) error {
	email := strings.ToLower(u.Profile.Email)
	if _, dup := m.byEmail[email]; dup {
		return ErrDuplicate
	}
	id := m.nextUser
	if u.Handle == "" {
		u.Handle = fmt.Sprintf("m_%d", id)
	}
	if _, dup := m.byHandle[u.Handle]; dup {
		return ErrDuplicate
	}
	m.nextUser++
	u.ID, u.Profile.ID = id, id
	if u.Status == "" {
		u.Status = "active"
	}
	if u.Role == "" {
		u.Role = "traveler"
	}
	if u.Profile.Level == "" {
		u.Profile.Level = "Explorer"
	}
	cp := *u
	m.users[id] = &cp
	m.byEmail[email] = id
	m.byHandle[u.Handle] = id
	m.wallets[id] = &walletRec{}
	return nil
}

func (m *memRepo) UserByID(id int64, _ bool) (*userRec, error) {
	u, ok := m.users[id]
	if !ok || u.Status != "active" {
		return nil, ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *memRepo) UserByEmail(email string) (*userRec, error) {
	id, ok := m.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, ErrNotFound
	}
	return m.UserByID(id, false)
}

func (m *memRepo) SaveUser(u *userRec) error {
	if _, ok := m.users[u.ID]; !ok {
		return ErrNotFound
	}
	cp := *u
	m.users[u.ID] = &cp
	return nil
}

func (m *memRepo) DeleteUser(id int64, _ time.Time) error {
	u, ok := m.users[id]
	if !ok {
		return ErrNotFound
	}
	delete(m.byEmail, strings.ToLower(u.Profile.Email))
	delete(m.byHandle, u.Handle)
	u.Status = "deleted"
	u.Profile = UserProfile{ID: id, Level: "Explorer", Email: fmt.Sprintf("deleted+%d@tep.invalid", id), FirstName: "ผู้ใช้ที่ลบแล้ว"}
	u.Handle = fmt.Sprintf("deleted_%d", id)
	u.PassHash = ""
	_ = m.RevokeAllRefresh(id)
	for t, uid := range m.push {
		if uid == id {
			delete(m.push, t)
		}
	}
	delete(m.settings, id)
	delete(m.granted, id)
	delete(m.notifs, id)
	delete(m.trips, id)
	delete(m.reviews, id)
	delete(m.tickets, id)
	delete(m.follows, id)
	for _, f := range m.follows {
		delete(f, id)
	}
	for pid, p := range m.posts {
		if p.authorUID == id {
			m.deletePost(pid)
		}
	}
	kept := m.comments[:0]
	for _, c := range m.comments {
		if c.authorUID != id {
			kept = append(kept, c)
		}
	}
	m.comments = kept
	for _, g := range m.groups {
		delete(g.members, id)
	}
	for _, l := range m.likes {
		delete(l, id)
	}
	for _, l := range m.saves {
		delete(l, id)
	}
	delete(m.hides, id)
	return nil
}

func (m *memRepo) AddRefresh(hash string, uid int64, expires, _ time.Time) error {
	m.sess[hash] = &sessRec{uid: uid, expires: expires}
	return nil
}

func (m *memRepo) UseRefresh(hash string, now time.Time) (int64, RefreshState, error) {
	s, ok := m.sess[hash]
	if !ok {
		return 0, RefreshUnknown, nil
	}
	if s.used {
		return s.uid, RefreshReused, nil
	}
	s.used = true
	if now.After(s.expires) {
		return s.uid, RefreshExpired, nil
	}
	return s.uid, RefreshOK, nil
}

func (m *memRepo) RevokeRefresh(hash string, uid int64) error {
	if s, ok := m.sess[hash]; ok && s.uid == uid {
		delete(m.sess, hash)
	}
	return nil
}

func (m *memRepo) RevokeAllRefresh(uid int64) error {
	for h, s := range m.sess {
		if s.uid == uid {
			delete(m.sess, h)
		}
	}
	return nil
}

func (m *memRepo) UpsertPushToken(uid int64, token, _ string, _ time.Time) error {
	m.push[token] = uid
	return nil
}

// ---- wallet ----

func (m *memRepo) wallet(uid int64) *walletRec {
	w := m.wallets[uid]
	if w == nil {
		w = &walletRec{}
		m.wallets[uid] = w
	}
	return w
}

func (m *memRepo) WalletBalance(uid int64, _ bool) (float64, error) { return m.wallet(uid).bal, nil }

func (m *memRepo) WalletAdjust(uid int64, delta float64, label string, now time.Time) (float64, error) {
	w := m.wallet(uid)
	nb := round2(w.bal + delta)
	if nb < 0 {
		return w.bal, ErrInsufficient
	}
	dir, amt := "in", delta
	if delta < 0 {
		dir, amt = "out", -delta
	}
	w.bal = nb
	w.txs = append(w.txs, WalletTx{ID: m.nextTx, Direction: dir, Label: label, Amount: round2(amt), CreatedAt: now})
	m.nextTx++
	return nb, nil
}

func (m *memRepo) WalletTxs(uid int64) ([]WalletTx, error) {
	out := append([]WalletTx{}, m.wallet(uid).txs...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

// ---- bookings & payments ----

func (m *memRepo) CreateBooking(b *bookingRec) error {
	b.ID = m.nextBk
	m.nextBk++
	b.BookingNumber = bookingNumber(b.ID, b.Date)
	cp := *b
	m.bookings[b.ID] = &cp
	return nil
}

func (m *memRepo) GetBooking(id int64, _ bool) (*bookingRec, error) {
	b, ok := m.bookings[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (m *memRepo) SaveBooking(b *bookingRec) error {
	if _, ok := m.bookings[b.ID]; !ok {
		return ErrNotFound
	}
	cp := *b
	m.bookings[b.ID] = &cp
	return nil
}

func (m *memRepo) ListBookings(uid int64, status string) ([]BookingSummary, error) {
	out := []BookingSummary{}
	for _, b := range m.bookings {
		if b.UserID == uid && (status == "" || status == "all" || b.Status == status) {
			out = append(out, b.BookingSummary)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date.Equal(out[j].Date) {
			return out[i].ID > out[j].ID
		}
		return out[i].Date.After(out[j].Date)
	})
	return out, nil
}

func (m *memRepo) CreateIntent(pi *intentRec) error {
	if _, dup := m.intents[pi.ID]; dup {
		return ErrDuplicate
	}
	cp := *pi
	m.intents[pi.ID] = &cp
	return nil
}

func (m *memRepo) GetIntent(id string, _ bool) (*intentRec, error) {
	pi, ok := m.intents[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *pi
	return &cp, nil
}

func (m *memRepo) SaveIntent(pi *intentRec) error {
	if _, ok := m.intents[pi.ID]; !ok {
		return ErrNotFound
	}
	cp := *pi
	m.intents[pi.ID] = &cp
	return nil
}

func (m *memRepo) PendingIntents(uid int64) ([]*intentRec, error) {
	var out []*intentRec
	for _, pi := range m.intents {
		if pi.UserID == uid && pi.Status == "pending" {
			cp := *pi
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *memRepo) RecordPayment(bookingID int64, intentID, method string, amount float64, status string, _ time.Time) error {
	m.pays[bookingID] = &payRec{intent: intentID, method: method, status: status, amount: amount}
	return nil
}

func (m *memRepo) MarkPaymentRefunded(bookingID int64) error {
	if p := m.pays[bookingID]; p != nil {
		p.status = "refunded"
	}
	return nil
}

func idemKey(uid int64, endpoint, key string) string {
	return fmt.Sprintf("%d|%s|%s", uid, endpoint, key)
}

func (m *memRepo) GetIdem(uid int64, endpoint, key string) (*idemRec, error) {
	r, ok := m.idem[idemKey(uid, endpoint, key)]
	if !ok {
		return nil, ErrNotFound
	}
	return &r, nil
}

func (m *memRepo) PutIdem(uid int64, endpoint, key string, rec idemRec, _ time.Time) error {
	k := idemKey(uid, endpoint, key)
	if _, dup := m.idem[k]; dup {
		return ErrDuplicate
	}
	m.idem[k] = rec
	return nil
}

func (m *memRepo) AddWebhookEvent(provider, eventID string, _ time.Time) error {
	k := provider + "|" + eventID
	if m.webhooks[k] {
		return ErrDuplicate
	}
	m.webhooks[k] = true
	return nil
}

// ---- engagement ----

func (m *memRepo) Settings(uid int64) (Settings, error) {
	if s, ok := m.settings[uid]; ok {
		return s, nil
	}
	return DefaultSettings(), nil
}

func (m *memRepo) SaveSettings(uid int64, s Settings) error {
	m.settings[uid] = s
	return nil
}

func (m *memRepo) UpsertCoupon(c couponDef) error {
	cp := c
	if _, ok := m.couponIDs[c.Code]; !ok {
		m.couponIDs[c.Code] = m.nextCoupon
		m.nextCoupon++
	}
	m.coupons[c.Code] = &cp
	return nil
}

func (m *memRepo) GrantCoupon(uid int64, code string) error {
	if _, ok := m.coupons[code]; !ok {
		return ErrNotFound
	}
	for _, c := range m.granted[uid] {
		if c == code {
			return nil
		}
	}
	m.granted[uid] = append(m.granted[uid], code)
	return nil
}

func (m *memRepo) UserCoupons(uid int64, today string) ([]UserCoupon, error) {
	out := []UserCoupon{}
	for _, code := range m.granted[uid] {
		out = append(out, toUserCoupon(m.couponIDs[code], *m.coupons[code], today))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func toUserCoupon(id int64, c couponDef, today string) UserCoupon {
	st, pct := "available", float64(c.PercentOff)/100
	if c.PercentOff == 0 {
		pct = 1 // fixed amount: MaxDiscount
	}
	if c.ExpiresOn != "" && c.ExpiresOn < today {
		st = "expired"
	}
	return UserCoupon{ID: id, Code: c.Code, Title: c.Title, Detail: c.Detail, MinSpend: c.MinSpend, MaxDiscount: c.MaxDiscount, Expiry: c.ExpiresOn, Status: st, pct: pct}
}

func (m *memRepo) AddNotification(uid int64, n Notification) (Notification, error) {
	n.ID = m.nextNotif
	m.nextNotif++
	m.notifs[uid] = append(m.notifs[uid], n)
	return n, nil
}

func (m *memRepo) Notifications(uid int64) ([]Notification, error) {
	out := append([]Notification{}, m.notifs[uid]...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (m *memRepo) MarkNotificationRead(uid, id int64) error {
	for i := range m.notifs[uid] {
		if m.notifs[uid][i].ID == id {
			m.notifs[uid][i].Read = true
			return nil
		}
	}
	return ErrNotFound
}

func (m *memRepo) MarkAllNotificationsRead(uid int64) error {
	for i := range m.notifs[uid] {
		m.notifs[uid][i].Read = true
	}
	return nil
}

func (m *memRepo) AddTrip(uid int64, in SavedTripInput, _ time.Time) (SavedTrip, error) {
	t := SavedTrip{ID: m.nextTrip, SavedTripInput: in}
	m.nextTrip++
	m.trips[uid] = append(m.trips[uid], t)
	return t, nil
}

func (m *memRepo) Trips(uid int64) ([]SavedTrip, error) {
	return append([]SavedTrip{}, m.trips[uid]...), nil
}

func (m *memRepo) Trip(uid, id int64) (*SavedTrip, error) {
	for _, t := range m.trips[uid] {
		if t.ID == id {
			cp := t
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *memRepo) AddReview(uid, listingID int64, rating int, comment string, now time.Time) (reviewRec, error) {
	r := reviewRec{ID: m.nextRev, ListingID: listingID, Rating: rating, Comment: comment, CreatedAt: now}
	m.nextRev++
	m.reviews[uid] = append(m.reviews[uid], r)
	return r, nil
}

func (m *memRepo) Reviews(uid int64) ([]reviewRec, error) {
	return append([]reviewRec{}, m.reviews[uid]...), nil
}

func (m *memRepo) AddTicket(uid int64, subject, message string, now time.Time) (SupportTicket, error) {
	t := SupportTicket{ID: m.nextTicket, Subject: subject, Message: message, CreatedAt: now}
	m.nextTicket++
	m.tickets[uid] = append(m.tickets[uid], t)
	return t, nil
}

func (m *memRepo) Tickets(uid int64) ([]SupportTicket, error) {
	return append([]SupportTicket{}, m.tickets[uid]...), nil
}

func (m *memRepo) AddJob(uid int64, j jobRec, _ time.Time) error {
	m.jobs[j.JobID] = j
	m.jobOwner[j.JobID] = uid
	return nil
}

func (m *memRepo) Job(uid int64, id string) (*ItineraryJob, error) {
	j, ok := m.jobs[id]
	if !ok || m.jobOwner[id] != uid {
		return nil, ErrNotFound
	}
	cp := j.ItineraryJob
	return &cp, nil
}

// ---- community ----

func (m *memRepo) UpsertGroup(g GroupDetail) error {
	for _, ex := range m.groups {
		if ex.Name == g.Name {
			ex.Description, ex.Destination, ex.GradientCSS, ex.Rules = g.Description, g.Destination, g.GradientCSS, g.Rules
			return nil
		}
	}
	g.ID = m.nextGroup
	m.nextGroup++
	m.groups[g.ID] = &groupRec{GroupDetail: g, members: map[int64]bool{}}
	return nil
}

func (m *memRepo) groupOut(g *groupRec, uid int64) GroupDetail {
	out := g.GroupDetail
	out.MemberCount = len(g.members)
	out.Joined = g.members[uid]
	return out
}

func (m *memRepo) Groups(uid int64) ([]Group, error) {
	out := []Group{}
	for _, g := range m.groups {
		out = append(out, m.groupOut(g, uid).Group)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *memRepo) GroupByID(uid, id int64) (*GroupDetail, error) {
	g, ok := m.groups[id]
	if !ok {
		return nil, ErrNotFound
	}
	out := m.groupOut(g, uid)
	return &out, nil
}

// JoinGroup is used by seeding (idempotent); ToggleGroup is the API operation.
func (m *memRepo) ToggleGroup(uid, id int64) (bool, error) {
	g, ok := m.groups[id]
	if !ok {
		return false, ErrNotFound
	}
	if g.members[uid] {
		delete(g.members, uid)
		return false, nil
	}
	g.members[uid] = true
	return true, nil
}

func (m *memRepo) postOut(p *postRec, viewer int64) Post {
	out := p.Post
	if a := m.users[p.authorUID]; a != nil {
		out.AuthorID, out.AuthorName = a.Handle, memberName(a.Profile)
	}
	out.Liked, out.Saved = m.likes[p.ID][viewer], m.saves[p.ID][viewer]
	out.HasImage = p.imageKey != ""
	n := 0
	for _, c := range m.comments {
		if c.postID == p.ID {
			n++
		}
	}
	out.CommentCount = n
	return out
}

func (m *memRepo) Feed(viewer int64, mode, destination string) ([]Post, error) {
	var recs []*postRec
	for _, p := range m.posts {
		if m.hides[viewer][p.ID] {
			continue
		}
		if mode == "following" && !m.follows[viewer][p.authorUID] {
			continue
		}
		if destination != "" && p.Destination != destination {
			continue
		}
		recs = append(recs, p)
	}
	sort.Slice(recs, func(i, j int) bool {
		if recs[i].CreatedAt.Equal(recs[j].CreatedAt) {
			return recs[i].ID > recs[j].ID
		}
		return recs[i].CreatedAt.After(recs[j].CreatedAt)
	})
	out := make([]Post, 0, len(recs))
	for _, p := range recs {
		out = append(out, m.postOut(p, viewer))
	}
	return out, nil
}

func (m *memRepo) PostByID(viewer, id int64) (*Post, error) {
	p, ok := m.posts[id]
	if !ok {
		return nil, ErrNotFound
	}
	out := m.postOut(p, viewer)
	return &out, nil
}

func (m *memRepo) Comments(postID int64) ([]Comment, error) {
	out := []Comment{}
	for _, c := range m.comments {
		if c.postID == postID {
			cm := c.Comment
			if a := m.users[c.authorUID]; a != nil {
				cm.AuthorID, cm.AuthorName = a.Handle, memberName(a.Profile)
			}
			out = append(out, cm)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *memRepo) AddPost(authorUID int64, p Post, imageKey string, now time.Time) (*Post, error) {
	if p.GroupID != nil {
		if _, ok := m.groups[*p.GroupID]; !ok {
			return nil, ErrNotFound
		}
	}
	p.ID = m.nextPost
	m.nextPost++
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	m.posts[p.ID] = &postRec{Post: p, authorUID: authorUID, imageKey: imageKey}
	out := m.postOut(m.posts[p.ID], authorUID)
	return &out, nil
}

func (m *memRepo) deletePost(id int64) {
	delete(m.posts, id)
	delete(m.likes, id)
	delete(m.saves, id)
	kept := m.comments[:0]
	for _, c := range m.comments {
		if c.postID != id {
			kept = append(kept, c)
		}
	}
	m.comments = kept
}

func (m *memRepo) DeletePost(uid, id int64) error {
	p, ok := m.posts[id]
	if !ok {
		return ErrNotFound
	}
	if p.authorUID != uid {
		return ErrForbidden
	}
	m.deletePost(id)
	return nil
}

func (m *memRepo) TogglePostLike(uid, id int64) (bool, int, error) {
	p, ok := m.posts[id]
	if !ok {
		return false, 0, ErrNotFound
	}
	if m.likes[id] == nil {
		m.likes[id] = map[int64]bool{}
	}
	if m.likes[id][uid] {
		delete(m.likes[id], uid)
		p.LikeCount = max(p.LikeCount-1, 0)
		return false, p.LikeCount, nil
	}
	m.likes[id][uid] = true
	p.LikeCount++
	return true, p.LikeCount, nil
}

func (m *memRepo) TogglePostSave(uid, id int64) (bool, error) {
	if _, ok := m.posts[id]; !ok {
		return false, ErrNotFound
	}
	if m.saves[id] == nil {
		m.saves[id] = map[int64]bool{}
	}
	if m.saves[id][uid] {
		delete(m.saves[id], uid)
		return false, nil
	}
	m.saves[id][uid] = true
	return true, nil
}

func (m *memRepo) HidePost(uid, id int64) error {
	if _, ok := m.posts[id]; !ok {
		return ErrNotFound
	}
	if m.hides[uid] == nil {
		m.hides[uid] = map[int64]bool{}
	}
	m.hides[uid][id] = true
	return nil
}

func (m *memRepo) ReportPost(_, id int64, _ string, _ time.Time) error {
	if _, ok := m.posts[id]; !ok {
		return ErrNotFound
	}
	m.reports++
	return nil
}

func (m *memRepo) AddComment(uid, postID int64, text string, now time.Time) (Comment, error) {
	if _, ok := m.posts[postID]; !ok {
		return Comment{}, ErrNotFound
	}
	c := Comment{ID: m.nextCmt, Text: text, CreatedAt: now}
	m.nextCmt++
	if a := m.users[uid]; a != nil {
		c.AuthorID, c.AuthorName = a.Handle, memberName(a.Profile)
	}
	m.comments = append(m.comments, commentRec{Comment: c, postID: postID, authorUID: uid})
	return c, nil
}

func (m *memRepo) Member(viewer int64, handle string) (*MemberProfile, error) {
	id, ok := m.byHandle[handle]
	if !ok {
		return nil, ErrNotFound
	}
	u := m.users[id]
	n := 0
	for _, f := range m.follows {
		if f[id] {
			n++
		}
	}
	return &MemberProfile{ID: u.Handle, Name: memberName(u.Profile), Level: u.Profile.Level, Bio: u.Bio, FollowerCount: n, Following: m.follows[viewer][id]}, nil
}

func (m *memRepo) ToggleFollow(viewer int64, handle string) (bool, error) {
	id, ok := m.byHandle[handle]
	if !ok {
		return false, ErrNotFound
	}
	if m.follows[viewer] == nil {
		m.follows[viewer] = map[int64]bool{}
	}
	if m.follows[viewer][id] {
		delete(m.follows[viewer], id)
		return false, nil
	}
	m.follows[viewer][id] = true
	return true, nil
}
