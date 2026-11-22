package api

import (
	"sort"
	"time"
)

type postReportRec struct {
	postID, userID int64
	reason         string
	at             time.Time
}

func (m *memRepo) AdminCoupons(today string) ([]CouponAdmin, error) {
	out := make([]CouponAdmin, 0, len(m.coupons))
	for code, c := range m.coupons {
		n := 0
		for _, held := range m.granted {
			for _, h := range held {
				if h == code {
					n++
				}
			}
		}
		out = append(out, CouponAdmin{Code: c.Code, Title: c.Title, Detail: c.Detail, PercentOff: c.PercentOff, MinSpend: c.MinSpend,
			MaxDiscount: c.MaxDiscount, ExpiresOn: c.ExpiresOn, Granted: n, Expired: c.ExpiresOn != "" && c.ExpiresOn < today})
	}
	sort.Slice(out, func(i, j int) bool { return m.couponIDs[out[i].Code] < m.couponIDs[out[j].Code] })
	return out, nil
}

func (m *memRepo) DeleteCoupon(code string) error {
	if _, ok := m.coupons[code]; !ok {
		return ErrNotFound
	}
	for _, held := range m.granted {
		for _, h := range held {
			if h == code {
				return ErrDuplicate
			}
		}
	}
	delete(m.coupons, code)
	delete(m.couponIDs, code)
	return nil
}

func (m *memRepo) initFaqs() {
	if m.faqsInit {
		return
	}
	m.faqsInit = true
	for i, f := range seedFaqs {
		m.faqs = append(m.faqs, FaqAdmin{ID: f.ID, Question: f.Question, Answer: f.Answer, SortNo: i})
		if f.ID >= m.nextFaq {
			m.nextFaq = f.ID + 1
		}
	}
}

func (m *memRepo) Faqs() ([]FaqAdmin, error) {
	m.initFaqs()
	out := append([]FaqAdmin{}, m.faqs...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SortNo != out[j].SortNo {
			return out[i].SortNo < out[j].SortNo
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (m *memRepo) SaveFaq(f *FaqAdmin) error {
	m.initFaqs()
	if f.ID == 0 {
		f.ID = m.nextFaq
		m.nextFaq++
		m.faqs = append(m.faqs, *f)
		return nil
	}
	for i := range m.faqs {
		if m.faqs[i].ID == f.ID {
			m.faqs[i] = *f
			return nil
		}
	}
	return ErrNotFound
}

func (m *memRepo) DeleteFaq(id int) error {
	m.initFaqs()
	for i := range m.faqs {
		if m.faqs[i].ID == id {
			m.faqs = append(m.faqs[:i], m.faqs[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *memRepo) PostReports(offset, limit int) ([]PostReportRow, int, error) {
	byPost := map[int64]*PostReportRow{}
	var order []int64
	for _, r := range m.reports {
		p, ok := m.posts[r.postID]
		if !ok {
			continue
		}
		row := byPost[r.postID]
		if row == nil {
			author := ""
			if u := m.users[p.authorUID]; u != nil {
				author = u.Handle
			}
			row = &PostReportRow{PostID: r.postID, PostText: p.Text, Author: author, Reasons: []string{}}
			byPost[r.postID] = row
			order = append(order, r.postID)
		}
		row.Reports++
		row.Reasons = append(row.Reasons, r.reason)
		if r.at.After(row.LastReport) {
			row.LastReport = r.at
		}
	}
	rows := make([]PostReportRow, 0, len(order))
	for _, id := range order {
		rows = append(rows, *byPost[id])
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Reports != rows[j].Reports {
			return rows[i].Reports > rows[j].Reports
		}
		return rows[i].LastReport.After(rows[j].LastReport)
	})
	total := len(rows)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return rows[offset:end], total, nil
}

func (m *memRepo) DismissPostReports(postID int64) (int, error) {
	kept, n := m.reports[:0], 0
	for _, r := range m.reports {
		if r.postID == postID {
			n++
			continue
		}
		kept = append(kept, r)
	}
	m.reports = kept
	if n == 0 {
		return 0, ErrNotFound
	}
	return n, nil
}

func (m *memRepo) AdminDeletePost(postID int64) error {
	if _, ok := m.posts[postID]; !ok {
		return ErrNotFound
	}
	m.deletePost(postID)
	kept := m.reports[:0]
	for _, r := range m.reports {
		if r.postID != postID {
			kept = append(kept, r)
		}
	}
	m.reports = kept
	return nil
}
