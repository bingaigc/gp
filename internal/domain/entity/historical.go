package entity

import "time"

// HistoricalData represents historical market data for a stock
type HistoricalData struct {
	StockCode string    `json:"stock_code"`
	StockName string    `json:"stock_name"`
	Date      time.Time `json:"date"`
	
	// OHLCV data
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"` // 成交量(手)
	Amount float64 `json:"amount"` // 成交额(元)
	
	// Technical indicators
	MA5       float64 `json:"ma5"`
	MA10      float64 `json:"ma10"`
	MA20      float64 `json:"ma20"`
	MA60      float64 `json:"ma60"`
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
	RSI       float64 `json:"rsi"`
	KDJ_K     float64 `json:"kdj_k"`
	KDJ_D     float64 `json:"kdj_d"`
	KDJ_J     float64 `json:"kdj_j"`
	
	// Volume indicators
	VolumeRatio float64 `json:"volume_ratio"` // 量比
	Turnover    float64 `json:"turnover"`     // 换手率
	
	// Capital flow
	MainInflow  float64 `json:"main_inflow"`  // 主力净流入
	RetailFlow  float64 `json:"retail_flow"`  // 散户净流入
	
	// Calculated fields
	Change        float64 `json:"change"`         // 涨跌额
	ChangePercent float64 `json:"change_percent"` // 涨跌幅
	
	CreatedAt time.Time `json:"created_at"`
}

// NewHistoricalData creates a new historical data record
func NewHistoricalData(stockCode, stockName string, date time.Time) *HistoricalData {
	return &HistoricalData{
		StockCode: stockCode,
		StockName: stockName,
		Date:      date,
		CreatedAt: time.Now(),
	}
}

// CalculateChange calculates price change and percentage
func (hd *HistoricalData) CalculateChange(prevClose float64) {
	if prevClose > 0 {
		hd.Change = hd.Close - prevClose
		hd.ChangePercent = (hd.Change / prevClose) * 100
	}
}

// IsValid validates the historical data
func (hd *HistoricalData) IsValid() bool {
	return hd.StockCode != "" &&
		hd.Open > 0 &&
		hd.High > 0 &&
		hd.Low > 0 &&
		hd.Close > 0 &&
		hd.Volume > 0 &&
		hd.High >= hd.Low &&
		hd.High >= hd.Open &&
		hd.High >= hd.Close &&
		hd.Low <= hd.Open &&
		hd.Low <= hd.Close
}

// GetTypicalPrice calculates the typical price (HLC/3)
func (hd *HistoricalData) GetTypicalPrice() float64 {
	return (hd.High + hd.Low + hd.Close) / 3
}

// GetAveragePrice calculates average price for the day
func (hd *HistoricalData) GetAveragePrice() float64 {
	if hd.Volume > 0 {
		return hd.Amount / (hd.Volume * 100) // 成交量单位是手(100股)
	}
	return hd.Close
}

// IsBullish returns true if close > open
func (hd *HistoricalData) IsBullish() bool {
	return hd.Close > hd.Open
}

// IsBearish returns true if close < open
func (hd *HistoricalData) IsBearish() bool {
	return hd.Close < hd.Open
}

// GetBodySize returns the size of the candle body
func (hd *HistoricalData) GetBodySize() float64 {
	if hd.Close > hd.Open {
		return hd.Close - hd.Open
	}
	return hd.Open - hd.Close
}

// GetUpperShadow returns the upper shadow length
func (hd *HistoricalData) GetUpperShadow() float64 {
	if hd.Close > hd.Open {
		return hd.High - hd.Close
	}
	return hd.High - hd.Open
}

// GetLowerShadow returns the lower shadow length
func (hd *HistoricalData) GetLowerShadow() float64 {
	if hd.Close > hd.Open {
		return hd.Open - hd.Low
	}
	return hd.Close - hd.Low
}

// HistoricalDataSeries represents a time series of historical data
type HistoricalDataSeries struct {
	StockCode string             `json:"stock_code"`
	Data      []*HistoricalData  `json:"data"`
	StartDate time.Time          `json:"start_date"`
	EndDate   time.Time          `json:"end_date"`
	Count     int                `json:"count"`
	CreatedAt time.Time          `json:"created_at"`
}

// NewHistoricalDataSeries creates a new time series
func NewHistoricalDataSeries(stockCode string, data []*HistoricalData) *HistoricalDataSeries {
	series := &HistoricalDataSeries{
		StockCode: stockCode,
		Data:      data,
		Count:     len(data),
		CreatedAt: time.Now(),
	}
	
	if len(data) > 0 {
		series.StartDate = data[0].Date
		series.EndDate = data[len(data)-1].Date
	}
	
	return series
}

// GetCloses returns all closing prices
func (hds *HistoricalDataSeries) GetCloses() []float64 {
	closes := make([]float64, len(hds.Data))
	for i, data := range hds.Data {
		closes[i] = data.Close
	}
	return closes
}

// GetVolumes returns all volumes
func (hds *HistoricalDataSeries) GetVolumes() []float64 {
	volumes := make([]float64, len(hds.Data))
	for i, data := range hds.Data {
		volumes[i] = data.Volume
	}
	return volumes
}

// GetLatest returns the most recent data point
func (hds *HistoricalDataSeries) GetLatest() *HistoricalData {
	if len(hds.Data) == 0 {
		return nil
	}
	return hds.Data[len(hds.Data)-1]
}

// GetRange returns data within a date range
func (hds *HistoricalDataSeries) GetRange(start, end time.Time) []*HistoricalData {
	var result []*HistoricalData
	for _, data := range hds.Data {
		if (data.Date.Equal(start) || data.Date.After(start)) &&
			(data.Date.Equal(end) || data.Date.Before(end)) {
			result = append(result, data)
		}
	}
	return result
}
