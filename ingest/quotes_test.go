package ingest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digitalwayhk/pulse/cmc"
)

func TestLookupQuoteHitsQuotesLatest(t *testing.T) {
	var sawPath, sawID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		sawID = r.URL.Query().Get("id")
		if r.Header.Get("X-CMC_PRO_API_KEY") != "live-key" {
			t.Errorf("missing X-CMC_PRO_API_KEY")
		}
		_, _ = w.Write(mustSampleQuotes(t))
	}))
	t.Cleanup(server.Close)

	client := &cmc.Client{BaseURL: server.URL, APIKey: "live-key", HTTP: server.Client()}
	lookup, err := LookupQuoteWithClient(context.Background(), client, 1, "BTC")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if sawPath != cmc.QuotesLatestPath || sawID != "1" {
		t.Fatalf("expected quotes/latest?id=1, got %s ?id=%s", sawPath, sawID)
	}
	if lookup == nil || !lookup.Live || lookup.Endpoint != cmc.QuotesLatestPath {
		t.Fatalf("expected live quotes/latest result, got %#v", lookup)
	}
	if lookup.Coin == nil || lookup.Coin.Symbol != "BTC" {
		t.Fatalf("expected BTC coin, got %#v", lookup.Coin)
	}
}

func TestRefreshQuotesAndEvaluateHitsQuotesLatest(t *testing.T) {
	var sawPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write(mustSampleQuotes(t))
	}))
	t.Cleanup(server.Close)

	client := &cmc.Client{BaseURL: server.URL, APIKey: "live-key", HTTP: server.Client()}
	result, err := RefreshQuotesAndEvaluateWithClient(context.Background(), client, []int{1})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if sawPath != cmc.QuotesLatestPath {
		t.Fatalf("expected %s, got %s", cmc.QuotesLatestPath, sawPath)
	}
	if result == nil || !result.Live || result.Quotes < 1 {
		t.Fatalf("expected live quote confirm, got %#v", result)
	}
}

func TestLookupQuoteFallsBackWithoutKey(t *testing.T) {
	client := &cmc.Client{BaseURL: "http://127.0.0.1:1", APIKey: "placeholder", HTTP: http.DefaultClient}
	lookup, err := LookupQuoteWithClient(context.Background(), client, 1, "BTC")
	if err != nil {
		t.Fatalf("fallback should serve sample/local: %v", err)
	}
	if lookup.Live {
		t.Fatal("placeholder key must not be reported as live CMC")
	}
	if lookup.Endpoint != cmc.QuotesLatestPath {
		t.Fatalf("endpoint should still be quotes/latest, got %s", lookup.Endpoint)
	}
}

func mustSampleQuotes(t *testing.T) []byte {
	t.Helper()
	quotes, err := cmc.SampleQuotes()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(quotes)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
