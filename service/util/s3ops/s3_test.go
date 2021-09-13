package s3ops

import (
	"testing"
)

func TestGetReceiptHtmlTemplate(t *testing.T) {
	svc := InitSesh()
	tmpl, err := GetReceiptHtmlTemplate(svc)
	if err != nil {
		t.Errorf("FAIL: %v", err)
	}
	t.Logf("result: %s", tmpl)
}

func TestGetOrderNotificationHtmlTemplate(t *testing.T) {
	svc := InitSesh()
	tmpl, err := GetOrderNotificationHtmlTemplate(svc)
	if err != nil {
		t.Errorf("FAIL: %v", err)
	}
	t.Logf("result: %s", tmpl)
}
