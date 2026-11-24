package api

import "time"

// sprintfDay buckets a UTC DATETIME column into Asia/Bangkok calendar days (fixed +07:00, no tz tables needed).
func sprintfDay(col string) string {
	return "DATE_FORMAT(DATE_ADD(" + col + ", INTERVAL 7 HOUR),'%Y-%m-%d')"
}

func (t *sqlTx) Analytics(from, to time.Time) (AnalyticsRaw, error) {
	out := AnalyticsRaw{ByStatus: map[string]int{}}
	dayAggs := func(q string, args ...any) ([]DayAgg, error) {
		rows, err := t.query(q, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var res []DayAgg
		for rows.Next() {
			var d DayAgg
			if err := rows.Scan(&d.Day, &d.Count, &d.Revenue); err != nil {
				return nil, err
			}
			res = append(res, d)
		}
		return res, rows.Err()
	}
	var err error
	bd := sprintfDay("b.booked_at")
	if out.PaidByDay, err = dayAggs(`SELECT `+bd+` d, COUNT(*), COALESCE(SUM(b.total),0) FROM bookings b
		WHERE b.status IN ('confirmed','completed') AND b.booked_at>=? AND b.booked_at<? GROUP BY d ORDER BY d`, utc(from), utc(to)); err != nil {
		return out, err
	}
	if out.UsersByDay, err = dayAggs(`SELECT `+sprintfDay("u.created_at")+` d, COUNT(*), 0 FROM users u
		WHERE u.created_at>=? AND u.created_at<? GROUP BY d ORDER BY d`, utc(from), utc(to)); err != nil {
		return out, err
	}
	rows, err := t.query(`SELECT b.item_type, COUNT(*), COALESCE(SUM(b.total),0) FROM bookings b
		WHERE b.status IN ('confirmed','completed') AND b.booked_at>=? AND b.booked_at<? GROUP BY b.item_type ORDER BY SUM(b.total) DESC`, utc(from), utc(to))
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var a TypeAgg
		if err := rows.Scan(&a.Type, &a.Count, &a.Revenue); err != nil {
			rows.Close()
			return out, err
		}
		out.PaidByType = append(out.PaidByType, a)
	}
	rows.Close()
	rows, err = t.query(`SELECT status, COUNT(*) FROM bookings WHERE booked_at>=? AND booked_at<? GROUP BY status`, utc(from), utc(to))
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			return out, err
		}
		out.ByStatus[st] = n
	}
	return out, rows.Err()
}
