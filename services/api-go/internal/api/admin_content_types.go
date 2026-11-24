package api

import "time"

// CouponAdmin is a coupon as managed by admins (with usage counters).
type CouponAdmin struct {
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	Detail      string  `json:"detail"`
	PercentOff  int     `json:"percentOff"` // 0 => fixed amount (maxDiscount)
	MinSpend    float64 `json:"minSpend"`
	MaxDiscount float64 `json:"maxDiscount"`
	ExpiresOn   string  `json:"expiresOn"` // YYYY-MM-DD, empty = never
	Granted     int     `json:"granted"`   // wallets holding it
	Expired     bool    `json:"expired"`
}

// FaqAdmin is an FAQ entry with its sort position.
type FaqAdmin struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	SortNo   int    `json:"sortNo"`
}

// PostReportRow is one reported community post (all its reports grouped).
type PostReportRow struct {
	PostID     int64     `json:"postId"`
	PostText   string    `json:"postText"`
	Author     string    `json:"author"` // author handle
	Reports    int       `json:"reports"`
	Reasons    []string  `json:"reasons"`
	LastReport time.Time `json:"lastReportAt"`
}
