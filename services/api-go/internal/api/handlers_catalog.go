package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

func (s *Server) listDestinations(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, 200, map[string]any{"items": seedDestinations})
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
	var all []ListingSummary
	for _, l := range seedListings {
		if typ != "" && l.Type != typ {
			continue
		}
		if destID != 0 && l.DestinationID != destID {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(l.Name+" "+l.Location), q) {
			continue
		}
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

func findListing(id int64) (listingRow, bool) {
	for _, l := range seedListings {
		if l.ID == id {
			return l, true
		}
	}
	return listingRow{}, false
}

func (s *Server) getListing(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "listingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	l, found := findListing(id)
	if !found {
		httpx.NotFound(w, r)
		return
	}
	httpx.WriteJSON(w, 200, l.ListingDetail)
}
