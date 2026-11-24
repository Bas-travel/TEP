package api

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

func (s *Server) adminAnalyticsRoutes(m *http.ServeMux) {
	p := s.portal(RoleAdmin)
	m.HandleFunc("GET /admin/v1/analytics", p(s.adminAnalytics))
	m.HandleFunc("GET /admin/v1/analytics/export", p(s.adminAnalyticsExport))
}

// analyticsWindow parses ?days= (1..365, default 30) into [from, to): the last N Bangkok days including today.
func (s *Server) analyticsWindow(r *http.Request) (days int, from, to time.Time, errs []httpx.FieldError) {
	days = 30
	if v := r.URL.Query().Get("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 365 {
			errs = append(errs, fe("days", "must be 1..365"))
			return
		}
		days = n
	}
	today := startOfDay(s.now())
	to = today.AddDate(0, 0, 1)
	from = today.AddDate(0, 0, -(days - 1))
	return
}

type analyticsResult struct {
	Days      int
	Series    []AnalyticsPoint
	ByType    []TypeAgg
	ByStatus  map[string]int
	Bookings  int
	Revenue   float64
	NewUsers  int
	AvgOrder  float64
	CancelPct float64
}

func (s *Server) buildAnalytics(r *http.Request) (*analyticsResult, []httpx.FieldError, error) {
	days, from, to, errs := s.analyticsWindow(r)
	if len(errs) > 0 {
		return nil, errs, nil
	}
	var raw AnalyticsRaw
	if err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { raw, e = tx.Analytics(from, to); return }); err != nil {
		return nil, nil, err
	}
	idx := map[string]*AnalyticsPoint{}
	res := &analyticsResult{Days: days, ByType: raw.PaidByType, ByStatus: raw.ByStatus}
	res.Series = make([]AnalyticsPoint, days)
	for i := 0; i < days; i++ {
		d := from.AddDate(0, 0, i).In(ictZone).Format("2006-01-02")
		res.Series[i] = AnalyticsPoint{Date: d}
		idx[d] = &res.Series[i]
	}
	for _, d := range raw.PaidByDay {
		if p := idx[d.Day]; p != nil {
			p.Bookings, p.Revenue = d.Count, round2(d.Revenue)
			res.Bookings += d.Count
			res.Revenue += d.Revenue
		}
	}
	for _, d := range raw.UsersByDay {
		if p := idx[d.Day]; p != nil {
			p.NewUsers = d.Count
			res.NewUsers += d.Count
		}
	}
	res.Revenue = round2(res.Revenue)
	if res.Bookings > 0 {
		res.AvgOrder = round2(res.Revenue / float64(res.Bookings))
	}
	all := 0
	for _, n := range raw.ByStatus {
		all += n
	}
	if all > 0 {
		res.CancelPct = round2(float64(raw.ByStatus["cancelled"]) * 100 / float64(all))
	}
	for i := range res.ByType {
		res.ByType[i].Revenue = round2(res.ByType[i].Revenue)
	}
	if res.ByType == nil {
		res.ByType = []TypeAgg{}
	}
	if res.ByStatus == nil {
		res.ByStatus = map[string]int{}
	}
	return res, nil, nil
}

func (s *Server) adminAnalytics(w http.ResponseWriter, r *http.Request, a actor) {
	res, errs, err := s.buildAnalytics(r)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{
		"days": res.Days, "series": res.Series, "byType": res.ByType, "byStatus": res.ByStatus,
		"totals": map[string]any{"bookings": res.Bookings, "revenue": res.Revenue, "newUsers": res.NewUsers,
			"avgOrderValue": res.AvgOrder, "cancellationRatePct": res.CancelPct},
	})
}

// adminAnalyticsExport streams the daily series as CSV (numbers and ISO dates only, so no spreadsheet-formula injection).
func (s *Server) adminAnalyticsExport(w http.ResponseWriter, r *http.Request, a actor) {
	res, errs, err := s.buildAnalytics(r)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	if s.fail(w, r, err) {
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="tep-analytics-`+strconv.Itoa(res.Days)+`d.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"date", "paid_bookings", "revenue_thb", "new_users"})
	for _, p := range res.Series {
		_ = cw.Write([]string{p.Date, strconv.Itoa(p.Bookings), strconv.FormatFloat(p.Revenue, 'f', 2, 64), strconv.Itoa(p.NewUsers)})
	}
	cw.Flush()
}
