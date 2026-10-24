package api

import "time"

type Destination struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Subtitle    string `json:"subtitle"`
	GradientCSS string `json:"gradientCss"`
}

type ListingSummary struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"reviewCount"`
	PriceFrom   float64 `json:"priceFrom"`
	ImageURL    string  `json:"imageUrl,omitempty"`
}

type ListingDetail struct {
	ListingSummary
	Description string   `json:"description"`
	Amenities   []string `json:"amenities"`
	MapQuery    string   `json:"mapQuery"`
}

type listingRow struct {
	ListingDetail
	DestinationID int64
}

type Faq struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type AuthTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

type RegisterRequest struct {
	Email               string   `json:"email"`
	Password            string   `json:"password"`
	Phone               string   `json:"phone,omitempty"`
	FirstName           string   `json:"firstName"`
	LastName            string   `json:"lastName"`
	Gender              string   `json:"gender,omitempty"`
	DateOfBirth         string   `json:"dateOfBirth,omitempty"`
	Nationality         string   `json:"nationality,omitempty"`
	Language            string   `json:"language,omitempty"`
	BudgetLevel         string   `json:"budgetLevel,omitempty"`
	TravelStyles        []string `json:"travelStyles,omitempty"`
	FavoriteDestination string   `json:"favoriteDestination,omitempty"`
	TravelWith          string   `json:"travelWith,omitempty"`
}

type UserProfile struct {
	ID                  int64    `json:"id"`
	Level               string   `json:"level"`
	Email               string   `json:"email"`
	Phone               string   `json:"phone,omitempty"`
	FirstName           string   `json:"firstName"`
	LastName            string   `json:"lastName"`
	Gender              string   `json:"gender,omitempty"`
	DateOfBirth         string   `json:"dateOfBirth,omitempty"`
	Nationality         string   `json:"nationality,omitempty"`
	Language            string   `json:"language,omitempty"`
	BudgetLevel         string   `json:"budgetLevel,omitempty"`
	TravelStyles        []string `json:"travelStyles,omitempty"`
	FavoriteDestination string   `json:"favoriteDestination,omitempty"`
	TravelWith          string   `json:"travelWith,omitempty"`
	AvatarURL           string   `json:"avatarUrl,omitempty"`
}

type BookingSummary struct {
	ID            int64     `json:"id"`
	BookingNumber string    `json:"bookingNumber"`
	Type          string    `json:"type"`
	Title         string    `json:"title"`
	Subtitle      string    `json:"subtitle"`
	Amount        float64   `json:"amount"`
	Method        string    `json:"method"`
	Date          time.Time `json:"date"`
	Status        string    `json:"status"`
}

type BookingDetail struct {
	BookingSummary
	Subtotal     float64  `json:"subtotal"`
	ServiceFee   float64  `json:"serviceFee"`
	CancelReason string   `json:"cancelReason,omitempty"`
	RefundAmount *float64 `json:"refundAmount,omitempty"`
}

type CreateBookingRequest struct {
	Lines []struct {
		ListingID int64 `json:"listingId"`
		Quantity  int   `json:"quantity"`
	} `json:"lines"`
	CouponCode  string `json:"couponCode,omitempty"`
	TripPackage *struct {
		Title    string  `json:"title"`
		Subtitle string  `json:"subtitle"`
		Price    float64 `json:"price"`
	} `json:"tripPackage,omitempty"`
}

type PaymentIntent struct {
	ID             string  `json:"id"`
	Status         string  `json:"status"`
	Method         string  `json:"method"`
	Amount         float64 `json:"amount"`
	PromptPayQrURL string  `json:"promptPayQrUrl,omitempty"`
	ClientSecret   string  `json:"clientSecret,omitempty"`
}

type WalletTx struct {
	ID        int64     `json:"id"`
	Direction string    `json:"direction"`
	Label     string    `json:"label"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

type Wallet struct {
	Balance      float64    `json:"balance"`
	Currency     string     `json:"currency"`
	Transactions []WalletTx `json:"transactions"`
}
