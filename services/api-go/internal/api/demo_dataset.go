package api

import (
	"fmt"
	"math/rand"
	"time"
)

// Demo dataset (DEV/BETA only): 20-30 realistic rows for every module and every menu of the four roles, written
// through the Tx primitives so it lands in MySQL (or the in-memory repo) exactly like production data would.
//
// It is deterministic (fixed RNG seed, times relative to "now") and idempotent: seedDemo runs it once per database,
// detected by the marker member below. Existing databases created before this dataset existed are topped up on the
// next start. Every synthetic account's password is the demo password except community members (cannot log in).

const datasetMarkerEmail = "m_aom@members.tep.invalid"

type demoMember struct{ handle, first, last, bio, dest, budget string }

var demoMembers = []demoMember{
	{"m_aom", "อ้อม", "รักทะเล", "หลงรักทะเลอันดามัน ดำน้ำทุกปี", "ภูเก็ต", "mid"},
	{"m_bank", "แบงค์", "สายปั่น", "ปั่นจักรยานเที่ยวเมืองเก่า", "อยุธยา", "budget"},
	{"m_cream", "ครีม", "คาเฟ่ฮอป", "ตามล่าคาเฟ่ลับทั่วไทย", "เชียงใหม่", "mid"},
	{"m_dream", "ดรีม", "เก็บภาพ", "ช่างภาพสายธรรมชาติ", "กระบี่", "luxury"},
	{"m_earth", "เอิร์ธ", "สายแคมป์", "กางเต็นท์ทุกเดือน", "เขาใหญ่", "budget"},
	{"m_fern", "เฟิร์น", "สปาเดย์", "พักผ่อนสปาและโยคะ", "เกาะสมุย", "luxury"},
	{"m_gift", "กิ๊ฟ", "เที่ยวเรื่อย", "เที่ยวคนเดียวบ่อย ๆ", "ปาย", "budget"},
	{"m_hana", "ฮานะ", "ใจกล้า", "ชอบกิจกรรมผจญภัย", "กระบี่", "mid"},
	{"m_ice", "ไอซ์", "กินจุ", "ตะลุยร้านอร่อยทั่วกรุงเทพฯ", "กรุงเทพฯ", "mid"},
	{"m_jane", "เจน", "ครอบครัวสุข", "เที่ยวกับลูก ๆ ทุกปิดเทอม", "ภูเก็ต", "mid"},
	{"m_kong", "คอง", "สายลุย", "เดินป่าและขึ้นดอย", "เชียงใหม่", "budget"},
	{"m_lily", "ลิลลี่", "ฮันนีมูน", "วางแผนทริปคู่รัก", "เกาะสมุย", "luxury"},
	{"m_mark", "มาร์ค", "นักธุรกิจ", "เดินทางเพื่อธุรกิจบ่อย", "กรุงเทพฯ", "luxury"},
	{"m_namtan", "น้ำตาล", "หวานใจ", "ชอบขนมและคาเฟ่", "เชียงใหม่", "mid"},
	{"m_oat", "โอ๊ต", "นักเดินทาง", "สะสมแสตมป์พาสปอร์ต", "ภูเก็ต", "mid"},
	{"m_ploy", "พลอย", "เที่ยวสนุก", "เที่ยวกับแก๊งเพื่อนสาว", "กระบี่", "mid"},
	{"m_quan", "ควน", "ใต้ดี๊ดี", "คนใต้แท้ แนะนำของกินถูกและดี", "กระบี่", "budget"},
	{"m_rin", "ริน", "ใจเย็น", "ชอบท่องเที่ยวเชิงวัฒนธรรม", "อยุธยา", "mid"},
	{"m_sky", "สกาย", "ฟ้าใส", "หลงใหลการดูดาวบนยอดดอย", "ปาย", "budget"},
	{"m_tan", "แทน", "สายเกาะ", "กระโดดเกาะทุกสุดสัปดาห์", "เกาะสมุย", "mid"},
	{"m_ung", "อุ้ม", "เที่ยวเงียบ", "ชอบที่พักสงบ ห่างผู้คน", "เขาใหญ่", "luxury"},
	{"m_view", "วิว", "ทิวทัศน์", "วิวสวยต้องมา", "เชียงใหม่", "mid"},
	{"m_wan", "หวาน", "เจ้าสำราญ", "สายช้อปปิ้งและสตรีทฟู้ด", "กรุงเทพฯ", "mid"},
	{"m_yok", "หยก", "ใจดี", "นักท่องเที่ยวมือใหม่", "ภูเก็ต", "budget"},
}

// seedDataset fills every module. Idempotent via the marker member; safe to call on a database that already holds the
// base demo data (users are looked up by e-mail).
func (s *Server) seedDataset(tx Tx, hash string) error {
	if _, err := tx.UserByEmail(datasetMarkerEmail); err == nil {
		return nil
	}
	d := &demoDS{s: s, tx: tx, hash: hash, now: s.now(), r: rand.New(rand.NewSource(20261121)),
		byHandle: map[string]*userRec{}, listingOwner: map[int64]int64{}}
	steps := []struct {
		name string
		run  func() error
	}{{"accounts", d.accounts}, {"partners", d.partnersAndListings}, {"coupons", d.coupons}, {"bookings", d.makeBookings},
		{"checks", d.checks}, {"reviews", d.reviews}, {"trips", d.trips}, {"notifications", d.notifications},
		{"tickets", d.tickets}, {"community", d.community}, {"audits", d.audits}}
	for _, st := range steps {
		if err := st.run(); err != nil {
			return fmt.Errorf("demo dataset (%s): %w", st.name, err)
		}
	}
	return nil
}

type demoDS struct {
	s     *Server
	tx    Tx
	hash  string
	now   time.Time
	r     *rand.Rand
	demo  *userRec
	admin *userRec
	staff []*userRec
	all   []*userRec // every traveler incl. demo, oldest members first
	// partner owners and records, indexed by order of creation
	owners   []*userRec
	partners []*partnerRec
	byHandle map[string]*userRec
	// bookings kept for the checks / reviews steps
	books        []*bookingRec
	listingOwner map[int64]int64
}

func (d *demoDS) ago(days int, hours int) time.Time {
	return d.now.Add(-time.Duration(days)*24*time.Hour - time.Duration(hours)*time.Hour)
}

func (d *demoDS) pick(n int) int { return d.r.Intn(n) }

// ---- accounts: 24 more travelers, 2 more staff, 5 partner owners, 1 moderator -------------------------------------

