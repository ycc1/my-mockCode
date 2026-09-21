package controller

import (
	"net/http"
	"strings"

	apimodel "advertiser-api/api/model"
	"advertiser-api/model"
	"advertiser-api/service"
)

type AdsPartnerController struct{ partners *service.AdsPartnerService }

func NewAdsPartnerController(partners *service.AdsPartnerService) *AdsPartnerController {
	return &AdsPartnerController{partners: partners}
}

func (c *AdsPartnerController) Collection(w http.ResponseWriter, r *http.Request) {
	username, _ := service.UsernameFromContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: c.partners.List()})
	case http.MethodPost:
		var request model.CreateAdsPartnerRequest
		if !apimodel.DecodeJSON(w, r, &request) {
			return
		}
		partner, err := c.partners.Create(request, username)
		if err != nil {
			apimodel.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		apimodel.WriteJSON(w, http.StatusCreated, model.APIResponse{Code: 0, Message: "Ads partner created successfully", Data: partner})
	default:
		apimodel.MethodNotAllowed(w)
	}
}

func (c *AdsPartnerController) ByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/ads-partners/")
	if id == "" || strings.Contains(id, "/") {
		apimodel.ErrorResponse(w, http.StatusNotFound, "ads partner not found")
		return
	}
	username, _ := service.UsernameFromContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		partner, ok := c.partners.Get(id)
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "ads partner not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: partner})
	case http.MethodPatch:
		var request model.UpdateAdsPartnerRequest
		if !apimodel.DecodeJSON(w, r, &request) {
			return
		}
		partner, ok, err := c.partners.Update(id, request, username)
		if err != nil {
			apimodel.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "ads partner not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Ads partner updated successfully", Data: partner})
	case http.MethodDelete:
		if !c.partners.Delete(id) {
			apimodel.ErrorResponse(w, http.StatusNotFound, "ads partner not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Ads partner deleted successfully"})
	default:
		apimodel.MethodNotAllowed(w)
	}
}
