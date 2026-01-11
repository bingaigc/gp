package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bingaigc/gp/internal/application/risk"
	"github.com/bingaigc/gp/internal/application/scorer"
	"github.com/bingaigc/gp/internal/config"
	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/domain/valueobject"
	"github.com/bingaigc/gp/internal/interfaces/sink"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "扫描市场并生成交易信号",
	Long:  "扫描当前市场活跃股票，进行CAMP-I多因子分析，生成交易信号",
	Run:   runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// 加载配置
	cfg, err := config.Load(GetConfigFile())
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   CAMP-I Distributed Alpha Signal Detector v1.0.0        ║")
	fmt.Println("║   企业级分布式Alpha信号探测系统                              ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("📋 配置: %s (%s)\n", cfg.System.Name, cfg.System.Environment)
	fmt.Printf("🔧 数据源: %s\n", cfg.DataSource.Primary)
	fmt.Printf("🤖 AI模型: %s (%s)\n", cfg.AI.Provider, cfg.AI.Model)
	fmt.Printf("⚙️  并发数: %d workers\n", cfg.Pipeline.Workers)
	fmt.Println()

	// 初始化组件
	scorer := scorer.NewCAMPIScorer()
	riskCtrl := risk.NewController(
		cfg.Risk.DailyMaxLoss,
		cfg.Risk.MaxSignalsPerDay,
		cfg.Risk.MaxConsecutiveLosses,
	)
	signalSink := sink.NewConsoleSink(true)

	fmt.Println("🚀 开始扫描市场...")
	fmt.Println()

	// 演示：创建模拟股票数据并进行评分
	demoStocks := createDemoStocks()

	signalCount := 0
	for _, stock := range demoStocks {
		// 评分
		score, factors, err := scorer.Score(ctx, stock)
		if err != nil {
			log.Printf("评分失败 %s: %v", stock.Code, err)
			continue
		}

		// 生成信号
		signal := createSignalFromScore(stock, score, factors)

		// 风控验证
		if err := riskCtrl.Validate(ctx, signal); err != nil {
			log.Printf("风控拦截 %s: %v", stock.Code, err)
			continue
		}

		// 输出信号
		if err := signalSink.Emit(ctx, signal); err != nil {
			log.Printf("输出信号失败 %s: %v", stock.Code, err)
			continue
		}

		signalCount++
	}

	fmt.Println()
	fmt.Printf("✅ 扫描完成！生成信号数: %d\n", signalCount)
	fmt.Println()

	// 显示风控统计
	stats := riskCtrl.GetStats()
	fmt.Println("📊 风控统计:")
	fmt.Printf("  日内信号数: %d/%d\n", stats["daily_signal_count"], cfg.Risk.MaxSignalsPerDay)
	fmt.Printf("  日内亏损: ¥%.2f/¥%.2f\n", stats["daily_loss"], cfg.Risk.DailyMaxLoss)
	fmt.Printf("  连续亏损: %d/%d\n", stats["consecutive_losses"], cfg.Risk.MaxConsecutiveLosses)
}

// createDemoStocks 创建演示股票数据
func createDemoStocks() []*entity.Stock {
	return []*entity.Stock{
		{
			Code:     "600519",
			Name:     "贵州茅台",
			Market:   "SH",
			Industry: "白酒",
			Price:    decimal.NewFromFloat(1680.50),
			Change:   decimal.NewFromFloat(2.3),
			Volume:   850000,
			Amount:   decimal.NewFromFloat(1400000000),
			Capital: entity.CapitalFlow{
				MainNetInflow:   decimal.NewFromFloat(85000000),
				SuperNetInflow:  decimal.NewFromFloat(50000000),
				BigNetInflow:    decimal.NewFromFloat(35000000),
				ConsecutiveDays: 3,
			},
			Technical: entity.Technical{
				VolumeRatio:  decimal.NewFromFloat(1.8),
				TurnoverRate: decimal.NewFromFloat(0.8),
				Amplitude:    decimal.NewFromFloat(3.5),
			},
			Valuation: entity.Valuation{
				PE:        decimal.NewFromFloat(35.5),
				PB:        decimal.NewFromFloat(12.8),
				MarketCap: decimal.NewFromFloat(2100000000000), // 2.1万亿
			},
			Sector: &entity.Sector{
				Name:      "白酒",
				Rank:      2,
				Change:    decimal.NewFromFloat(1.8),
				NetInflow: decimal.NewFromFloat(2500000000),
			},
			Timestamp: time.Now(),
			Source:    "demo",
		},
		{
			Code:     "300750",
			Name:     "宁德时代",
			Market:   "SZ",
			Industry: "新能源",
			Price:    decimal.NewFromFloat(185.20),
			Change:   decimal.NewFromFloat(5.2),
			Volume:   28000000,
			Amount:   decimal.NewFromFloat(5200000000),
			Capital: entity.CapitalFlow{
				MainNetInflow:   decimal.NewFromFloat(320000000),
				SuperNetInflow:  decimal.NewFromFloat(180000000),
				BigNetInflow:    decimal.NewFromFloat(140000000),
				ConsecutiveDays: 2,
			},
			Technical: entity.Technical{
				VolumeRatio:  decimal.NewFromFloat(2.5),
				TurnoverRate: decimal.NewFromFloat(6.2),
				Amplitude:    decimal.NewFromFloat(7.8),
			},
			Valuation: entity.Valuation{
				PE:        decimal.NewFromFloat(28.3),
				PB:        decimal.NewFromFloat(5.6),
				MarketCap: decimal.NewFromFloat(850000000000), // 8500亿
			},
			Sector: &entity.Sector{
				Name:      "新能源电池",
				Rank:      1,
				Change:    decimal.NewFromFloat(4.5),
				NetInflow: decimal.NewFromFloat(4800000000),
			},
			LHB: &entity.LHBData{
				Reason:       "日涨幅偏离值达7%",
				BuyAmount:    decimal.NewFromFloat(85000000),
				SellAmount:   decimal.NewFromFloat(25000000),
				NetAmount:    decimal.NewFromFloat(60000000),
				BuyOrgs:      2,
				SellOrgs:     0,
				OrgNetInflow: decimal.NewFromFloat(60000000),
			},
			Timestamp: time.Now(),
			Source:    "demo",
		},
	}
}

