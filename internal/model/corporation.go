package model

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"unicode"
)

// XMLResponse represents the raw XML response from the API.
type XMLResponse struct {
	XMLName      xml.Name         `xml:"corporations"`
	Count        int              `xml:"count"`
	DivideNumber int              `xml:"divideNumber"`
	DivideSize   int              `xml:"divideSize"`
	Corporations []XMLCorporation `xml:"corporation"`
}

// XMLCorporation represents a single corporation entry in the XML response.
type XMLCorporation struct {
	CorporateNumber string `xml:"corporateNumber"`
	Name            string `xml:"name"`
	Furigana        string `xml:"furigana"`
	EnName          string `xml:"enName"`
	Kind            string `xml:"kind"`
	PrefectureName  string `xml:"prefectureName"`
	CityName        string `xml:"cityName"`
	StreetNumber    string `xml:"streetNumber"`
	PrefectureCode  string `xml:"prefectureCode"`
	CityCode        string `xml:"cityCode"`
	PostCode        string `xml:"postCode"`
	AddressOutside  string `xml:"addressOutside"`
	AssignmentDate  string `xml:"assignmentDate"`
	UpdateDate      string `xml:"updateDate"`
	ChangeDate      string `xml:"changeDate"`
	CloseDate       string `xml:"closeDate"`
	CloseCause      string `xml:"closeCause"`
	Latest          string `xml:"latest"`
	Process         string `xml:"process"`
	Correct         string `xml:"correct"`
	Hihyoji         string `xml:"hihyoji"`
}

// Corporation is the public-facing corporation data structure.
type Corporation struct {
	CorporateNumber string `json:"corporate_number"`
	Name            string `json:"name"`
	NameKana        string `json:"name_kana"`
	NameEnglish     string `json:"name_english"`
	Kind            string `json:"kind"`
	Prefecture      string `json:"prefecture"`
	City            string `json:"city"`
	Address         string `json:"address"`
	PostalCode      string `json:"postal_code"`
	AssignmentDate  string `json:"assignment_date"`
	UpdateDate      string `json:"update_date"`
	ChangeDate      string `json:"change_date"`
	CloseDate       string `json:"close_date,omitempty"`
	CloseCause      string `json:"close_cause,omitempty"`
}

// Response is the parsed API response.
type Response struct {
	Count        int           `json:"count"`
	DivideNumber int           `json:"divide_number"`
	DivideSize   int           `json:"divide_size"`
	Corporations []Corporation `json:"corporations"`
}

// ToResponse converts the XML response to the public Response type.
func (x *XMLResponse) ToResponse() *Response {
	corps := make([]Corporation, len(x.Corporations))
	for i, xc := range x.Corporations {
		corps[i] = Corporation{
			CorporateNumber: xc.CorporateNumber,
			Name:            xc.Name,
			NameKana:        xc.Furigana,
			NameEnglish:     xc.EnName,
			Kind:            xc.Kind,
			Prefecture:      xc.PrefectureName,
			City:            xc.CityName,
			Address:         xc.StreetNumber,
			PostalCode:      xc.PostCode,
			AssignmentDate:  xc.AssignmentDate,
			UpdateDate:      xc.UpdateDate,
			ChangeDate:      xc.ChangeDate,
			CloseDate:       xc.CloseDate,
			CloseCause:      xc.CloseCause,
		}
	}
	return &Response{
		Count:        x.Count,
		DivideNumber: x.DivideNumber,
		DivideSize:   x.DivideSize,
		Corporations: corps,
	}
}

// ValidateCorporateNumber validates a 13-digit corporate number including check digit.
func ValidateCorporateNumber(number string) error {
	if len(number) != 13 {
		return fmt.Errorf("corporate number must be 13 digits, got %d", len(number))
	}
	for _, r := range number {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("corporate number must contain only digits")
		}
	}

	digits := make([]int, 13)
	for i, r := range number {
		digits[i], _ = strconv.Atoi(string(r))
	}

	// Check digit is the first digit (position 1, 1-indexed).
	// Formula: check = 9 - (sum mod 9)
	// sum = Σ(i=2..13) Q_i × P_i
	// P_i = 1 if i is even, 2 if i is odd
	// Position from right: n = 13 - i (where i is 1-indexed array position)
	// P_n = 1 if n is odd, 2 if n is even
	// Since n = 13-i: n odd ↔ i even, n even ↔ i odd
	sum := 0
	for i := 1; i < 13; i++ {
		p := 1
		if i%2 == 1 {
			p = 2
		}
		sum += digits[i] * p
	}
	remainder := sum % 9
	expected := 9 - remainder

	if digits[0] != expected {
		return fmt.Errorf("invalid check digit: expected %d, got %d", expected, digits[0])
	}
	return nil
}
