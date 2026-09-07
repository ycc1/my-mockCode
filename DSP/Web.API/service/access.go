package service

import (
	"advertiser-api/model"
	"advertiser-api/repository"
	"errors"
)

var ErrInvalidAccessPayload = errors.New("invalid role or feature payload")

type AccessService struct {
	roles    repository.RoleRepository
	features repository.FeatureRepository
}

func NewAccessService(roles repository.RoleRepository, features repository.FeatureRepository) *AccessService {
	return &AccessService{roles: roles, features: features}
}
func (s *AccessService) CreateRole(request model.CreateRoleRequest) (model.Role, error) {
	if request.Name == "" {
		return model.Role{}, ErrInvalidAccessPayload
	}
	return s.roles.Create(request), nil
}
func (s *AccessService) ListRoles() []model.Role              { return s.roles.List() }
func (s *AccessService) GetRole(id string) (model.Role, bool) { return s.roles.Get(id) }
func (s *AccessService) UpdateRole(id string, request model.UpdateRoleRequest) (model.Role, bool) {
	return s.roles.Update(id, request)
}
func (s *AccessService) DeleteRole(id string) bool { return s.roles.Delete(id) }
func (s *AccessService) AssignFeatures(id string, featureIDs []string) bool {
	return s.roles.AssignFeatures(id, featureIDs)
}
func (s *AccessService) RoleFeatures(id string) []model.Feature { return s.roles.Features(id) }
func (s *AccessService) CreateFeature(request model.CreateFeatureRequest) (model.Feature, error) {
	if request.Code == "" || request.Name == "" {
		return model.Feature{}, ErrInvalidAccessPayload
	}
	return s.features.Create(request), nil
}
func (s *AccessService) ListFeatures() []model.Feature              { return s.features.List() }
func (s *AccessService) GetFeature(id string) (model.Feature, bool) { return s.features.Get(id) }
func (s *AccessService) UpdateFeature(id string, request model.UpdateFeatureRequest) (model.Feature, bool) {
	return s.features.Update(id, request)
}
func (s *AccessService) DeleteFeature(id string) bool { return s.features.Delete(id) }
