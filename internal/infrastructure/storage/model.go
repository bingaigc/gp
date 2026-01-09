package storage

import (
	"fmt"
	"sync"

	"github.com/bingaigc/gp/internal/domain/entity"
)

// ModelStorage stores ML models
type ModelStorage struct {
	mu     sync.RWMutex
	models map[string]*entity.PredictionModel // ID -> model
	active map[string]*entity.PredictionModel // targetMetric -> active model
}

// NewModelStorage creates a new model storage
func NewModelStorage() *ModelStorage {
	return &ModelStorage{
		models: make(map[string]*entity.PredictionModel),
		active: make(map[string]*entity.PredictionModel),
	}
}

// Save saves a model
func (ms *ModelStorage) Save(model *entity.PredictionModel) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if model.ID == "" {
		return fmt.Errorf("model ID cannot be empty")
	}
	
	ms.models[model.ID] = model
	
	// Set as active if it's the first model for this metric
	if _, exists := ms.active[model.TargetMetric]; !exists || model.IsActive {
		ms.active[model.TargetMetric] = model
	}
	
	return nil
}

// Get retrieves a model by ID
func (ms *ModelStorage) Get(id string) (*entity.PredictionModel, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	model, exists := ms.models[id]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", id)
	}
	
	return model, nil
}

// GetActive retrieves the active model for a target metric
func (ms *ModelStorage) GetActive(targetMetric string) (*entity.PredictionModel, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	model, exists := ms.active[targetMetric]
	if !exists {
		return nil, fmt.Errorf("no active model for metric: %s", targetMetric)
	}
	
	return model, nil
}

// Update updates a model
func (ms *ModelStorage) Update(model *entity.PredictionModel) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if _, exists := ms.models[model.ID]; !exists {
		return fmt.Errorf("model not found: %s", model.ID)
	}
	
	ms.models[model.ID] = model
	
	// Update active if this is the active model
	if activeModel, exists := ms.active[model.TargetMetric]; exists && activeModel.ID == model.ID {
		ms.active[model.TargetMetric] = model
	}
	
	return nil
}

// SetActive sets a model as active for its target metric
func (ms *ModelStorage) SetActive(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	model, exists := ms.models[id]
	if !exists {
		return fmt.Errorf("model not found: %s", id)
	}
	
	// Deactivate old active model
	if oldActive, exists := ms.active[model.TargetMetric]; exists {
		oldActive.Deactivate()
	}
	
	// Activate new model
	model.Activate()
	ms.active[model.TargetMetric] = model
	
	return nil
}

// List retrieves all models
func (ms *ModelStorage) List() []*entity.PredictionModel {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	models := make([]*entity.PredictionModel, 0, len(ms.models))
	for _, model := range ms.models {
		models = append(models, model)
	}
	
	return models
}

// ListByMetric retrieves all models for a target metric
func (ms *ModelStorage) ListByMetric(targetMetric string) []*entity.PredictionModel {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	var models []*entity.PredictionModel
	for _, model := range ms.models {
		if model.TargetMetric == targetMetric {
			models = append(models, model)
		}
	}
	
	return models
}

// Delete deletes a model
func (ms *ModelStorage) Delete(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	model, exists := ms.models[id]
	if !exists {
		return fmt.Errorf("model not found: %s", id)
	}
	
	// Remove from active if it's the active model
	if activeModel, exists := ms.active[model.TargetMetric]; exists && activeModel.ID == id {
		delete(ms.active, model.TargetMetric)
	}
	
	delete(ms.models, id)
	return nil
}

// Count returns the number of models
func (ms *ModelStorage) Count() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	return len(ms.models)
}

// Clear clears all models
func (ms *ModelStorage) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	ms.models = make(map[string]*entity.PredictionModel)
	ms.active = make(map[string]*entity.PredictionModel)
}
