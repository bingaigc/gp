package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
)

// WebhookSink Webhook输出
type WebhookSink struct {
	url        string
	secret     string
	httpClient *http.Client
	format     string // "dingtalk", "feishu", "generic"
}

// NewWebhookSink 创建Webhook输出
func NewWebhookSink(url, secret, format string) *WebhookSink {
	if format == "" {
		format = "generic"
	}

	return &WebhookSink{
		url:    url,
		secret: secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		format: format,
	}
}

// Emit 输出信号到Webhook
func (s *WebhookSink) Emit(ctx context.Context, signal *entity.Signal) error {
	var payload interface{}
	var err error

	switch s.format {
	case "dingtalk":
		payload = s.buildDingTalkPayload(signal)
	case "feishu":
		payload = s.buildFeishuPayload(signal)
	default:
		payload = signal
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.secret != "" {
		req.Header.Set("X-Secret", s.secret)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// buildDingTalkPayload 构建钉钉消息格式
func (s *WebhookSink) buildDingTalkPayload(signal *entity.Signal) map[string]interface{} {
	signalEmoji := getSignalEmoji(signal.SignalType)
	riskEmoji := getRiskEmoji(signal.RiskLevel)

	text := fmt.Sprintf(`## %s %s 交易信号

**%s (%s)** - %s

---

### 📊 信号详情
- 信号类型: %s
- 综合评分: %d/100
- 置信度: %.1f%%
- 风险等级: %s %s

### 💰 价格目标
- 当前价格: %.2f
- 止损价: %.2f
- 止盈价: %.2f

### 📋 仓位建议
- 建议操作: %s
- 建议仓位: %.1f%%
- 持有周期: %s

---

生成时间: %s
`,
		signalEmoji, signal.SignalType,
		signal.StockName, signal.StockCode, signal.Market,
		signal.SignalType,
		signal.TotalScore,
		signal.Confidence.InexactFloat64()*100,
		riskEmoji, signal.RiskLevel,
		signal.PriceTargets.Current.InexactFloat64(),
		signal.PriceTargets.StopLoss.InexactFloat64(),
		signal.PriceTargets.TakeProfit.InexactFloat64(),
		signal.PositionAdvice.Action,
		signal.PositionAdvice.SizePct.InexactFloat64(),
		signal.PositionAdvice.HoldingDays,
		signal.GeneratedAt.Format("2006-01-02 15:04:05"),
	)

	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": fmt.Sprintf("%s %s", signal.StockName, signal.SignalType),
			"text":  text,
		},
	}
}

// buildFeishuPayload 构建飞书消息格式
func (s *WebhookSink) buildFeishuPayload(signal *entity.Signal) map[string]interface{} {
	signalEmoji := getSignalEmoji(signal.SignalType)
	riskEmoji := getRiskEmoji(signal.RiskLevel)

	text := fmt.Sprintf(`%s %s 交易信号

%s (%s) - %s

📊 信号详情
信号类型: %s
综合评分: %d/100
置信度: %.1f%%
风险等级: %s %s

💰 价格目标
当前价格: %.2f
止损价: %.2f
止盈价: %.2f

📋 仓位建议
建议操作: %s
建议仓位: %.1f%%
持有周期: %s

生成时间: %s`,
		signalEmoji, signal.SignalType,
		signal.StockName, signal.StockCode, signal.Market,
		signal.SignalType,
		signal.TotalScore,
		signal.Confidence.InexactFloat64()*100,
		riskEmoji, signal.RiskLevel,
		signal.PriceTargets.Current.InexactFloat64(),
		signal.PriceTargets.StopLoss.InexactFloat64(),
		signal.PriceTargets.TakeProfit.InexactFloat64(),
		signal.PositionAdvice.Action,
		signal.PositionAdvice.SizePct.InexactFloat64(),
		signal.PositionAdvice.HoldingDays,
		signal.GeneratedAt.Format("2006-01-02 15:04:05"),
	)

	return map[string]interface{}{
		"msg_type": "text",
		"content": map[string]interface{}{
			"text": text,
		},
	}
}

// Close 关闭
func (s *WebhookSink) Close() error {
	return nil
}

// Name 名称
func (s *WebhookSink) Name() string {
	return "webhook"
}
