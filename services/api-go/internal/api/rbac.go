package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/auth"
	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

// ictZone is the business time zone: "today" and month boundaries of the portals are Asia/Bangkok.
var ictZone = time.FixedZone("ICT", 7*3600)

// actor is the authenticated portal user, loaded from the database on every request (never trusted from the JWT).
type actor struct {
	ID      int64
	Role    string
	Name    string
	Partner *partnerRec // set for /partner/v1 only
}

type portalHandler func(w http.ResponseWriter, r *http.Request, a actor)

// portal guards a handler: valid token, user still active, role (read from the DB) in roles.
// Partner routes (roles == RolePartner) additionally require an approved partner record.
func (s *Server) portal(roles ...string) func(portalHandler) http.HandlerFunc {
	return func(h portalHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			hdr := r.Header.Get("Authorization")
			if !strings.HasPrefix(hdr, "Bearer ") {
				httpx.Unauthorized(w, r, "missing bearer token")
				return
			}
			c, err := auth.ParseJWT(s.cfg.JWTSecret, strings.TrimPrefix(hdr, "Bearer "), s.now())
			if err != nil {
				httpx.Unauthorized(w, r, "invalid or expired access token")
				return
			}
			var a actor
			err = s.repo.Tx(r.Context(), func(tx Tx) error {
				u, e := tx.UserByID(c.Sub, false)
				if e != nil {
					return e
				}
				a = actor{ID: u.ID, Role: u.Role, Name: memberName(u.Profile)}
				allowed := false
				for _, ro := range roles {
					if ro == u.Role {
						allowed = true
					}
				}
				if !allowed {
					return forbidden("this area requires role " + strings.Join(roles, " or "))
				}
				if u.Role == RolePartner {
					p, e := tx.PartnerByUser(u.ID, false)
					if e != nil {
						return forbidden("no partner profile")
					}
					if p.Status != PartnerApproved {
						return forbidden("partner account is " + p.Status)
					}
					a.Partner = p
				}
				return nil
			})
			if errors.Is(err, ErrNotFound) {
				httpx.Unauthorized(w, r, "unknown user")
				return
			}
			if s.fail(w, r, err) {
				return
			}
			h(w, r, a)
		}
	}
}

// audit records a portal mutation inside the caller's transaction.
func (s *Server) audit(tx Tx, a actor, action, entity string, id int64, meta map[string]any) error {
	return tx.AddAudit(auditRec{ActorID: a.ID, Action: action, Entity: entity, EntityID: strconv.FormatInt(id, 10), Meta: meta, CreatedAt: s.now()})
}

// pageParams reads offset/limit (limit default 20, max 100).
func pageParams(r *http.Request) (offset, limit int, errs []httpx.FieldError) {
	limit = 20
	q := r.URL.Query()
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			errs = append(errs, fe("limit", "must be 1..100"))
		} else {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			errs = append(errs, fe("offset", "must be >= 0"))
		} else {
			offset = n
		}
	}
	return
}

func pageJSON(w http.ResponseWriter, items any, total, offset, limit int) {
	httpx.WriteJSON(w, 200, map[string]any{"items": items, "total": total, "offset": offset, "limit": limit})
}

func oneOf(v string, set ...string) bool {
	for _, s := range set {
		if v == s {
			return true
		}
	}
	return false
}

// refundBooking cancels a pending/confirmed booking and returns the money to the traveler's wallet when it was paid.
// The caller holds the booking lock and has checked the status.
func (s *Server) refundBooking(tx Tx, b *bookingRec, reason string) error {
	refund := 0.0
	if b.Status == "confirmed" {
		refund = b.Amount
		if _, e := tx.WalletAdjust(b.UserID, refund, "คืนเงินการจอง "+b.BookingNumber, s.now()); e != nil {
			return e
		}
		if e := tx.MarkPaymentRefunded(b.ID); e != nil {
			return e
		}
	}
	b.Status, b.CancelReason, b.RefundAmount = "cancelled", reason, &refund
	return tx.SaveBooking(b)
}

func atoi64(v string) (int64, bool) {
	n, err := strconv.ParseInt(v, 10, 64)
	return n, err == nil && n > 0
}
