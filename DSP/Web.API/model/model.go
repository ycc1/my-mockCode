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

type ChannelPartner struct {
	Code                  string    `json:"code"`
	ChannelPartnerID      string    `json:"channel_partner_id"`
	Name                  string    `json:"name"`
	MerchantID            string    `json:"merchant_id"`
	APIKey                string    `json:"api_key"`
	SecurityType          string    `json:"security_type"`
	PartnerType           string    `json:"partner_type"`
	PartnerTypeOther      string    `json:"partner_type_other,omitempty"`
	ServiceCategory       string    `json:"service_category"`
	ServiceCategoryOther  string    `json:"service_category_other,omitempty"`
	TrafficModel          string    `json:"traffic_model"`
	TrafficModelOther     string    `json:"traffic_model_other,omitempty"`
	BillingModel          string    `json:"billing_model"`
	BillingModelOther     string    `json:"billing_model_other,omitempty"`
	PrimaryChannel        string    `json:"primary_channel"`
	PrimaryChannelOther   string    `json:"primary_channel_other,omitempty"`
	SecondaryChannel      string    `json:"secondary_channel"`
	SecondaryChannelOther string    `json:"secondary_channel_other,omitempty"`
	MarketContract        string    `json:"market_contract"`
	MarketContractOther   string    `json:"market_contract_other,omitempty"`
	DeliveryPackage       []string  `json:"delivery_package"`
	DeliveryPackageOther  string    `json:"delivery_package_other,omitempty"`
	DataSystem            []string  `json:"data_system"`
	DataSystemOther       string    `json:"data_system_other,omitempty"`
	Status                bool      `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	ModifiedBy            string    `json:"modified_by"`
}

type CreateChannelPartnerRequest struct {
	Code                  string   `json:"code"`
	Name                  string   `json:"name"`
	MerchantID            string   `json:"merchant_id"`
	APIKey                string   `json:"api_key"`
	SecurityType          string   `json:"security_type"`
	PartnerType           string   `json:"partner_type"`
	PartnerTypeOther      string   `json:"partner_type_other"`
	ServiceCategory       string   `json:"service_category"`
	ServiceCategoryOther  string   `json:"service_category_other"`
	TrafficModel          string   `json:"traffic_model"`
	TrafficModelOther     string   `json:"traffic_model_other"`
	BillingModel          string   `json:"billing_model"`
	BillingModelOther     string   `json:"billing_model_other"`
	PrimaryChannel        string   `json:"primary_channel"`
	PrimaryChannelOther   string   `json:"primary_channel_other"`
	SecondaryChannel      string   `json:"secondary_channel"`
	SecondaryChannelOther string   `json:"secondary_channel_other"`
	MarketContract        string   `json:"market_contract"`
	MarketContractOther   string   `json:"market_contract_other"`
	DeliveryPackage       []string `json:"delivery_package"`
	DeliveryPackageOther  string   `json:"delivery_package_other"`
	DataSystem            []string `json:"data_system"`
	DataSystemOther       string   `json:"data_system_other"`
	Status                bool     `json:"status"`
}

type UpdateChannelPartnerRequest struct {
	Code                  *string   `json:"code"`
	Name                  *string   `json:"name"`
	MerchantID            *string   `json:"merchant_id"`
	APIKey                *string   `json:"api_key"`
	SecurityType          *string   `json:"security_type"`
	PartnerType           *string   `json:"partner_type"`
	PartnerTypeOther      *string   `json:"partner_type_other"`
	ServiceCategory       *string   `json:"service_category"`
	ServiceCategoryOther  *string   `json:"service_category_other"`
	TrafficModel          *string   `json:"traffic_model"`
	TrafficModelOther     *string   `json:"traffic_model_other"`
	BillingModel          *string   `json:"billing_model"`
	BillingModelOther     *string   `json:"billing_model_other"`
	PrimaryChannel        *string   `json:"primary_channel"`
	PrimaryChannelOther   *string   `json:"primary_channel_other"`
	SecondaryChannel      *string   `json:"secondary_channel"`
	SecondaryChannelOther *string   `json:"secondary_channel_other"`
	MarketContract        *string   `json:"market_contract"`
	MarketContractOther   *string   `json:"market_contract_other"`
	DeliveryPackage       *[]string `json:"delivery_package"`
	DeliveryPackageOther  *string   `json:"delivery_package_other"`
	DataSystem            *[]string `json:"data_system"`
	DataSystemOther       *string   `json:"data_system_other"`
	Status                *bool     `json:"status"`
}

type AdsPartner struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	AdsMerchantID string    `json:"ads_merchant_id"`
	APIKey        string    `json:"api_key"`
	SecurityType  string    `json:"security_type"`
	CreateTime    time.Time `json:"create_time"`
	UpdateTime    time.Time `json:"update_time"`
	CreateBy      string    `json:"create_by"`
	UpdateBy      string    `json:"update_by"`
}

type CreateAdsPartnerRequest struct {
	Name          string `json:"name"`
	AdsMerchantID string `json:"ads_merchant_id"`
	APIKey        string `json:"api_key"`
	SecurityType  string `json:"security_type"`
}

type UpdateAdsPartnerRequest struct {
	Name          *string `json:"name"`
	AdsMerchantID *string `json:"ads_merchant_id"`
	APIKey        *string `json:"api_key"`
	SecurityType  *string `json:"security_type"`
}
