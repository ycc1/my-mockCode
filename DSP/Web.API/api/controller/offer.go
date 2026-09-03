package controller

import (
	"net/http"
	"strings"

	apimodel "advertiser-api/api/model"
	"advertiser-api/model"
	"advertiser-api/service"
)

type OfferController struct{ offers *service.OfferService }

func NewOfferController(offers *service.OfferService) *OfferController {
	return &OfferController{offers: offers}
}

func (c *OfferController) Collection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apimodel.MethodNotAllowed(w)
		return
	}
	var request model.CreateOfferRequest
	if !apimodel.DecodeJSON(w, r, &request) {
		return
	}
	offer, err := c.offers.Create(request)
	if err != nil {
		apimodel.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	apimodel.WriteJSON(w, http.StatusCreated, model.APIResponse{Code: 0, Message: "Offer created successfully", Data: map[string]interface{}{"offer_id": offer.OfferID, "status": offer.Status, "created_at": offer.CreatedAt}})
}

func (c *OfferController) ByID(w http.ResponseWriter, r *http.Request) {
	id := service.OfferIDFromPath(r.URL.Path)
	if id == "" || strings.Contains(id, "/") {
		apimodel.ErrorResponse(w, http.StatusNotFound, "offer not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		offer, ok := c.offers.Get(id)
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "offer not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "success", Data: offer})
	case http.MethodPatch:
		var request model.UpdateOfferRequest
		if !apimodel.DecodeJSON(w, r, &request) {
			return
		}
		offer, ok := c.offers.Update(id, request)
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "offer not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Offer updated successfully", Data: offer})
	case http.MethodDelete:
		offer, ok := c.offers.Archive(id)
		if !ok {
			apimodel.ErrorResponse(w, http.StatusNotFound, "offer not found")
			return
		}
		apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Offer archived successfully", Data: map[string]interface{}{"offer_id": offer.OfferID, "status": offer.Status, "archived_at": offer.ArchivedAt}})
	default:
		apimodel.MethodNotAllowed(w)
	}
}