// createSignalFromScore 从评分创建信号
func createSignalFromScore(stock *entity.Stock, score int, factors []entity.FactorScore) *entity.Signal {
	signalType := valueobject.FromScore(score)
	confidence := decimal.NewFromFloat(float64(score) / 100.0)

	// 计算风险等级
	riskLevel := valueobject.RiskMedium
	riskScore := 100 - score
	if score >= 85 {
		riskLevel = valueobject.RiskLow
	} else if score < 60 {
		riskLevel = valueobject.RiskHigh
	}

	// 价格目标
	currentPrice := stock.Price.InexactFloat64()
	priceTargets := entity.PriceTargets{
		Current:    stock.Price,
		Support:    decimal.NewFromFloat(currentPrice * 0.95),
		Resistance: decimal.NewFromFloat(currentPrice * 1.10),
		StopLoss:   decimal.NewFromFloat(currentPrice * 0.97),
		TakeProfit: decimal.NewFromFloat(currentPrice * 1.08),
	}

	// 仓位建议
	positionAdvice := entity.PositionAdvice{
		Action:        "BUY",
		SizePct:       decimal.NewFromFloat(5.0),
		EntryPrice:    fmt.Sprintf("%.2f-%.2f", currentPrice*0.99, currentPrice*1.01),
		StopLossPct:   decimal.NewFromFloat(3.0),
		TakeProfitPct: decimal.NewFromFloat(8.0),
		HoldingDays:   "5-10天",
		Reason:        "多因子共振，建议5%仓位试仓",
	}

	if signalType == valueobject.SignalStrongBuy {
		positionAdvice.SizePct = decimal.NewFromFloat(10.0)
		positionAdvice.Reason = "强买入信号，建议10%仓位"
	}

	return &entity.Signal{
		ID:           fmt.Sprintf("SIG-%d", time.Now().UnixNano()),
		StockCode:    stock.Code,
		StockName:    stock.Name,
		Market:       stock.Market,
		SignalType:   signalType,
		Confidence:   confidence,
		FactorScores: factors,
		TotalScore:   score,
		AIReasoning: entity.AIReasoning{
			BullishFactors: []string{
				fmt.Sprintf("主力资金连续%d日流入", stock.Capital.ConsecutiveDays),
				fmt.Sprintf("板块排名第%d", stock.Sector.Rank),
			},
			BearishFactors: []string{},
			KeyInsight:     fmt.Sprintf("%s处于强势状态，多因子共振", stock.Name),
			Confidence:     confidence,
		},
		PriceTargets:   priceTargets,
		RiskLevel:      riskLevel,
		RiskScore:      riskScore,
		PositionAdvice: positionAdvice,
		TimeHorizon:    "SHORT",
		GeneratedAt:    time.Now(),
		ExpiresAt:      time.Now().Add(4 * time.Hour),
		TraceID:        fmt.Sprintf("TRACE-%d", time.Now().UnixNano()),
		LatencyMS:      150,
		AITokensUsed:   0,
		AICostYuan:     decimal.Zero,
	}
}
