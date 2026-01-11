package ml

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/internal/infrastructure/storage"
)

// HistoricalCollector collects historical market data
type HistoricalCollector struct {
	storage *storage.HistoricalStorage
	// In production, inject EastMoney client here
}

// NewHistoricalCollector creates a new historical collector
func NewHistoricalCollector(storage *storage.HistoricalStorage) *HistoricalCollector {
	return &HistoricalCollector{
		storage: storage,
	}
}

// Collect collects historical data for a stock
func (hc *HistoricalCollector) Collect(ctx context.Context, stockCode string, days int) (*entity.HistoricalDataSeries, error) {
	log.Printf("📊 开始收集历史数据: %s, %d天", stockCode, days)

	// In production, fetch from EastMoney API
	// For now, generate sample data
	data := hc.generateSampleData(stockCode, days)

	// Store data
	for _, record := range data {
		if err := hc.storage.Save(record); err != nil {
			return nil, fmt.Errorf("failed to save historical data: %w", err)
		}
	}

	series := entity.NewHistoricalDataSeries(stockCode, data)
	log.Printf("✅ 收集完成: %d条记录", len(data))

	return series, nil
}

// CollectBatch collects historical data for multiple stocks
func (hc *HistoricalCollector) CollectBatch(ctx context.Context, stockCodes []string, days int) (map[string]*entity.HistoricalDataSeries, error) {
	result := make(map[string]*entity.HistoricalDataSeries)

	for _, code := range stockCodes {
		series, err := hc.Collect(ctx, code, days)
		if err != nil {
			log.Printf("❌ 收集失败: %s, %v", code, err)
			continue
		}
		result[code] = series
	}

	return result, nil
}

// GetHistorical retrieves historical data from storage
func (hc *HistoricalCollector) GetHistorical(stockCode string, start, end time.Time) ([]*entity.HistoricalData, error) {
	return hc.storage.GetByTimeRange(stockCode, start, end)
}

// generateSampleData generates sample historical data for testing
func (hc *HistoricalCollector) generateSampleData(stockCode string, days int) []*entity.HistoricalData {
	data := make([]*entity.HistoricalData, days)
	now := time.Now()
	basePrice := 100.0

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -days+i)

		// Simulate price movement
		change := (float64(i%10) - 5) * 0.5
		open := basePrice + change
		high := open * 1.02
		low := open * 0.98
		close := open + (float64(i%3)-1)*0.3
		volume := 1000000.0 + float64(i%1000)*1000
		amount := volume * close * 100

		record := entity.NewHistoricalData(stockCode, "Sample Stock", date)
		record.Open = open
		record.High = high
		record.Low = low
		record.Close = close
		record.Volume = volume
		record.Amount = amount

		// Calculate technical indicators (simplified)
		record.MA5 = close
		record.MA10 = close * 0.99
		record.MA20 = close * 0.98
		record.MA60 = close * 0.97
		record.RSI = 50.0 + float64(i%20)
		record.MACD = close * 0.001
		record.Signal = record.MACD * 0.9
		record.Histogram = record.MACD - record.Signal

		// Volume indicators
		record.VolumeRatio = 1.0 + float64(i%10)*0.1
		record.Turnover = 2.0 + float64(i%5)*0.2

		// Capital flow (sample)
		record.MainInflow = float64((i%10)-5) * 1000000
		record.RetailFlow = -record.MainInflow * 0.8

		// Calculate change
		if i > 0 {
			record.CalculateChange(data[i-1].Close)
		}

		data[i] = record
		basePrice = close
	}

	return data
}

// UpdateRealtime updates storage with real-time data
func (hc *HistoricalCollector) UpdateRealtime(ctx context.Context, stockCode string, data *entity.HistoricalData) error {
	return hc.storage.Save(data)
}

// GetLatest gets the latest historical data for a stock
func (hc *HistoricalCollector) GetLatest(stockCode string) (*entity.HistoricalData, error) {
	now := time.Now()
	start := now.AddDate(0, 0, -1)
	data, err := hc.storage.GetByTimeRange(stockCode, start, now)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no data found for %s", stockCode)
	}
	return data[len(data)-1], nil
}
