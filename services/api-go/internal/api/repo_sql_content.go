package api

import (
	"database/sql"
	"strings"
)

func (t *sqlTx) AdminCoupons(today string) ([]CouponAdmin, error) {
	rows, err := t.query(`SELECT c.code, c.title, COALESCE(c.detail,''), COALESCE(c.percent_off,0), c.min_spend, COALESCE(c.max_discount,0),
		COALESCE(DATE_FORMAT(c.expires_on,'%Y-%m-%d'),''), (SELECT COUNT(*) FROM user_coupons uc WHERE uc.coupon_id=c.id)
		FROM coupons c ORDER BY c.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CouponAdmin{}
	for rows.Next() {
		var c CouponAdmin
		if err := rows.Scan(&c.Code, &c.Title, &c.Detail, &c.PercentOff, &c.MinSpend, &c.MaxDiscount, &c.ExpiresOn, &c.Granted); err != nil {
			return nil, err
		}
		c.Expired = c.ExpiresOn != "" && c.ExpiresOn < today
		out = append(out, c)
	}
	return out, rows.Err()
}

func (t *sqlTx) DeleteCoupon(code string) error {
	var id int64
	if err := t.row(`SELECT id FROM coupons WHERE code=? FOR UPDATE`, code).Scan(&id); err != nil {
		return notFound(err)
	}
	var held int
	if err := t.row(`SELECT COUNT(*) FROM user_coupons WHERE coupon_id=?`, id).Scan(&held); err != nil {
		return err
	}
	if held > 0 {
		return ErrDuplicate
	}
	_, err := t.exec(`DELETE FROM coupons WHERE id=?`, id)
	return err
}

func (t *sqlTx) Faqs() ([]FaqAdmin, error) {
	rows, err := t.query(`SELECT id, question, answer, sort_no FROM faqs ORDER BY sort_no, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []FaqAdmin{}
	for rows.Next() {
		var f FaqAdmin
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &f.SortNo); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (t *sqlTx) SaveFaq(f *FaqAdmin) error {
	if f.ID == 0 {
		res, err := t.exec(`INSERT INTO faqs (question, answer, sort_no) VALUES (?,?,?)`, f.Question, f.Answer, f.SortNo)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		f.ID = int(id)
		return err
	}
	if err := t.exists(`SELECT 1 FROM faqs WHERE id=? FOR UPDATE`, f.ID); err != nil {
		return err
	}
	_, err := t.exec(`UPDATE faqs SET question=?, answer=?, sort_no=? WHERE id=?`, f.Question, f.Answer, f.SortNo, f.ID)
	return err
}

func (t *sqlTx) DeleteFaq(id int) error {
	res, err := t.exec(`DELETE FROM faqs WHERE id=?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (t *sqlTx) PostReports(offset, limit int) ([]PostReportRow, int, error) {
	var total int
	if err := t.row(`SELECT COUNT(DISTINCT post_id) FROM post_reports`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.query(`SELECT r.post_id, p.body, COALESCE(u.handle,''), COUNT(*), GROUP_CONCAT(r.reason ORDER BY r.created_at SEPARATOR '\n'), MAX(r.created_at)
		FROM post_reports r JOIN community_posts p ON p.id=r.post_id JOIN users u ON u.id=p.author_id
		GROUP BY r.post_id, p.body, u.handle ORDER BY COUNT(*) DESC, MAX(r.created_at) DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []PostReportRow{}
	for rows.Next() {
		var r PostReportRow
		var reasons sql.NullString
		if err := rows.Scan(&r.PostID, &r.PostText, &r.Author, &r.Reports, &reasons, &r.LastReport); err != nil {
			return nil, 0, err
		}
		r.Reasons = []string{}
		if reasons.Valid && reasons.String != "" {
			r.Reasons = strings.Split(reasons.String, "\n")
		}
		r.LastReport = r.LastReport.UTC()
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func (t *sqlTx) DismissPostReports(postID int64) (int, error) {
	res, err := t.exec(`DELETE FROM post_reports WHERE post_id=?`, postID)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return 0, ErrNotFound
	}
	return int(n), nil
}

func (t *sqlTx) AdminDeletePost(postID int64) error {
	res, err := t.exec(`DELETE FROM community_posts WHERE id=?`, postID) // reports/likes/comments cascade
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
