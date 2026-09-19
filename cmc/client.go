package cmc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/digitalwayhk/pulse/contract"
)

const (
	// DefaultBaseURL 是 CoinMarketCap Pro API 官方主机。
	DefaultBaseURL = "https://pro-api.coinmarketcap.com"
	// ListingsLatestPath 是 Startup / Basic 可用的最新榜单端点。
	ListingsLatestPath = "/v1/cryptocurrency/listings/latest"
	// QuotesLatestPath 是按 ID/代码取最新报价的端点。
	QuotesLatestPath = "/v1/cryptocurrency/quotes/latest"
	// GlobalMetricsPath 是全球市场指标端点。
	GlobalMetricsPath = "/v1/global-metrics/quotes/latest"
	// KeyInfoPath 用于展示套餐额度，不回传密钥。
	KeyInfoPath = "/v1/key/info"
)

// ErrMissingAPIKey 表示未配置真实 CMC 密钥。
var ErrMissingAPIKey = errors.New("CMC_API_KEY is missing or is a placeholder. Set a real CoinMarketCap Pro API key from https://pro.coinmarketcap.com/account")

// Client 以 X-CMC_PRO_API_KEY 调用官方 Pro API。
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

// NewFromEnv 从环境变量构造客户端；占位密钥会保留以便调用时给出明确错误。
func NewFromEnv() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		APIKey:  strings.TrimSpace(os.Getenv(contract.CMCAPIKeyEnv)),
		HTTP:    &http.Client{Timeout: 15 * time.Second},
	}
}

// KeyConfigured 判断环境变量是否看起来像真实密钥。
func (c *Client) KeyConfigured() bool {
	return !IsPlaceholderKey(c.APIKey)
}

// MaskedKey 返回打码后的密钥，永不输出原文。
func (c *Client) MaskedKey() string {
	key := strings.TrimSpace(c.APIKey)
	if key == "" {
		return ""
	}
	if IsPlaceholderKey(key) {
		return "placeholder"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "…" + key[len(key)-4:]
}

// IsPlaceholderKey 识别空值和常见占位字符串。
func IsPlaceholderKey(key string) bool {
	value := strings.ToLower(strings.TrimSpace(key))
	switch value {
	case "", "your-key-here", "changeme", "placeholder", "xxx", "test", "demo", "cmc_api_key", "your_cmc_api_key":
		return true
	default:
		return false
	}
}

// ListingsLatest 拉取按市值排序的最新榜单。
func (c *Client) ListingsLatest(ctx context.Context, start, limit int) (*ListingsResponse, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	if start <= 0 {
		start = 1
	}
	query := url.Values{}
	query.Set("start", fmt.Sprintf("%d", start))
	query.Set("limit", fmt.Sprintf("%d", limit))
	query.Set("convert", "USD")
	var out ListingsResponse
	if err := c.get(ctx, ListingsLatestPath, query, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// QuotesLatest 按 CMC ID 批量取最新报价。
func (c *Client) QuotesLatest(ctx context.Context, ids []int) (*QuotesResponse, error) {
	if len(ids) == 0 {
		return &QuotesResponse{}, nil
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			parts = append(parts, fmt.Sprintf("%d", id))
		}
	}
	query := url.Values{}
	query.Set("id", strings.Join(parts, ","))
	query.Set("convert", "USD")
	var out QuotesResponse
	if err := c.get(ctx, QuotesLatestPath, query, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// QuotesLatestBySymbol 按交易代码取最新报价，供点选详情在仅有 symbol 时使用。
func (c *Client) QuotesLatestBySymbol(ctx context.Context, symbol string) (*QuotesResponse, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return &QuotesResponse{}, nil
	}
	query := url.Values{}
	query.Set("symbol", symbol)
	query.Set("convert", "USD")
	var out QuotesResponse
	if err := c.get(ctx, QuotesLatestPath, query, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GlobalMetrics 拉取全球市场指标。
func (c *Client) GlobalMetrics(ctx context.Context) (*GlobalResponse, error) {
	query := url.Values{}
	query.Set("convert", "USD")
	var out GlobalResponse
	if err := c.get(ctx, GlobalMetricsPath, query, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// KeyInfo 读取套餐额度，失败时不阻断行情采集。
func (c *Client) KeyInfo(ctx context.Context) (*KeyInfoResponse, error) {
	var out KeyInfoResponse
	if err := c.get(ctx, KeyInfoPath, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, dest any) error {
	if c == nil || !c.KeyConfigured() {
		return ErrMissingAPIKey
	}
	endpoint := strings.TrimRight(c.BaseURL, "/") + path
	if query != nil {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-CMC_PRO_API_KEY", c.APIKey)
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("CMC request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	var envelope struct {
		Status Status `json:"status"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("CMC returned HTTP %d and a non-JSON body", resp.StatusCode)
	}
	if envelope.Status.ErrorCode != 0 || resp.StatusCode >= 400 {
		msg := strings.TrimSpace(envelope.Status.ErrorMessage)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return fmt.Errorf("CMC %s failed: %s", path, msg)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode CMC %s: %w", path, err)
	}
	return nil
}
