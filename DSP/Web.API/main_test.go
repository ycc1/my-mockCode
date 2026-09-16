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

func TestRoleCRUDAndFeatureAssignment(t *testing.T) {
	handler := routes()
	cookie := login(t, handler)

	featureResponse := httptest.NewRecorder()
	featureRequest := httptest.NewRequest(http.MethodPost, "/api/v1/features", bytes.NewReader([]byte(`{"code":"offer.read","name":"Read offers"}`)))
	featureRequest.AddCookie(cookie)
	handler.ServeHTTP(featureResponse, featureRequest)
	if featureResponse.Code != http.StatusCreated {
		t.Fatalf("create feature status = %d, want %d", featureResponse.Code, http.StatusCreated)
	}
	var createdFeature struct {
		Data struct {
			FeatureID string `json:"feature_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(featureResponse.Body).Decode(&createdFeature); err != nil {
		t.Fatal(err)
	}

	roleResponse := httptest.NewRecorder()
	roleRequest := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(`{"name":"Advertiser","description":"Advertiser access"}`)))
	roleRequest.AddCookie(cookie)
	handler.ServeHTTP(roleResponse, roleRequest)
	if roleResponse.Code != http.StatusCreated {
		t.Fatalf("create role status = %d, want %d", roleResponse.Code, http.StatusCreated)
	}
	var createdRole struct {
		Data struct {
			RoleID string `json:"role_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(roleResponse.Body).Decode(&createdRole); err != nil {
		t.Fatal(err)
	}
	if createdRole.Data.RoleID == "" || createdFeature.Data.FeatureID == "" {
		t.Fatal("role or feature response did not include an ID")
	}

	assignResponse := httptest.NewRecorder()
	assignRequest := httptest.NewRequest(http.MethodPut, "/api/v1/roles/"+createdRole.Data.RoleID, bytes.NewReader([]byte(`{"feature_ids":["`+createdFeature.Data.FeatureID+`"]}`)))
	assignRequest.AddCookie(cookie)
	handler.ServeHTTP(assignResponse, assignRequest)
	if assignResponse.Code != http.StatusOK {
		t.Fatalf("assign features status = %d, want %d", assignResponse.Code, http.StatusOK)
	}

	updateResponse := httptest.NewRecorder()
	updateRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/roles/"+createdRole.Data.RoleID, bytes.NewReader([]byte(`{"description":"Updated access"}`)))
	updateRequest.AddCookie(cookie)
	handler.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("update role status = %d, want %d", updateResponse.Code, http.StatusOK)
	}

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+createdRole.Data.RoleID, nil)
	getRequest.AddCookie(cookie)
	handler.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("get role status = %d, want %d", getResponse.Code, http.StatusOK)
	}

	deleteResponse := httptest.NewRecorder()
	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/"+createdRole.Data.RoleID, nil)
	deleteRequest.AddCookie(cookie)
	handler.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("delete role status = %d, want %d", deleteResponse.Code, http.StatusOK)
	}
}

func TestChannelPartnerCRUD(t *testing.T) {
	handler := routes()
	cookie := login(t, handler)
	payload := []byte(`{"name":"Partner A","code":"123","partner_type":"其他","partner_type_other":"合作伙伴","service_category":"DSP","traffic_model":"cpc","billing_model":"cpm","primary_channel":"Google","secondary_channel":"其他","secondary_channel_other":"私域","market_contract":"是","delivery_package":["苹果","安卓"],"data_system":["GA4","adjust"],"status":true}`)

	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/channel-partners", bytes.NewReader(payload))
	createRequest.AddCookie(cookie)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createRequest)
	if create.Code != http.StatusCreated { t.Fatalf("create channel partner status = %d, want %d", create.Code, http.StatusCreated) }
	var created struct { Data struct { ID string `json:"channel_partner_id"`; Status bool `json:"status"`; ModifiedBy string `json:"modified_by"` } `json:"data"` }
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil { t.Fatal(err) }
	if created.Data.ID == "" || !created.Data.Status || created.Data.ModifiedBy != "admin" { t.Fatalf("unexpected created partner: %+v", created.Data) }

	updateRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/channel-partners/"+created.Data.ID, bytes.NewReader([]byte(`{"status":false}`)))
	updateRequest.AddCookie(cookie)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateRequest)
	if update.Code != http.StatusOK { t.Fatalf("update channel partner status = %d, want %d", update.Code, http.StatusOK) }

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/channel-partners/"+created.Data.ID, nil)
	getRequest.AddCookie(cookie)
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, getRequest)
	if get.Code != http.StatusOK { t.Fatalf("get channel partner status = %d, want %d", get.Code, http.StatusOK) }
	var current struct { Data struct { Status bool `json:"status"`; SecondaryOther string `json:"secondary_channel_other"` } `json:"data"` }
	if err := json.NewDecoder(get.Body).Decode(&current); err != nil { t.Fatal(err) }
	if current.Data.Status || current.Data.SecondaryOther != "私域" { t.Fatalf("unexpected partner after update: %+v", current.Data) }

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/channel-partners/"+created.Data.ID, nil)
	deleteRequest.AddCookie(cookie)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteRequest)
	if deleted.Code != http.StatusOK { t.Fatalf("delete channel partner status = %d, want %d", deleted.Code, http.StatusOK) }
}

func TestChannelPartnerRequiresLogin(t *testing.T) {
	response := httptest.NewRecorder()
	routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/channel-partners", nil))
	if response.Code != http.StatusUnauthorized { t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized) }
}
