package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

var couponCodeRe = regexp.MustCompile(`^[A-Z0-9_-]{3,40}$`)

// adminContentRoutes: coupons, FAQs and the community moderation queue (all admin-only, all audited).
func (s *Server) adminContentRoutes(m *http.ServeMux) {
	p := s.portal(RoleAdmin)
	s.adminDestinationRoutes(m)
	s.adminPaymentRoutes(m)
	s.adminAnalyticsRoutes(m)
	s.platformSettingsRoutes(m)
	m.HandleFunc("GET /admin/v1/coupons", p(s.adminListCoupons))
	m.HandleFunc("POST /admin/v1/coupons", p(s.adminCreateCoupon))
	m.HandleFunc("PUT /admin/v1/coupons/{code}", p(s.adminUpdateCoupon))
	m.HandleFunc("DELETE /admin/v1/coupons/{code}", p(s.adminDeleteCoupon))
	m.HandleFunc("GET /admin/v1/faqs", p(s.adminListFaqs))
	m.HandleFunc("POST /admin/v1/faqs", p(s.adminCreateFaq))
	m.HandleFunc("PUT /admin/v1/faqs/{faqId}", p(s.adminUpdateFaq))
	m.HandleFunc("DELETE /admin/v1/faqs/{faqId}", p(s.adminDeleteFaq))
	m.HandleFunc("GET /admin/v1/reports", p(s.adminListReports))
	m.HandleFunc("POST /admin/v1/reports/{postId}/dismiss", p(s.adminDismissReports))
	m.HandleFunc("POST /admin/v1/reports/{postId}/remove", p(s.adminRemoveReportedPost))
}

// ---- coupons ----

type couponIn struct {
	Code        string   `json:"code"`
	Title       string   `json:"title"`
	Detail      string   `json:"detail"`
	PercentOff  *int     `json:"percentOff"`
	MinSpend    *float64 `json:"minSpend"`
	MaxDiscount *float64 `json:"maxDiscount"`
	ExpiresOn   string   `json:"expiresOn"`
}

func (in couponIn) validate(needCode bool) (couponDef, []httpx.FieldError) {
	var errs []httpx.FieldError
	c := couponDef{Code: strings.ToUpper(strings.TrimSpace(in.Code)), Title: strings.TrimSpace(in.Title), Detail: strings.TrimSpace(in.Detail), ExpiresOn: in.ExpiresOn}
	if needCode && !couponCodeRe.MatchString(c.Code) {
		errs = append(errs, fe("code", "3-40 characters: A-Z, 0-9, _ or -"))
	}
	if n := utf8.RuneCountInString(c.Title); n < 2 || n > 190 {
		errs = append(errs, fe("title", "2-190 characters"))
	}
	if utf8.RuneCountInString(c.Detail) > 255 {
		errs = append(errs, fe("detail", "at most 255 characters"))
	}
	if in.PercentOff != nil {
		if *in.PercentOff < 0 || *in.PercentOff > 100 {
			errs = append(errs, fe("percentOff", "0..100 (0 = fixed amount)"))
		}
		c.PercentOff = *in.PercentOff
	}
	if in.MinSpend != nil {
		if *in.MinSpend < 0 {
			errs = append(errs, fe("minSpend", "must be >= 0"))
		}
		c.MinSpend = *in.MinSpend
	}
	if in.MaxDiscount != nil {
		if *in.MaxDiscount < 0 {
			errs = append(errs, fe("maxDiscount", "must be >= 0"))
		}
		c.MaxDiscount = *in.MaxDiscount
	}
	if c.PercentOff == 0 && c.MaxDiscount <= 0 {
		errs = append(errs, fe("maxDiscount", "required for a fixed-amount coupon (percentOff = 0)"))
	}
	if c.ExpiresOn != "" {
		if _, err := time.Parse("2006-01-02", c.ExpiresOn); err != nil {
			errs = append(errs, fe("expiresOn", "use YYYY-MM-DD"))
		}
	}
	return c, errs
}

func (s *Server) adminListCoupons(w http.ResponseWriter, r *http.Request, a actor) {
	var rows []CouponAdmin
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { rows, e = tx.AdminCoupons(s.today()); return })
	if s.fail(w, r, err) {
		return
	}
	if q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q"))); q != "" {
		kept := rows[:0]
		for _, c := range rows {
			if strings.Contains(strings.ToLower(c.Code+" "+c.Title), q) {
				kept = append(kept, c)
			}
		}
		rows = kept
	}
	off, lim, errs := pageParams(r)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	total := len(rows)
	if off > total {
		off = total
	}
	end := min(off+lim, total)
	pageJSON(w, rows[off:end], total, off, lim)
}

func (s *Server) adminCreateCoupon(w http.ResponseWriter, r *http.Request, a actor) {
	var in couponIn
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", "invalid JSON"))
		return
	}
	c, errs := in.validate(true)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out []CouponAdmin
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		cur, err := tx.AdminCoupons(s.today())
		if err != nil {
			return err
		}
		for _, x := range cur {
			if x.Code == c.Code {
				return conflict("coupon code already exists")
			}
		}
		if err := tx.UpsertCoupon(c); err != nil {
			return err
		}
		out, err = tx.AdminCoupons(s.today())
		if err != nil {
			return err
		}
		return s.audit(tx, a, "admin.coupon.create", "coupon", 0, map[string]any{"code": c.Code})
	})
	if s.fail(w, r, err) {
		return
	}
	for _, x := range out {
		if x.Code == c.Code {
			httpx.WriteJSON(w, 201, x)
			return
		}
	}
}

