package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"advertiser-api/model"
)

func TestFeatureCatalogIncludesEveryCRUDAction(t *testing.T) {
	handler := routes()
	cookie := login(t, handler)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/features", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("list features status = %d", response.Code)
	}
	var payload struct {
		Data []model.Feature `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	codes := make(map[string]string)
	for _, feature := range payload.Data {
		codes[feature.Code] = feature.FeatureID
	}
	for _, resource := range []string{"offer", "channel_partner", "channel_number", "channel_link", "role", "feature", "account"} {
		for _, action := range []string{"create", "read", "update", "delete"} {
			if codes[resource+"."+action] == "" {
				t.Errorf("missing assignable permission: %s.%s", resource, action)
			}
		}
	}
	for _, resource := range []string{"report", "report.upload_log", "report.settlement", "report.attribution"} {
		if codes[resource+".read"] == "" {
			t.Errorf("missing report query permission: %s", resource)
		}
		for _, action := range []string{"create", "update", "delete"} {
			if codes[resource+"."+action] != "" {
				t.Errorf("report must be read-only: %s.%s", resource, action)
			}
		}
	}
}
