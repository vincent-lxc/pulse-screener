package cmc

import (
	_ "embed"
	"encoding/json"
	"os"
	"strings"

	"github.com/digitalwayhk/pulse/contract"
)

//go:embed testdata/sample-listings.json
var sampleListings []byte

//go:embed testdata/sample-global.json
var sampleGlobal []byte

//go:embed testdata/sample-quotes.json
var sampleQuotes []byte

// AllowSample 在缺少真实密钥时是否允许离线样例。默认允许，便于无密钥演示。
func AllowSample() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(contract.AllowSampleEnv)))
	if value == "0" || value == "false" || value == "no" {
		return false
	}
	return true
}

// SampleListings 返回嵌入的脱敏 listings 样例。
func SampleListings() (*ListingsResponse, error) {
	var out ListingsResponse
	if err := json.Unmarshal(sampleListings, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SampleGlobal 返回嵌入的脱敏全球指标样例。
func SampleGlobal() (*GlobalResponse, error) {
	var out GlobalResponse
	if err := json.Unmarshal(sampleGlobal, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SampleQuotes 返回嵌入的脱敏 quotes/latest 样例。
func SampleQuotes() (*QuotesResponse, error) {
	var out QuotesResponse
	if err := json.Unmarshal(sampleQuotes, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
