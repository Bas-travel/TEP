package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

var destSlugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// adminDestinationRoutes: destination master data (admin-only, audited).
func (s *Server) adminDestinationRoutes(m *http.ServeMux) {
	p := s.portal(RoleAdmin)
	m.HandleFunc("GET /admin/v1/destinations", p(s.adminListDestinations))
	m.HandleFunc("POST /admin/v1/destinations", p(s.adminCreateDestination))
	m.HandleFunc("PUT /admin/v1/destinations/{destId}", p(s.adminUpdateDestination))
	m.HandleFunc("DELETE /admin/v1/destinations/{destId}", p(s.adminDeleteDestination))
}

type destinationIn struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Subtitle    string `json:"subtitle"`
	GradientCSS string `json:"gradientCss"`
}

func (in destinationIn) validate() (Destination, []httpx.FieldError) {
	var errs []httpx.FieldError
	d := Destination{Slug: strings.ToLower(strings.TrimSpace(in.Slug)), Name: strings.TrimSpace(in.Name),
		Subtitle: strings.TrimSpace(in.Subtitle), GradientCSS: strings.TrimSpace(in.GradientCSS)}
	if len(d.Slug) < 2 || len(d.Slug) > 80 || !destSlugRe.MatchString(d.Slug) {
		errs = append(errs, fe("slug", "2-80 characters: a-z, 0-9 and single hyphens"))
	}
	if n := utf8.RuneCountInString(d.Name); n < 2 || n > 120 {
		errs = append(errs, fe("name", "2-120 characters"))
	}
	if utf8.RuneCountInString(d.Subtitle) > 500 {
		errs = append(errs, fe("subtitle", "at most 500 characters"))
	}
	// the value is rendered by clients as a gradient; keep it to a plain CSS gradient function
	if len(d.GradientCSS) > 255 || (d.GradientCSS != "" && (!strings.Contains(d.GradientCSS, "gradient(") || strings.ContainsAny(d.GradientCSS, ";{}<>\\"))) {
		errs = append(errs, fe("gradientCss", "a CSS gradient such as linear-gradient(135deg,#0ea5e9,#22d3ee)"))
	}
	return d, errs
}

func destID(r *http.Request) (int64, bool) {
	n, err := strconv.ParseInt(r.PathValue("destId"), 10, 64)
	return n, err == nil && n > 0
}

func (s *Server) adminListDestinations(w http.ResponseWriter, r *http.Request, a actor) {
	var items []Destination
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.Destinations(); return })) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items, "total": len(items), "offset": 0, "limit": len(items)})
}

func (s *Server) saveDestination(w http.ResponseWriter, r *http.Request, a actor, id int64) {
	var in destinationIn
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", "invalid JSON"))
		return
	}
	d, errs := in.validate()
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	d.ID = id
	action, status := "admin.destination.create", 201
	if id != 0 {
		action, status = "admin.destination.update", 200
	}
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		if err := tx.SaveDestination(&d); err != nil {
			if err == ErrDuplicate {
				return conflict("a destination with this slug already exists")
			}
			return err
		}
		return s.audit(tx, a, action, "destination", d.ID, map[string]any{"slug": d.Slug})
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, status, d)
}

func (s *Server) adminCreateDestination(w http.ResponseWriter, r *http.Request, a actor) {
	s.saveDestination(w, r, a, 0)
}

func (s *Server) adminUpdateDestination(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := destID(r)
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	s.saveDestination(w, r, a, id)
}

func (s *Server) adminDeleteDestination(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := destID(r)
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		if err := tx.DeleteDestination(id); err != nil {
			if err == ErrDuplicate {
				return conflict("listings still belong to this destination; move or remove them first")
			}
			return err
		}
		return s.audit(tx, a, "admin.destination.delete", "destination", id, nil)
	})
	if s.fail(w, r, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
