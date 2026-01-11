package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateID 生成唯一ID
func GenerateID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return fmt.Sprintf("%d-%s", timestamp, hex.EncodeToString(randomBytes))
}

// FormatMoney 格式化金额（万元）
func FormatMoney(value float64) string {
	wan := value / 10000
	if wan >= 10000 {
		return fmt.Sprintf("%.2f亿", wan/10000)
	}
	return fmt.Sprintf("%.2f万", wan)
}

// Min 返回最小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max 返回最大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinFloat 返回最小值
func MinFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// MaxFloat 返回最大值
func MaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Clamp 限制值在范围内
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
