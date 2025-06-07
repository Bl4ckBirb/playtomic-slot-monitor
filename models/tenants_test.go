package models

import "testing"

func TestSearchTenantsParamsWithProperties(t *testing.T) {
	v := (&SearchTenantsParams{WithProperties: []string{"ALLOWS_CASH_PAYMENT", "LIVE_TV_URL"}}).ToURLValues()
	if got := v.Get("with_properties"); got != "ALLOWS_CASH_PAYMENT,LIVE_TV_URL" {
		t.Errorf("with_properties = %q", got)
	}
}
