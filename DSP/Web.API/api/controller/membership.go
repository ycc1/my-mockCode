package controller

import (
	"net/http"

	apimodel "advertiser-api/api/model"
	"advertiser-api/model"
	"advertiser-api/service"
)

type MembershipController struct{ auth *service.AuthService }

func NewMembershipController(auth *service.AuthService) *MembershipController {
	return &MembershipController{auth: auth}
}

func (c *MembershipController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apimodel.MethodNotAllowed(w)
		return
	}
	var request model.LoginRequest
	if !apimodel.DecodeJSON(w, r, &request) {
		return
	}
	token, valid, err := c.auth.Login(request.Username, request.Password)
	if err != nil {
		apimodel.ErrorResponse(w, http.StatusInternalServerError, "could not create session")
		return
	}
	if !valid {
		apimodel.ErrorResponse(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: service.SessionCookieName, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Login successful", Data: map[string]string{"username": request.Username}})
}

func (c *MembershipController) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apimodel.MethodNotAllowed(w)
		return
	}
	if cookie, err := r.Cookie(service.SessionCookieName); err == nil {
		c.auth.Logout(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: service.SessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	apimodel.WriteJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "Logout successful"})
}
