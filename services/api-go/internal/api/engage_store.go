package api

import (
	"fmt"
	"time"
)

type Settings struct {
	Language      string        `json:"language"`
	Currency      string        `json:"currency"`
	Theme         string        `json:"theme"`
	Notifications NotifySetting `json:"notifications"`
}

type NotifySetting struct {
	Push  bool `json:"push"`
	Email bool `json:"email"`
	Trip  bool `json:"trip"`
	Promo bool `json:"promo"`
}

type UserCoupon struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	Detail      string  `json:"detail"`
	MinSpend    float64 `json:"minSpend"`
	MaxDiscount float64 `json:"maxDiscount"`
	Expiry      string  `json:"expiry"`
	Status      string  `json:"status"`
	pct         float64
}

type Notification struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}

type ItineraryItem struct {
	Time  string `json:"time"`
	Text  string `json:"text"`
	Place string `json:"place"`
}

type ItineraryDay struct {
	Title string          `json:"title"`
	Items []ItineraryItem `json:"items"`
}

type BudgetLine struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type ItineraryResult struct {
	Days            []ItineraryDay `json:"days"`
	BudgetBreakdown []BudgetLine   `json:"budgetBreakdown"`
}

type SavedTripInput struct {
	Destination string           `json:"destination"`
	Days        int              `json:"days"`
	Travelers   int              `json:"travelers"`
	Budget      float64          `json:"budget"`
	Itinerary   *ItineraryResult `json:"itinerary,omitempty"`
	Source      string           `json:"source"`
}

type SavedTrip struct {
	ID int64 `json:"id"`
	SavedTripInput
}

type Review struct {
	ID        int64     `json:"id"`
	ItemName  string    `json:"itemName"`
	ItemType  string    `json:"itemType"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

type SupportTicket struct {
	ID        int64     `json:"id"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type SharedTrip struct {
	Days   int      `json:"days"`
	Budget float64  `json:"budget"`
	Styles []string `json:"styles"`
}

type Post struct {
	ID          int64       `json:"id"`
	AuthorID    string      `json:"authorId"`
	Type        string      `json:"type"`
	Destination string      `json:"destination"`
	GroupID     *int64      `json:"groupId"`
	Text        string      `json:"text"`
	LikeCount   int         `json:"likeCount"`
	HasImage    bool        `json:"hasImage"`
	SharedTrip  *SharedTrip `json:"sharedTrip"`
	CreatedAt   time.Time   `json:"createdAt"`
}

type ItineraryJob struct {
	JobID    string           `json:"jobId"`
	Status   string           `json:"status"`
	Degraded bool             `json:"degraded"`
	Result   *ItineraryResult `json:"result,omitempty"`
}

// userExtra is per-user engagement state (settings, coupons, notifications...). Guarded by store.mu.
type userExtra struct {
	Settings Settings
	Coupons  []UserCoupon
	Notifs   []Notification
	Trips    []SavedTrip
	Reviews  []Review
	Tickets  []SupportTicket
	Liked    map[int64]bool
	Jobs     map[string]ItineraryJob
}

// extra returns (creating) the per-user state. Caller must hold s.st.mu.
func (s *Server) extra(uid int64) *userExtra {
	if s.st.extras == nil {
		s.st.extras = map[int64]*userExtra{}
	}
	x := s.st.extras[uid]
	if x == nil {
		x = &userExtra{
			Settings: Settings{Language: "th", Currency: "THB", Theme: "light", Notifications: NotifySetting{true, true, true, true}},
			Liked:    map[int64]bool{}, Jobs: map[string]ItineraryJob{},
		}
		s.st.extras[uid] = x
	}
	return x
}

func (s *Server) nextID() int64 {
	s.st.engageSeq++
	return s.st.engageSeq
}

func buildItinerary(dest string, days int, budget float64, travelers int) *ItineraryResult {
	if travelers < 1 {
		travelers = 1
	}
	res := &ItineraryResult{}
	for d := 1; d <= days; d++ {
		res.Days = append(res.Days, ItineraryDay{
			Title: fmt.Sprintf("วันที่ %d · %s", d, dest),
			Items: []ItineraryItem{
				{Time: "09:00", Text: "อาหารเช้าและเดินเล่นย่านท้องถิ่น", Place: dest},
				{Time: "13:00", Text: "มื้อกลางวันร้านแนะนำ", Place: dest},
				{Time: "16:00", Text: "เที่ยวจุดชมวิว/แลนด์มาร์ก", Place: dest},
				{Time: "19:00", Text: "มื้อเย็นและพักผ่อน", Place: dest},
			},
		})
	}
	res.BudgetBreakdown = []BudgetLine{
		{"ที่พัก", round2(budget * 0.40)}, {"อาหาร", round2(budget * 0.25)},
		{"เดินทาง", round2(budget * 0.20)}, {"กิจกรรม", round2(budget * 0.15)},
	}
	return res
}