func (s *Server) adminUpdateCoupon(w http.ResponseWriter, r *http.Request, a actor) {
	code := strings.ToUpper(r.PathValue("code"))
	var in couponIn
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", "invalid JSON"))
		return
	}
	c, errs := in.validate(false)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	c.Code = code
	var got *CouponAdmin
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		cur, err := tx.AdminCoupons(s.today())
		if err != nil {
			return err
		}
		found := false
		for _, x := range cur {
			found = found || x.Code == code
		}
		if !found {
			return ErrNotFound
		}
		if err := tx.UpsertCoupon(c); err != nil {
			return err
		}
		if cur, err = tx.AdminCoupons(s.today()); err != nil {
			return err
		}
		for i := range cur {
			if cur[i].Code == code {
				got = &cur[i]
			}
		}
		return s.audit(tx, a, "admin.coupon.update", "coupon", 0, map[string]any{"code": code})
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, got)
}

func (s *Server) adminDeleteCoupon(w http.ResponseWriter, r *http.Request, a actor) {
	code := strings.ToUpper(r.PathValue("code"))
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		if err := tx.DeleteCoupon(code); err != nil {
			if err == ErrDuplicate {
				return conflict("coupon is held by travelers; set an expiry date instead of deleting it")
			}
			return err
		}
		return s.audit(tx, a, "admin.coupon.delete", "coupon", 0, map[string]any{"code": code})
	})
	if s.fail(w, r, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- FAQs ----

type faqIn struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	SortNo   *int   `json:"sortNo"`
}

func (in faqIn) validate() (FaqAdmin, []httpx.FieldError) {
	var errs []httpx.FieldError
	f := FaqAdmin{Question: strings.TrimSpace(in.Question), Answer: strings.TrimSpace(in.Answer)}
	if n := utf8.RuneCountInString(f.Question); n < 3 || n > 255 {
		errs = append(errs, fe("question", "3-255 characters"))
	}
	if n := utf8.RuneCountInString(f.Answer); n < 3 || n > 4000 {
		errs = append(errs, fe("answer", "3-4000 characters"))
	}
	if in.SortNo != nil {
		if *in.SortNo < 0 || *in.SortNo > 100000 {
			errs = append(errs, fe("sortNo", "0..100000"))
		}
		f.SortNo = *in.SortNo
	}
	return f, errs
}

func (s *Server) adminListFaqs(w http.ResponseWriter, r *http.Request, a actor) {
	var rows []FaqAdmin
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { rows, e = tx.Faqs(); return })
	if s.fail(w, r, err) {
		return
	}
	off, lim, errs := pageParams(r)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	total := len(rows)
	if off > total {
		off = total
	}
	pageJSON(w, rows[off:min(off+lim, total)], total, off, lim)
}

func (s *Server) adminCreateFaq(w http.ResponseWriter, r *http.Request, a actor) {
	var in faqIn
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", "invalid JSON"))
		return
	}
	f, errs := in.validate()
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		if in.SortNo == nil { // append at the end
			all, err := tx.Faqs()
			if err != nil {
				return err
			}
			for _, x := range all {
				f.SortNo = max(f.SortNo, x.SortNo+1)
			}
		}
		if err := tx.SaveFaq(&f); err != nil {
			return err
		}
		return s.audit(tx, a, "admin.faq.create", "faq", int64(f.ID), nil)
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, f)
}

func faqID(r *http.Request) (int, bool) {
	n, err := strconv.Atoi(r.PathValue("faqId"))
	return n, err == nil && n > 0
}

func (s *Server) adminUpdateFaq(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := faqID(r)
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in faqIn
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", "invalid JSON"))
		return
	}
	f, errs := in.validate()
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	f.ID = id
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		if in.SortNo == nil { // keep the current position
			all, err := tx.Faqs()
			if err != nil {
				return err
			}
			for _, x := range all {
				if x.ID == id {
					f.SortNo = x.SortNo
				}
			}
		}
		if err := tx.SaveFaq(&f); err != nil {
			return err
		}
		return s.audit(tx, a, "admin.faq.update", "faq", int64(id), nil)
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, f)
}

func (s *Server) adminDeleteFaq(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := faqID(r)
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		if err := tx.DeleteFaq(id); err != nil {
			return err
		}
		return s.audit(tx, a, "admin.faq.delete", "faq", int64(id), nil)
	})
	if s.fail(w, r, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- moderation queue ----

func (s *Server) adminListReports(w http.ResponseWriter, r *http.Request, a actor) {
	off, lim, errs := pageParams(r)
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var rows []PostReportRow
	var total int
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { rows, total, e = tx.PostReports(off, lim); return })
	if s.fail(w, r, err) {
		return
	}
	pageJSON(w, rows, total, off, lim)
}

func (s *Server) adminDismissReports(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var n int
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		var e error
		if n, e = tx.DismissPostReports(id); e != nil {
			return e
		}
		return s.audit(tx, a, "admin.report.dismiss", "post", id, map[string]any{"reports": n})
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"postId": id, "dismissed": n})
}

func (s *Server) adminRemoveReportedPost(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		if err := tx.AdminDeletePost(id); err != nil {
			return err
		}
		return s.audit(tx, a, "admin.post.remove", "post", id, nil)
	})
	if s.fail(w, r, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
