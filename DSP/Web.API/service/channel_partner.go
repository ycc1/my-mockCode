package service

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"

	"advertiser-api/model"
	"advertiser-api/repository"
)

var ErrInvalidChannelPartner = errors.New("invalid channel partner payload")

type ChannelPartnerService struct {
	partners repository.ChannelPartnerRepository
	mu       sync.Mutex
}

func NewChannelPartnerService(partners repository.ChannelPartnerRepository) *ChannelPartnerService {
	return &ChannelPartnerService{partners: partners}
}
func (s *ChannelPartnerService) Create(request model.CreateChannelPartnerRequest, username string) (model.ChannelPartner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request.Name = strings.TrimSpace(request.Name)
	if !validChannelPartner(request) || !regexp.MustCompile(`^[0-9]{3}$`).MatchString(request.Code) {
		return model.ChannelPartner{}, ErrInvalidChannelPartner
	}
	for _, partner := range s.partners.List() {
		if partner.Code == request.Code {
			return model.ChannelPartner{}, errors.New("渠道商代号已存在")
		}
	}
	return s.partners.Create(request, username), nil
}
func (s *ChannelPartnerService) List() []model.ChannelPartner {
	items := s.partners.List()
	sort.Slice(items, func(i, j int) bool { return items[i].ChannelPartnerID > items[j].ChannelPartnerID })
	return items
}
func (s *ChannelPartnerService) Get(id string) (model.ChannelPartner, bool) {
	return s.partners.Get(id)
}
func (s *ChannelPartnerService) Update(id string, request model.UpdateChannelPartnerRequest, username string) (model.ChannelPartner, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.partners.Get(id)
	if !ok {
		return model.ChannelPartner{}, false, nil
	}
	if request.Name != nil {
		name := strings.TrimSpace(*request.Name)
		request.Name = &name
	}
	// Validate the complete merged record, including partial changes to "other" values.
	base, _ := json.Marshal(current)
	var merged map[string]any
	_ = json.Unmarshal(base, &merged)
	patch, _ := json.Marshal(request)
	var changes map[string]any
	_ = json.Unmarshal(patch, &changes)
	for key, value := range changes {
		if value != nil {
			merged[key] = value
		}
	}
	data, _ := json.Marshal(merged)
	var candidate model.CreateChannelPartnerRequest
	_ = json.Unmarshal(data, &candidate)
	if !validChannelPartner(candidate) || !regexp.MustCompile(`^[0-9]{3}$`).MatchString(candidate.Code) {
		return model.ChannelPartner{}, false, ErrInvalidChannelPartner
	}
	for _, partner := range s.partners.List() {
		if partner.ChannelPartnerID != id && partner.Code == candidate.Code {
			return model.ChannelPartner{}, false, errors.New("渠道商代号已存在")
		}
	}
	partner, ok := s.partners.Update(id, request, username)
	return partner, ok, nil
}
func (s *ChannelPartnerService) Delete(id string) bool { return s.partners.Delete(id) }

func validChoice(field, value, other string) bool {
	if value == "" {
		return false
	}
	allowed := false
	for _, option := range model.ChannelPartnerOptions()[field] {
		if option == value {
			allowed = true
			break
		}
	}
	if !allowed {
		return false
	}
	if value == "其他" {
		return strings.TrimSpace(other) != ""
	}
	return strings.TrimSpace(other) == ""
}

func validChoices(field string, values []string, other string) bool {
	if len(values) == 0 {
		return false
	}
	seen := make(map[string]bool)
	for _, value := range values {
		if seen[value] {
			return false
		}
		seen[value] = true
		detail := ""
		if value == "其他" {
			detail = other
		}
		if !validChoice(field, value, detail) {
			return false
		}
	}
	return seen["其他"] || strings.TrimSpace(other) == ""
}

func validChannelPartner(request model.CreateChannelPartnerRequest) bool {
	return strings.TrimSpace(request.Name) != "" && validChoice("partner_type", request.PartnerType, request.PartnerTypeOther) && validChoice("service_category", request.ServiceCategory, request.ServiceCategoryOther) && validChoice("traffic_model", request.TrafficModel, request.TrafficModelOther) && validChoice("billing_model", request.BillingModel, request.BillingModelOther) && validChoice("primary_channel", request.PrimaryChannel, request.PrimaryChannelOther) && validChoice("secondary_channel", request.SecondaryChannel, request.SecondaryChannelOther) && validChoice("market_contract", request.MarketContract, request.MarketContractOther) && validChoices("delivery_package", request.DeliveryPackage, request.DeliveryPackageOther) && validChoices("data_system", request.DataSystem, request.DataSystemOther)
}

func validChannelPartnerUpdate(request model.UpdateChannelPartnerRequest) bool {
	if request.PartnerType != nil && !validChoice("partner_type", *request.PartnerType, valueOrEmpty(request.PartnerTypeOther)) {
		return false
	}
	if request.ServiceCategory != nil && !validChoice("service_category", *request.ServiceCategory, valueOrEmpty(request.ServiceCategoryOther)) {
		return false
	}
	if request.TrafficModel != nil && !validChoice("traffic_model", *request.TrafficModel, valueOrEmpty(request.TrafficModelOther)) {
		return false
	}
	if request.BillingModel != nil && !validChoice("billing_model", *request.BillingModel, valueOrEmpty(request.BillingModelOther)) {
		return false
	}
	if request.PrimaryChannel != nil && !validChoice("primary_channel", *request.PrimaryChannel, valueOrEmpty(request.PrimaryChannelOther)) {
		return false
	}
	if request.SecondaryChannel != nil && !validChoice("secondary_channel", *request.SecondaryChannel, valueOrEmpty(request.SecondaryChannelOther)) {
		return false
	}
	if request.MarketContract != nil && !validChoice("market_contract", *request.MarketContract, valueOrEmpty(request.MarketContractOther)) {
		return false
	}
	if request.DeliveryPackage != nil && !validChoices("delivery_package", *request.DeliveryPackage, valueOrEmpty(request.DeliveryPackageOther)) {
		return false
	}
	if request.DataSystem != nil && !validChoices("data_system", *request.DataSystem, valueOrEmpty(request.DataSystemOther)) {
		return false
	}
	return request.Name == nil || strings.TrimSpace(*request.Name) != ""
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
