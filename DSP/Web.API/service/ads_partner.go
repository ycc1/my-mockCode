package service

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"advertiser-api/model"
	"advertiser-api/repository"
)

var ErrInvalidAdsPartner = errors.New("invalid ads partner payload")

type AdsPartnerService struct {
	partners repository.AdsPartnerRepository
	mu       sync.Mutex
}

func NewAdsPartnerService(partners repository.AdsPartnerRepository) *AdsPartnerService {
	return &AdsPartnerService{partners: partners}
}

func validAdsPartner(request model.CreateAdsPartnerRequest) bool {
	lengths := map[string]int{"MD5": 32, "SHA312": 78, "AES": 32, "DES": 8}
	expected, ok := lengths[request.SecurityType]
	merchantID := strings.TrimSpace(request.AdsMerchantID)
	return strings.TrimSpace(request.Name) != "" && merchantID != "" && utf8.RuneCountInString(merchantID) <= 20 && ok && len(request.APIKey) == expected
}

func (s *AdsPartnerService) Create(request model.CreateAdsPartnerRequest, username string) (model.AdsPartner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request.Name, request.AdsMerchantID = strings.TrimSpace(request.Name), strings.TrimSpace(request.AdsMerchantID)
	if !validAdsPartner(request) {
		return model.AdsPartner{}, ErrInvalidAdsPartner
	}
	for _, item := range s.partners.List() {
		if item.AdsMerchantID == request.AdsMerchantID {
			return model.AdsPartner{}, errors.New("广告商户 ID 已存在")
		}
	}
	return s.partners.Create(request, username), nil
}

func (s *AdsPartnerService) List() []model.AdsPartner {
	items := s.partners.List()
	sort.Slice(items, func(i, j int) bool { return items[i].CreateTime.After(items[j].CreateTime) })
	return items
}
func (s *AdsPartnerService) Get(id string) (model.AdsPartner, bool) { return s.partners.Get(id) }
func (s *AdsPartnerService) Delete(id string) bool                  { return s.partners.Delete(id) }

func (s *AdsPartnerService) Update(id string, request model.UpdateAdsPartnerRequest, username string) (model.AdsPartner, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.partners.Get(id)
	if !ok {
		return model.AdsPartner{}, false, nil
	}
	base, _ := json.Marshal(current)
	patch, _ := json.Marshal(request)
	var merged, changes map[string]any
	_ = json.Unmarshal(base, &merged)
	_ = json.Unmarshal(patch, &changes)
	for key, value := range changes {
		if value != nil {
			merged[key] = value
		}
	}
	data, _ := json.Marshal(merged)
	var candidate model.CreateAdsPartnerRequest
	_ = json.Unmarshal(data, &candidate)
	candidate.Name, candidate.AdsMerchantID = strings.TrimSpace(candidate.Name), strings.TrimSpace(candidate.AdsMerchantID)
	if !validAdsPartner(candidate) {
		return model.AdsPartner{}, false, ErrInvalidAdsPartner
	}
	for _, item := range s.partners.List() {
		if item.ID != id && item.AdsMerchantID == candidate.AdsMerchantID {
			return model.AdsPartner{}, false, errors.New("广告商户 ID 已存在")
		}
	}
	request.Name, request.AdsMerchantID = &candidate.Name, &candidate.AdsMerchantID
	partner, ok := s.partners.Update(id, request, username)
	return partner, ok, nil
}
