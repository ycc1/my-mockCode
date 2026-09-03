package repository

import "advertiser-api/model"

type OfferRepository interface {
	Create(model.CreateOfferRequest) model.Offer
	Get(string) (model.Offer, bool)
	Update(string, model.UpdateOfferRequest) (model.Offer, bool)
	Archive(string) (model.Offer, bool)
}

type CredentialRepository interface {
	Validate(username, password string) (bool, error)
}

type LoginRepository interface {
	CredentialRepository
}

type SessionRepository interface {
	Create(username string) (string, error)
	Delete(token string)
	Get(token string) (string, bool)
}
