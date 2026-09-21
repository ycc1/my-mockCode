package service

import (
	"strings"
	"testing"

	"advertiser-api/model"
	"advertiser-api/repository"
)

func TestAdsPartnerCRUDAndValidation(t *testing.T) {
	service := NewAdsPartnerService(repository.NewMemoryAdsPartnerRepository())
	request := model.CreateAdsPartnerRequest{Name: "Advertiser A", AdsMerchantID: "Ab12Cd34Ef56Gh78", APIKey: strings.Repeat("a", 32), SecurityType: "MD5"}
	created, err := service.Create(request, "creator")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.CreateBy != "creator" || created.UpdateBy != "creator" {
		t.Fatalf("unexpected audit users: %#v", created)
	}
	if _, err := service.Create(request, "creator"); err == nil {
		t.Fatal("accepted duplicate ads merchant ID")
	}
	invalid := request
	invalid.AdsMerchantID = strings.Repeat("x", 21)
	if _, err := service.Create(invalid, "creator"); err == nil {
		t.Fatal("accepted ads merchant ID longer than 20 characters")
	}
	securityType, apiKey := "DES", strings.Repeat("b", 8)
	updated, ok, err := service.Update(created.ID, model.UpdateAdsPartnerRequest{SecurityType: &securityType, APIKey: &apiKey}, "editor")
	if err != nil || !ok {
		t.Fatalf("update: ok=%v err=%v", ok, err)
	}
	if updated.SecurityType != "DES" || updated.UpdateBy != "editor" {
		t.Fatalf("unexpected update: %#v", updated)
	}
	if !service.Delete(created.ID) {
		t.Fatal("delete returned false")
	}
}