func (d *demoDS) accounts() error {
	var err error
	if d.demo, err = d.tx.UserByEmail("demo@tep.dev"); err != nil {
		return err
	}
	if d.admin, err = d.tx.UserByEmail("admin@tep.dev"); err != nil {
		return err
	}
	st1, err := d.tx.UserByEmail("staff@tep.dev")
	if err != nil {
		return err
	}
	d.staff = []*userRec{st1}
	d.all = []*userRec{d.demo}
	for _, h := range []string{"m_nok", "m_ton", "m_fah", "m_pim"} {
		u, err := d.tx.UserByEmail(h + "@members.tep.invalid")
		if err != nil {
			return err
		}
		d.byHandle[h] = u
		d.all = append(d.all, u)
	}
	for i, m := range demoMembers {
		u := &userRec{Handle: m.handle, Bio: m.bio, Profile: UserProfile{Email: m.handle + "@members.tep.invalid", FirstName: m.first, LastName: m.last,
			Language: "th", BudgetLevel: m.budget, FavoriteDestination: m.dest, TravelStyles: []string{[]string{"ทะเล", "ธรรมชาติ", "คาเฟ่", "วัฒนธรรม", "อาหาร"}[i%5]}}}
		if err := d.tx.CreateUser(u, d.ago(90-i*3, i)); err != nil {
			return err
		}
		d.byHandle[m.handle] = u
		d.all = append(d.all, u)
	}
	mk := func(email, first, last, role string, createdDaysAgo int) (*userRec, error) {
		u := &userRec{PassHash: d.hash, Role: role, Profile: UserProfile{Email: email, FirstName: first, LastName: last, Language: "th"}}
		return u, d.tx.CreateUser(u, d.ago(createdDaysAgo, 0))
	}
	for i, n := range []struct{ first, last string }{{"นภา", "ใจเย็น"}, {"ธนา", "ตอบไว"}} {
		u, err := mk(fmt.Sprintf("staff%d@tep.dev", i+2), n.first, n.last, RoleStaff, 60-i*5)
		if err != nil {
			return err
		}
		d.staff = append(d.staff, u)
	}
	for i, n := range []struct{ first, last string }{{"ศิริพร", "ล้านนา"}, {"ปกรณ์", "บลูกระบี่"}, {"มณีรัตน์", "เมืองกรุง"}, {"ชัยวัฒน์", "เกาะสมุย"}, {"ปริญญา", "ปายวัลเลย์"}} {
		u, err := mk(fmt.Sprintf("partner%d@tep.dev", i+2), n.first, n.last, RolePartner, 80-i*6)
		if err != nil {
			return err
		}
		d.owners = append(d.owners, u)
	}
	if _, err := mk("moderator@tep.dev", "นภัส", "ผู้ดูแลชุมชน", RoleModerator, 70); err != nil {
		return err
	}
	// One suspended traveler for the admin "users" demo.
	yok := d.byHandle["m_yok"]
	yok.Status = "suspended"
	return d.tx.SaveUser(yok)
}

// ---- partners & catalog ownership ---------------------------------------------------------------------------------

func (d *demoDS) partnersAndListings() error {
	p1, err := d.tx.PartnerByUser(mustUID(d.tx, "partner@tep.dev"), false)
	if err != nil {
		return err
	}
	d.partners = []*partnerRec{p1}
	decided := d.ago(20, 0)
	type def struct {
		owner         *userRec
		name, phone   string
		desc          string
		status        string
		rate          float64
		note          string
		created       time.Time
		listings      []int64
		listingStatus map[int64]string
	}
	defs := []def{
		{d.owners[0], "Lanna Heritage Tours", "053-100-200", "ทัวร์วัฒนธรรมและที่พักสไตล์ล้านนา เชียงใหม่", PartnerApproved, 10, "", d.ago(75, 0), []int64{5, 6, 7, 16, 17, 18}, nil},
		{d.owners[1], "Krabi Blue Adventures", "075-300-400", "ทัวร์เกาะและกิจกรรมทางน้ำกระบี่", PartnerApproved, 12, "", d.ago(60, 0), []int64{8, 19, 20}, map[int64]string{20: "archived"}},
		{d.owners[2], "Bangkok Urban Hospitality", "02-555-0100", "ที่พักและประสบการณ์อาหารในกรุงเทพฯ", PartnerApproved, 15, "", d.ago(50, 0), []int64{10, 12, 21, 22}, nil},
		{d.owners[3], "Samui Island Co.", "077-111-222", "รีสอร์ตและทัวร์เกาะสมุย", PartnerApproved, 12, "", d.ago(40, 0), []int64{11, 25, 26}, nil},
		{d.owners[4], "Pai Valley Stays", "053-699-888", "บังกะโลริมน้ำปาย", PartnerSuspended, 12, "พบข้อร้องเรียนเรื่องการยืนยันห้องพัก อยู่ระหว่างตรวจสอบ", d.ago(35, 0), []int64{27, 28}, nil},
		{d.byHandle["m_bank"], "Ayutthaya Bike & Heritage", "035-200-300", "ทัวร์จักรยานเมืองมรดกโลก", PartnerPending, 12, "", d.ago(1, 3), nil, nil},
		{d.byHandle["m_quan"], "Quan Southern Kitchen", "081-222-3333", "ทัวร์อาหารถิ่นใต้", PartnerRejected, 12, "เอกสารธุรกิจไม่ครบ กรุณายื่นใหม่พร้อมใบทะเบียนพาณิชย์", d.ago(9, 0), nil, nil},
	}
	for _, df := range defs {
		p := &partnerRec{UserID: df.owner.ID, BusinessName: df.name, ContactPhone: df.phone, Description: df.desc, Status: df.status,
			CommissionRate: df.rate, DecisionNote: df.note, CreatedAt: df.created}
		if df.status != PartnerPending {
			t := decided
			p.DecidedBy, p.DecidedAt = d.admin.ID, &t
		}
		if err := d.tx.CreatePartner(p); err != nil {
			return err
		}
		d.partners = append(d.partners, p)
		for _, id := range df.listings {
			if err := d.assign(id, p.ID, df.listingStatus[id]); err != nil {
				return err
			}
		}
	}
	// The first partner also owns three of the extension listings; one of them is still a draft.
	for _, id := range []int64{13, 14, 15} {
		st := ""
		if id == 15 {
			st = "draft"
		}
		if err := d.assign(id, p1.ID, st); err != nil {
			return err
		}
	}
	// Approved partners get their role promoted at approval time (see handlers_admin); owners were created with the role.
	return nil
}

func mustUID(tx Tx, email string) int64 {
	u, err := tx.UserByEmail(email)
	if err != nil {
		return 0
	}
	return u.ID
}

func (d *demoDS) assign(listingID, partnerID int64, status string) error {
	l, err := d.tx.ListingByID(listingID, false)
	if err != nil {
		return err
	}
	l.PartnerID = partnerID
	if status != "" {
		l.Status = status
	}
	d.listingOwner[listingID] = partnerID
	return d.tx.SaveListing(l)
}

// ---- coupons --------------------------------------------------------------------------------------------------------

