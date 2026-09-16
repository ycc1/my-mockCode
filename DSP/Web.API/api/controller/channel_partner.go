package controller

import (
	"net/http"
	"strings"

	apimodel "advertiser-api/api/model"
	"advertiser-api/model"
	"advertiser-api/service"
)

type ChannelPartnerController struct{ partners *service.ChannelPartnerService }

func (c *ChannelPartnerController) Options(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { apimodel.MethodNotAllowed(w); return }
	apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: model.ChannelPartnerOptions()})
}

func NewChannelPartnerController(partners *service.ChannelPartnerService) *ChannelPartnerController { return &ChannelPartnerController{partners: partners} }

func (c *ChannelPartnerController) Collection(w http.ResponseWriter, r *http.Request) {
	username, _ := service.UsernameFromContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: c.partners.List()})
	case http.MethodPost:
		var request model.CreateChannelPartnerRequest
		if !apimodel.DecodeJSON(w, r, &request) { return }
		partner, err := c.partners.Create(request, username)
		if err != nil { apimodel.ErrorResponse(w, http.StatusBadRequest, err.Error()); return }
		apimodel.WriteJSON(w, http.StatusCreated, model.APIResponse{Code: 0, Message: "Channel partner created successfully", Data: partner})
	default:
		apimodel.MethodNotAllowed(w)
	}
}

func (c *ChannelPartnerController) ByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/channel-partners/")
	if id == "" || strings.Contains(id, "/") { apimodel.ErrorResponse(w, http.StatusNotFound, "channel partner not found"); return }
	username, _ := service.UsernameFromContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		partner, ok := c.partners.Get(id); if !ok { apimodel.ErrorResponse(w, http.StatusNotFound, "channel partner not found"); return }; apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: partner})
	case http.MethodPatch:
		var request model.UpdateChannelPartnerRequest; if !apimodel.DecodeJSON(w, r, &request) { return }; partner, ok, err := c.partners.Update(id, request, username); if err != nil { apimodel.ErrorResponse(w, http.StatusBadRequest, err.Error()); return }; if !ok { apimodel.ErrorResponse(w, http.StatusNotFound, "channel partner not found"); return }; apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Channel partner updated successfully", Data: partner})
	case http.MethodDelete:
		if !c.partners.Delete(id) { apimodel.ErrorResponse(w, http.StatusNotFound, "channel partner not found"); return }; apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Channel partner deleted successfully"})
	default:
		apimodel.MethodNotAllowed(w)
	}
}
