package repository

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"advertiser-api/model"
)

type MemoryAdsPartnerRepository struct {
	mu       sync.RWMutex
	partners map[string]model.AdsPartner
	seq      uint64
}

func NewMemoryAdsPartnerRepository() *MemoryAdsPartnerRepository {
	return &MemoryAdsPartnerRepository{partners: make(map[string]model.AdsPartner)}
}

func (r *MemoryAdsPartnerRepository) Create(request model.CreateAdsPartnerRequest, username string) model.AdsPartner {
	now := time.Now().UTC()
	partner := model.AdsPartner{ID: fmt.Sprintf("AP_%06d", atomic.AddUint64(&r.seq, 1)), Name: request.Name, AdsMerchantID: request.AdsMerchantID, APIKey: request.APIKey, SecurityType: request.SecurityType, CreateTime: now, UpdateTime: now, CreateBy: username, UpdateBy: username}
	r.mu.Lock()
	r.partners[partner.ID] = partner
	r.mu.Unlock()
	return partner
}

func (r *MemoryAdsPartnerRepository) List() []model.AdsPartner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.AdsPartner, 0, len(r.partners))
	for _, partner := range r.partners {
		result = append(result, partner)
	}
	return result
}

func (r *MemoryAdsPartnerRepository) Get(id string) (model.AdsPartner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	partner, ok := r.partners[id]
	return partner, ok
}

func (r *MemoryAdsPartnerRepository) Update(id string, request model.UpdateAdsPartnerRequest, username string) (model.AdsPartner, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	partner, ok := r.partners[id]
	if !ok {
		return model.AdsPartner{}, false
	}
	if request.Name != nil {
		partner.Name = *request.Name
	}
	if request.AdsMerchantID != nil {
		partner.AdsMerchantID = *request.AdsMerchantID
	}
	if request.APIKey != nil {
		partner.APIKey = *request.APIKey
	}
	if request.SecurityType != nil {
		partner.SecurityType = *request.SecurityType
	}
	partner.UpdateTime, partner.UpdateBy = time.Now().UTC(), username
	r.partners[id] = partner
	return partner, true
}

func (r *MemoryAdsPartnerRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.partners[id]; !ok {
		return false
	}
	delete(r.partners, id)
	return true
}