func (d *demoDS) coupons() error {
	exp := func(months int) string { return d.now.AddDate(0, months, 0).Format("2006-01-02") }
	defs := []couponDef{
		{Code: "SONGKRAN15", Title: "สงกรานต์ ลด 15%", Detail: "ที่พักและทัวร์ทุกประเภท", PercentOff: 15, MinSpend: 2000, MaxDiscount: 800, ExpiresOn: exp(5)},
		{Code: "FLY300", Title: "ลด 300 บาท ตั๋วเครื่องบิน", Detail: "เฉพาะเที่ยวบินภายในประเทศ", MinSpend: 1500, MaxDiscount: 300, ExpiresOn: exp(2)},
		{Code: "HOTEL12", Title: "โรงแรม ลด 12%", Detail: "ที่พักทั่วประเทศ ขั้นต่ำ 3,000 บาท", PercentOff: 12, MinSpend: 3000, MaxDiscount: 1000, ExpiresOn: exp(4)},
		{Code: "FOODIE100", Title: "ร้านอาหาร ลด 100 บาท", Detail: "ร้านอาหารพาร์ทเนอร์ทุกแห่ง", MinSpend: 600, MaxDiscount: 100, ExpiresOn: exp(1)},
		{Code: "WEEKEND8", Title: "เที่ยววันธรรมดา ลด 8%", Detail: "เข้าพักอาทิตย์–พฤหัสบดี", PercentOff: 8, MinSpend: 1500, MaxDiscount: 400, ExpiresOn: exp(6)},
		{Code: "TRIP500", Title: "แพ็กเกจทริป ลด 500 บาท", Detail: "แพ็กเกจเที่ยวมากกว่า 2 วัน", MinSpend: 5000, MaxDiscount: 500, ExpiresOn: exp(3)},
		{Code: "NEWYEAR20", Title: "ปีใหม่ ลด 20%", Detail: "จองล่วงหน้า 30 วัน", PercentOff: 20, MinSpend: 4000, MaxDiscount: 1500, ExpiresOn: exp(2)},
		{Code: "STUDENT7", Title: "นักศึกษา ลด 7%", Detail: "แสดงบัตรนักศึกษาเมื่อเช็กอิน", PercentOff: 7, MinSpend: 800, MaxDiscount: 300, ExpiresOn: exp(8)},
		{Code: "BEACH9", Title: "ทริปทะเล ลด 9%", Detail: "ที่พักริมทะเลทุกแห่ง", PercentOff: 9, MinSpend: 2500, MaxDiscount: 700, ExpiresOn: exp(3)},
		{Code: "NORTH150", Title: "เที่ยวเหนือ ลด 150 บาท", Detail: "กิจกรรมและที่พักภาคเหนือ", MinSpend: 1200, MaxDiscount: 150, ExpiresOn: exp(4)},
		{Code: "COUPLE10", Title: "คู่รัก ลด 10%", Detail: "แพ็กเกจสำหรับ 2 ท่าน", PercentOff: 10, MinSpend: 3500, MaxDiscount: 900, ExpiresOn: exp(5)},
		{Code: "BKK200", Title: "กรุงเทพฯ ลด 200 บาท", Detail: "โรงแรมและทัวร์ในกรุงเทพฯ", MinSpend: 1800, MaxDiscount: 200, ExpiresOn: exp(2)},
		{Code: "ISLAND11", Title: "ทัวร์เกาะ ลด 11%", Detail: "ทัวร์เกาะพีพี สมุย ราชา", PercentOff: 11, MinSpend: 1500, MaxDiscount: 450, ExpiresOn: exp(3)},
		{Code: "FAMILY6", Title: "ครอบครัว ลด 6%", Detail: "จองตั้งแต่ 3 ท่านขึ้นไป", PercentOff: 6, MinSpend: 4000, MaxDiscount: 600, ExpiresOn: exp(7)},
		{Code: "LASTMIN", Title: "จองด่วน ลด 300 บาท", Detail: "จองก่อนเดินทางไม่เกิน 3 วัน", MinSpend: 2200, MaxDiscount: 300, ExpiresOn: exp(1)},
		{Code: "MEMBER5", Title: "สมาชิก ลด 5%", Detail: "สำหรับสมาชิกทุกระดับ", PercentOff: 5, MinSpend: 500, MaxDiscount: 200, ExpiresOn: exp(10)},
		{Code: "EXPIRED25", Title: "โปรเก่า ลด 25%", Detail: "หมดอายุแล้ว", PercentOff: 25, MinSpend: 1000, MaxDiscount: 600, ExpiresOn: d.now.AddDate(0, -2, 0).Format("2006-01-02")},
	}
	for _, c := range defs {
		if err := d.tx.UpsertCoupon(c); err != nil {
			return err
		}
	}
	// The demo traveler holds most of them, other travelers a couple each.
	for _, c := range defs[:16] {
		if err := d.tx.GrantCoupon(d.demo.ID, c.Code); err != nil {
			return err
		}
	}
	for i, u := range d.all[1:] {
		for k := 0; k < 2; k++ {
			if err := d.tx.GrantCoupon(u.ID, defs[(i+k)%len(defs)].Code); err != nil && err != ErrDuplicate {
				return err
			}
		}
	}
	return nil
}

// ---- bookings, payments, wallet -----------------------------------------------------------------------------------

type bookSpec struct {
	status string
	days   int // booked this many days ago
	method string
}

