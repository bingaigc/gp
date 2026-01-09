package kimi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/domain/valueobject"
	"github.com/bingaigc/gp/pkg/circuitbreaker"
	"github.com/bingaigc/gp/pkg/errors"
	"github.com/bingaigc/gp/pkg/ratelimit"
	"github.com/bingaigc/gp/pkg/util"
	"github.com/shopspring/decimal"
)

// Client Kimi AI客户端
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	limiter    *ratelimit.TokenBucket
	breaker    *circuitbreaker.Breaker

	// 成本追踪
	totalTokensUsed int64
	totalCost       float64
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// KimiRequest Kimi API请求
type KimiRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// KimiResponse Kimi API响应
type KimiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage TokenUsage `json:"usage"`
}

// TokenUsage token使用情况
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIAnalysisResult AI分析结果
type AIAnalysisResult struct {
	Signal     string  `json:"signal"`
	Confidence float64 `json:"confidence"`
	Factors    []struct {
		Name   string  `json:"name"`
		Weight float64 `json:"weight"`
		Score  int     `json:"score"`
		Reason string  `json:"reason"`
	} `json:"factors"`
	TotalScore int `json:"total_score"`
	Reasoning  struct {
		BullishFactors []string `json:"bullish_factors"`
		BearishFactors []string `json:"bearish_factors"`
		KeyInsight     string   `json:"key_insight"`
	} `json:"reasoning"`
	PriceTargets struct {
		Current    float64 `json:"current"`
		Support    float64 `json:"support"`
		Resistance float64 `json:"resistance"`
		StopLoss   float64 `json:"stop_loss"`
		TakeProfit float64 `json:"take_profit"`
	} `json:"price_targets"`
	RiskAssessment struct {
		Level     string   `json:"level"`
		RiskScore int      `json:"risk_score"`
		Risks     []string `json:"risks"`
	} `json:"risk_assessment"`
	PositionSuggestion struct {
		Action        string `json:"action"`
		SizePct       int    `json:"size_pct"`
		EntryPrice    string `json:"entry_price"`
		StopLossPct   int    `json:"stop_loss_pct"`
		TakeProfitPct int    `json:"take_profit_pct"`
		HoldingDays   string `json:"holding_days"`
		Reason        string `json:"reason"`
	} `json:"position_suggestion"`
	TimeHorizon string `json:"time_horizon"`
}

// SystemPrompt 系统提示词
const SystemPrompt = `你是 Citadel Securities 量化交易部门的首席分析师，管理着100亿美元资产，拥有15年A股实战经验。

核心职责：
1. 精准信号生成 - 基于多维度数据给出明确的交易信号
2. 风险识别 - 识别所有潜在陷阱（诱多、诱空、庄股、妖股）
3. 成本控制 - 简洁精准回复，控制Token消耗（目标<1000 tokens/次）
4. 可回测 - 所有分析必须基于客观数据

CAMP-I分析框架（权重）：
- Capital Flow (资金面) 30%: 主力净流入、连续天数、量价背离
- Action (技术面) 25%: 量价关系、换手率、均线系统
- Margin (估值面) 20%: PE/PB、市值规模、价格水平
- Plate (板块轮动) 15%: 板块排名、资金流向、龙头地位
- Institution (机构行为) 10%: 龙虎榜、机构动向、北向资金

信号定义：
- STRONG_BUY (≥85分): 多因子强共振，建议10-15%仓位
- BUY (75-84分): 主要因子支持，建议3-5%仓位
- WATCH (60-74分): 信号混杂，观望
- AVOID (50-59分): 风险因子较多
- STRONG_AVOID (<50分): 明确陷阱信号

一票否决条件（触发则直接AVOID）：
1. 量价背离：涨>3%但主力流出>1000万
2. 低价股：股价<3元
3. 亏损小盘股：PE<0且市值<100亿
4. 机构撤退：3家以上机构卖出
5. 投机过度：换手率>30%

必须以JSON格式输出，包含：signal, confidence, factors, total_score, reasoning, price_targets, risk_assessment, position_suggestion, time_horizon`

