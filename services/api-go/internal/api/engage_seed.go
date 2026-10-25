package api

import "time"

var seedPosts = []Post{
	{ID: 1, AuthorID: "m_nok", Type: "story", Destination: "ภูเก็ต", Text: "พระอาทิตย์ตกที่หาดกะตะสวยมาก แนะนำให้ไปก่อน 17:30", LikeCount: 42, HasImage: true},
	{ID: 2, AuthorID: "m_ton", Type: "tip", Destination: "เชียงใหม่", Text: "ขึ้นดอยสุเทพช่วงเช้าหมอกสวย อากาศเย็นสบาย พกเสื้อกันหนาว", LikeCount: 31},
	{ID: 3, AuthorID: "m_fah", Type: "question", Destination: "กระบี่", Text: "ไร่เลย์ไปช่วงเดือนไหนคลื่นสงบที่สุดคะ?", LikeCount: 7},
	{ID: 4, AuthorID: "m_nok", Type: "trip", Destination: "เกาะสมุย", Text: "แผน 3 วัน 2 คืน งบไม่บาน", LikeCount: 55, SharedTrip: &SharedTrip{Days: 3, Budget: 12000, Styles: []string{"ทะเล", "คาเฟ่"}}},
	{ID: 5, AuthorID: "m_pim", Type: "story", Destination: "กรุงเทพฯ", Text: "ปูผัดผงกะหรี่เจ๊ไฝ คุ้มกับการต่อคิวจริง ๆ", LikeCount: 90, HasImage: true},
}

// followedAuthors are the authors the demo user follows (feed mode=following).
var followedAuthors = map[string]bool{"m_nok": true, "m_pim": true}

func (s *Server) seedEngage(uid int64) {
	now := s.now()
	x := s.extra(uid)
	x.Coupons = []UserCoupon{
		{ID: 1, Code: "WELCOME10", Title: "ส่วนลด 10%", Detail: "สำหรับการจองครั้งแรก", MinSpend: 1000, MaxDiscount: 500, Expiry: now.AddDate(0, 3, 0).Format("2006-01-02"), Status: "available", pct: 0.10},
		{ID: 2, Code: "SEA200", Title: "ลด 200 บาท", Detail: "ที่พักและทัวร์ทะเล", MinSpend: 2000, MaxDiscount: 200, Expiry: now.AddDate(0, 1, 0).Format("2006-01-02"), Status: "available", pct: 1},
		{ID: 3, Code: "OLD50", Title: "ส่วนลด 5%", Detail: "คูปองหมดอายุแล้ว", MinSpend: 500, MaxDiscount: 100, Expiry: now.AddDate(0, -1, 0).Format("2006-01-02"), Status: "expired", pct: 0.05},
	}
	mk := func(typ, title, msg string, read bool, ago time.Duration) {
		x.Notifs = append(x.Notifs, Notification{ID: s.nextID(), Type: typ, Title: title, Message: msg, Read: read, CreatedAt: now.Add(-ago)})
	}
	mk("promo", "คูปองต้อนรับ", "รับส่วนลด 10% สำหรับการจองครั้งแรก", false, 2*time.Hour)
	mk("trip", "เตรียมเดินทาง", "ตรวจสอบรายละเอียดทริปของคุณก่อนออกเดินทาง", false, 26*time.Hour)
	mk("booking", "การจองสำเร็จ", "การจองของคุณได้รับการยืนยันแล้ว", true, 48*time.Hour)
	mk("system", "ยินดีต้อนรับ", "ขอบคุณที่ใช้ TEP Travel Super App", true, 72*time.Hour)
}
