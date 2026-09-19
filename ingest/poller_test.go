package ingest

import (
	"testing"

	"github.com/digitalwayhk/pulse/cmc"
)

func TestQuoteIDsMergesWatchlistAndTopListings(t *testing.T) {
	listings := &cmc.ListingsResponse{Data: []cmc.ListingCoin{
		{ID: 1}, {ID: 1027}, {ID: 52}, {ID: 1839}, {ID: 5426}, {ID: 3408},
	}}
	ids := quoteIDs(listings, []int{52, 1, 0})
	if len(ids) != 5 {
		t.Fatalf("expected 5 unique IDs, got %v", ids)
	}
	if ids[0] != 52 || ids[1] != 1 {
		t.Fatalf("watchlist IDs should lead: %v", ids)
	}
}

func TestMissingKeyMessageIsStable(t *testing.T) {
	if cmc.ErrMissingAPIKey == nil || cmc.ErrMissingAPIKey.Error() == "" {
		t.Fatal("missing key error must be a clear public message")
	}
}