func (d *demoDS) makeBookings() error {
	// 30 bookings: 12 completed, 10 confirmed, 4 cancelled (refunded to wallet), 4 awaiting payment.
	var specs []bookSpec
	for i := 0; i < 12; i++ {
		specs = append(specs, bookSpec{"completed", 70 - i*5, []string{"PromptPay", "บัตรเครดิต/เดบิต", "Wallet"}[i%3]})
	}
	for i := 0; i < 10; i++ {
		specs = append(specs, bookSpec{"confirmed", 1 + i*2, []string{"PromptPay", "Wallet", "บัตรเครดิต/เดบิต"}[i%3]})
	}
	for i := 0; i < 4; i++ {
		specs = append(specs, bookSpec{"cancelled", 14 + i*9, []string{"PromptPay", "Wallet"}[i%2]})
	}
	for i := 0; i < 4; i++ {
		specs = append(specs, bookSpec{"pending_payment", 0, ""})
	}
	// Listings a traveler can actually book (published, partner not suspended/rejected).
	bookable := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 16, 17, 18, 19, 21, 22, 23, 24, 25, 26, 29, 30}
	reasons := []string{"เปลี่ยนแผนการเดินทาง", "ติดภารกิจเร่งด่วน", "ผู้ร่วมทางไม่สะดวก", "พบข้อเสนอที่เหมาะสมกว่า"}
	for i, sp := range specs {
		owner := d.all[0]
		if i%2 != 0 {
			owner = d.all[1+i%(len(d.all)-1)]
		}
		lid := bookable[(i*7+3)%len(bookable)]
		l, err := d.tx.ListingByID(lid, false)
		if err != nil {
			return err
		}
		qty := 1 + i%3
		if l.Type == "restaurant" {
			qty = 1
		}
		unit := l.PriceFrom
		sub := round2(unit * float64(qty))
		fee := round2(sub * 0.03)
		at := d.ago(sp.days, i%7)
		start := at.AddDate(0, 0, 5+i%9)
		if sp.status == "completed" {
			start = at.AddDate(0, 0, 3)
		}
		if sp.status == "confirmed" {
			start = d.now.AddDate(0, 0, 4+i*2)
		}
		b := &bookingRec{UserID: owner.ID, Lines: []bookingLine{{ListingID: l.ID, Description: l.Name, Quantity: qty, UnitPrice: unit}},
			BookingDetail: BookingDetail{BookingSummary: BookingSummary{Type: l.Type, Title: l.Name, Subtitle: l.Location, Amount: round2(sub + fee), Date: at,
				Status: sp.status, StartDate: start.Format("2006-01-02"), EndDate: start.AddDate(0, 0, 1+i%3).Format("2006-01-02"), Guests: 1 + i%4},
				Subtotal: sub, ServiceFee: fee}}
		if l.Type == "restaurant" || l.Type == "attraction" || l.Type == "flight" {
			b.EndDate = ""
		}
		b.Method = sp.method
		if sp.status == "cancelled" {
			b.CancelReason = reasons[i%len(reasons)]
			ref := b.Amount
			b.RefundAmount = &ref
		}
		if err := d.tx.CreateBooking(b); err != nil {
			return err
		}
		if sp.status != "pending_payment" {
			if err := d.tx.SaveBooking(b); err != nil {
				return err
			}
			code := map[string]string{"PromptPay": "promptpay", "Wallet": "wallet"}[sp.method]
			if code == "" {
				code = "card"
			}
			if code == "wallet" {
				if _, err := d.tx.WalletAdjust(owner.ID, b.Amount, "เติมเงิน PromptPay", at.Add(-time.Hour)); err != nil {
					return err
				}
				if _, err := d.tx.WalletAdjust(owner.ID, -b.Amount, "ชำระการจอง "+b.BookingNumber, at); err != nil {
					return err
				}
			}
			if err := d.tx.RecordPayment(b.ID, "", code, b.Amount, "succeeded", at); err != nil {
				return err
			}
			if sp.status == "cancelled" {
				if err := d.tx.MarkPaymentRefunded(b.ID); err != nil {
					return err
				}
				if _, err := d.tx.WalletAdjust(owner.ID, b.Amount, "คืนเงินการจอง "+b.BookingNumber, at.Add(26*time.Hour)); err != nil {
					return err
				}
			}
		}
		d.books = append(d.books, b)
	}
	// Wallet history for the demo traveler: top-ups and spends over the last two months (>= 20 rows with the above).
	for i, t := range []struct {
		delta float64
		label string
	}{{2000, "เติมเงิน PromptPay"}, {-350, "ชำระค่าอาหาร"}, {1000, "เติมเงิน บัตรเครดิต/เดบิต"}, {-780, "ชำระทัวร์"}, {500, "โบนัสแนะนำเพื่อน"},
		{-1200, "ชำระที่พัก"}, {3000, "เติมเงิน PromptPay"}, {-640, "ชำระค่าอาหาร"}, {800, "คืนเงินการจอง"}, {-2100, "ชำระที่พัก"},
		{250, "Cashback จากการจอง"}, {-180, "ชำระค่าเดินทาง"}, {1500, "เติมเงิน PromptPay"}, {-420, "ชำระทัวร์"}, {300, "โบนัสรีวิว"}} {
		if _, err := d.tx.WalletAdjust(d.demo.ID, t.delta, t.label, d.ago(60-i*4, 0)); err != nil {
			// A spend larger than the balance at that moment is skipped rather than failing the whole seed.
			if err == ErrInsufficient {
				continue
			}
			return err
		}
	}
	return nil
}

// ---- staff booking checks ------------------------------------------------------------------------------------------

func (d *demoDS) checks() error {
	n := 0
	for _, b := range d.books {
		if b.Status != "confirmed" && b.Status != "completed" {
			continue
		}
		if n >= 22 {
			break
		}
		res, note := "verified", []string{"ตรวจสอบการชำระเงินแล้ว", "ข้อมูลผู้เดินทางครบถ้วน", "ยืนยันกับพาร์ทเนอร์แล้ว", "เอกสารถูกต้อง"}[n%4]
		if n%5 == 4 {
			res, note = "flagged", []string{"ชื่อผู้เดินทางไม่ตรงกับบัตร รอติดต่อลูกค้า", "ยอดชำระผิดปกติ ตรวจสอบกับธนาคาร", "จองซ้ำภายในวันเดียวกัน ตรวจสอบก่อนยืนยัน"}[n%3]
		}
		if err := d.tx.AddBookingCheck(&bookingCheck{BookingID: b.ID, StaffID: d.staff[n%len(d.staff)].ID, Result: res, Note: note, CreatedAt: b.Date.Add(2 * time.Hour)}); err != nil {
			return err
		}
		n++
	}
	return nil
}

// ---- reviews: 30 across listings ---------------------------------------------------------------------------------

func (d *demoDS) reviews() error {
	comments := []string{
		"บริการดีมาก พนักงานน่ารัก สะอาด คุ้มราคา", "ทำเลดี เดินทางสะดวก แต่ช่วงเย็นคนเยอะนิดหน่อย", "ประทับใจมาก จะกลับมาอีกแน่นอน",
		"อาหารอร่อย รสชาติจัดจ้านถูกปาก", "ไกด์ดูแลดี ให้ข้อมูลเยอะ สนุกตลอดทริป", "ห้องพักสะอาด วิวสวย อาหารเช้าหลากหลาย",
		"ราคาสมเหตุสมผล แต่รอคิวนานไปหน่อย", "เหมาะกับครอบครัว เด็ก ๆ ชอบมาก", "ตรงเวลา จัดการเรียบร้อย ไม่มีปัญหา",
		"บรรยากาศดี เหมาะถ่ายรูป", "พอใช้ได้ ไม่ได้ตื่นเต้นมาก แต่ก็ไม่มีอะไรต้องติ", "แอร์ในห้องเสียงดังเล็กน้อย แต่ทีมงานรีบแก้ให้ทันที",
	}
	seen := map[[2]int64]bool{}
	add := func(uid, lid int64, i int) error {
		k := [2]int64{uid, lid}
		if seen[k] {
			return nil
		}
		seen[k] = true
		rating := []int{5, 5, 4, 5, 4, 3, 5, 4, 4, 2, 5, 4}[i%12]
		_, err := d.tx.AddReview(uid, lid, rating, comments[i%len(comments)], d.ago(2+i*2, i%5))
		return err
	}
	n := 0
	for _, b := range d.books { // reviews come first from travelers who completed a stay
		if b.Status == "completed" && len(b.Lines) > 0 {
			if err := add(b.UserID, b.Lines[0].ListingID, n); err != nil {
				return err
			}
			n++
		}
	}
	bookable := []int64{1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 16, 17, 18, 19, 21, 22, 25, 26, 29, 30}
	for i := 0; i < len(bookable) && demoReviews(seen, d.demo.ID) < 12; i++ { // the demo account's "my reviews" menu
		if err := add(d.demo.ID, bookable[(i*5+2)%len(bookable)], n+i); err != nil {
			return err
		}
	}
	for i := 0; len(seen) < 30 && i < 400; i++ {
		u := d.all[(i*5+2)%len(d.all)]
		if err := add(u.ID, bookable[(i*3+1)%len(bookable)], n+i); err != nil {
			return err
		}
	}
	return nil
}

