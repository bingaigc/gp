package ml

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/bingaigc/gp/internal/domain/entity"
)

// IndicatorAnalyzer analyzes prediction-actual differences
type IndicatorAnalyzer struct {
	// Dependencies would be injected here
}

// NewIndicatorAnalyzer creates a new indicator analyzer
func NewIndicatorAnalyzer() *IndicatorAnalyzer {
	return &IndicatorAnalyzer{}
}

// ErrorAnalysis represents analysis of prediction errors
type ErrorAnalysis struct {
	StockCode       string         `json:"stock_code"`
	IndicatorType   string         `json:"indicator_type"`
	ErrorMetrics    ErrorMetrics   `json:"error_metrics"`
	RootCauses      []RootCause    `json:"root_causes"`
	Patterns        []ErrorPattern `json:"patterns"`
	Recommendations []string       `json:"recommendations"`
}

// ErrorMetrics contains error statistics
type ErrorMetrics struct {
	MAE         float64 `json:"mae"`
	RMSE        float64 `json:"rmse"`
	MAPE        float64 `json:"mape"`
	MaxError    float64 `json:"max_error"`
	MinError    float64 `json:"min_error"`
	StdDev      float64 `json:"std_dev"`
	SampleCount int     `json:"sample_count"`
}

// RootCause represents a root cause of prediction errors
type RootCause struct {
	Type        string  `json:"type"` // market_regime, volatility, volume, news
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`     // 0-1
	Confidence  float64 `json:"confidence"` // 0-1
}

// ErrorPattern represents a detected error pattern
type ErrorPattern struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Frequency   float64 `json:"frequency"` // How often it occurs
	Severity    float64 `json:"severity"`  // Impact magnitude
}

// Analyze performs comprehensive error analysis
func (ia *IndicatorAnalyzer) Analyze(ctx context.Context, indicators []*entity.TechnicalIndicator) (*ErrorAnalysis, error) {
	if len(indicators) == 0 {
		return nil, fmt.Errorf("no indicators to analyze")
	}

	log.Printf("🔍 开始误差分析...")

	// Calculate error metrics
	metrics := ia.calculateErrorMetrics(indicators)

	// Identify root causes
	rootCauses := ia.identifyRootCauses(indicators, metrics)

	// Detect patterns
	patterns := ia.detectPatterns(indicators)

	// Generate recommendations
	recommendations := ia.generateRecommendations(rootCauses, patterns, metrics)

	analysis := &ErrorAnalysis{
		StockCode:       indicators[0].StockCode,
		IndicatorType:   indicators[0].IndicatorType,
		ErrorMetrics:    metrics,
		RootCauses:      rootCauses,
		Patterns:        patterns,
		Recommendations: recommendations,
	}

	log.Printf("✅ 分析完成:")
	log.Printf("   MAPE: %.2f%%", metrics.MAPE)
	log.Printf("   根本原因: %d个", len(rootCauses))
	log.Printf("   错误模式: %d个", len(patterns))
	log.Printf("   优化建议: %d条", len(recommendations))

	return analysis, nil
}

// calculateErrorMetrics calculates aggregate error metrics
func (ia *IndicatorAnalyzer) calculateErrorMetrics(indicators []*entity.TechnicalIndicator) ErrorMetrics {
	var sumAE, sumSE, sumAPE float64
	var maxError, minError float64 = 0, math.MaxFloat64
	validCount := 0

	for _, ind := range indicators {
		if !ind.HasActual() {
			continue
		}

		validCount++
		sumAE += ind.AbsoluteError
		sumSE += ind.SquaredError
		sumAPE += ind.PercentageError

		if ind.AbsoluteError > maxError {
			maxError = ind.AbsoluteError
		}
		if ind.AbsoluteError < minError {
			minError = ind.AbsoluteError
		}
	}

	if validCount == 0 {
		return ErrorMetrics{}
	}

	mae := sumAE / float64(validCount)
	rmse := math.Sqrt(sumSE / float64(validCount))
	mape := sumAPE / float64(validCount)

	// Calculate standard deviation
	var sumSquaredDiff float64
	for _, ind := range indicators {
		if ind.HasActual() {
			diff := ind.AbsoluteError - mae
			sumSquaredDiff += diff * diff
		}
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(validCount))

	return ErrorMetrics{
		MAE:         mae,
		RMSE:        rmse,
		MAPE:        mape,
		MaxError:    maxError,
		MinError:    minError,
		StdDev:      stdDev,
		SampleCount: validCount,
	}
}

// identifyRootCauses identifies root causes of errors
func (ia *IndicatorAnalyzer) identifyRootCauses(indicators []*entity.TechnicalIndicator, metrics ErrorMetrics) []RootCause {
	var causes []RootCause

	// High volatility
	if metrics.StdDev > metrics.MAE {
		causes = append(causes, RootCause{
			Type:        "high_volatility",
			Description: "市场波动率显著高于预期，导致预测偏差增大",
			Impact:      0.8,
			Confidence:  0.9,
		})
	}

	// Systematic bias
	overestimateCount := 0
	for _, ind := range indicators {
		if ind.HasActual() && ind.PredictedValue > ind.ActualValue {
			overestimateCount++
		}
	}
	bias := float64(overestimateCount) / float64(len(indicators))
	if bias > 0.7 || bias < 0.3 {
		causes = append(causes, RootCause{
			Type:        "systematic_bias",
			Description: fmt.Sprintf("模型存在系统性偏差（高估比例: %.1f%%）", bias*100),
			Impact:      0.6,
			Confidence:  0.85,
		})
	}

	// Market regime change
	if metrics.MAPE > 10 {
		causes = append(causes, RootCause{
			Type:        "market_regime_change",
			Description: "市场环境可能发生变化，模型需要重新适应",
			Impact:      0.7,
			Confidence:  0.75,
		})
	}

	return causes
}

// detectPatterns detects error patterns
func (ia *IndicatorAnalyzer) detectPatterns(indicators []*entity.TechnicalIndicator) []ErrorPattern {
	var patterns []ErrorPattern

	// Increasing error trend
	if len(indicators) >= 5 {
		firstHalf := indicators[:len(indicators)/2]
		secondHalf := indicators[len(indicators)/2:]

		var firstAvgError, secondAvgError float64
		for _, ind := range firstHalf {
			if ind.HasActual() {
				firstAvgError += ind.AbsoluteError
			}
		}
		for _, ind := range secondHalf {
			if ind.HasActual() {
				secondAvgError += ind.AbsoluteError
			}
		}
		firstAvgError /= float64(len(firstHalf))
		secondAvgError /= float64(len(secondHalf))

		if secondAvgError > firstAvgError*1.2 {
			patterns = append(patterns, ErrorPattern{
				Name:        "increasing_error_trend",
				Description: "预测误差呈上升趋势，模型性能下降",
				Frequency:   0.8,
				Severity:    0.7,
			})
		}
	}

	// High variance pattern
	variancePattern := ErrorPattern{
		Name:        "high_variance",
		Description: "预测误差波动较大，稳定性不足",
		Frequency:   0.6,
		Severity:    0.5,
	}
	patterns = append(patterns, variancePattern)

	return patterns
}

// generateRecommendations generates actionable recommendations
func (ia *IndicatorAnalyzer) generateRecommendations(causes []RootCause, patterns []ErrorPattern, metrics ErrorMetrics) []string {
	var recs []string

	// Based on error magnitude
	if metrics.MAPE > 15 {
		recs = append(recs, "❗ MAPE过高(>15%)，建议立即重新训练模型")
	} else if metrics.MAPE > 10 {
		recs = append(recs, "⚠️  MAPE较高(>10%)，建议考虑重新训练或调整特征")
	}

	// Based on root causes
	for _, cause := range causes {
		switch cause.Type {
		case "high_volatility":
			recs = append(recs, "📊 增加波动率相关特征，提高对市场波动的适应性")
		case "systematic_bias":
			recs = append(recs, "🎯 调整模型参数，修正系统性偏差")
		case "market_regime_change":
			recs = append(recs, "🔄 使用近期数据重新训练，适应新的市场环境")
		}
	}

	// Based on patterns
	for _, pattern := range patterns {
		if pattern.Name == "increasing_error_trend" {
			recs = append(recs, "📈 误差上升趋势明显，建议增加训练频率")
		}
		if pattern.Name == "high_variance" {
			recs = append(recs, "⚡ 考虑使用集成模型，降低预测方差")
		}
	}

	// General recommendations
	if len(recs) == 0 {
		recs = append(recs, "✅ 模型表现良好，继续保持当前策略")
	}

	return recs
}
