package public

import "testing"

func TestGetQuoteRequiresIDOrSymbol(t *testing.T) {
	api := &GetQuote{}
	if err := api.Validation(nil); err == nil {
		t.Fatal("expected validation error when id and symbol are empty")
	}
	api.CmcID = 1
	if err := api.Validation(nil); err != nil {
		t.Fatalf("id should be enough: %v", err)
	}
	api = &GetQuote{Symbol: "BTC"}
	if err := api.Validation(nil); err != nil {
		t.Fatalf("symbol should be enough: %v", err)
	}
}