// ---- trips (22) and AI planner jobs (6) --------------------------------------------------------------------------

func (d *demoDS) trips() error {
	dests := []string{"ภูเก็ต", "เชียงใหม่", "กระบี่", "กรุงเทพฯ", "เกาะสมุย", "ปาย", "อยุธยา", "เขาใหญ่"}
	for i := 0; i < 26; i++ {
		owner := d.demo
		if i >= 12 {
			owner = d.all[1+i%(len(d.all)-1)]
		}
		dest := dests[i%len(dests)]
		days, trav := 2+i%4, 1+i%4
		budget := float64(6000 + (i%7)*2500)
		src := []string{"ai", "manual", "community"}[i%3]
		if _, err := d.tx.AddTrip(owner.ID, SavedTripInput{Destination: dest, Days: days, Travelers: trav, Budget: budget,
			Itinerary: buildItinerary(dest, days, budget, trav), Source: src}, d.ago(i*3, i)); err != nil {
			return err
		}
	}
	for i := 0; i < 6; i++ {
		dest := dests[(i*3)%len(dests)]
		days, budget := 3+i%3, float64(9000+i*1500)
		job := ItineraryJob{JobID: fmt.Sprintf("job_demo_%02d", i+1), Status: "completed", Degraded: i%2 == 0, Result: buildItinerary(dest, days, budget, 2)}
		in := map[string]any{"destination": dest, "days": days, "budget": budget, "travelers": 2}
		if err := d.tx.AddJob(d.demo.ID, jobRec{ItineraryJob: job, Input: in}, d.ago(i*4, 0)); err != nil {
			return err
		}
	}
	return nil
}

// ---- notifications (24 for the demo traveler) --------------------------------------------------------------------

func (d *demoDS) notifications() error {
	items := []struct{ typ, title, msg string }{
		{"booking", "ชำระเงินสำเร็จ", "เราได้รับการชำระเงินของคุณแล้ว"}, {"booking", "การจองได้รับการยืนยัน", "ผู้ให้บริการยืนยันการจองของคุณแล้ว"},
		{"trip", "อีก 3 วันถึงวันเดินทาง", "เตรียมเอกสารและตรวจสอบเวลานัดหมาย"}, {"promo", "คูปองใหม่สำหรับคุณ", "รับส่วนลด 15% สำหรับทริปสงกรานต์"},
		{"system", "อัปเดตแอปเวอร์ชันใหม่", "ปรับปรุงการค้นหาและการชำระเงิน"}, {"booking", "การจองถูกยกเลิก", "เงินคืนเข้า Wallet เรียบร้อยแล้ว"},
		{"promo", "โปรโมชันสุดสัปดาห์", "โรงแรมเลือกพิเศษ ลดสูงสุด 12%"}, {"trip", "ทริปของคุณพร้อมแล้ว", "AI วางแผนทริปให้เรียบร้อยแล้ว ดูรายละเอียดได้เลย"},
		{"booking", "ถึงเวลาเช็กอิน", "แสดงเลขที่การจองเมื่อถึงที่พัก"}, {"system", "ความปลอดภัยบัญชี", "เราตรวจพบการเข้าสู่ระบบจากอุปกรณ์ใหม่"},
		{"promo", "แต้มสะสมใกล้หมดอายุ", "ใช้แต้มก่อนสิ้นเดือนนี้"}, {"trip", "สภาพอากาศเปลี่ยนแปลง", "ภูเก็ตมีฝนตกช่วงบ่าย พกร่มไปด้วย"},
		{"booking", "รีวิวการเดินทางของคุณ", "เล่าประสบการณ์ให้เพื่อน ๆ ทราบ"}, {"promo", "ชวนเพื่อนรับ 500 บาท", "แชร์โค้ดให้เพื่อนและรับโบนัสเข้า Wallet"},
		{"system", "ขอบคุณที่ใช้บริการ", "ทีมงานพร้อมช่วยเหลือตลอด 24 ชั่วโมง"}, {"booking", "ใบเสร็จพร้อมแล้ว", "ดาวน์โหลดใบเสร็จได้ที่รายละเอียดการจอง"},
		{"trip", "เพิ่มกิจกรรมในแผน", "แนะนำร้านอาหารใกล้ที่พักของคุณ"}, {"promo", "FLY300 ใกล้หมดอายุ", "ลด 300 บาทสำหรับเที่ยวบินในประเทศ"},
		{"system", "นโยบายความเป็นส่วนตัวอัปเดต", "โปรดตรวจสอบรายละเอียดที่เมนูตั้งค่า"}, {"booking", "คำขอพิเศษได้รับแล้ว", "ที่พักรับทราบคำขอห้องเตียงเดี่ยว"},
		{"trip", "เพื่อนร่วมทริปตอบรับแล้ว", "สมาชิกในแผนทริปเพิ่มขึ้น 1 คน"}, {"promo", "ส่วนลดวันเกิด", "รับคูปองพิเศษในเดือนเกิดของคุณ"},
		{"booking", "เปลี่ยนวันเดินทางสำเร็จ", "วันเดินทางใหม่ได้รับการยืนยัน"}, {"system", "บำรุงรักษาระบบ", "ระบบจะปิดปรับปรุงชั่วคราวเวลา 02:00–03:00 น."},
	}
	for i, it := range items {
		if _, err := d.tx.AddNotification(d.demo.ID, Notification{Type: it.typ, Title: it.title, Message: it.msg, Read: i >= 8, CreatedAt: d.ago(0, 3+i*7)}); err != nil {
			return err
		}
	}
	// A few for other travelers so the admin/staff views see varied accounts.
	for i, u := range d.all[1:7] {
		if _, err := d.tx.AddNotification(u.ID, Notification{Type: "booking", Title: "การจองได้รับการยืนยัน", Message: "ผู้ให้บริการยืนยันการจองของคุณแล้ว", CreatedAt: d.ago(i, 2)}); err != nil {
			return err
		}
	}
	return nil
}

