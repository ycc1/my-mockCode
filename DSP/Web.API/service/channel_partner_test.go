package service

import (
	"advertiser-api/model"
	"advertiser-api/repository"
	"testing"
)

func TestMultiSelectChoices(t *testing.T) {
	for _, field := range []string{"delivery_package", "data_system"} {
		first, second := "苹果", "安卓"
		if field == "data_system" {
			first, second = "GA4", "adjust"
		}
		for _, test := range []struct {
			values []string
			other  string
			valid  bool
		}{
			{[]string{first, second}, "", true},
			{[]string{first, "其他"}, "自定义", true},
			{nil, "", false},
			{[]string{}, "", false},
			{[]string{""}, "", false},
			{[]string{first, first}, "", false},
			{[]string{"invalid"}, "", false},
			{[]string{"其他"}, "  ", false},
			{[]string{first}, "多余内容", false},
		} {
			if got := validChoices(field, test.values, test.other); got != test.valid {
				t.Errorf("%s %v: got %v, want %v", field, test.values, got, test.valid)
			}
		}
	}
}

func TestChannelPartnerValidationAndUpdates(t *testing.T) {
	s := NewChannelPartnerService(repository.NewMemoryChannelPartnerRepository())
	request := model.CreateChannelPartnerRequest{Code: "001", Name: "Test", MerchantID: "merchant-001", APIKey: "12345678901234567890123456789012", SecurityType: "MD5", PartnerType: "外部", ServiceCategory: "DSP", TrafficModel: "cpc", BillingModel: "cpm", PrimaryChannel: "Google", SecondaryChannel: "TikTok", MarketContract: "是", DeliveryPackage: []string{"苹果", "安卓"}, DataSystem: []string{"GA4", "adjust"}, Status: true}
	tooLongMerchantID := request
	tooLongMerchantID.MerchantID = "123456789012345678901"
	if _, err := s.Create(tooLongMerchantID, "admin"); err == nil {
		t.Fatal("accepted merchant ID longer than 20 characters")
	}
	for _, code := range []string{"", "12", "abcd", "１２３"} {
		invalid := request
		invalid.Code = code
		if _, err := s.Create(invalid, "admin"); err == nil {
			t.Fatalf("accepted invalid code %q", code)
		}
	}
	partner, err := s.Create(request, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if partner.Code != "001" {
		t.Fatal("lost leading zeros")
	}
	if _, err := s.Create(request, "admin"); err == nil {
		t.Fatal("accepted duplicate code")
	}
	request.Code = "002"
	second, err := s.Create(request, "admin")
	if err != nil {
		t.Fatal(err)
	}
	duplicate := "001"
	if _, _, err := s.Update(second.ChannelPartnerID, model.UpdateChannelPartnerRequest{Code: &duplicate}, "admin"); err == nil {
		t.Fatal("accepted duplicate code update")
	}
	other := "其他"
	detail := "私域"
	if _, _, err := s.Update(partner.ChannelPartnerID, model.UpdateChannelPartnerRequest{SecondaryChannel: &other}, "admin"); err == nil {
		t.Fatal("accepted other without detail")
	}
	if _, ok, err := s.Update(partner.ChannelPartnerID, model.UpdateChannelPartnerRequest{SecondaryChannel: &other, SecondaryChannelOther: &detail}, "admin"); !ok || err != nil {
		t.Fatal("valid other update failed", err)
	}
	empty := ""
	if _, _, err := s.Update(partner.ChannelPartnerID, model.UpdateChannelPartnerRequest{SecondaryChannelOther: &empty}, "admin"); err == nil {
		t.Fatal("accepted clearing required other detail")
	}
	disabled := false
	updated, ok, err := s.Update(partner.ChannelPartnerID, model.UpdateChannelPartnerRequest{Status: &disabled}, "operator")
	if !ok || err != nil || updated.Status || updated.SecondaryChannelOther != detail || updated.ModifiedBy != "operator" {
		t.Fatal("partial status update failed")
	}
	if !s.Delete(partner.ChannelPartnerID) {
		t.Fatal("delete failed")
	}
	if _, ok := s.Get(partner.ChannelPartnerID); ok {
		t.Fatal("deleted record still exists")
	}
}
