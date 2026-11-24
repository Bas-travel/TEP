package api

import "time"

// Roles (users.role). The board's "Staff" is `staff`; `moderator` is kept from docs/architecture/security-model.md.
const (
	RoleTraveler  = "traveler"
	RolePartner   = "partner"
	RoleStaff     = "staff"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
)

// Partner statuses.
const (
	PartnerPending   = "pending"
	PartnerApproved  = "approved"
	PartnerRejected  = "rejected"
	PartnerSuspended = "suspended"
)

// DefaultCommissionRate is the percentage the platform keeps from a partner's gross sales until admin changes it.
const DefaultCommissionRate = 12.0

type partnerRec struct {
	ID, UserID     int64
	BusinessName   string
	ContactPhone   string
	Description    string
	Status         string
	CommissionRate float64
	DecisionNote   string
	DecidedBy      int64
	DecidedAt      *time.Time
	CreatedAt      time.Time
}

// ---- filters (offset pagination for portal lists, per docs/architecture/README.md) ----

type UserFilter struct {
	Q, Role, Status string
	Offset, Limit   int
}

type PartnerFilter struct {
	Q, Status     string
	Offset, Limit int
}

type ListingFilter struct {
	Q, Type       string
	DestinationID int64
	PartnerID     int64
	Status        string // exact status (partner/admin views)
	PublicOnly    bool   // published and owned by nobody or by an approved partner
}

type BookingFilter struct {
	Status        string
	Statuses      []string
	Q             string // booking number or title, case-insensitive substring
	PartnerID     int64  // bookings with at least one line of this partner's listings
	Checked       *bool  // staff verification: true = has a check, false = none yet
	Since         time.Time
	Offset, Limit int // Limit 0 = no paging
}

type TicketFilter struct {
	Status, Q     string
	UserID        int64
	Since         time.Time // activity (created / first reply / resolved) at or after
	Offset, Limit int
}

type AuditFilter struct {
	Action        string // prefix match, e.g. "admin."
	ActorID       int64
	Offset, Limit int
}

// ---- records ----

type bookingCheck struct {
	ID, BookingID, StaffID int64
	StaffName              string
	Result                 string // verified | flagged
	Note                   string
	CreatedAt              time.Time
}

type ticketRec struct {
	SupportTicket
	UserID         int64
	ReplyCount     int
	FirstRepliedAt *time.Time
	ResolvedAt     *time.Time
}

type auditRec struct {
	ID        int64
	ActorID   int64
	ActorName string
	Action    string
	Entity    string
	EntityID  string
	Meta      map[string]any
	CreatedAt time.Time
}

type PlatformStats struct {
	Users             int
	UsersSuspended    int
	ByRole            map[string]int // active users per role
	Partners          map[string]int // per partner status
	ListingsPublished int
	BookingsSince     int // bookings created since the given time
	BookingsTotal     int
	Revenue           float64 // sum of totals of confirmed + completed bookings
}

// ---- wire models of the portals ----

type PartnerView struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"userId"`
	BusinessName   string     `json:"businessName"`
	ContactPhone   string     `json:"contactPhone,omitempty"`
	Description    string     `json:"description,omitempty"`
	Status         string     `json:"status"`
	CommissionRate float64    `json:"commissionRate"`
	DecisionNote   string     `json:"decisionNote,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	DecidedAt      *time.Time `json:"decidedAt,omitempty"`
	OwnerName      string     `json:"ownerName,omitempty"`
	OwnerEmail     string     `json:"ownerEmail,omitempty"`
	ListingCount   int        `json:"listingCount"`
}

type ListingAdminView struct {
	ListingDetail
	DestinationID int64  `json:"destinationId"`
	PartnerID     int64  `json:"partnerId,omitempty"`
	Status        string `json:"status"`
}

type PartnerBookingView struct {
	ID            int64     `json:"id"`
	BookingNumber string    `json:"bookingNumber"`
	Type          string    `json:"type"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	Date          time.Time `json:"date"`
	StartDate     string    `json:"startDate,omitempty"`
	EndDate       string    `json:"endDate,omitempty"`
	Guests        int       `json:"guests"`
	TravelerName  string    `json:"travelerName"`
	Gross         float64   `json:"gross"`
	Commission    float64   `json:"commission"`
	Net           float64   `json:"net"`
	CanAct        bool      `json:"canAct"` // every line of the booking belongs to this partner
	CancelReason  string    `json:"cancelReason,omitempty"`
}

type EarningsMonth struct {
	Month      string  `json:"month"` // YYYY-MM
	Bookings   int     `json:"bookings"`
	Gross      float64 `json:"gross"`
	Commission float64 `json:"commission"`
	Net        float64 `json:"net"`
}

type Earnings struct {
	CommissionRate float64         `json:"commissionRate"`
	Gross          float64         `json:"gross"`
	Commission     float64         `json:"commission"`
	Net            float64         `json:"net"`
	Months         []EarningsMonth `json:"months"`
}

type AdminUserView struct {
	ID        int64     `json:"id"`
	Handle    string    `json:"handle"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone,omitempty"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type BookingCheckView struct {
	ID        int64     `json:"id"`
	StaffName string    `json:"staffName"`
	Result    string    `json:"result"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type StaffBookingView struct {
	BookingDetail
	TravelerName  string             `json:"travelerName"`
	TravelerEmail string             `json:"travelerEmail"`
	Checks        []BookingCheckView `json:"checks"`
	Verification  string             `json:"verification"` // none | verified | flagged (latest check)
}

type StaffTicketView struct {
	SupportTicket
	TravelerName  string `json:"travelerName"`
	TravelerEmail string `json:"travelerEmail"`
	ReplyCount    int    `json:"replyCount"`
}

type AuditView struct {
	ID        int64          `json:"id"`
	ActorID   int64          `json:"actorId"`
	ActorName string         `json:"actorName"`
	Action    string         `json:"action"`
	Entity    string         `json:"entity"`
	EntityID  string         `json:"entityId"`
	Meta      map[string]any `json:"meta,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}
