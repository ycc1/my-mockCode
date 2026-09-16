package repository

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"advertiser-api/model"
)

type MemoryChannelPartnerRepository struct {
	mu       sync.RWMutex
	partners map[string]model.ChannelPartner
	seq      uint64
}

func NewMemoryChannelPartnerRepository() *MemoryChannelPartnerRepository {
	return &MemoryChannelPartnerRepository{partners: make(map[string]model.ChannelPartner)}
}

func (r *MemoryChannelPartnerRepository) Create(request model.CreateChannelPartnerRequest, username string) model.ChannelPartner {
	now := time.Now().UTC()
	partner := model.ChannelPartner{
		Code:             request.Code,
		ChannelPartnerID: fmt.Sprintf("CP_%06d", atomic.AddUint64(&r.seq, 1)), Name: request.Name,
		PartnerType: request.PartnerType, PartnerTypeOther: request.PartnerTypeOther,
		ServiceCategory: request.ServiceCategory, ServiceCategoryOther: request.ServiceCategoryOther,
		TrafficModel: request.TrafficModel, TrafficModelOther: request.TrafficModelOther,
		BillingModel: request.BillingModel, BillingModelOther: request.BillingModelOther,
		PrimaryChannel: request.PrimaryChannel, PrimaryChannelOther: request.PrimaryChannelOther,
		SecondaryChannel: request.SecondaryChannel, SecondaryChannelOther: request.SecondaryChannelOther,
		MarketContract: request.MarketContract, MarketContractOther: request.MarketContractOther,
		DeliveryPackage: request.DeliveryPackage, DeliveryPackageOther: request.DeliveryPackageOther,
		DataSystem: request.DataSystem, DataSystemOther: request.DataSystemOther,
		Status: request.Status, CreatedAt: now, UpdatedAt: now, ModifiedBy: username,
	}
	r.mu.Lock()
	r.partners[partner.ChannelPartnerID] = partner
	r.mu.Unlock()
	return partner
}

func (r *MemoryChannelPartnerRepository) List() []model.ChannelPartner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.ChannelPartner, 0, len(r.partners))
	for _, partner := range r.partners {
		result = append(result, partner)
	}
	return result
}

func (r *MemoryChannelPartnerRepository) Get(id string) (model.ChannelPartner, bool) {
	r.mu.RLock()
	partner, ok := r.partners[id]
	r.mu.RUnlock()
	return partner, ok
}

func (r *MemoryChannelPartnerRepository) Update(id string, request model.UpdateChannelPartnerRequest, username string) (model.ChannelPartner, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	partner, ok := r.partners[id]
	if !ok {
		return model.ChannelPartner{}, false
	}
	if request.Code != nil {
		partner.Code = *request.Code
	}
	if request.Name != nil {
		partner.Name = *request.Name
	}
	if request.PartnerType != nil {
		partner.PartnerType = *request.PartnerType
	}
	if request.PartnerTypeOther != nil {
		partner.PartnerTypeOther = *request.PartnerTypeOther
	}
	if request.ServiceCategory != nil {
		partner.ServiceCategory = *request.ServiceCategory
	}
	if request.ServiceCategoryOther != nil {
		partner.ServiceCategoryOther = *request.ServiceCategoryOther
	}
	if request.TrafficModel != nil {
		partner.TrafficModel = *request.TrafficModel
	}
	if request.TrafficModelOther != nil {
		partner.TrafficModelOther = *request.TrafficModelOther
	}
	if request.BillingModel != nil {
		partner.BillingModel = *request.BillingModel
	}
	if request.BillingModelOther != nil {
		partner.BillingModelOther = *request.BillingModelOther
	}
	if request.PrimaryChannel != nil {
		partner.PrimaryChannel = *request.PrimaryChannel
	}
	if request.PrimaryChannelOther != nil {
		partner.PrimaryChannelOther = *request.PrimaryChannelOther
	}
	if request.SecondaryChannel != nil {
		partner.SecondaryChannel = *request.SecondaryChannel
	}
	if request.SecondaryChannelOther != nil {
		partner.SecondaryChannelOther = *request.SecondaryChannelOther
	}
	if request.MarketContract != nil {
		partner.MarketContract = *request.MarketContract
	}
	if request.MarketContractOther != nil {
		partner.MarketContractOther = *request.MarketContractOther
	}
	if request.DeliveryPackage != nil {
		partner.DeliveryPackage = *request.DeliveryPackage
	}
	if request.DeliveryPackageOther != nil {
		partner.DeliveryPackageOther = *request.DeliveryPackageOther
	}
	if request.DataSystem != nil {
		partner.DataSystem = *request.DataSystem
	}
	if request.DataSystemOther != nil {
		partner.DataSystemOther = *request.DataSystemOther
	}
	if request.Status != nil {
		partner.Status = *request.Status
	}
	partner.UpdatedAt = time.Now().UTC()
	partner.ModifiedBy = username
	r.partners[id] = partner
	return partner, true
}

func (r *MemoryChannelPartnerRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.partners[id]; !ok {
		return false
	}
	delete(r.partners, id)
	return true
}
