package model_test

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/planitaicojp/houjin-cli/internal/model"
)

func TestValidateCorporateNumber_valid(t *testing.T) {
	// check digit for base 180301018771 = 1
	if err := model.ValidateCorporateNumber("1180301018771"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestValidateCorporateNumber_invalidCheckDigit(t *testing.T) {
	err := model.ValidateCorporateNumber("2180301018771")
	if err == nil {
		t.Error("expected error for invalid check digit")
	}
}

func TestValidateCorporateNumber_wrongLength(t *testing.T) {
	err := model.ValidateCorporateNumber("123")
	if err == nil {
		t.Error("expected error for wrong length")
	}
}

func TestValidateCorporateNumber_nonNumeric(t *testing.T) {
	err := model.ValidateCorporateNumber("abcdefghijklm")
	if err == nil {
		t.Error("expected error for non-numeric")
	}
}

func TestParseXMLResponse(t *testing.T) {
	data, err := os.ReadFile("../../testdata/num_response.xml")
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	var xmlResp model.XMLResponse
	if err := xml.Unmarshal(data, &xmlResp); err != nil {
		t.Fatalf("failed to parse XML: %v", err)
	}

	resp := xmlResp.ToResponse()
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if len(resp.Corporations) != 1 {
		t.Fatalf("expected 1 corporation, got %d", len(resp.Corporations))
	}
	corp := resp.Corporations[0]
	if corp.CorporateNumber != "2180301018771" {
		t.Errorf("unexpected corporate number: %s", corp.CorporateNumber)
	}
	if corp.Name != "トヨタ自動車株式会社" {
		t.Errorf("unexpected name: %s", corp.Name)
	}
	if corp.NameKana != "トヨタジドウシャ" {
		t.Errorf("unexpected kana: %s", corp.NameKana)
	}
	if corp.Prefecture != "愛知県" {
		t.Errorf("unexpected prefecture: %s", corp.Prefecture)
	}
}
