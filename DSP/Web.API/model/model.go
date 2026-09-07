package model

import "time"

type Offer struct {
	OfferID             string     `json:"offer_id"`
	Name                string     `json:"name"`
	AdvertiserID        string     `json:"advertiser_id,omitempty"`
	Status              string     `json:"status"`
	Payout              Payout     `json:"payout"`
	Targeting           Targeting  `json:"targeting,omitempty"`
	Caps                Caps       `json:"caps"`
	LandingPageURL      string     `json:"landing_page_url,omitempty"`
	TrackingURLTemplate string     `json:"tracking_url_template,omitempty"`
	TodayConversions    int        `json:"today_conversions"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	ArchivedAt          *time.Time `json:"archived_at,omitempty"`
}

type Payout struct {
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Targeting struct {
	Countries    []string `json:"countries,omitempty"`
	OS           string   `json:"os,omitempty"`
	MinOSVersion string   `json:"min_os_version,omitempty"`
}

type Caps struct {
	DailyCap int `json:"daily_cap"`
}

type CreateOfferRequest struct {
	Name                string    `json:"name"`
	AdvertiserID        string    `json:"advertiser_id"`
	Status              string    `json:"status"`
	Payout              Payout    `json:"payout"`
	Targeting           Targeting `json:"targeting"`
	Caps                Caps      `json:"caps"`
	LandingPageURL      string    `json:"landing_page_url"`
	TrackingURLTemplate string    `json:"tracking_url_template"`
}

type UpdateOfferRequest struct {
	Name                *string       `json:"name"`
	Status              *string       `json:"status"`
	Payout              *UpdatePayout `json:"payout"`
	Targeting           *Targeting    `json:"targeting"`
	Caps                *Caps         `json:"caps"`
	LandingPageURL      *string       `json:"landing_page_url"`
	TrackingURLTemplate *string       `json:"tracking_url_template"`
}

type UpdatePayout struct {
	Type     *string  `json:"type"`
	Amount   *float64 `json:"amount"`
	Currency *string  `json:"currency"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Role struct {
	RoleID      string    `json:"role_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Feature struct {
	FeatureID   string    `json:"feature_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateFeatureRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateFeatureRequest struct {
	Code        *string `json:"code"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AssignFeaturesRequest struct {
	FeatureIDs []string `json:"feature_ids"`
}
