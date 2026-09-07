package controller

import (
	"net/http"
	"strings"

	apimodel "advertiser-api/api/model"
	"advertiser-api/model"
	"advertiser-api/service"
)

type AccessController struct{ access *service.AccessService }

func NewAccessController(access *service.AccessService) *AccessController {
	return &AccessController{access: access}
}

func (c *AccessController) Roles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: c.access.ListRoles()})
	case http.MethodPost:
		var request model.CreateRoleRequest
		if !apimodel.DecodeJSON(w, r, &request) {
			return
		}
		role, err := c.access.CreateRole(request)
		if err != nil {
			apimodel.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		apimodel.WriteJSON(w, http.StatusCreated, model.APIResponse{Code: 0, Message: "Role created successfully", Data: role})
	default:
		apimodel.MethodNotAllowed(w)
	}
}

func (c *AccessController) RoleByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/roles/")
	if id == "" || strings.Contains(id, "/") {
		apimodel.ErrorResponse(w, http.StatusNotFound, "role not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		role, ok := c.access.GetRole(id)
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "role not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: map[string]interface{}{"role": role, "features": c.access.RoleFeatures(id)}})
	case http.MethodPatch:
		var request model.UpdateRoleRequest
		if !apimodel.DecodeJSON(w, r, &request) {
			return
		}
		role, ok := c.access.UpdateRole(id, request)
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "role not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Role updated successfully", Data: role})
	case http.MethodPut:
		var request model.AssignFeaturesRequest
		if !apimodel.DecodeJSON(w, r, &request) {
			return
		}
		if !c.access.AssignFeatures(id, request.FeatureIDs) {
			apimodel.ErrorResponse(w, http.StatusNotFound, "role not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Role features updated successfully", Data: c.access.RoleFeatures(id)})
	case http.MethodDelete:
		if !c.access.DeleteRole(id) {
			apimodel.ErrorResponse(w, http.StatusNotFound, "role not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Role deleted successfully"})
	default:
		apimodel.MethodNotAllowed(w)
	}
}

func (c *AccessController) Features(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: c.access.ListFeatures()})
	case http.MethodPost:
		var request model.CreateFeatureRequest
		if !apimodel.DecodeJSON(w, r, &request) {
			return
		}
		feature, err := c.access.CreateFeature(request)
		if err != nil {
			apimodel.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		apimodel.WriteJSON(w, http.StatusCreated, model.APIResponse{Code: 0, Message: "Feature created successfully", Data: feature})
	default:
		apimodel.MethodNotAllowed(w)
	}
}

func (c *AccessController) FeatureByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/features/")
	if id == "" || strings.Contains(id, "/") {
		apimodel.ErrorResponse(w, http.StatusNotFound, "feature not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		feature, ok := c.access.GetFeature(id)
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "feature not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: feature})
	case http.MethodPatch:
		var request model.UpdateFeatureRequest
		if !apimodel.DecodeJSON(w, r, &request) {
			return
		}
		feature, ok := c.access.UpdateFeature(id, request)
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "feature not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Feature updated successfully", Data: feature})
	case http.MethodDelete:
		if !c.access.DeleteFeature(id) {
			apimodel.ErrorResponse(w, http.StatusNotFound, "feature not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Feature deleted successfully"})
	default:
		apimodel.MethodNotAllowed(w)
	}
}
