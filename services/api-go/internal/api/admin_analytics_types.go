package api

// Raw aggregates a repository returns for the admin analytics window [from, to).
// Days are Asia/Bangkok calendar days (YYYY-MM-DD); zero-filling happens in the handler.
type DayAgg struct {
	Day     string
	Count   int
	Revenue float64
}

type TypeAgg struct {
	Type    string  `json:"type"`
	Count   int     `json:"bookings"`
	Revenue float64 `json:"revenue"`
}

type AnalyticsRaw struct {
	PaidByDay  []DayAgg       // paid bookings (confirmed or completed) per day
	UsersByDay []DayAgg       // sign-ups per day (Count only)
	PaidByType []TypeAgg      // paid bookings per item type
	ByStatus   map[string]int // every booking in the window by status
}

// AnalyticsPoint is one day of the zero-filled series returned to clients.
type AnalyticsPoint struct {
	Date     string  `json:"date"`
	Bookings int     `json:"bookings"`
	Revenue  float64 `json:"revenue"`
	NewUsers int     `json:"newUsers"`
}
