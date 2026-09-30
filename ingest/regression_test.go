package ingest

import (
	"context"
	"encoding/json"
	"github.com/digitalwayhk/pulse/cmc"
	"github.com/digitalwayhk/pulse/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAlertRefreshNeverUsesStaleOrDemoQuotes(t *testing.T) {
	sample, _ := cmc.SampleListings()
	coins := validCoins(sample.Data)
	if len(coins) < 2 {
		t.Fatal("need two fixtures")
	}
	now := time.Now().UTC().Add(-time.Hour)
	if err := persistCoins(coins, now, "cmc"); err != nil {
		t.Fatal(err)
	}
	for _, coin := range coins[:2] {
		rule := models.NewAlertRule()
		rule.UserID = "regression"
		rule.CmcID = coin.ID
		rule.Kind = models.AlertPriceAbove
		rule.Threshold = 0.01
		if err := rule.Insert(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = rule.Delete() })
	}
	for _, allow := range []string{"true", "false"} {
		t.Setenv("PULSE_ALLOW_SAMPLE", allow)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"status":{"error_code":1001,"error_message":"denied"}}`))
		}))
		result, err := RefreshQuotesAndEvaluateWithClient(context.Background(), &cmc.Client{BaseURL: server.URL, APIKey: "stub-key", HTTP: server.Client()}, []int{coins[0].ID})
		server.Close()
		if err != nil || result.Hits != 0 || result.Live {
			t.Fatalf("failure evaluated: %#v %v", result, err)
		}
		stored, _ := models.NewCoinQuote().FindByCmcID(coins[0].ID)
		if stored.Source != "cmc" || !stored.FetchedAt.Equal(now) {
			t.Fatalf("failure overwrote snapshot: %#v", stored)
		}
	}
	for _, partial := range []bool{false, true} {
		response := &cmc.QuotesResponse{Data: map[string]json.RawMessage{}}
		if partial {
			raw, _ := json.Marshal(coins[0])
			response.Data["one"] = raw
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(response) }))
		result, err := RefreshQuotesAndEvaluateWithClient(context.Background(), &cmc.Client{BaseURL: server.URL, APIKey: "stub-key", HTTP: server.Client()}, []int{coins[0].ID, coins[1].ID})
		server.Close()
		expected := 0
		if partial {
			expected = 1
		}
		if err != nil || result.Hits != expected || result.Quotes != expected {
			t.Fatalf("partial=%v: %#v %v", partial, result, err)
		}
	}
}

func TestRunRecordsEachEndpointOutcome(t *testing.T) {
	listings, _ := cmc.SampleListings()
	quotes, _ := cmc.SampleQuotes()
	global, _ := cmc.SampleGlobal()
	fail := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail && (r.URL.Path == cmc.QuotesLatestPath || r.URL.Path == cmc.GlobalMetricsPath) {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"status":{"error_code":1001}}`))
			return
		}
		switch r.URL.Path {
		case cmc.ListingsLatestPath:
			_ = json.NewEncoder(w).Encode(listings)
		case cmc.QuotesLatestPath:
			_ = json.NewEncoder(w).Encode(quotes)
		case cmc.GlobalMetricsPath:
			_ = json.NewEncoder(w).Encode(global)
		default:
			_, _ = w.Write([]byte(`{"data":{}}`))
		}
	}))
	defer server.Close()
	client := &cmc.Client{BaseURL: server.URL, APIKey: "stub-key", HTTP: server.Client()}
	if status := runOnceWithClient(context.Background(), client); !status.OK {
		t.Fatalf("success: %#v", status)
	}
	fail = true
	status := runOnceWithClient(context.Background(), client)
	if status.OK || status.AlertHits != 0 || len(status.Endpoints) != 4 {
		t.Fatalf("failure: %#v", status)
	}
	runs, err := models.NewIngestRun().QueryRecent(100)
	if err != nil {
		t.Fatal(err)
	}
	latest := models.LatestByEndpoint(runs)
	for endpoint, want := range map[string]string{cmc.ListingsLatestPath: "ok", cmc.QuotesLatestPath: "error", cmc.GlobalMetricsPath: "error", cmc.KeyInfoPath: "ok"} {
		if latest[endpoint] == nil || latest[endpoint].Status != want {
			t.Fatalf("%s: %#v", endpoint, latest[endpoint])
		}
	}
}

func TestSampleCannotReplaceLiveQuote(t *testing.T) {
	sample, _ := cmc.SampleListings()
	coin := sample.Data[0]
	now := time.Now().UTC()
	if err := persistCoins([]cmc.ListingCoin{coin}, now, "cmc"); err != nil {
		t.Fatal(err)
	}
	coin.Symbol = "DEMO"
	if err := persistCoins([]cmc.ListingCoin{coin}, now.Add(time.Minute), "sample"); err != nil {
		t.Fatal(err)
	}
	stored, err := models.NewCoinQuote().FindByCmcID(coin.ID)
	if err != nil || stored.Source != "cmc" || stored.Symbol == "DEMO" {
		t.Fatalf("demo replaced live quote: %#v %v", stored, err)
	}
}

func TestValidCoinsRejectsMissingUSD(t *testing.T) {
	coins := validCoins([]cmc.ListingCoin{{ID: 1}, {ID: 2, Quote: map[string]cmc.QuoteUSD{"USD": {Price: 2}}}})
	if len(coins) != 1 || coins[0].ID != 2 {
		t.Fatalf("invalid quotes accepted: %v", coins)
	}
}
