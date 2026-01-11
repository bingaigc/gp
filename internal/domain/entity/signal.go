package entity

import (
	"time"

	"github.com/bingaigc/gp/internal/domain/valueobject"
	"github.com/shopspring/decimal"
)

// Signal 交易信号实体
type Signal struct {
	ID string `json:"id"`

	// 股票信息
	StockCode string `json:"stock_code"`
	StockName string `json:"stock_name"`
	Market    string `json:"market"`

	// 信号信息
	SignalType valueobject.SignalType `json:"signal_type"`
	Confidence decimal.Decimal        `json:"confidence"`

	// 评分信息
	FactorScores []FactorScore `json:"factor_scores"`
	TotalScore   int           `json:"total_score"`

	// AI分析
	AIReasoning AIReasoning `json:"ai_reasoning"`

	// 价格目标
	PriceTargets PriceTargets `json:"price_targets"`

	// 风险信息
	RiskLevel valueobject.RiskLevel `json:"risk_level"`
	RiskScore int                   `json:"risk_score"`
	Risks     []string              `json:"risks,omitempty"`

	// 仓位建议
	PositionAdvice PositionAdvice `json:"position_advice"`

	// 时效性
	TimeHorizon string    `json:"time_horizon"`
	GeneratedAt time.Time `json:"generated_at"`
	ExpiresAt   time.Time `json:"expires_at"`

	// 追踪
	TraceID   string `json:"trace_id"`
	LatencyMS int64  `json:"latency_ms"`

	// 成本
	AITokensUsed int             `json:"ai_tokens_used"`
	AICostYuan   decimal.Decimal `json:"ai_cost_yuan"`
}

// FactorScore 因子评分
type FactorScore struct {
	Name   string          `json:"name"`
	Weight decimal.Decimal `json:"weight"`
	Score  int             `json:"score"`
	Reason string          `json:"reason"`
}

// AIReasoning AI推理结果
type AIReasoning struct {
	BullishFactors []string        `json:"bullish_factors"`
	BearishFactors []string        `json:"bearish_factors"`
	KeyInsight     string          `json:"key_insight"`
	Confidence     decimal.Decimal `json:"confidence"`
}

// PriceTargets 价格目标
type PriceTargets struct {
	Current    decimal.Decimal `json:"current"`
	Support    decimal.Decimal `json:"support"`
	Resistance decimal.Decimal `json:"resistance"`
	StopLoss   decimal.Decimal `json:"stop_loss"`
	TakeProfit decimal.Decimal `json:"take_profit"`
}

// PositionAdvice 仓位建议
type PositionAdvice struct {
	Action        string          `json:"action"`          // BUY/HOLD/SELL
	SizePct       decimal.Decimal `json:"size_pct"`        // 建议仓位%
	EntryPrice    string          `json:"entry_price"`     // 入场价格区间
	StopLossPct   decimal.Decimal `json:"stop_loss_pct"`   // 止损%
	TakeProfitPct decimal.Decimal `json:"take_profit_pct"` // 止盈%
	HoldingDays   string          `json:"holding_days"`    // 持有天数
	Reason        string          `json:"reason"`          // 建议理由
}

// IsValid 验证信号有效性
func (s *Signal) IsValid() bool {
	return time.Now().Before(s.ExpiresAt)
}

// ShouldAlert 是否需要告警
func (s *Signal) ShouldAlert() bool {
	return s.SignalType == valueobject.SignalStrongBuy ||
		s.SignalType == valueobject.SignalStrongAvoid
}

// IsActionable 是否可操作
func (s *Signal) IsActionable() bool {
	return s.SignalType.IsBullish() && s.IsValid()
}
