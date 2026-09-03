package model

import (
	"encoding/json"
	"net/http"

	domainmodel "advertiser-api/model"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func WriteJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, domainmodel.APIResponse{Code: status, Message: message})
}

func MethodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, POST, PATCH, DELETE")
	ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
}
