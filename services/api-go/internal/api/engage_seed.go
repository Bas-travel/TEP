package api

import (
	"time"
)

// seedAll populates a fresh DEV database: the demo traveler, four community members, groups, coupons and sample posts.
// It runs once (see seedDemo) and uses only Tx primitives, so it works on every backend.
func (s *Server) seedAll(tx Tx, demoHash string) error {
	now := s.now()
	for _, c := range []couponDef{
		{Code: "WELCOME10", Title: "ส่วนลด 10%", Detail: "สำหรับการจองครั้งแรก", PercentOff: 10, MinSpend: 1000, MaxDiscount: 500, ExpiresOn: now.AddDate(0, 3, 0).Format("2006-01-02")},
		{Code: "SEA200", Title: "ลด 200 บาท", Detail: "ที่พักและทัวร์ทะเล", MinSpend: 2000, MaxDiscount: 200, ExpiresOn: now.AddDate(0, 1, 0).Format("2006-01-02")},
		{Code: "OLD50", Title: "ส่วนลด 5%", Detail: "คูปองหมดอายุแล้ว", PercentOff: 5, MinSpend: 500, MaxDiscount: 100, ExpiresOn: now.AddDate(0, -1, 0).Format("2006-01-02")},
	} {
		if err := tx.UpsertCoupon(c); err != nil {
			return err
		}
	}

	demo := &userRec{PassHash: demoHash, Profile: UserProfile{Email: "demo@tep.dev", FirstName: "สวัสดี", LastName: "อนันต์", Language: "th",
		BudgetLevel: "mid", TravelStyles: []string{"ทะเล", "คาเฟ่"}, FavoriteDestination: "ภูเก็ต"}}
	if err := tx.CreateUser(demo, now); err != nil {
		return err
	}
	uid := demo.ID
	for _, c := range []string{"WELCOME10", "SEA200", "OLD50"} {
		if err := tx.GrantCoupon(uid, c); err != nil {
			return err
		}
	}
	if _, err := tx.WalletAdjust(uid, 1500, "เติมเงิน PromptPay", now.Add(-48*time.Hour)); err != nil {
		return err
	}
	for _, n := range []Notification{
		{Type: "promo", Title: "คูปองต้อนรับ", Message: "รับส่วนลด 10% สำหรับการจองครั้งแรก", CreatedAt: now.Add(-2 * time.Hour)},
		{Type: "trip", Title: "เตรียมเดินทาง", Message: "ตรวจสอบรายละเอียดทริปของคุณก่อนออกเดินทาง", CreatedAt: now.Add(-26 * time.Hour)},
		{Type: "booking", Title: "การจองสำเร็จ", Message: "การจองของคุณได้รับการยืนยันแล้ว", Read: true, CreatedAt: now.Add(-48 * time.Hour)},
		{Type: "system", Title: "ยินดีต้อนรับ", Message: "ขอบคุณที่ใช้ Sabai jai Travel", Read: true, CreatedAt: now.Add(-72 * time.Hour)},
	} {
		if _, err := tx.AddNotification(uid, n); err != nil {
			return err
		}
	}

	// Community members (synthetic accounts that cannot log in: empty password hash never verifies).
	members := map[string]*userRec{}
	for _, m := range []struct{ handle, first, last, bio string }{
		{"m_nok", "นก", "ใจดี", "ชอบทะเลและพระอาทิตย์ตก"},
		{"m_ton", "ต้น", "สายเที่ยว", "นักเดินป่า ชอบเชียงใหม่"},
		{"m_fah", "ฟ้า", "ใสใจ", "มือใหม่หัดเที่ยว"},
		{"m_pim", "พิม", "กินเก่ง", "ตามล่าอาหารริมทาง"},
	} {
		u := &userRec{Handle: m.handle, Bio: m.bio, Profile: UserProfile{Email: m.handle + "@members.tep.invalid", FirstName: m.first, LastName: m.last, Language: "th"}}
		if err := tx.CreateUser(u, now); err != nil {
			return err
		}
		members[m.handle] = u
	}

	dest := func(s string) *string { return &s }
	groups := []GroupDetail{
		{Group: Group{Name: "สายทะเลอันดามัน", Description: "แชร์ที่เที่ยวและร้านอาหารฝั่งอันดามัน", Destination: dest("ภูเก็ต"), GradientCSS: "linear-gradient(135deg,#0b5ed7,#2bb3c0)"},
			Rules: []string{"สุภาพและให้เกียรติกัน", "ห้ามโฆษณาโดยไม่ได้รับอนุญาต", "แชร์ข้อมูลที่ตรวจสอบแล้ว"}},
		{Group: Group{Name: "เที่ยวเหนือ เชียงใหม่-ปาย", Description: "คาเฟ่ ดอย และหมอกยามเช้า", Destination: dest("เชียงใหม่"), GradientCSS: "linear-gradient(135deg,#1b7f5c,#8bc34a)"},
			Rules: []string{"สุภาพและให้เกียรติกัน", "ไม่โพสต์ข้อมูลส่วนตัวของผู้อื่น"}},
		{Group: Group{Name: "เที่ยวคนเดียว", Description: "เคล็ดลับและความปลอดภัยสำหรับ solo traveler", GradientCSS: "linear-gradient(135deg,#7c3aed,#22c55e)"},
			Rules: []string{"สุภาพและให้เกียรติกัน", "แชร์ประสบการณ์จริง"}},
	}
	for _, g := range groups {
		if err := tx.UpsertGroup(g); err != nil {
			return err
		}
	}
	gs, err := tx.Groups(uid)
	if err != nil {
		return err
	}
	for _, g := range gs[:2] { // demo and two members join the first two groups
		for _, u := range []int64{uid, members["m_nok"].ID, members["m_ton"].ID} {
			if joined, err := tx.ToggleGroup(u, g.ID); err != nil || !joined {
				return err
			}
		}
	}

	seedPosts := []struct {
		author string
		p      Post
		img    bool
	}{
		{"m_nok", Post{Type: "story", Destination: "ภูเก็ต", Text: "พระอาทิตย์ตกที่หาดกะตะสวยมาก แนะนำให้ไปก่อน 17:30", LikeCount: 42}, true},
		{"m_ton", Post{Type: "tip", Destination: "เชียงใหม่", Text: "ขึ้นดอยสุเทพช่วงเช้าหมอกสวย อากาศเย็นสบาย พกเสื้อกันหนาวไปด้วย", LikeCount: 18}, false},
		{"m_fah", Post{Type: "question", Destination: "กระบี่", Text: "ไร่เลย์ไปช่วงเดือนไหนคลื่นสงบที่สุดคะ?", LikeCount: 7}, false},
		{"m_nok", Post{Type: "trip", Destination: "เกาะสมุย", Text: "แผน 3 วัน 2 คืน งบไม่บาน", LikeCount: 55, SharedTrip: &SharedTrip{Days: 3, Budget: 12000, Styles: []string{"ทะเล", "ผ่อนคลาย"}}}, true},
		{"m_pim", Post{Type: "story", Destination: "กรุงเทพฯ", Text: "ปูผัดผงกะหรี่เจ๊ไฝ คุ้มกับการต่อคิวจริง ๆ", LikeCount: 31}, true},
	}
	for i, sp := range seedPosts {
		sp.p.CreatedAt = now.Add(-time.Duration(i+1) * 5 * time.Hour)
		key := ""
		if sp.img {
			key = "post_image/seed.jpg"
		}
		if _, err := tx.AddPost(members[sp.author].ID, sp.p, key, now); err != nil {
			return err
		}
	}
	// The demo user follows nok and pim (feed mode=following).
	for _, h := range []string{"m_nok", "m_pim"} {
		if _, err := tx.ToggleFollow(uid, h); err != nil {
			return err
		}
	}
	return nil
}
