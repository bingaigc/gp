package entity

import (
	"math"
	"time"
)

// TechnicalIndicator represents a technical indicator with predicted and actual values
type TechnicalIndicator struct {
	StockCode     string `json:"stock_code"`
	IndicatorType string `json:"indicator_type"` // MA5, MA10, MA20, MACD, RSI, KDJ, etc.

	// Predicted values
	PredictedValue float64   `json:"predicted_value"`
	PredictedAt    time.Time `json:"predicted_at"`
	Confidence     float64   `json:"confidence"` // 0-1

	// Actual values
	ActualValue float64   `json:"actual_value"`
	ActualAt    time.Time `json:"actual_at"`

	// Error metrics
	AbsoluteError   float64 `json:"absolute_error"`   // |predicted - actual|
	PercentageError float64 `json:"percentage_error"` // |(predicted - actual) / actual| * 100
	SquaredError    float64 `json:"squared_error"`    // (predicted - actual)^2

	// Model information
	ModelVersion string `json:"model_version"`
	ModelType    string `json:"model_type"` // linear, arima, lstm, ensemble

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTechnicalIndicator creates a new technical indicator
func NewTechnicalIndicator(stockCode, indicatorType string, predictedValue float64, confidence float64, modelType, modelVersion string) *TechnicalIndicator {
	now := time.Now()
	return &TechnicalIndicator{
		StockCode:      stockCode,
		IndicatorType:  indicatorType,
		PredictedValue: predictedValue,
		PredictedAt:    now,
		Confidence:     confidence,
		ModelType:      modelType,
		ModelVersion:   modelVersion,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UpdateActual updates the indicator with actual value and calculates errors
func (ti *TechnicalIndicator) UpdateActual(actualValue float64) {
	ti.ActualValue = actualValue
	ti.ActualAt = time.Now()
	ti.UpdatedAt = time.Now()

	ti.calculateErrors()
}

// calculateErrors calculates error metrics
func (ti *TechnicalIndicator) calculateErrors() {
	// Absolute error
	ti.AbsoluteError = math.Abs(ti.PredictedValue - ti.ActualValue)

	// Percentage error (avoid division by zero)
	if ti.ActualValue != 0 {
		ti.PercentageError = math.Abs((ti.PredictedValue - ti.ActualValue) / ti.ActualValue * 100)
	}

	// Squared error
	diff := ti.PredictedValue - ti.ActualValue
	ti.SquaredError = diff * diff
}

// IsAccurate checks if prediction is within acceptable error range
func (ti *TechnicalIndicator) IsAccurate(maxPercentageError float64) bool {
	return ti.PercentageError <= maxPercentageError
}

// GetDirection returns whether prediction was directionally correct
func (ti *TechnicalIndicator) GetDirection() string {
	if ti.PredictedValue > ti.ActualValue {
		return "overestimated"
	} else if ti.PredictedValue < ti.ActualValue {
		return "underestimated"
	}
	return "exact"
}

// HasActual checks if actual value has been recorded
func (ti *TechnicalIndicator) HasActual() bool {
	return !ti.ActualAt.IsZero()
}

// GetAge returns how old the prediction is
func (ti *TechnicalIndicator) GetAge() time.Duration {
	return time.Since(ti.PredictedAt)
}

// IndicatorBatch represents a batch of indicators for analysis
type IndicatorBatch struct {
	Indicators []*TechnicalIndicator `json:"indicators"`

	// Aggregate metrics
	MAE  float64 `json:"mae"`  // Mean Absolute Error
	RMSE float64 `json:"rmse"` // Root Mean Square Error
	MAPE float64 `json:"mape"` // Mean Absolute Percentage Error

	// Statistics
	TotalCount    int     `json:"total_count"`
	AccurateCount int     `json:"accurate_count"`
	Accuracy      float64 `json:"accuracy"` // Percentage

	CreatedAt time.Time `json:"created_at"`
}

// NewIndicatorBatch creates a new batch of indicators
func NewIndicatorBatch(indicators []*TechnicalIndicator) *IndicatorBatch {
	batch := &IndicatorBatch{
		Indicators: indicators,
		CreatedAt:  time.Now(),
	}
	batch.calculateMetrics()
	return batch
}

// calculateMetrics calculates aggregate metrics for the batch
func (ib *IndicatorBatch) calculateMetrics() {
	if len(ib.Indicators) == 0 {
		return
	}

	var sumAE, sumSE, sumAPE float64
	accurateCount := 0
	validCount := 0

	for _, ind := range ib.Indicators {
		if !ind.HasActual() {
			continue
		}

		validCount++
		sumAE += ind.AbsoluteError
		sumSE += ind.SquaredError
		sumAPE += ind.PercentageError

		if ind.IsAccurate(5.0) { // 5% threshold
			accurateCount++
		}
	}

	if validCount > 0 {
		ib.MAE = sumAE / float64(validCount)
		ib.RMSE = math.Sqrt(sumSE / float64(validCount))
		ib.MAPE = sumAPE / float64(validCount)
		ib.TotalCount = validCount
		ib.AccurateCount = accurateCount
		ib.Accuracy = float64(accurateCount) / float64(validCount) * 100
	}
}

// GetBestIndicator returns the indicator with lowest error
func (ib *IndicatorBatch) GetBestIndicator() *TechnicalIndicator {
	if len(ib.Indicators) == 0 {
		return nil
	}

	var best *TechnicalIndicator
	minError := math.MaxFloat64

	for _, ind := range ib.Indicators {
		if ind.HasActual() && ind.AbsoluteError < minError {
			minError = ind.AbsoluteError
			best = ind
		}
	}

	return best
}

// GetWorstIndicator returns the indicator with highest error
func (ib *IndicatorBatch) GetWorstIndicator() *TechnicalIndicator {
	if len(ib.Indicators) == 0 {
		return nil
	}

	var worst *TechnicalIndicator
	maxError := 0.0

	for _, ind := range ib.Indicators {
		if ind.HasActual() && ind.AbsoluteError > maxError {
			maxError = ind.AbsoluteError
			worst = ind
		}
	}

	return worst
}
