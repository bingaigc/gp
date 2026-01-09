package ml

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/infrastructure/storage"
)

// IndicatorPredictor predicts technical indicators using ML models
type IndicatorPredictor struct {
	modelStorage *storage.ModelStorage
	histStorage  *storage.HistoricalStorage
}

// NewIndicatorPredictor creates a new indicator predictor
func NewIndicatorPredictor(modelStorage *storage.ModelStorage, histStorage *storage.HistoricalStorage) *IndicatorPredictor {
	return &IndicatorPredictor{
		modelStorage: modelStorage,
		histStorage:  histStorage,
	}
}

// Predict predicts technical indicators for a stock
func (ip *IndicatorPredictor) Predict(ctx context.Context, stockCode string, indicatorType string) (*entity.TechnicalIndicator, error) {
	log.Printf("🔮 开始预测: %s - %s", stockCode, indicatorType)
	
	// Get active model for this indicator
	model, err := ip.modelStorage.GetActive(indicatorType)
	if err != nil {
		// Create default model if none exists
		model = entity.NewPredictionModel("default-"+indicatorType, "linear", "v1.0", indicatorType)
		ip.modelStorage.Save(model)
	}
	
	// Get historical data for features
	historicalData, err := ip.getRecentHistorical(stockCode, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}
	
	// Extract features
	features := ip.extractFeatures(historicalData)
	
	// Make prediction based on model type
	var predictedValue float64
	var confidence float64
	
	switch model.Type {
	case "linear":
		predictedValue, confidence = ip.predictLinear(features, indicatorType)
	case "arima":
		predictedValue, confidence = ip.predictARIMA(features, indicatorType)
	case "lstm":
		predictedValue, confidence = ip.predictLSTM(features, indicatorType)
	default:
		predictedValue, confidence = ip.predictLinear(features, indicatorType)
	}
	
	// Create indicator
	indicator := entity.NewTechnicalIndicator(stockCode, indicatorType, predictedValue, confidence, model.Type, model.Version)
	
	// Record prediction
	model.RecordPrediction()
	ip.modelStorage.Update(model)
	
	log.Printf("✅ 预测完成: %.2f (置信度: %.2f%%)", predictedValue, confidence*100)
	
	return indicator, nil
}

// extractFeatures extracts features from historical data
func (ip *IndicatorPredictor) extractFeatures(data []*entity.HistoricalData) map[string]float64 {
	if len(data) == 0 {
		return make(map[string]float64)
	}
	
	features := make(map[string]float64)
	latest := data[len(data)-1]
	
	// Price features
	features["close"] = latest.Close
	features["open"] = latest.Open
	features["high"] = latest.High
	features["low"] = latest.Low
	features["change_pct"] = latest.ChangePercent
	
	// Volume features
	features["volume"] = latest.Volume
	features["amount"] = latest.Amount
	features["turnover"] = latest.Turnover
	features["volume_ratio"] = latest.VolumeRatio
	
	// Technical indicators
	features["ma5"] = latest.MA5
	features["ma10"] = latest.MA10
	features["ma20"] = latest.MA20
	features["rsi"] = latest.RSI
	features["macd"] = latest.MACD
	
	// Derived features
	if len(data) >= 5 {
		features["volatility"] = ip.calculateVolatility(data[len(data)-5:])
		features["momentum"] = ip.calculateMomentum(data[len(data)-5:])
	}
	
	return features
}

// predictLinear simple linear prediction
func (ip *IndicatorPredictor) predictLinear(features map[string]float64, indicatorType string) (float64, float64) {
	// Simple linear model: predict based on recent trend
	close := features["close"]
	changePct := features["change_pct"]
	
	// Predict next value based on trend
	predicted := close * (1 + changePct/100*0.5) // Dampen the trend
	
	// Confidence based on volatility
	volatility := features["volatility"]
	confidence := math.Max(0.5, 1.0-volatility/10.0)
	
	return predicted, confidence
}

// predictARIMA ARIMA-based prediction (simplified)
func (ip *IndicatorPredictor) predictARIMA(features map[string]float64, indicatorType string) (float64, float64) {
	// Simplified ARIMA: weighted moving average
	close := features["close"]
	ma5 := features["ma5"]
	ma10 := features["ma10"]
	
	// Weighted prediction
	predicted := close*0.5 + ma5*0.3 + ma10*0.2
	confidence := 0.75
	
	return predicted, confidence
}

// predictLSTM LSTM-based prediction (placeholder for future implementation)
func (ip *IndicatorPredictor) predictLSTM(features map[string]float64, indicatorType string) (float64, float64) {
	// Placeholder: use linear for now
	return ip.predictLinear(features, indicatorType)
}

// calculateVolatility calculates price volatility
func (ip *IndicatorPredictor) calculateVolatility(data []*entity.HistoricalData) float64 {
	if len(data) < 2 {
		return 0
	}
	
	var sum float64
	for i := 1; i < len(data); i++ {
		change := math.Abs(data[i].ChangePercent)
		sum += change
	}
	
	return sum / float64(len(data)-1)
}

// calculateMomentum calculates price momentum
func (ip *IndicatorPredictor) calculateMomentum(data []*entity.HistoricalData) float64 {
	if len(data) < 2 {
		return 0
	}
	
	first := data[0].Close
	last := data[len(data)-1].Close
	
	return (last - first) / first * 100
}

// getRecentHistorical gets recent historical data
func (ip *IndicatorPredictor) getRecentHistorical(stockCode string, days int) ([]*entity.HistoricalData, error) {
	end := entity.Now()
	start := end.AddDate(0, 0, -days)
	return ip.histStorage.GetByTimeRange(stockCode, start, end)
}

// PredictBatch predicts indicators for multiple stocks
func (ip *IndicatorPredictor) PredictBatch(ctx context.Context, stockCodes []string, indicatorType string) ([]*entity.TechnicalIndicator, error) {
	var indicators []*entity.TechnicalIndicator
	
	for _, code := range stockCodes {
		indicator, err := ip.Predict(ctx, code, indicatorType)
		if err != nil {
			log.Printf("❌ 预测失败: %s, %v", code, err)
			continue
		}
		indicators = append(indicators, indicator)
	}
	
	return indicators, nil
}