// ---- support tickets (26) ------------------------------------------------------------------------------------------

func (d *demoDS) tickets() error {
	subjects := []struct{ subject, msg string }{
		{"ขอเปลี่ยนวันเข้าพัก", "ต้องการเลื่อนวันเข้าพักออกไป 2 วัน ทำได้หรือไม่คะ"}, {"ยังไม่ได้รับเงินคืน", "ยกเลิกการจองไปแล้ว 2 วัน เงินยังไม่เข้า Wallet"},
		{"ชำระเงินแล้วแต่สถานะยังรอชำระ", "สแกน PromptPay แล้วแต่ระบบยังไม่ยืนยัน"}, {"ขอใบกำกับภาษี", "ขอใบกำกับภาษีเต็มรูปแบบสำหรับการจองเดือนที่แล้ว"},
		{"คูปองใช้ไม่ได้", "ใส่โค้ด SEA200 แล้วระบบแจ้งไม่เข้าเงื่อนไข"}, {"สอบถามนโยบายเด็ก", "เด็กอายุ 5 ขวบ ต้องจ่ายค่าทัวร์ไหมคะ"},
		{"แก้ไขชื่อผู้เดินทาง", "พิมพ์ชื่อผิด ต้องการแก้ให้ตรงกับพาสปอร์ต"}, {"ที่พักไม่ตรงตามที่โฆษณา", "ห้องไม่มีสระว่ายน้ำตามที่ระบุ"},
		{"รีวิวของฉันหายไป", "เขียนรีวิวไว้แต่ไม่แสดงในหน้ารายการ"}, {"สมัครพาร์ทเนอร์ใช้เวลานานไหม", "ส่งใบสมัครไปแล้ว 3 วัน ยังไม่ได้รับการตอบกลับ"},
		{"ลืมรหัสผ่าน", "รีเซ็ตรหัสผ่านแล้วแต่ไม่ได้รับอีเมล"}, {"ตั๋วเครื่องบินเปลี่ยนเวลา", "สายการบินเลื่อนเที่ยวบิน ต้องการเปลี่ยนที่พักตาม"},
		{"ขอข้อมูลการเดินทางจากสนามบิน", "มีรถรับส่งจากสนามบินกระบี่หรือไม่"}, {"บัตรเครดิตถูกตัดซ้ำ", "เห็นรายการตัดบัตร 2 ครั้งในการจองเดียว"},
		{"ปุ่มคอมเมนต์ใช้ไม่ได้", "กดส่งคอมเมนต์แล้วไม่มีอะไรเกิดขึ้น"}, {"ขอลบบัญชี", "ต้องการลบบัญชีและข้อมูลส่วนตัวทั้งหมด"},
		{"สอบถามแพ็กเกจเหมาจ่าย", "มีแพ็กเกจสำหรับกรุ๊ป 10 คนหรือไม่"}, {"ปัญหาการสแกน QR", "QR PromptPay หมดอายุเร็วเกินไป"},
		{"ขอหลักฐานการจองไปยื่นวีซ่า", "ต้องการเอกสารยืนยันการจองภาษาอังกฤษ"}, {"ผู้ให้บริการยกเลิกกะทันหัน", "โรงแรมแจ้งยกเลิกก่อนเดินทาง 1 วัน"},
		{"คะแนนสะสมไม่เข้า", "จองเสร็จสิ้นแล้วแต่ไม่ได้รับแต้ม"}, {"แจ้งโพสต์ไม่เหมาะสม", "พบโพสต์โฆษณาในกลุ่มชุมชน"},
		{"สอบถามประกันการเดินทาง", "มีประกันการเดินทางให้เลือกไหม"}, {"แอปค้างหน้าชำระเงิน", "กดชำระแล้วแอปค้าง ต้องปิดเปิดใหม่"},
		{"ขอเปลี่ยนอีเมลบัญชี", "ต้องการเปลี่ยนอีเมลที่ใช้ล็อกอิน"}, {"ชมเชยพนักงาน", "ขอบคุณทีมงานที่ช่วยแก้ปัญหาเรื่องห้องพักอย่างรวดเร็ว"},
	}
	replies := []string{"ทีมงานได้รับเรื่องแล้ว กำลังตรวจสอบให้ค่ะ", "ตรวจสอบเรียบร้อย ดำเนินการแก้ไขให้แล้วค่ะ", "รบกวนแจ้งเลขที่การจองเพิ่มเติมเพื่อตรวจสอบค่ะ",
		"ประสานกับผู้ให้บริการแล้ว คาดว่าจะได้รับคำตอบภายใน 24 ชั่วโมงค่ะ"}
	for i, sj := range subjects {
		owner := d.all[1+i%(len(d.all)-1)]
		if i%3 == 0 {
			owner = d.demo
		}
		at := d.ago(i/2, i%24+1)
		tk, err := d.tx.AddTicket(owner.ID, sj.subject, sj.msg, at)
		if err != nil {
			return err
		}
		st := d.staff[i%len(d.staff)]
		switch i % 4 {
		case 0: // open, no reply yet
		case 1: // answered
			if _, err := d.tx.AddTicketReply(tk.ID, st.ID, RoleStaff, replies[i%len(replies)], at.Add(time.Hour)); err != nil {
				return err
			}
			if err := d.tx.SetTicketStatus(tk.ID, "answered", at.Add(time.Hour)); err != nil {
				return err
			}
		case 2: // traveler followed up -> open again
			if _, err := d.tx.AddTicketReply(tk.ID, st.ID, RoleStaff, replies[(i+1)%len(replies)], at.Add(time.Hour)); err != nil {
				return err
			}
			if _, err := d.tx.AddTicketReply(tk.ID, owner.ID, RoleTraveler, "ขอบคุณค่ะ ยังไม่ได้รับการแก้ไข รบกวนตรวจสอบอีกครั้งได้ไหมคะ", at.Add(5*time.Hour)); err != nil {
				return err
			}
		case 3: // resolved
			if _, err := d.tx.AddTicketReply(tk.ID, st.ID, RoleStaff, replies[1], at.Add(2*time.Hour)); err != nil {
				return err
			}
			if err := d.tx.SetTicketStatus(tk.ID, "resolved", at.Add(3*time.Hour)); err != nil {
				return err
			}
		}
	}
	return nil
}

// ---- community: 7 more groups (10), 30 posts (35), comments, likes, saves, follows, reports ----------------------