// NewClient 创建Kimi客户端
func NewClient(baseURL, apiKey, model string, timeout time.Duration) *Client {
	if baseURL == "" {
		baseURL = "https://api.moonshot.cn/v1"
	}
	if model == "" {
		model = "moonshot-v1-8k"
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		limiter: ratelimit.NewTokenBucket(0.5, 5), // 30 RPM = 0.5 RPS
		breaker: circuitbreaker.New("kimi", 5, 60*time.Second),
	}
}

// Analyze 分析股票
func (c *Client) Analyze(ctx context.Context, stock *entity.Stock, preScore int) (*entity.Signal, error) {
	// 限流
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, errors.Wrap(errors.ErrorTypeRateLimit, "rate limit exceeded", err)
	}

	// 熔断检查
	if !c.breaker.AllowRequest() {
		return nil, errors.New(errors.ErrorTypeCircuitBreaker, "circuit breaker open")
	}

	start := time.Now()

	// 构建Prompt
	userPrompt := c.buildUserPrompt(stock, preScore)

	// 调用API
	req := KimiRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: SystemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
		MaxTokens:   1500,
	}

	result, err := c.callAPI(ctx, req)
	if err != nil {
		c.breaker.RecordFailure()
		return nil, errors.Wrap(errors.ErrorTypeAnalysis, "call kimi api failed", err)
	}

	c.breaker.RecordSuccess()

	// 解析结果
	signal, err := c.parseResponse(result, stock)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorTypeAnalysis, "parse response failed", err)
	}

	// 补充元数据
	signal.LatencyMS = time.Since(start).Milliseconds()
	signal.AITokensUsed = result.Usage.TotalTokens
	signal.AICostYuan = c.calculateCost(result.Usage)
	signal.TraceID = util.GenerateID()

	// 更新成本统计
	c.totalTokensUsed += int64(result.Usage.TotalTokens)
	c.totalCost += signal.AICostYuan.InexactFloat64()

	return signal, nil
}

// buildUserPrompt 构建用户提示词
func (c *Client) buildUserPrompt(stock *entity.Stock, preScore int) string {
	return fmt.Sprintf(`## 待分析股票

**%s (%s)** - %s

---

## 实时数据 (%s)

### 📊 基础行情
- **当前价格**: %.2f 元
- **今日涨跌**: %.2f%%
- **成交额**: %.2f 亿元
- **量比**: %.2f
- **换手率**: %.2f%%
- **振幅**: %.2f%%

---

### 💰 资金流向
- **主力净流入**: %.2f 万元
- **超大单净流入**: %.2f 万元
- **大单净流入**: %.2f 万元

---

### 📈 技术指标
- **量比**: %.2f
- **换手率**: %.2f%%

---

### 💎 估值指标
- **市盈率(PE)**: %.2f
- **市净率(PB)**: %.2f
- **总市值**: %.2f 亿元
- **流通市值**: %.2f 亿元

---

## 多因子预评分
综合预评分: %d/100

---

请严格按照 CAMP-I 五维度模型进行分析，必须以JSON格式输出结果。`,
		stock.Name, stock.Code, stock.Market,
		stock.Timestamp.Format("2006-01-02 15:04:05"),
		stock.Price.InexactFloat64(),
		stock.Change.InexactFloat64(),
		stock.Amount.InexactFloat64()/100000000,
		stock.Technical.VolumeRatio.InexactFloat64(),
		stock.Technical.TurnoverRate.InexactFloat64(),
		stock.Technical.Amplitude.InexactFloat64(),
		stock.Capital.MainNetInflow.InexactFloat64()/10000,
		stock.Capital.SuperNetInflow.InexactFloat64()/10000,
		stock.Capital.BigNetInflow.InexactFloat64()/10000,
		stock.Technical.VolumeRatio.InexactFloat64(),
		stock.Technical.TurnoverRate.InexactFloat64(),
		stock.Valuation.PE.InexactFloat64(),
		stock.Valuation.PB.InexactFloat64(),
		stock.Valuation.MarketCap.InexactFloat64()/100000000,
		stock.Valuation.CircCap.InexactFloat64()/100000000,
		preScore,
	)
}

