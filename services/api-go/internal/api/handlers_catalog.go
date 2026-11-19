package api

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

func (s *Server) listDestinations(w http.ResponseWriter, r *http.Request) {
	var items []Destination
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.Destinations(); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

// publicListing resolves a listing that travelers may see (published, owner not suspended).
func (s *Server) publicListing(ctx context.Context, id int64) (*listingRow, error) {
	var l *listingRow
	err := s.repo.Tx(ctx, func(tx Tx) (e error) { l, e = tx.ListingByID(id, true); return })
	return l, err
}

func (s *Server) listFaqs(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, 200, map[string]any{"items": seedFaqs})
}

func (s *Server) searchListings(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	typ := r.URL.Query().Get("type")
	var errs []httpx.FieldError
	switch typ {
	case "", "hotel", "restaurant", "flight", "attraction", "trip":
	default:
		errs = append(errs, httpx.FieldError{Field: "type", Message: "must be hotel|restaurant|flight|attraction|trip"})
	}
	var destID int64
	if v := r.URL.Query().Get("destinationId"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			errs = append(errs, httpx.FieldError{Field: "destinationId", Message: "must be an integer"})
		}
		destID = n
	}
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 50 {
			errs = append(errs, httpx.FieldError{Field: "limit", Message: "must be 1..50"})
		} else {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("cursor"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			errs = append(errs, httpx.FieldError{Field: "cursor", Message: "invalid cursor"})
		} else {
			offset = n
		}
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var rows []listingRow
	if err := s.repo.Tx(r.Context(), func(tx Tx) (e error) {
		rows, e = tx.Listings(ListingFilter{PublicOnly: true, Q: q, Type: typ, DestinationID: destID})
		return
	}); s.fail(w, r, err) {
		return
	}
	var all []ListingSummary
	for _, l := range rows {
		all = append(all, l.ListingSummary)
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].Rating > all[j].Rating })
	items := []ListingSummary{}
	next := ""
	if offset < len(all) {
		end := offset + limit
		if end < len(all) {
			next = strconv.Itoa(end)
		} else {
			end = len(all)
		}
		items = all[offset:end]
	}
	out := map[string]any{"items": items}
	if next != "" {
		out["nextCursor"] = next
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) getListing(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "listingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	l, err := s.publicListing(r.Context(), id)
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, l.ListingDetail)
}
