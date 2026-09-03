package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func login(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/membership/login", bytes.NewReader([]byte(`{"username":"admin","password":"admin123"}`)))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", response.Code, http.StatusOK)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value == "" {
		t.Fatal("login did not set a session cookie")
	}
	return cookies[0]
}

func TestAuthFilterProtectsOffers(t *testing.T) {
	handler := routes()
	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/advertiser/offers/OFF_missing", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	cookie := login(t, handler)
	authorized := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/advertiser/offers/OFF_missing", nil)
	request.AddCookie(cookie)
	handler.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusNotFound {
		t.Fatalf("authorized status = %d, want %d", authorized.Code, http.StatusNotFound)
	}
}

func TestOfferCRUD(t *testing.T) {
	handler := routes()
	cookie := login(t, handler)
	body := []byte(`{"name":"Test CPI","advertiser_id":"ADV_TEST","status":"active","payout":{"type":"CPI","amount":2.5,"currency":"USD"},"caps":{"daily_cap":100}}`)

	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/advertiser/offers", bytes.NewReader(body))
	createRequest.AddCookie(cookie)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createRequest)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", create.Code, http.StatusCreated)
	}
	var created struct {
		Data struct {
			OfferID string `json:"offer_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.Data.OfferID == "" {
		t.Fatal("create response did not include offer_id")
	}

	patch := []byte(`{"payout":{"amount":3},"caps":{"daily_cap":200}}`)
	updateRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/advertiser/offers/"+created.Data.OfferID, bytes.NewReader(patch))
	updateRequest.AddCookie(cookie)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateRequest)
	if update.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d", update.Code, http.StatusOK)
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/api/v1/advertiser/offers/"+created.Data.OfferID, nil)
	readRequest.AddCookie(cookie)
	read := httptest.NewRecorder()
	handler.ServeHTTP(read, readRequest)
	if read.Code != http.StatusOK {
		t.Fatalf("read status = %d, want %d", read.Code, http.StatusOK)
	}
	var current struct {
		Data struct {
			Payout struct {
				Amount float64 `json:"amount"`
			} `json:"payout"`
			Caps struct {
				DailyCap int `json:"daily_cap"`
			} `json:"caps"`
		} `json:"data"`
	}
	if err := json.NewDecoder(read.Body).Decode(&current); err != nil {
		t.Fatal(err)
	}
	if current.Data.Payout.Amount != 3 || current.Data.Caps.DailyCap != 200 {
		t.Fatalf("read returned stale offer: %+v", current.Data)
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/api/v1/membership/logout", nil)
	logoutRequest.AddCookie(cookie)
	logout := httptest.NewRecorder()
	handler.ServeHTTP(logout, logoutRequest)
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want %d", logout.Code, http.StatusOK)
	}
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/membership/login", bytes.NewReader([]byte(`{"username":"admin","password":"wrong"}`)))
	routes().ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
