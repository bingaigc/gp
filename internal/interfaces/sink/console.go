package sink

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/domain/valueobject"
)

// ConsoleSink 控制台输出
type ConsoleSink struct {
	pretty bool
}

// NewConsoleSink 创建控制台输出
func NewConsoleSink(pretty bool) *ConsoleSink {
	return &ConsoleSink{
		pretty: pretty,
	}
}

// Emit 输出信号
func (s *ConsoleSink) Emit(ctx context.Context, signal *entity.Signal) error {
	if s.pretty {
		return s.emitPretty(signal)
	}
	return s.emitJSON(signal)
}

// emitPretty 美化输出
func (s *ConsoleSink) emitPretty(signal *entity.Signal) error {
	fmt.Println("\n" + "═══════════════════════════════════════════════════════════")
	fmt.Printf("📊 %s (%s) - %s\n", signal.StockName, signal.StockCode, signal.Market)
	fmt.Println("═══════════════════════════════════════════════════════════")

	// 信号类型
	signalEmoji := getSignalEmoji(signal.SignalType)
	fmt.Printf("\n🎯 信号类型: %s %s\n", signalEmoji, signal.SignalType)
	fmt.Printf("📈 综合评分: %d/100\n", signal.TotalScore)
	fmt.Printf("🎲 置信度: %.2f%%\n", signal.Confidence.InexactFloat64()*100)

	// 因子评分
	fmt.Println("\n📊 因子评分:")
	for _, factor := range signal.FactorScores {
		weight := factor.Weight.InexactFloat64() * 100
		fmt.Printf("  • %s (权重%.0f%%): %d/100 - %s\n",
			factor.Name, weight, factor.Score, factor.Reason)
	}

	// AI分析
	fmt.Println("\n🤖 AI分析:")
	if len(signal.AIReasoning.BullishFactors) > 0 {
		fmt.Println("  看涨因素:")
		for _, factor := range signal.AIReasoning.BullishFactors {
			fmt.Printf("    ✓ %s\n", factor)
		}
	}
	if len(signal.AIReasoning.BearishFactors) > 0 {
		fmt.Println("  看跌因素:")
		for _, factor := range signal.AIReasoning.BearishFactors {
			fmt.Printf("    ✗ %s\n", factor)
		}
	}
	if signal.AIReasoning.KeyInsight != "" {
		fmt.Printf("  核心观点: %s\n", signal.AIReasoning.KeyInsight)
	}

	// 价格目标
	fmt.Println("\n💰 价格目标:")
	fmt.Printf("  当前价格: %.2f\n", signal.PriceTargets.Current.InexactFloat64())
	fmt.Printf("  支撑位: %.2f\n", signal.PriceTargets.Support.InexactFloat64())
	fmt.Printf("  压力位: %.2f\n", signal.PriceTargets.Resistance.InexactFloat64())
	fmt.Printf("  止损价: %.2f\n", signal.PriceTargets.StopLoss.InexactFloat64())
	fmt.Printf("  止盈价: %.2f\n", signal.PriceTargets.TakeProfit.InexactFloat64())

	// 风险评估
	riskEmoji := getRiskEmoji(signal.RiskLevel)
	fmt.Printf("\n⚠️  风险等级: %s %s (评分: %d/100)\n",
		riskEmoji, signal.RiskLevel, signal.RiskScore)
	if len(signal.Risks) > 0 {
		fmt.Println("  风险提示:")
		for _, risk := range signal.Risks {
			fmt.Printf("    • %s\n", risk)
		}
	}

	// 仓位建议
	fmt.Println("\n📋 仓位建议:")
	fmt.Printf("  操作: %s\n", signal.PositionAdvice.Action)
	fmt.Printf("  建议仓位: %.1f%%\n", signal.PositionAdvice.SizePct.InexactFloat64())
	fmt.Printf("  入场价格: %s\n", signal.PositionAdvice.EntryPrice)
	fmt.Printf("  止损比例: %.1f%%\n", signal.PositionAdvice.StopLossPct.InexactFloat64())
	fmt.Printf("  止盈比例: %.1f%%\n", signal.PositionAdvice.TakeProfitPct.InexactFloat64())
	fmt.Printf("  持有周期: %s\n", signal.PositionAdvice.HoldingDays)
	fmt.Printf("  建议理由: %s\n", signal.PositionAdvice.Reason)

	// 元数据
	fmt.Println("\n📝 元数据:")
	fmt.Printf("  生成时间: %s\n", signal.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  有效期至: %s\n", signal.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  延迟: %dms\n", signal.LatencyMS)
	fmt.Printf("  AI成本: ¥%.2f (%d tokens)\n",
		signal.AICostYuan.InexactFloat64(), signal.AITokensUsed)

	fmt.Println("═══════════════════════════════════════════════════════════\n")

	return nil
}

// emitJSON 输出JSON
func (s *ConsoleSink) emitJSON(signal *entity.Signal) error {
	data, err := json.MarshalIndent(signal, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// Close 关闭
func (s *ConsoleSink) Close() error {
	return nil
}

// Name 名称
func (s *ConsoleSink) Name() string {
	return "console"
}

// getSignalEmoji 获取信号emoji
func getSignalEmoji(signalType valueobject.SignalType) string {
	switch signalType {
	case valueobject.SignalStrongBuy:
		return "🚀🚀🚀"
	case valueobject.SignalBuy:
		return "📈📈"
	case valueobject.SignalWatch:
		return "👀"
	case valueobject.SignalAvoid:
		return "⚠️"
	case valueobject.SignalStrongAvoid:
		return "🚫🚫🚫"
	default:
		return "❓"
	}
}

// getRiskEmoji 获取风险emoji
func getRiskEmoji(riskLevel valueobject.RiskLevel) string {
	switch riskLevel {
	case valueobject.RiskLow:
		return "🟢"
	case valueobject.RiskMedium:
		return "🟡"
	case valueobject.RiskHigh:
		return "🔴"
	default:
		return "⚪"
	}
}
