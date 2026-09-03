package repository

import (
	"crypto/rand"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"advertiser-api/model"
)

type MemoryOfferRepository struct {
	mu     sync.RWMutex
	offers map[string]model.Offer
	seq    uint64
}

func NewMemoryOfferRepository() *MemoryOfferRepository {
	return &MemoryOfferRepository{offers: make(map[string]model.Offer)}
}

func (r *MemoryOfferRepository) Create(request model.CreateOfferRequest) model.Offer {
	now := time.Now().UTC()
	id := fmt.Sprintf("OFF_%06d", atomic.AddUint64(&r.seq, 1)+889011)
	offer := model.Offer{OfferID: id, Name: request.Name, AdvertiserID: request.AdvertiserID, Status: request.Status, Payout: request.Payout, Targeting: request.Targeting, Caps: request.Caps, LandingPageURL: request.LandingPageURL, TrackingURLTemplate: request.TrackingURLTemplate, CreatedAt: now, UpdatedAt: now}
	r.mu.Lock()
	r.offers[id] = offer
	r.mu.Unlock()
	return offer
}

func (r *MemoryOfferRepository) Get(id string) (model.Offer, bool) {
	r.mu.RLock()
	offer, ok := r.offers[id]
	r.mu.RUnlock()
	return offer, ok
}

func (r *MemoryOfferRepository) Update(id string, request model.UpdateOfferRequest) (model.Offer, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	offer, ok := r.offers[id]
	if !ok {
		return model.Offer{}, false
	}
	if request.Name != nil {
		offer.Name = *request.Name
	}
	if request.Status != nil {
		offer.Status = *request.Status
	}
	if request.Targeting != nil {
		offer.Targeting = *request.Targeting
	}
	if request.Caps != nil {
		offer.Caps = *request.Caps
	}
	if request.LandingPageURL != nil {
		offer.LandingPageURL = *request.LandingPageURL
	}
	if request.TrackingURLTemplate != nil {
		offer.TrackingURLTemplate = *request.TrackingURLTemplate
	}
	if request.Payout != nil {
		if request.Payout.Type != nil {
			offer.Payout.Type = *request.Payout.Type
		}
		if request.Payout.Amount != nil {
			offer.Payout.Amount = *request.Payout.Amount
		}
		if request.Payout.Currency != nil {
			offer.Payout.Currency = *request.Payout.Currency
		}
	}
	offer.UpdatedAt = time.Now().UTC()
	r.offers[id] = offer
	return offer, true
}

func (r *MemoryOfferRepository) Archive(id string) (model.Offer, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	offer, ok := r.offers[id]
	if !ok {
		return model.Offer{}, false
	}
	now := time.Now().UTC()
	offer.Status = "archived"
	offer.ArchivedAt = &now
	offer.UpdatedAt = now
	r.offers[id] = offer
	return offer, true
}

type MemoryCredentialRepository struct{ username, password string }

func NewMemoryCredentialRepository(username, password string) *MemoryCredentialRepository {
	return &MemoryCredentialRepository{username: username, password: password}
}

func (r *MemoryCredentialRepository) Validate(username, password string) (bool, error) {
	return r.username == username && r.password == password, nil
}

type MemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]string
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{sessions: make(map[string]string)}
}

func (r *MemorySessionRepository) Create(username string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := fmt.Sprintf("%x", tokenBytes)
	r.mu.Lock()
	r.sessions[token] = username
	r.mu.Unlock()
	return token, nil
}

func (r *MemorySessionRepository) Delete(token string) {
	r.mu.Lock()
	delete(r.sessions, token)
	r.mu.Unlock()
}

func (r *MemorySessionRepository) Get(token string) (string, bool) {
	r.mu.RLock()
	username, ok := r.sessions[token]
	r.mu.RUnlock()
	return username, ok
}
