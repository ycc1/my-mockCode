package repository

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"advertiser-api/model"
)

type MemoryFeatureRepository struct {
	mu       sync.RWMutex
	features map[string]model.Feature
	seq      uint64
}

func NewMemoryFeatureRepository() *MemoryFeatureRepository {
	return &MemoryFeatureRepository{features: make(map[string]model.Feature)}
}
func (r *MemoryFeatureRepository) Create(request model.CreateFeatureRequest) model.Feature {
	now := time.Now().UTC()
	feature := model.Feature{FeatureID: fmt.Sprintf("FEAT_%06d", atomic.AddUint64(&r.seq, 1)), Code: request.Code, Name: request.Name, Description: request.Description, CreatedAt: now, UpdatedAt: now}
	r.mu.Lock()
	r.features[feature.FeatureID] = feature
	r.mu.Unlock()
	return feature
}
func (r *MemoryFeatureRepository) List() []model.Feature {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Feature, 0, len(r.features))
	for _, item := range r.features {
		result = append(result, item)
	}
	return result
}
func (r *MemoryFeatureRepository) Get(id string) (model.Feature, bool) {
	r.mu.RLock()
	item, ok := r.features[id]
	r.mu.RUnlock()
	return item, ok
}
func (r *MemoryFeatureRepository) Update(id string, request model.UpdateFeatureRequest) (model.Feature, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.features[id]
	if !ok {
		return model.Feature{}, false
	}
	if request.Code != nil {
		item.Code = *request.Code
	}
	if request.Name != nil {
		item.Name = *request.Name
	}
	if request.Description != nil {
		item.Description = *request.Description
	}
	item.UpdatedAt = time.Now().UTC()
	r.features[id] = item
	return item, true
}
func (r *MemoryFeatureRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.features[id]; !ok {
		return false
	}
	delete(r.features, id)
	return true
}

type MemoryRoleRepository struct {
	mu           sync.RWMutex
	roles        map[string]model.Role
	roleFeatures map[string]map[string]bool
	features     FeatureRepository
	seq          uint64
}

func NewMemoryRoleRepository(features FeatureRepository) *MemoryRoleRepository {
	return &MemoryRoleRepository{roles: make(map[string]model.Role), roleFeatures: make(map[string]map[string]bool), features: features}
}
func (r *MemoryRoleRepository) Create(request model.CreateRoleRequest) model.Role {
	now := time.Now().UTC()
	role := model.Role{RoleID: fmt.Sprintf("ROLE_%06d", atomic.AddUint64(&r.seq, 1)), Name: request.Name, Description: request.Description, CreatedAt: now, UpdatedAt: now}
	r.mu.Lock()
	r.roles[role.RoleID] = role
	r.roleFeatures[role.RoleID] = make(map[string]bool)
	r.mu.Unlock()
	return role
}
func (r *MemoryRoleRepository) List() []model.Role {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Role, 0, len(r.roles))
	for _, item := range r.roles {
		result = append(result, item)
	}
	return result
}
func (r *MemoryRoleRepository) Get(id string) (model.Role, bool) {
	r.mu.RLock()
	item, ok := r.roles[id]
	r.mu.RUnlock()
	return item, ok
}
func (r *MemoryRoleRepository) Update(id string, request model.UpdateRoleRequest) (model.Role, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.roles[id]
	if !ok {
		return model.Role{}, false
	}
	if request.Name != nil {
		item.Name = *request.Name
	}
	if request.Description != nil {
		item.Description = *request.Description
	}
	item.UpdatedAt = time.Now().UTC()
	r.roles[id] = item
	return item, true
}
func (r *MemoryRoleRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.roles[id]; !ok {
		return false
	}
	delete(r.roles, id)
	delete(r.roleFeatures, id)
	return true
}
func (r *MemoryRoleRepository) AssignFeatures(roleID string, featureIDs []string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.roles[roleID]; !ok {
		return false
	}
	assigned := make(map[string]bool)
	for _, id := range featureIDs {
		if _, exists := r.features.Get(id); exists {
			assigned[id] = true
		}
	}
	r.roleFeatures[roleID] = assigned
	return true
}
func (r *MemoryRoleRepository) Features(roleID string) []model.Feature {
	r.mu.RLock()
	ids := r.roleFeatures[roleID]
	r.mu.RUnlock()
	result := []model.Feature{}
	for id := range ids {
		if item, ok := r.features.Get(id); ok {
			result = append(result, item)
		}
	}
	return result
}

type MemoryAuthorizationRepository struct {
	roles RoleRepository
	users map[string]string
}

func NewMemoryAuthorizationRepository(roles RoleRepository) *MemoryAuthorizationRepository {
	return &MemoryAuthorizationRepository{roles: roles, users: map[string]string{"admin": "ROLE_ADMIN"}}
}
func (r *MemoryAuthorizationRepository) HasFeature(username, featureCode string) (bool, error) {
	if username == "admin" {
		return true, nil
	}
	return false, nil
}