func (d *demoDS) community() error {
	// Suspended members take no part in community activity (the API hides them).
	act := make([]*userRec, 0, len(d.all))
	for _, u := range d.all {
		if u.Status != "suspended" {
			act = append(act, u)
		}
	}
	type g struct {
		name, desc, dest, css string
	}
	extra := []g{
		{"คาเฟ่ฮอปปิ้งทั่วไทย", "แชร์พิกัดคาเฟ่และร้านขนมน่านั่ง", "", "linear-gradient(135deg,#f59e0b,#ef4444)"},
		{"สายกิน กรุงเทพฯ", "ร้านเด็ดในกรุงเทพฯ ตั้งแต่สตรีทฟู้ดถึงมิชลิน", "กรุงเทพฯ", "linear-gradient(135deg,#002765,#6366f1)"},
		{"ครอบครัวเที่ยวสนุก", "เคล็ดลับเที่ยวกับเด็กเล็กและผู้สูงอายุ", "", "linear-gradient(135deg,#0f766e,#38bdf8)"},
		{"ดำน้ำและกิจกรรมทางทะเล", "แหล่งดำน้ำ ฤดูกาลที่เหมาะสม และอุปกรณ์", "กระบี่", "linear-gradient(135deg,#0b5ed7,#2bb3c0)"},
		{"แคมป์ปิ้ง เขาใหญ่-ภูกระดึง", "ลานกางเต็นท์ การเตรียมตัว และสภาพอากาศ", "เขาใหญ่", "linear-gradient(135deg,#166534,#84cc16)"},
		{"เที่ยวเมืองรอง", "ที่เที่ยวเงียบสงบ ไม่แออัด", "", "linear-gradient(135deg,#92400e,#f59e0b)"},
		{"ถ่ายภาพท่องเที่ยว", "แชร์ภาพและมุมถ่ายภาพสวย ๆ", "", "linear-gradient(135deg,#7c3aed,#22c55e)"},
	}
	for _, x := range extra {
		gd := GroupDetail{Group: Group{Name: x.name, Description: x.desc, GradientCSS: x.css}, Rules: []string{"สุภาพและให้เกียรติกัน", "ห้ามโฆษณาโดยไม่ได้รับอนุญาต", "แชร์ข้อมูลที่ตรวจสอบแล้ว"}}
		if x.dest != "" {
			dest := x.dest
			gd.Destination = &dest
		}
		if err := d.tx.UpsertGroup(gd); err != nil {
			return err
		}
	}
	groups, err := d.tx.Groups(d.demo.ID)
	if err != nil {
		return err
	}
	// Memberships: each traveler joins 2-4 groups, the demo traveler joins several new ones.
	for i, u := range act {
		for k := 0; k < 2+i%3; k++ {
			grp := groups[(i+k*3)%len(groups)]
			if i == 0 && k == 0 {
				continue // demo already belongs to the first two seeded groups
			}
			if joined, err := d.tx.ToggleGroup(u.ID, grp.ID); err != nil {
				return err
			} else if !joined { // toggled off an existing membership: join again
				if _, err := d.tx.ToggleGroup(u.ID, grp.ID); err != nil {
					return err
				}
			}
		}
	}

	texts := []struct{ typ, dest, text string }{
		{"story", "ภูเก็ต", "ไปดำน้ำเกาะราชามา น้ำใสมาก ปลาเยอะ แนะนำให้ไปช่วงเช้า"}, {"tip", "เชียงใหม่", "คาเฟ่ลับย่านสันติธรรม เปิดแต่เช้า กาแฟดี ที่นั่งเยอะ"},
		{"question", "กระบี่", "ทัวร์ 4 เกาะ ควรจองช่วงไหนถึงจะไม่แออัดคะ?"}, {"trip", "เกาะสมุย", "แผน 4 วัน 3 คืน เน้นผ่อนคลาย งบ 15,000"},
		{"story", "กรุงเทพฯ", "ล่องเรือเจ้าพระยามื้อค่ำ วิวสวยจริง แนะนำโต๊ะริมหน้าต่าง"}, {"tip", "ปาย", "เช่ามอเตอร์ไซค์ในปาย ตรวจเบรกก่อนออกทุกครั้ง"},
		{"story", "อยุธยา", "ปั่นจักรยานรอบเกาะเมือง วันเดียวไหว แนะนำออกก่อน 8 โมง"}, {"question", "เขาใหญ่", "ลานกางเต็นท์ไหนมีห้องน้ำสะอาดและปลั๊กไฟบ้างครับ?"},
		{"tip", "ภูเก็ต", "ร้านอาหารทะเลเมืองเก่า ไปก่อน 18:00 ไม่ต้องรอคิว"}, {"trip", "เชียงใหม่", "แผน 3 วัน 2 คืน เน้นดอยและคาเฟ่ งบ 9,000"},
		{"story", "กระบี่", "ไร่เลย์ช่วงพระอาทิตย์ตกสวยที่สุด เดินจากหาดอ่าวนางไม่ถึง 20 นาที"}, {"question", "กรุงเทพฯ", "แนะนำที่พักใกล้ BTS งบไม่เกิน 2,000 ค่ะ"},
		{"tip", "เกาะสมุย", "ขับรถรอบเกาะ ควรเติมน้ำมันก่อนเข้าฝั่งตะวันตก"}, {"story", "ปาย", "ดูหมอกยามเช้าที่ปายแคนยอน ต้องไปก่อนตะวันขึ้น"},
		{"trip", "ภูเก็ต", "แผน 2 วัน 1 คืนสายกิน เมืองเก่า-กะตะ"}, {"story", "เขาใหญ่", "ไวน์เทสติ้งที่ไร่ในปากช่อง บรรยากาศดีมาก"},
		{"tip", "กรุงเทพฯ", "ซื้อบัตรรถไฟฟ้าแบบเที่ยวต่อเนื่อง ประหยัดกว่าซื้อรายครั้ง"}, {"question", "อยุธยา", "ไหว้พระเก้าวัด ควรเริ่มที่วัดไหนก่อนดีคะ?"},
		{"story", "เชียงใหม่", "ปางช้างจริยธรรม ได้ใกล้ชิดช้างโดยไม่ต้องขี่ ประทับใจมาก"}, {"tip", "กระบี่", "ฤดูมรสุม พ.ค.–ต.ค. คลื่นแรง ตรวจสอบสภาพทะเลก่อนจองทัวร์"},
		{"trip", "ปาย", "แผน 3 วัน 2 คืน งบ 6,500 พักริมน้ำ"}, {"story", "ภูเก็ต", "ตื่นเช้าไปใส่บาตรที่วัดฉลอง บรรยากาศสงบมาก"},
		{"question", "เกาะสมุย", "ช่วงเดือนไหนคลื่นสงบที่สุดสำหรับดำน้ำ?"}, {"tip", "เชียงใหม่", "ขึ้นดอยอินทนนท์ พกเสื้อกันหนา อุณหภูมิต่ำกว่าในเมือง 10 องศา"},
		{"story", "กรุงเทพฯ", "ตลาดนัดจตุจักร เดินทั้งวันก็ไม่หมด ซื้อของที่ระลึกได้ครบ"}, {"trip", "อยุธยา", "แผนวันเดียว วัด-อาหาร-ล่องเรือ งบ 1,500"},
		{"tip", "เขาใหญ่", "ช่วงปลายปีอากาศเย็น จองที่พักล่วงหน้า 1 เดือน"}, {"story", "กระบี่", "ทะเลแหวกตอนน้ำลด สวยมาก แต่ต้องดูเวลาน้ำให้ดี"},
		{"question", "ปาย", "รถประจำทางจากเชียงใหม่ไปปายออกกี่โมงบ้างคะ?"}, {"story", "เกาะสมุย", "อ่างทอง 1 วัน คายัคสนุกมาก น้ำใสเห็นปลา"},
	}
	var postIDs []int64
	for i, t := range texts {
		author := act[1+i%(len(act)-1)]
		p := Post{Type: t.typ, Destination: t.dest, Text: t.text, CreatedAt: d.ago(0, 2+i*9)}
		if t.typ == "trip" {
			p.SharedTrip = &SharedTrip{Days: 2 + i%4, Budget: float64(6000 + i*400), Styles: []string{"ผ่อนคลาย", "ธรรมชาติ"}}
		}
		if i%3 == 0 {
			gid := groups[i%len(groups)].ID
			p.GroupID = &gid
		}
		key := ""
		if i%2 == 0 {
			key = "post_image/seed.jpg"
		}
		np, err := d.tx.AddPost(author.ID, p, key, d.now)
		if err != nil {
			return err
		}
		postIDs = append(postIDs, np.ID)
	}
	commentTexts := []string{"ขอบคุณสำหรับข้อมูลค่ะ เป็นประโยชน์มาก", "เคยไปมาแล้ว แนะนำเหมือนกันครับ", "ขอพิกัดหน่อยได้ไหมคะ", "ช่วงนี้ไปได้ไหม ฝนตกหรือเปล่า?", "ภาพสวยมาก อยากไปบ้าง",
		"บันทึกไว้เลย ปีหน้าจะไป", "ราคาประมาณเท่าไรครับ", "ไปกับครอบครัวได้ไหมคะ", "ตรงกับที่ไปมาเลย ชอบมาก", "ขอบคุณที่แชร์ครับ"}
	for i, pid := range postIDs {
		for k := 0; k < 1+i%3; k++ {
			u := act[(i*3+k*5+1)%len(act)]
			if _, err := d.tx.AddComment(u.ID, pid, commentTexts[(i+k)%len(commentTexts)], d.ago(0, 1+i*8-k)); err != nil {
				return err
			}
		}
		// likes from a deterministic subset of travelers, saves from a smaller one
		for k := 0; k < 4+i%9; k++ {
			if _, _, err := d.tx.TogglePostLike(act[(i+k*2)%len(act)].ID, pid); err != nil {
				return err
			}
		}
		if i%3 == 0 {
			if _, err := d.tx.TogglePostSave(d.demo.ID, pid); err != nil {
				return err
			}
		}
	}
	// Follows: everyone follows 3-6 others; the demo traveler follows 5 more members.
	handles := make([]string, 0, len(act))
	for _, u := range act[1:] {
		if u.Status != "suspended" { // suspended members cannot be followed (the API treats them as unknown)
			handles = append(handles, u.Handle)
		}
	}
	for i, u := range act {
		for k := 0; k < 3+i%4; k++ {
			h := handles[(i*2+k*3)%len(handles)]
			if h == u.Handle {
				continue
			}
			if _, err := d.tx.ToggleFollow(u.ID, h); err != nil {
				return err
			}
		}
	}
	// Moderation queue: a handful of reports.
	for i, reason := range []string{"โฆษณาโดยไม่ได้รับอนุญาต", "ข้อมูลเท็จ", "ภาษาไม่สุภาพ", "สแปม", "ข้อมูลส่วนตัวของผู้อื่น"} {
		if err := d.tx.ReportPost(act[2+i].ID, postIDs[(i*5+2)%len(postIDs)], reason, d.ago(i, 3)); err != nil {
			return err
		}
	}
	return nil
}

