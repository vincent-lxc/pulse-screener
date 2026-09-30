package cmc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// This synthetic fixture follows the official Key Info example, not an account response.
func TestKeyInfoDecodesOfficialPlanShape(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample-key-info.json")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != KeyInfoPath {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write(raw)
	}))
	defer server.Close()
	client := &Client{BaseURL: server.URL, APIKey: "fixture-key", HTTP: server.Client()}
	info, err := client.KeyInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	plan := info.Data.Plan
	if plan.CreditLimitMonthly != 120000 || plan.RateLimitMinute != 60 || plan.CreditLimitMonthlyReset != "In 3 days, 19 hours, 56 minutes" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestLiveKeyInfoSkipWithoutKey(t *testing.T) {
	client := NewFromEnv()
	if !client.KeyConfigured() {
		t.Skip("CMC_API_KEY not set; skipping live Key Info")
	}
	info, err := client.KeyInfo(context.Background())
	// Keep account details and upstream error text out of test output.
	if err != nil {
		t.Fatal("live Key Info request or decoding failed")
	}
	if info.Data.Plan.CreditLimitMonthly <= 0 || info.Data.Plan.RateLimitMinute <= 0 || info.Data.Plan.CreditLimitMonthlyReset == "" {
		t.Fatal("live Key Info plan fields missing")
	}
	t.Log("official Key Info request and plan decoding succeeded")
}
