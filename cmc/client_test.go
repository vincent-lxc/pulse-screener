package cmc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsPlaceholderKey(t *testing.T) {
	if !IsPlaceholderKey("your-key-here") || !IsPlaceholderKey("") {
		t.Fatal("expected placeholders to be rejected")
	}
	if IsPlaceholderKey("b54bcf4d-1bca-4e8e-9a77-97example000") {
		t.Fatal("did not expect a UUID-like key to be treated as placeholder")
	}
}

func TestClientRejectsPlaceholderBeforeNetwork(t *testing.T) {
	client := &Client{BaseURL: "http://127.0.0.1:1", APIKey: "placeholder", HTTP: http.DefaultClient}
	_, err := client.ListingsLatest(context.Background(), 1, 10)
	if err == nil || err != ErrMissingAPIKey {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestClientParsesListingsAndSurfacesCMCError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(ListingsLatestPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CMC_PRO_API_KEY") != "live-key" {
			t.Errorf("missing API key header")
		}
		_, _ = w.Write(sampleListings)
	})
	mux.HandleFunc(GlobalMetricsPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": map[string]any{"error_code": 1001, "error_message": "This API Key is invalid."},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "live-key", HTTP: server.Client()}
	listings, err := client.ListingsLatest(context.Background(), 1, 5)
	if err != nil || listings == nil || len(listings.Data) != 5 || listings.Data[0].Symbol != "BTC" {
		t.Fatalf("listings parse failed: %#v %v", listings, err)
	}
	_, err = client.GlobalMetrics(context.Background())
	if err == nil || err.Error() == "" {
		t.Fatal("expected CMC error to surface")
	}
}

func TestQuotesLatestBySymbol(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(QuotesLatestPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("symbol") != "BTC" {
			t.Errorf("symbol=%q", r.URL.Query().Get("symbol"))
		}
		_, _ = w.Write(sampleQuotes)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := &Client{BaseURL: server.URL, APIKey: "live-key", HTTP: server.Client()}
	quotes, err := client.QuotesLatestBySymbol(context.Background(), "btc")
	if err != nil || len(quotes.Coins()) != 1 || quotes.Coins()[0].Symbol != "BTC" {
		t.Fatalf("by symbol: %#v %v", quotes, err)
	}
}

func TestQuotesLatestBySymbolRejectsPlaceholder(t *testing.T) {
	client := &Client{BaseURL: "http://127.0.0.1:1", APIKey: "placeholder", HTTP: http.DefaultClient}
	if _, err := client.QuotesLatestBySymbol(context.Background(), "BTC"); err != ErrMissingAPIKey {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestSamplePayloads(t *testing.T) {
	listings, err := SampleListings()
	if err != nil || len(listings.Data) == 0 {
		t.Fatalf("sample listings: %v", err)
	}
	global, err := SampleGlobal()
	if err != nil || global.Data.BtcDominance == 0 {
		t.Fatalf("sample global: %v", err)
	}
	quotes, err := SampleQuotes()
	if err != nil || len(quotes.Coins()) != 1 {
		t.Fatalf("sample quotes: %v", err)
	}
}
