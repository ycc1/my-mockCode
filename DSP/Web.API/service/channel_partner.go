package service

import (
	"errors"
	"strings"

	"advertiser-api/model"
	"advertiser-api/repository"
)

var ErrInvalidChannelPartner = errors.New("invalid channel partner payload")

var channelPartnerOptions = map[string][]string{
	"partner_type": {"外部", "内部", "其他"},
	"service_category": {"DSP", "大媒体采买", "网红/达人推广", "SEO", "SEM", "ASO", "社交媒体", "电子邮件营销", "短信营销", "信息流/原生广告", "其他"},
	"traffic_model": {"cpc", "cpm", "cpa", "cpl", "收入分成", "其他"},
	"billing_model": {"cpc", "cpm", "cpa", "cpl", "收入分成", "其他"},
	"primary_channel": {"Google", "TikTok", "Meta", "X", "YouTube", "原生广告", "程序化广告", "seo", "其他"},
	"secondary_channel": {"Google", "TikTok", "Meta", "X", "YouTube", "原生广告", "程序化广告", "seo", "其他"},
	"market_contract": {"是", "否", "其他"},
	"delivery_package": {"苹果", "H5", "PWA", "白标包体", "其他"},
	"data_system": {"GA4", "appsflyer", "adjust", "自研系统", "其他"},
}

type ChannelPartnerService struct{ partners repository.ChannelPartnerRepository }

func NewChannelPartnerService(partners repository.ChannelPartnerRepository) *ChannelPartnerService { return &ChannelPartnerService{partners: partners} }
func (s *ChannelPartnerService) Create(request model.CreateChannelPartnerRequest, username string) (model.ChannelPartner, error) { if !validChannelPartner(request) { return model.ChannelPartner{}, ErrInvalidChannelPartner }; return s.partners.Create(request, username), nil }
func (s *ChannelPartnerService) List() []model.ChannelPartner { return s.partners.List() }
func (s *ChannelPartnerService) Get(id string) (model.ChannelPartner, bool) { return s.partners.Get(id) }
func (s *ChannelPartnerService) Update(id string, request model.UpdateChannelPartnerRequest, username string) (model.ChannelPartner, bool, error) { if !validChannelPartnerUpdate(request) { return model.ChannelPartner{}, false, ErrInvalidChannelPartner }; partner, ok := s.partners.Update(id, request, username); return partner, ok, nil }
func (s *ChannelPartnerService) Delete(id string) bool { return s.partners.Delete(id) }

func validChoice(field, value, other string) bool {
	if value == "" { return false }
	allowed := false
	for _, option := range channelPartnerOptions[field] { if option == value { allowed = true; break } }
	if !allowed { return false }
	if value == "其他" { return strings.TrimSpace(other) != "" }
	return strings.TrimSpace(other) == ""
}

func validChannelPartner(request model.CreateChannelPartnerRequest) bool {
	return strings.TrimSpace(request.Name) != "" && validChoice("partner_type", request.PartnerType, request.PartnerTypeOther) && validChoice("service_category", request.ServiceCategory, request.ServiceCategoryOther) && validChoice("traffic_model", request.TrafficModel, request.TrafficModelOther) && validChoice("billing_model", request.BillingModel, request.BillingModelOther) && validChoice("primary_channel", request.PrimaryChannel, request.PrimaryChannelOther) && validChoice("secondary_channel", request.SecondaryChannel, request.SecondaryChannelOther) && validChoice("market_contract", request.MarketContract, request.MarketContractOther) && validChoice("delivery_package", request.DeliveryPackage, request.DeliveryPackageOther) && validChoice("data_system", request.DataSystem, request.DataSystemOther)
}

func validChannelPartnerUpdate(request model.UpdateChannelPartnerRequest) bool {
	if request.PartnerType != nil && !validChoice("partner_type", *request.PartnerType, valueOrEmpty(request.PartnerTypeOther)) { return false }
	if request.ServiceCategory != nil && !validChoice("service_category", *request.ServiceCategory, valueOrEmpty(request.ServiceCategoryOther)) { return false }
	if request.TrafficModel != nil && !validChoice("traffic_model", *request.TrafficModel, valueOrEmpty(request.TrafficModelOther)) { return false }
	if request.BillingModel != nil && !validChoice("billing_model", *request.BillingModel, valueOrEmpty(request.BillingModelOther)) { return false }
	if request.PrimaryChannel != nil && !validChoice("primary_channel", *request.PrimaryChannel, valueOrEmpty(request.PrimaryChannelOther)) { return false }
	if request.SecondaryChannel != nil && !validChoice("secondary_channel", *request.SecondaryChannel, valueOrEmpty(request.SecondaryChannelOther)) { return false }
	if request.MarketContract != nil && !validChoice("market_contract", *request.MarketContract, valueOrEmpty(request.MarketContractOther)) { return false }
	if request.DeliveryPackage != nil && !validChoice("delivery_package", *request.DeliveryPackage, valueOrEmpty(request.DeliveryPackageOther)) { return false }
	if request.DataSystem != nil && !validChoice("data_system", *request.DataSystem, valueOrEmpty(request.DataSystemOther)) { return false }
	return request.Name == nil || strings.TrimSpace(*request.Name) != ""
}

func valueOrEmpty(value *string) string { if value == nil { return "" }; return *value }
