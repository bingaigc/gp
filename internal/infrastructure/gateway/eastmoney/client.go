package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/pkg/errors"
	"github.com/bingaigc/gp/pkg/ratelimit"
	"github.com/shopspring/decimal"
)

// Client 东方财富数据客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
	limiter    *ratelimit.TokenBucket
}

// EastMoneyResponse 东方财富API响应
type EastMoneyResponse struct {
	RC   int           `json:"rc"`
	RT   int           `json:"rt"`
	Data EastMoneyData `json:"data"`
}

// EastMoneyData 数据部分
type EastMoneyData struct {
	Total int                      `json:"total"`
	Diff  []map[string]interface{} `json:"diff"`
}

// NewClient 创建东方财富客户端
func NewClient(baseURL string, timeout time.Duration) *Client {
	if baseURL == "" {
		baseURL = "https://push2.eastmoney.com"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		limiter: ratelimit.NewTokenBucket(10, 20), // 10 QPS, burst 20
	}
}

// FetchActiveStocks 获取活跃股票列表
func (c *Client) FetchActiveStocks(ctx context.Context, limit int) ([]*entity.Stock, error) {
	// 限流
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, errors.Wrap(errors.ErrorTypeRateLimit, "rate limit exceeded", err)
	}

	// 构建请求URL - 获取活跃股票（涨幅榜+成交额榜）
	url := fmt.Sprintf("%s/api/qt/clist/get?pn=1&pz=%d&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f62,f115,f128,f140,f141,f136,f152",
		c.baseURL, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorTypeDataFetch, "create request failed", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorTypeDataFetch, "request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errors.ErrorTypeDataFetch, fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	var result EastMoneyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(errors.ErrorTypeDataFetch, "decode response failed", err)
	}

	if result.RC != 0 {
		return nil, errors.New(errors.ErrorTypeDataFetch, fmt.Sprintf("API error: rc=%d", result.RC))
	}

	return c.parseStocks(result.Data.Diff), nil
}

// FetchByCode 获取单只股票
func (c *Client) FetchByCode(ctx context.Context, code string) (*entity.Stock, error) {
	// 限流
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, errors.Wrap(errors.ErrorTypeRateLimit, "rate limit exceeded", err)
	}

	// 构建请求URL
	market := "1" // 深圳
	if len(code) > 0 && code[0] == '6' {
		market = "0" // 上海
	}

	url := fmt.Sprintf("%s/api/qt/stock/get?secid=%s.%s&fields=f57,f58,f162,f163,f164,f165,f166,f167,f168,f169,f170,f171,f47,f48,f49,f50,f51,f52,f53,f54,f55,f56,f59,f60,f61,f62,f63,f64,f65,f66,f67,f68",
		c.baseURL, market, code)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorTypeDataFetch, "create request failed", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorTypeDataFetch, "request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errors.ErrorTypeDataFetch, fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(errors.ErrorTypeDataFetch, "decode response failed", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New(errors.ErrorTypeDataFetch, "invalid response data")
	}

	stocks := c.parseStocks([]map[string]interface{}{data})
	if len(stocks) == 0 {
		return nil, errors.New(errors.ErrorTypeDataFetch, "no stock data found")
	}

	return stocks[0], nil
}

// parseStocks 解析股票数据
func (c *Client) parseStocks(rawData []map[string]interface{}) []*entity.Stock {
	stocks := make([]*entity.Stock, 0, len(rawData))

	for _, item := range rawData {
		stock := &entity.Stock{
			Code:   getStringField(item, "f12"),
			Name:   getStringField(item, "f14"),
			Market: getMarket(getStringField(item, "f13")),
			Price:  decimal.NewFromFloat(getFloatField(item, "f2")),
			Change: decimal.NewFromFloat(getFloatField(item, "f3")),
			Volume: int64(getFloatField(item, "f5")),
			Amount: decimal.NewFromFloat(getFloatField(item, "f6")),
			Capital: entity.CapitalFlow{
				MainNetInflow:   decimal.NewFromFloat(getFloatField(item, "f62")),
				SuperNetInflow:  decimal.NewFromFloat(getFloatField(item, "f184") * 10000),
				BigNetInflow:    decimal.NewFromFloat(getFloatField(item, "f66")),
				MiddleNetInflow: decimal.Zero,
				SmallNetInflow:  decimal.Zero,
				ConsecutiveDays: 0,
			},
			Technical: entity.Technical{
				VolumeRatio:  decimal.NewFromFloat(getFloatField(item, "f10")),
				TurnoverRate: decimal.NewFromFloat(getFloatField(item, "f8")),
				Amplitude:    decimal.NewFromFloat(getFloatField(item, "f7")),
				MA5:          decimal.Zero,
				MA10:         decimal.Zero,
				MA20:         decimal.Zero,
			},
			Valuation: entity.Valuation{
				PE:        decimal.NewFromFloat(getFloatField(item, "f9")),
				PB:        decimal.NewFromFloat(getFloatField(item, "f23")),
				MarketCap: decimal.NewFromFloat(getFloatField(item, "f20")),
				CircCap:   decimal.NewFromFloat(getFloatField(item, "f21")),
			},
			Timestamp: time.Now(),
			Source:    "eastmoney",
		}

		// 过滤无效数据
		if stock.Code == "" || stock.Name == "" {
			continue
		}

		stocks = append(stocks, stock)
	}

	return stocks
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.FetchActiveStocks(ctx, 1)
	return err
}

// Name 数据源名称
func (c *Client) Name() string {
	return "eastmoney"
}

// Helper functions

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloatField(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		case string:
			if val == "-" {
				return 0
			}
		}
	}
	return 0
}

func getMarket(marketCode string) string {
	switch marketCode {
	case "0", "1":
		return "SH"
	case "90", "91":
		return "SZ"
	default:
		return "SH"
	}
}
