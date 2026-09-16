package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"advertiser-api/model"
)

func TestChannelPartnerOptionsEndpoint(t *testing.T) {
	handler := routes()
	url := "/api/v1/channel-partners/options"
	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, url, nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d", unauthorized.Code)
	}
	cookie := login(t, handler)
	request := httptest.NewRequest(http.MethodGet, url, nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("options status = %d", response.Code)
	}
	var payload struct {
		Data map[string][]string `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(payload.Data, model.ChannelPartnerOptions()) {
		t.Fatal("endpoint options differ from model configuration")
	}
	request = httptest.NewRequest(http.MethodPost, url, nil)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST options status = %d", response.Code)
	}
}