// ---- audit trail (32 entries across admin / staff / partner actions) --------------------------------------------

func (d *demoDS) audits() error {
	add := func(actor *userRec, action, entity string, id int64, meta map[string]any, ago time.Duration) error {
		return d.tx.AddAudit(auditRec{ActorID: actor.ID, Action: action, Entity: entity, EntityID: fmt.Sprint(id), Meta: meta, CreatedAt: d.now.Add(-ago)})
	}
	h := time.Hour
	for i, p := range d.partners[1:] { // the partner decisions (the first partner was approved by the base seed)
		if p.Status == PartnerPending {
			continue
		}
		if err := add(d.admin, "admin.partner.update", "partner", p.ID, map[string]any{"status": p.Status, "note": p.DecisionNote}, time.Duration(20+i)*24*h); err != nil {
			return err
		}
	}
	for i := 0; i < 6; i++ {
		u := d.all[3+i]
		if err := add(d.admin, "admin.user.update", "user", u.ID, map[string]any{"role": "traveler", "status": []string{"active", "suspended"}[i%2]}, time.Duration(2+i*4)*24*h); err != nil {
			return err
		}
	}
	for i, id := range []int64{15, 20, 28, 24} {
		if err := add(d.admin, "admin.listing.status", "listing", id, map[string]any{"status": []string{"draft", "archived", "archived", "published"}[i]}, time.Duration(3+i*6)*24*h); err != nil {
			return err
		}
	}
	for i, id := range []int64{13, 14, 15, 5, 16, 8, 19, 10} {
		action := "partner.listing.create"
		if i%2 == 1 {
			action = "partner.listing.update"
		}
		owner := d.owners[0]
		if pid := d.listingOwner[id]; pid != 0 {
			for k, p := range d.partners {
				if p.ID == pid && k > 0 && k-1 < len(d.owners) {
					owner = d.owners[k-1]
				}
			}
		}
		if err := add(owner, action, "listing", id, map[string]any{"name": fmt.Sprintf("listing-%d", id)}, time.Duration(1+i*3)*24*h); err != nil {
			return err
		}
	}
	n := 0
	for _, b := range d.books {
		if b.Status == "confirmed" || b.Status == "completed" {
			if err := add(d.staff[n%len(d.staff)], "staff.booking.check", "booking", b.ID, map[string]any{"result": "verified"}, time.Duration(1+n)*5*h); err != nil {
				return err
			}
			n++
			if n >= 8 {
				break
			}
		}
	}
	for i := 0; i < 6; i++ {
		action := []string{"staff.ticket.reply", "staff.ticket.status"}[i%2]
		if err := add(d.staff[i%len(d.staff)], action, "ticket", int64(i+1), map[string]any{"status": "answered"}, time.Duration(i*7+1)*h); err != nil {
			return err
		}
	}
	return nil
}

func demoReviews(seen map[[2]int64]bool, uid int64) (n int) {
	for k := range seen {
		if k[0] == uid {
			n++
		}
	}
	return
}
