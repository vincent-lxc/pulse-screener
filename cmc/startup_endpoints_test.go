package cmc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/digitalwayhk/pulse/contract"
)

func TestStartupTierEndpointsWithAPIKeyHeader(t *testing.T) {
	seen := map[string]bool{}
	mux := http.NewServeMux()
	mux.HandleFunc(ListingsLatestPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CMC_PRO_API_KEY") != "live-key" {
			t.Errorf("%s missing X-CMC_PRO_API_KEY", ListingsLatestPath)
		}
		if r.URL.Query().Get("convert") != "USD" {
			t.Errorf("listings convert=%q", r.URL.Query().Get("convert"))
		}
		seen[ListingsLatestPath] = true
		_, _ = w.Write(sampleListings)
	})
	mux.HandleFunc(QuotesLatestPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CMC_PRO_API_KEY") != "live-key" {
			t.Errorf("%s missing X-CMC_PRO_API_KEY", QuotesLatestPath)
		}
		if r.URL.Query().Get("id") == "" {
			t.Errorf("quotes missing id query")
		}
		seen[QuotesLatestPath] = true
		_, _ = w.Write(sampleQuotes)
	})
	mux.HandleFunc(GlobalMetricsPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CMC_PRO_API_KEY") != "live-key" {
			t.Errorf("%s missing X-CMC_PRO_API_KEY", GlobalMetricsPath)
		}
		seen[GlobalMetricsPath] = true
		_, _ = w.Write(sampleGlobal)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := &Client{BaseURL: server.URL, APIKey: "live-key", HTTP: server.Client()}
	ctx := context.Background()

	listings, err := client.ListingsLatest(ctx, 1, 100)
	if err != nil || len(listings.Data) == 0 {
		t.Fatalf("listings: %v", err)
	}
	quotes, err := client.QuotesLatest(ctx, []int{1})
	if err != nil {
		t.Fatalf("quotes: %v", err)
	}
	coins := quotes.Coins()
	if len(coins) != 1 || coins[0].Symbol != "BTC" {
		t.Fatalf("quotes coins: %#v", coins)
	}
	global, err := client.GlobalMetrics(ctx)
	if err != nil || global.Data.BtcDominance == 0 {
		t.Fatalf("global: %#v %v", global, err)
	}
	for _, path := range []string{ListingsLatestPath, QuotesLatestPath, GlobalMetricsPath} {
		if !seen[path] {
			t.Fatalf("did not call %s", path)
		}
	}
}

func TestMissingKeyFailsBeforeNetworkOnAllStartupEndpoints(t *testing.T) {
	client := NewFromEnv()
	if client.KeyConfigured() {
		t.Skip("CMC_API_KEY is set in this environment")
	}
	ctx := context.Background()
	if _, err := client.ListingsLatest(ctx, 1, 10); err != ErrMissingAPIKey {
		t.Fatalf("listings: %v", err)
	}
	if _, err := client.QuotesLatest(ctx, []int{1}); err != ErrMissingAPIKey {
		t.Fatalf("quotes: %v", err)
	}
	if _, err := client.GlobalMetrics(ctx); err != ErrMissingAPIKey {
		t.Fatalf("global: %v", err)
	}
}

func TestLiveStartupEndpointsSkipWithoutKey(t *testing.T) {
	if IsPlaceholderKey(os.Getenv(contract.CMCAPIKeyEnv)) {
		t.Skip("CMC_API_KEY not set; skipping live CoinMarketCap call")
	}
	client := NewFromEnv()
	ctx := context.Background()
	listings, err := client.ListingsLatest(ctx, 1, 5)
	if err != nil || listings == nil || len(listings.Data) == 0 {
		t.Fatalf("live listings: %v", err)
	}
	quotes, err := client.QuotesLatest(ctx, []int{listings.Data[0].ID})
	if err != nil || len(quotes.Coins()) == 0 {
		t.Fatalf("live quotes: %v", err)
	}
	if _, err := client.GlobalMetrics(ctx); err != nil {
		t.Fatalf("live global: %v", err)
	}
}

func TestQuotesResponseCoinsIgnoresBadEntries(t *testing.T) {
	raw, _ := json.Marshal(ListingCoin{ID: 1, Symbol: "BTC"})
	resp := &QuotesResponse{Data: map[string]json.RawMessage{
		"1": raw,
		"x": json.RawMessage(`{"id":0}`),
	}}
	if coins := resp.Coins(); len(coins) != 1 || coins[0].Symbol != "BTC" {
		t.Fatalf("got %#v", coins)
	}
}
