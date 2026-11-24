package api

import (
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

// PlatformSetting is one admin-editable platform value (defaults live in code, overrides in platform_settings).
type PlatformSetting struct {
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	Default     string     `json:"default"`
	Type        string     `json:"type"` // number | text
	Description string     `json:"description"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type settingSpec struct {
	Type, Default, Description string
	// validate returns the normalised value or a message.
	validate func(v string) (string, string)
}

const (
	SettingServiceFee = "service_fee_percent"
	SettingBanner     = "maintenance_banner"
	SettingSupport    = "support_email"
)

var settingSpecs = map[string]settingSpec{
	SettingServiceFee: {Type: "number", Default: "3", Description: "Service fee added to every booking, in percent (0-20)",
		validate: func(v string) (string, string) {
			f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil || f < 0 || f > 20 {
				return "", "a number between 0 and 20"
			}
			return strconv.FormatFloat(round2(f), 'f', -1, 64), ""
		}},
	SettingBanner: {Type: "text", Default: "", Description: "Notice shown to travelers (empty = hidden), max 200 characters",
		validate: func(v string) (string, string) {
			v = strings.TrimSpace(v)
			if utf8.RuneCountInString(v) > 200 {
				return "", "at most 200 characters"
			}
			return v, ""
		}},
	SettingSupport: {Type: "text", Default: "", Description: "Support e-mail address shown in the app (empty = hidden)",
		validate: func(v string) (string, string) {
			v = strings.TrimSpace(v)
			if v == "" {
				return "", ""
			}
			a, err := mail.ParseAddress(v)
			if err != nil || a.Address != v || len(v) > 120 {
				return "", "a valid e-mail address (max 120 characters)"
			}
			return v, ""
		}},
}

var settingKeys = []string{SettingServiceFee, SettingBanner, SettingSupport}

// settingValues merges stored overrides over the defaults.
func settingValues(stored map[string]PlatformSetting) []PlatformSetting {
	out := make([]PlatformSetting, 0, len(settingKeys))
	for _, k := range settingKeys {
		sp := settingSpecs[k]
		ps := PlatformSetting{Key: k, Value: sp.Default, Default: sp.Default, Type: sp.Type, Description: sp.Description}
		if st, ok := stored[k]; ok {
			ps.Value, ps.UpdatedAt = st.Value, st.UpdatedAt
		}
		out = append(out, ps)
	}
	return out
}

// serviceFeePercent is read inside the booking transaction so a change applies to the next booking.
func serviceFeePercent(tx Tx) (float64, error) {
	st, err := tx.PlatformSettings()
	if err != nil {
		return 0, err
	}
	v := settingSpecs[SettingServiceFee].Default
	if s, ok := st[SettingServiceFee]; ok {
		v = s.Value
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 3, nil
	}
	return f, nil
}

func (s *Server) platformSettingsRoutes(m *http.ServeMux) {
	p := s.portal(RoleAdmin)
	m.HandleFunc("GET /admin/v1/settings", p(s.adminListSettings))
	m.HandleFunc("PUT /admin/v1/settings/{key}", p(s.adminPutSetting))
	m.HandleFunc("GET /v1/config", s.publicConfig)
}

func (s *Server) adminListSettings(w http.ResponseWriter, r *http.Request, a actor) {
	var st map[string]PlatformSetting
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) (e error) { st, e = tx.PlatformSettings(); return })) {
		return
	}
	items := settingValues(st)
	httpx.WriteJSON(w, 200, map[string]any{"items": items, "total": len(items), "offset": 0, "limit": len(items)})
}

func (s *Server) adminPutSetting(w http.ResponseWriter, r *http.Request, a actor) {
	key := r.PathValue("key")
	sp, ok := settingSpecs[key]
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Value string `json:"value"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", "invalid JSON"))
		return
	}
	val, msg := sp.validate(in.Value)
	if msg != "" {
		httpx.Invalid(w, r, fe("value", msg))
		return
	}
	var out PlatformSetting
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		st, e := tx.PlatformSettings()
		if e != nil {
			return e
		}
		old := settingValues(st)
		var oldVal string
		for _, o := range old {
			if o.Key == key {
				oldVal = o.Value
			}
		}
		now := s.now()
		if e := tx.SetPlatformSetting(key, val, now); e != nil {
			return e
		}
		out = PlatformSetting{Key: key, Value: val, Default: sp.Default, Type: sp.Type, Description: sp.Description, UpdatedAt: &now}
		return s.audit(tx, a, "admin.setting.update", "setting", 0, map[string]any{"key": key, "from": oldVal, "to": val})
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

// publicConfig exposes the traveler-facing settings (no auth).
func (s *Server) publicConfig(w http.ResponseWriter, r *http.Request) {
	var st map[string]PlatformSetting
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) (e error) { st, e = tx.PlatformSettings(); return })) {
		return
	}
	v := map[string]string{}
	for _, p := range settingValues(st) {
		v[p.Key] = p.Value
	}
	fee, _ := strconv.ParseFloat(v[SettingServiceFee], 64)
	httpx.WriteJSON(w, 200, map[string]any{"serviceFeePercent": fee, "maintenanceBanner": v[SettingBanner], "supportEmail": v[SettingSupport]})
}
