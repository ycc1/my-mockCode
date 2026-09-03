package controller

import (
	"net/http"

	apimodel "advertiser-api/api/model"
)

type HealthController struct{}

func NewHealthController() *HealthController { return &HealthController{} }

func (c *HealthController) Check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apimodel.MethodNotAllowed(w)
		return
	}
	apimodel.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
