package api

import (
	"sort"
	"time"
)

func (m *memRepo) Analytics(from, to time.Time) (AnalyticsRaw, error) {
	out := AnalyticsRaw{ByStatus: map[string]int{}}
	day := func(t time.Time) string { return t.In(ictZone).Format("2006-01-02") }
	in := func(t time.Time) bool { return !t.Before(from) && t.Before(to) }
	paid, users := map[string]*DayAgg{}, map[string]*DayAgg{}
	types := map[string]*TypeAgg{}
	for _, b := range m.bookings {
		if !in(b.Date) {
			continue
		}
		out.ByStatus[b.Status]++
		if b.Status != "confirmed" && b.Status != "completed" {
			continue
		}
		d := paid[day(b.Date)]
		if d == nil {
			d = &DayAgg{Day: day(b.Date)}
			paid[d.Day] = d
		}
		d.Count++
		d.Revenue += b.Amount
		t := types[b.Type]
		if t == nil {
			t = &TypeAgg{Type: b.Type}
			types[b.Type] = t
		}
		t.Count++
		t.Revenue += b.Amount
	}
	for _, u := range m.users {
		if in(u.CreatedAt) {
			d := users[day(u.CreatedAt)]
			if d == nil {
				d = &DayAgg{Day: day(u.CreatedAt)}
				users[d.Day] = d
			}
			d.Count++
		}
	}
	for _, d := range paid {
		out.PaidByDay = append(out.PaidByDay, *d)
	}
	for _, d := range users {
		out.UsersByDay = append(out.UsersByDay, *d)
	}
	for _, t := range types {
		out.PaidByType = append(out.PaidByType, *t)
	}
	sort.Slice(out.PaidByType, func(i, j int) bool { return out.PaidByType[i].Revenue > out.PaidByType[j].Revenue })
	return out, nil
}