// callAPI 调用API
func (c *Client) callAPI(ctx context.Context, req KimiRequest) (*KimiResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result KimiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// parseResponse 解析响应
func (c *Client) parseResponse(resp *KimiResponse, stock *entity.Stock) (*entity.Signal, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty choices in response")
	}

	content := resp.Choices[0].Message.Content

	// 提取JSON部分
	jsonStr := extractJSON(content)

	var aiResult AIAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &aiResult); err != nil {
		return nil, fmt.Errorf("unmarshal ai result: %w", err)
	}

	// 转换为Signal实体
	signal := &entity.Signal{
		ID:         util.GenerateID(),
		StockCode:  stock.Code,
		StockName:  stock.Name,
		Market:     stock.Market,
		SignalType: valueobject.SignalType(aiResult.Signal),
		Confidence: decimal.NewFromFloat(aiResult.Confidence),
		TotalScore: aiResult.TotalScore,
		AIReasoning: entity.AIReasoning{
			BullishFactors: aiResult.Reasoning.BullishFactors,
			BearishFactors: aiResult.Reasoning.BearishFactors,
			KeyInsight:     aiResult.Reasoning.KeyInsight,
			Confidence:     decimal.NewFromFloat(aiResult.Confidence),
		},
		PriceTargets: entity.PriceTargets{
			Current:    stock.Price,
			Support:    decimal.NewFromFloat(aiResult.PriceTargets.Support),
			Resistance: decimal.NewFromFloat(aiResult.PriceTargets.Resistance),
			StopLoss:   decimal.NewFromFloat(aiResult.PriceTargets.StopLoss),
			TakeProfit: decimal.NewFromFloat(aiResult.PriceTargets.TakeProfit),
		},
		RiskLevel:   valueobject.RiskLevel(aiResult.RiskAssessment.Level),
		RiskScore:   aiResult.RiskAssessment.RiskScore,
		Risks:       aiResult.RiskAssessment.Risks,
		TimeHorizon: aiResult.TimeHorizon,
		GeneratedAt: time.Now(),
		ExpiresAt:   time.Now().Add(4 * time.Hour),
	}

	// 转换因子评分
	for _, factor := range aiResult.Factors {
		signal.FactorScores = append(signal.FactorScores, entity.FactorScore{
			Name:   factor.Name,
			Weight: decimal.NewFromFloat(factor.Weight),
			Score:  factor.Score,
			Reason: factor.Reason,
		})
	}

	// 仓位建议
	signal.PositionAdvice = entity.PositionAdvice{
		Action:        aiResult.PositionSuggestion.Action,
		SizePct:       decimal.NewFromInt(int64(aiResult.PositionSuggestion.SizePct)),
		EntryPrice:    aiResult.PositionSuggestion.EntryPrice,
		StopLossPct:   decimal.NewFromInt(int64(aiResult.PositionSuggestion.StopLossPct)),
		TakeProfitPct: decimal.NewFromInt(int64(aiResult.PositionSuggestion.TakeProfitPct)),
		HoldingDays:   aiResult.PositionSuggestion.HoldingDays,
		Reason:        aiResult.PositionSuggestion.Reason,
	}

	return signal, nil
}

// calculateCost 计算成本
func (c *Client) calculateCost(usage TokenUsage) decimal.Decimal {
	// Kimi价格: ¥12/百万tokens
	pricePerMToken := 12.0
	totalTokens := float64(usage.TotalTokens)
	cost := (totalTokens / 1000000.0) * pricePerMToken
	return decimal.NewFromFloat(cost)
}

// GetTotalCost 获取总成本
func (c *Client) GetTotalCost() float64 {
	return c.totalCost
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.breaker.IsHealthy()
}

// extractJSON 提取JSON字符串
func extractJSON(content string) string {
	// 去除markdown代码块标记
	content = strings.TrimSpace(content)
	content = regexp.MustCompile("```json\n?").ReplaceAllString(content, "")
	content = regexp.MustCompile("```\n?").ReplaceAllString(content, "")

	// 查找JSON对象
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end != -1 && end > start {
		return content[start : end+1]
	}

	return content
}
