package htmlops

import (
	"testing"

	"github.com/tpillz-presents/service/util/s3ops"
)

func TestCreateHtmlTemplate(t *testing.T) {
	var tests = []ReceiptTemplateData{
		ReceiptTemplateData{
			OrderID:    "0000001",
			Subtotal:   19.99,
			SalesTax:   0.88,
			Shipping:   5.99,
			OrderTotal: 26.78,
			FirstName:  "Daniel",
			LastName:   "Garcia",
			Address1:   "3250 Hollis St",
			Address2:   "Apt 319",
			City:       "Oakland",
			State:      "CA",
			Zip:        "94608",
			Phone:      "209-534-0739",
			Items: []ItemSummary{
				ItemSummary{"PawnWars Chess Set", 2},
				ItemSummary{"PawnWars Sniper Poster", 1},
			},
		},
	}
	for _, test := range tests {
		tmpl, err := s3ops.GetReceiptHtmlTemplate(s3ops.InitSesh()) // test only
		if err != nil {
			t.Errorf("FAIL - template: %v", err)
		} else {
			t.Log("template OK")
		}
		html, err := CreateHtmlTemplate(tmpl, test)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
		t.Logf("result: %s", html)
	}
}
