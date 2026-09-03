package service

import (
	"errors"
	"strings"

	"advertiser-api/model"
	"advertiser-api/repository"
)

var ErrInvalidOffer = errors.New("invalid offer payload")

type OfferService struct{ offers repository.OfferRepository }

func NewOfferService(offers repository.OfferRepository) *OfferService {
	return &OfferService{offers: offers}
}

func (s *OfferService) Create(request model.CreateOfferRequest) (model.Offer, error) {
	if request.Name == "" || request.AdvertiserID == "" || request.Payout.Type == "" || request.Payout.Amount <= 0 || request.Payout.Currency == "" || request.Caps.DailyCap < 0 {
		return model.Offer{}, ErrInvalidOffer
	}
	if request.Status == "" {
		request.Status = "active"
	}
	return s.offers.Create(request), nil
}

func (s *OfferService) Get(id string) (model.Offer, bool) { return s.offers.Get(id) }
func (s *OfferService) Update(id string, request model.UpdateOfferRequest) (model.Offer, bool) {
	return s.offers.Update(id, request)
}
func (s *OfferService) Archive(id string) (model.Offer, bool) { return s.offers.Archive(id) }
func OfferIDFromPath(path string) string {
	return strings.TrimPrefix(path, "/api/v1/advertiser/offers/")
}
