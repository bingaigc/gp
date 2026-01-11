package entity

import "time"

// PredictionModel represents a machine learning model for predictions
type PredictionModel struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"` // linear, arima, lstm, ensemble
	Version string `json:"version"`

	// Model metadata
	Description  string `json:"description"`
	TargetMetric string `json:"target_metric"` // which indicator to predict

	// Training info
	TrainedAt         time.Time `json:"trained_at"`
	TrainingSamples   int       `json:"training_samples"`
	ValidationSamples int       `json:"validation_samples"`

	// Performance metrics
	TrainingAccuracy   float64 `json:"training_accuracy"`
	ValidationAccuracy float64 `json:"validation_accuracy"`
	MAE                float64 `json:"mae"`      // Mean Absolute Error
	RMSE               float64 `json:"rmse"`     // Root Mean Square Error
	MAPE               float64 `json:"mape"`     // Mean Absolute Percentage Error
	R2Score            float64 `json:"r2_score"` // R-squared

	// Feature importance (for interpretability)
	FeatureImportance map[string]float64 `json:"feature_importance"`

	// Model state
	IsActive        bool      `json:"is_active"`
	LastUsedAt      time.Time `json:"last_used_at"`
	PredictionCount int64     `json:"prediction_count"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewPredictionModel creates a new prediction model
func NewPredictionModel(name, modelType, version, targetMetric string) *PredictionModel {
	now := time.Now()
	return &PredictionModel{
		ID:                GenerateID(),
		Name:              name,
		Type:              modelType,
		Version:           version,
		TargetMetric:      targetMetric,
		FeatureImportance: make(map[string]float64),
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// UpdateMetrics updates the model's performance metrics
func (pm *PredictionModel) UpdateMetrics(mae, rmse, mape, r2 float64) {
	pm.MAE = mae
	pm.RMSE = rmse
	pm.MAPE = mape
	pm.R2Score = r2
	pm.UpdatedAt = time.Now()
}

// RecordPrediction records that a prediction was made
func (pm *PredictionModel) RecordPrediction() {
	pm.PredictionCount++
	pm.LastUsedAt = time.Now()
	pm.UpdatedAt = time.Now()
}

// Deactivate marks the model as inactive
func (pm *PredictionModel) Deactivate() {
	pm.IsActive = false
	pm.UpdatedAt = time.Now()
}

// Activate marks the model as active
func (pm *PredictionModel) Activate() {
	pm.IsActive = true
	pm.UpdatedAt = time.Now()
}

// GetAccuracy returns the overall accuracy (1 - MAPE/100)
func (pm *PredictionModel) GetAccuracy() float64 {
	if pm.MAPE > 0 {
		return (100 - pm.MAPE) / 100
	}
	return 0
}

// IsHealthy checks if model performance is acceptable
func (pm *PredictionModel) IsHealthy(maxMAPE float64) bool {
	return pm.IsActive && pm.MAPE <= maxMAPE
}

// GetAge returns how old the model is
func (pm *PredictionModel) GetAge() time.Duration {
	return time.Since(pm.TrainedAt)
}

// NeedsRetraining checks if model should be retrained
func (pm *PredictionModel) NeedsRetraining(maxAge time.Duration, maxMAPE float64) bool {
	return pm.GetAge() > maxAge || pm.MAPE > maxMAPE || !pm.IsActive
}

// SetFeatureImportance sets the importance of a feature
func (pm *PredictionModel) SetFeatureImportance(feature string, importance float64) {
	pm.FeatureImportance[feature] = importance
	pm.UpdatedAt = time.Now()
}

// GetTopFeatures returns the N most important features
func (pm *PredictionModel) GetTopFeatures(n int) []string {
	type featureScore struct {
		name  string
		score float64
	}

	var features []featureScore
	for name, score := range pm.FeatureImportance {
		features = append(features, featureScore{name, score})
	}

	// Simple bubble sort (good enough for small n)
	for i := 0; i < len(features)-1; i++ {
		for j := 0; j < len(features)-i-1; j++ {
			if features[j].score < features[j+1].score {
				features[j], features[j+1] = features[j+1], features[j]
			}
		}
	}

	var topFeatures []string
	for i := 0; i < n && i < len(features); i++ {
		topFeatures = append(topFeatures, features[i].name)
	}

	return topFeatures
}

// GenerateID generates a unique ID for the model
func GenerateID() string {
	return time.Now().Format("20060102150405")
}

// ModelComparison compares multiple models
type ModelComparison struct {
	Models     []*PredictionModel `json:"models"`
	BestModel  *PredictionModel   `json:"best_model"`
	Metric     string             `json:"metric"` // which metric to compare (mae, rmse, mape)
	ComparedAt time.Time          `json:"compared_at"`
}

// NewModelComparison creates a new model comparison
func NewModelComparison(models []*PredictionModel, metric string) *ModelComparison {
	comparison := &ModelComparison{
		Models:     models,
		Metric:     metric,
		ComparedAt: time.Now(),
	}
	comparison.findBest()
	return comparison
}

// findBest finds the best performing model
func (mc *ModelComparison) findBest() {
	if len(mc.Models) == 0 {
		return
	}

	best := mc.Models[0]
	bestScore := mc.getScore(best)

	for _, model := range mc.Models[1:] {
		score := mc.getScore(model)
		if score < bestScore { // Lower is better for error metrics
			best = model
			bestScore = score
		}
	}

	mc.BestModel = best
}

// getScore gets the comparison score based on metric
func (mc *ModelComparison) getScore(model *PredictionModel) float64 {
	switch mc.Metric {
	case "mae":
		return model.MAE
	case "rmse":
		return model.RMSE
	case "mape":
		return model.MAPE
	default:
		return model.MAPE
	}
}
