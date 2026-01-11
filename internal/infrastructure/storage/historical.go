package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
)

// HistoricalStorage stores historical market data
type HistoricalStorage struct {
	mu   sync.RWMutex
	data map[string][]*entity.HistoricalData // stockCode -> data
}

// NewHistoricalStorage creates a new historical storage
func NewHistoricalStorage() *HistoricalStorage {
	return &HistoricalStorage{
		data: make(map[string][]*entity.HistoricalData),
	}
}

// Save saves historical data
func (hs *HistoricalStorage) Save(data *entity.HistoricalData) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if data.StockCode == "" {
		return fmt.Errorf("stock code cannot be empty")
	}

	hs.data[data.StockCode] = append(hs.data[data.StockCode], data)
	return nil
}

// GetByTimeRange retrieves data within a time range
func (hs *HistoricalStorage) GetByTimeRange(stockCode string, start, end time.Time) ([]*entity.HistoricalData, error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	allData, exists := hs.data[stockCode]
	if !exists {
		return nil, fmt.Errorf("no data found for stock: %s", stockCode)
	}

	var result []*entity.HistoricalData
	for _, d := range allData {
		if (d.Date.Equal(start) || d.Date.After(start)) &&
			(d.Date.Equal(end) || d.Date.Before(end)) {
			result = append(result, d)
		}
	}

	return result, nil
}

// GetLatest retrieves the latest data for a stock
func (hs *HistoricalStorage) GetLatest(stockCode string) (*entity.HistoricalData, error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	allData, exists := hs.data[stockCode]
	if !exists || len(allData) == 0 {
		return nil, fmt.Errorf("no data found for stock: %s", stockCode)
	}

	return allData[len(allData)-1], nil
}

// GetAll retrieves all data for a stock
func (hs *HistoricalStorage) GetAll(stockCode string) ([]*entity.HistoricalData, error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	allData, exists := hs.data[stockCode]
	if !exists {
		return nil, fmt.Errorf("no data found for stock: %s", stockCode)
	}

	return allData, nil
}

// Count returns the number of records for a stock
func (hs *HistoricalStorage) Count(stockCode string) int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	return len(hs.data[stockCode])
}

// Clear clears all data
func (hs *HistoricalStorage) Clear() {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.data = make(map[string][]*entity.HistoricalData)
}
