package utils

import (
	"math"
	"strconv"
)

// RoundToDecimal rounds a float64 to the specified number of decimal places
func RoundToDecimal(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// FormatFloat formats a float64 to a string with specified precision
func FormatFloat(value float64, precision int) string {
	return strconv.FormatFloat(value, 'f', precision, 64)
}

// CalculatePercentageChange calculates percentage change between two values
func CalculatePercentageChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		return 0
	}
	return ((newValue - oldValue) / oldValue) * 100
}

// Min returns the minimum of two float64 values
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two float64 values
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Abs returns the absolute value of a float64
func Abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

// IsValidPrice checks if a price is valid (positive and not NaN)
func IsValidPrice(price float64) bool {
	return price > 0 && !math.IsNaN(price) && !math.IsInf(price, 0)
}

// IsValidQuantity checks if a quantity is valid (positive and not NaN)
func IsValidQuantity(quantity float64) bool {
	return quantity > 0 && !math.IsNaN(quantity) && !math.IsInf(quantity, 0)
}

// NormalizeQuantity normalizes quantity to exchange precision
func NormalizeQuantity(quantity float64, stepSize float64) float64 {
	if stepSize == 0 {
		return quantity
	}
	
	steps := math.Floor(quantity / stepSize)
	return steps * stepSize
}

// NormalizePrice normalizes price to exchange precision
func NormalizePrice(price float64, tickSize float64) float64 {
	if tickSize == 0 {
		return price
	}
	
	ticks := math.Round(price / tickSize)
	return ticks * tickSize
}

// CalculateStandardDeviation calculates standard deviation of a slice of float64
func CalculateStandardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	mean := sum / float64(len(values))
	
	// Calculate variance
	variance := 0.0
	for _, value := range values {
		variance += math.Pow(value-mean, 2)
	}
	variance /= float64(len(values))
	
	return math.Sqrt(variance)
}

// CalculateMovingAverage calculates simple moving average
func CalculateMovingAverage(values []float64, period int) float64 {
	if len(values) < period {
		return 0
	}
	
	sum := 0.0
	for i := len(values) - period; i < len(values); i++ {
		sum += values[i]
	}
	
	return sum / float64(period)
}

// CalculateEMA calculates exponential moving average
func CalculateEMA(values []float64, period int) float64 {
	if len(values) == 0 {
		return 0
	}
	
	if len(values) == 1 {
		return values[0]
	}
	
	multiplier := 2.0 / (float64(period) + 1.0)
	ema := values[0]
	
	for i := 1; i < len(values); i++ {
		ema = (values[i] * multiplier) + (ema * (1 - multiplier))
	}
	
	return ema
}

// CalculateRSI calculates Relative Strength Index
func CalculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50 // Neutral RSI
	}
	
	gains := make([]float64, 0)
	losses := make([]float64, 0)
	
	// Calculate price changes
	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, -change)
		}
	}
	
	if len(gains) < period {
		return 50
	}
	
	// Calculate average gain and loss
	avgGain := CalculateMovingAverage(gains, period)
	avgLoss := CalculateMovingAverage(losses, period)
	
	if avgLoss == 0 {
		return 100
	}
	
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	
	return rsi
}

// CalculateVolatility calculates price volatility (standard deviation of returns)
func CalculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}
	
	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
	}
	
	return CalculateStandardDeviation(returns)
}

// CalculateSharpeRatio calculates Sharpe ratio (return/risk)
func CalculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	
	avgReturn := 0.0
	for _, ret := range returns {
		avgReturn += ret
	}
	avgReturn /= float64(len(returns))
	
	excessReturn := avgReturn - riskFreeRate
	volatility := CalculateStandardDeviation(returns)
	
	if volatility == 0 {
		return 0
	}
	
	return excessReturn / volatility
}

// CalculateMaxDrawdown calculates maximum drawdown from a series of cumulative returns
func CalculateMaxDrawdown(cumulativeReturns []float64) float64 {
	if len(cumulativeReturns) == 0 {
		return 0
	}
	
	peak := cumulativeReturns[0]
	maxDrawdown := 0.0
	
	for _, value := range cumulativeReturns {
		if value > peak {
			peak = value
		}
		
		drawdown := (peak - value) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	
	return maxDrawdown
}

// CalculateVaR calculates Value at Risk at specified confidence level
func CalculateVaR(returns []float64, confidenceLevel float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	
	// Sort returns in ascending order
	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)
	
	for i := 0; i < len(sortedReturns); i++ {
		for j := i + 1; j < len(sortedReturns); j++ {
			if sortedReturns[i] > sortedReturns[j] {
				sortedReturns[i], sortedReturns[j] = sortedReturns[j], sortedReturns[i]
			}
		}
	}
	
	// Calculate index for the confidence level
	index := int(float64(len(sortedReturns)) * (1 - confidenceLevel))
	if index >= len(sortedReturns) {
		index = len(sortedReturns) - 1
	}
	
	return math.Abs(sortedReturns[index])
}
