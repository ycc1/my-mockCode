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

type RoleRepository interface {
	Create(model.CreateRoleRequest) model.Role
	List() []model.Role
	Get(string) (model.Role, bool)
	Update(string, model.UpdateRoleRequest) (model.Role, bool)
	Delete(string) bool
	AssignFeatures(string, []string) bool
	Features(string) []model.Feature
}

type FeatureRepository interface {
	Create(model.CreateFeatureRequest) model.Feature
	List() []model.Feature
	Get(string) (model.Feature, bool)
	Update(string, model.UpdateFeatureRequest) (model.Feature, bool)
	Delete(string) bool
}

type AuthorizationRepository interface {
	HasFeature(username, featureCode string) (bool, error)
}
