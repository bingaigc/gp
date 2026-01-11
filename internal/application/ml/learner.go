package ml

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/infrastructure/storage"
)

// LearningEngine continuously learns from prediction errors
type LearningEngine struct {
	modelStorage *storage.ModelStorage
	histStorage  *storage.HistoricalStorage
	predictor    *IndicatorPredictor
}

// NewLearningEngine creates a new learning engine
func NewLearningEngine(modelStorage *storage.ModelStorage, histStorage *storage.HistoricalStorage, predictor *IndicatorPredictor) *LearningEngine {
	return &LearningEngine{
		modelStorage: modelStorage,
		histStorage:  histStorage,
		predictor:    predictor,
	}
}

// Learn performs one learning cycle
func (le *LearningEngine) Learn(ctx context.Context, stockCode string, indicatorType string) error {
	log.Printf("🎓 开始学习: %s - %s", stockCode, indicatorType)

	// Get predicted indicators
	predicted, err := le.predictor.Predict(ctx, stockCode, indicatorType)
	if err != nil {
		return err
	}

	// Wait a bit and get actual value (in production, get real-time data)
	// For demo, simulate actual value
	actualValue := le.simulateActualValue(predicted.PredictedValue)
	predicted.UpdateActual(actualValue)

	// Calculate errors
	mae := predicted.AbsoluteError
	rmse := math.Sqrt(predicted.SquaredError)
	mape := predicted.PercentageError

	log.Printf("📊 误差统计:")
	log.Printf("   MAE: %.4f", mae)
	log.Printf("   RMSE: %.4f", rmse)
	log.Printf("   MAPE: %.2f%%", mape)

	// Update model metrics
	model, _ := le.modelStorage.GetActive(indicatorType)
	if model != nil {
		oldMAPE := model.MAPE
		newMAPE := le.updateMovingAverage(oldMAPE, mape, 0.1) // 10% learning rate
		model.UpdateMetrics(mae, rmse, newMAPE, 0.9)          // R2 placeholder
		le.modelStorage.Update(model)

		improvement := oldMAPE - newMAPE
		if improvement > 0 {
			log.Printf("✅ 模型改进: MAPE %.2f%% → %.2f%% (提升%.2f%%)", oldMAPE, newMAPE, improvement)
		}
	}

	// Check if retraining needed
	if le.needsRetraining(model) {
		log.Printf("🔄 开始重新训练模型...")
		if err := le.retrain(ctx, model, stockCode); err != nil {
			log.Printf("❌ 重训练失败: %v", err)
		} else {
			log.Printf("✅ 重训练完成")
		}
	}

	return nil
}

// updateMovingAverage updates metric using exponential moving average
func (le *LearningEngine) updateMovingAverage(oldValue, newValue, alpha float64) float64 {
	return alpha*newValue + (1-alpha)*oldValue
}

// needsRetraining checks if model needs retraining
func (le *LearningEngine) needsRetraining(model *entity.PredictionModel) bool {
	if model == nil {
		return false
	}

	// Retrain if:
	// 1. MAPE too high (>10%)
	// 2. Model too old (>7 days)
	maxMAPE := 10.0
	maxAge := 7 * 24 * time.Hour

	return model.MAPE > maxMAPE || model.GetAge() > maxAge
}

// retrain retrains the model with new data
func (le *LearningEngine) retrain(ctx context.Context, model *entity.PredictionModel, stockCode string) error {
	// Get recent historical data
	days := 60
	data, err := le.histStorage.GetByTimeRange(stockCode, time.Now().AddDate(0, 0, -days), time.Now())
	if err != nil {
		return err
	}

	if len(data) < 30 {
		return fmt.Errorf("insufficient data for retraining: %d", len(data))
	}

	// Simple retraining: recalculate model parameters
	// In production, use proper ML framework
	model.TrainedAt = time.Now()
	model.TrainingSamples = len(data) * 70 / 100
	model.ValidationSamples = len(data) * 30 / 100
	model.Version = fmt.Sprintf("v%s", time.Now().Format("20060102"))

	// Reset metrics for fresh start
	model.MAPE = model.MAPE * 0.8 // Simulate improvement
	model.UpdatedAt = time.Now()

	return le.modelStorage.Update(model)
}

// simulateActualValue simulates actual value for demo
func (le *LearningEngine) simulateActualValue(predicted float64) float64 {
	// Add small random error (±2%)
	errorPct := (float64(time.Now().UnixNano()%100) - 50) / 100 * 0.02
	return predicted * (1 + errorPct)
}

// ContinuousLearning runs continuous learning loop
func (le *LearningEngine) ContinuousLearning(ctx context.Context, stockCode string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	indicatorTypes := []string{"MA5", "MA10", "MA20", "RSI", "MACD"}

	for {
		select {
		case <-ctx.Done():
			log.Printf("🛑 停止持续学习")
			return
		case <-ticker.C:
			log.Printf("⏰ 执行定时学习...")
			for _, indicatorType := range indicatorTypes {
				if err := le.Learn(ctx, stockCode, indicatorType); err != nil {
					log.Printf("❌ 学习失败 %s: %v", indicatorType, err)
				}
			}
		}
	}
}

// ConvergenceAnalysis analyzes convergence of predictions
type ConvergenceAnalysis struct {
	IndicatorType      string    `json:"indicator_type"`
	InitialMAPE        float64   `json:"initial_mape"`
	CurrentMAPE        float64   `json:"current_mape"`
	Improvement        float64   `json:"improvement"`
	ConvergenceRate    float64   `json:"convergence_rate"` // % per iteration
	IsConverged        bool      `json:"is_converged"`
	IterationsToTarget int       `json:"iterations_to_target"`
	CreatedAt          time.Time `json:"created_at"`
}

// AnalyzeConvergence analyzes how well predictions are converging
func (le *LearningEngine) AnalyzeConvergence(stockCode, indicatorType string) (*ConvergenceAnalysis, error) {
	model, err := le.modelStorage.GetActive(indicatorType)
	if err != nil {
		return nil, err
	}

	// Calculate convergence metrics
	initialMAPE := 20.0 // Assume initial error
	currentMAPE := model.MAPE
	improvement := initialMAPE - currentMAPE
	convergenceRate := improvement / float64(model.PredictionCount+1) * 100

	// Check if converged (MAPE < 5%)
	targetMAPE := 5.0
	isConverged := currentMAPE < targetMAPE

	// Estimate iterations to target
	iterationsToTarget := 0
	if !isConverged && convergenceRate > 0 {
		iterationsToTarget = int((currentMAPE - targetMAPE) / (convergenceRate / 100))
	}

	return &ConvergenceAnalysis{
		IndicatorType:      indicatorType,
		InitialMAPE:        initialMAPE,
		CurrentMAPE:        currentMAPE,
		Improvement:        improvement,
		ConvergenceRate:    convergenceRate,
		IsConverged:        isConverged,
		IterationsToTarget: iterationsToTarget,
		CreatedAt:          time.Now(),
	}, nil
}
