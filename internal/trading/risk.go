package trading

import (
	"context"
	"math"
	"time"

	"github.com/sirupsen/logrus"
)

// RiskManager handles risk management and validation
type RiskManager struct {
	config    *RiskConfig
	logger    *logrus.Logger
	
	// Track daily metrics
	dailyLoss     float64
	dailyTrades   int
	lastResetDate time.Time
	
	// Position tracking
	totalExposure float64
	maxExposure   float64
}

// RiskConfig holds risk management configuration
type RiskConfig struct {
	MaxPositionSize   float64 `json:"max_position_size"`
	StopLossPercent   float64 `json:"stop_loss_percent"`
	TakeProfitPercent float64 `json:"take_profit_percent"`
	MaxDailyLoss      float64 `json:"max_daily_loss"`
	MaxLeverage       int     `json:"max_leverage"`
	RiskPerTrade      float64 `json:"risk_per_trade_percent"`
	MaxDrawdown       float64 `json:"max_drawdown"`
	MaxOpenPositions  int     `json:"max_open_positions"`
	MinOrderValue     float64 `json:"min_order_value"`
	MaxOrderValue     float64 `json:"max_order_value"`
	VaRLimit          float64 `json:"var_limit"`          // Value at Risk limit
	CorrelationLimit  float64 `json:"correlation_limit"`  // Max correlation between positions
}

// NewRiskManager creates a new risk manager
func NewRiskManager(config *RiskConfig) *RiskManager {
	return &RiskManager{
		config:        config,
		logger:        logrus.New(),
		lastResetDate: time.Now(),
		maxExposure:   config.MaxPositionSize * 10, // Default max exposure
	}
}

// ValidateOrder validates if an order meets risk criteria
func (rm *RiskManager) ValidateOrder(ctx context.Context, order *OrderInfo) bool {
	// Reset daily counters if new day
	rm.resetDailyCountersIfNeeded()
	
	// Check if trading is allowed
	if !rm.isTradingAllowed() {
		rm.logger.Warn("Trading not allowed due to risk limits")
		return false
	}
	
	// Validate order size
	if !rm.validateOrderSize(order) {
		rm.logger.Warnf("Order size validation failed for %s", order.Symbol)
		return false
	}
	
	// Validate position size
	if !rm.validatePositionSize(order) {
		rm.logger.Warnf("Position size validation failed for %s", order.Symbol)
		return false
	}
	
	// Validate daily loss limit
	if !rm.validateDailyLossLimit(order) {
		rm.logger.Warn("Daily loss limit validation failed")
		return false
	}
	
	// Validate exposure limits
	if !rm.validateExposureLimit(order) {
		rm.logger.Warnf("Exposure limit validation failed for %s", order.Symbol)
		return false
	}
	
	// Validate risk per trade
	if !rm.validateRiskPerTrade(order) {
		rm.logger.Warnf("Risk per trade validation failed for %s", order.Symbol)
		return false
	}
	
	rm.logger.Infof("Order validation passed for %s", order.Symbol)
	return true
}

// validateOrderSize checks if order size is within limits
func (rm *RiskManager) validateOrderSize(order *OrderInfo) bool {
	orderValue := order.Quantity * order.Price
	
	// Check minimum order value
	if orderValue < rm.config.MinOrderValue {
		rm.logger.Debugf("Order value %.2f below minimum %.2f", orderValue, rm.config.MinOrderValue)
		return false
	}
	
	// Check maximum order value
	if rm.config.MaxOrderValue > 0 && orderValue > rm.config.MaxOrderValue {
		rm.logger.Debugf("Order value %.2f exceeds maximum %.2f", orderValue, rm.config.MaxOrderValue)
		return false
	}
	
	return true
}

// validatePositionSize checks if position size is within limits
func (rm *RiskManager) validatePositionSize(order *OrderInfo) bool {
	orderValue := order.Quantity * order.Price
	
	if orderValue > rm.config.MaxPositionSize {
		rm.logger.Debugf("Position size %.2f exceeds maximum %.2f", orderValue, rm.config.MaxPositionSize)
		return false
	}
	
	return true
}

// validateDailyLossLimit checks daily loss limits
func (rm *RiskManager) validateDailyLossLimit(order *OrderInfo) bool {
	if rm.dailyLoss >= rm.config.MaxDailyLoss {
		rm.logger.Debugf("Daily loss %.2f exceeds limit %.2f", rm.dailyLoss, rm.config.MaxDailyLoss)
		return false
	}
	
	return true
}

// validateExposureLimit checks total exposure limits
func (rm *RiskManager) validateExposureLimit(order *OrderInfo) bool {
	orderValue := order.Quantity * order.Price
	newExposure := rm.totalExposure + orderValue
	
	if newExposure > rm.maxExposure {
		rm.logger.Debugf("New exposure %.2f would exceed limit %.2f", newExposure, rm.maxExposure)
		return false
	}
	
	return true
}

// validateRiskPerTrade checks risk per trade limits
func (rm *RiskManager) validateRiskPerTrade(order *OrderInfo) bool {
	orderValue := order.Quantity * order.Price
	riskAmount := orderValue * (rm.config.RiskPerTrade / 100.0)
	
	// This is a simplified check - in reality you'd want to factor in stop loss distance
	maxRiskPerTrade := rm.config.MaxPositionSize * (rm.config.RiskPerTrade / 100.0)
	
	if riskAmount > maxRiskPerTrade {
		rm.logger.Debugf("Risk amount %.2f exceeds limit %.2f", riskAmount, maxRiskPerTrade)
		return false
	}
	
	return true
}

// isTradingAllowed checks if trading is currently allowed
func (rm *RiskManager) isTradingAllowed() bool {
	// Check if max daily trades reached
	if rm.config.MaxOpenPositions > 0 && rm.dailyTrades >= rm.config.MaxOpenPositions {
		return false
	}
	
	// Check if daily loss limit reached
	if rm.dailyLoss >= rm.config.MaxDailyLoss {
		return false
	}
	
	return true
}

// resetDailyCountersIfNeeded resets daily counters at start of new day
func (rm *RiskManager) resetDailyCountersIfNeeded() {
	now := time.Now()
	if now.Day() != rm.lastResetDate.Day() || now.Month() != rm.lastResetDate.Month() || now.Year() != rm.lastResetDate.Year() {
		rm.dailyLoss = 0
		rm.dailyTrades = 0
		rm.lastResetDate = now
		rm.logger.Info("Daily risk counters reset")
	}
}

// UpdateDailyLoss updates the daily loss tracking
func (rm *RiskManager) UpdateDailyLoss(loss float64) {
	rm.resetDailyCountersIfNeeded()
	rm.dailyLoss += loss
	rm.logger.Debugf("Daily loss updated: %.2f", rm.dailyLoss)
}

// UpdateDailyTrades updates the daily trade count
func (rm *RiskManager) UpdateDailyTrades() {
	rm.resetDailyCountersIfNeeded()
	rm.dailyTrades++
	rm.logger.Debugf("Daily trades updated: %d", rm.dailyTrades)
}

// UpdateExposure updates total exposure tracking
func (rm *RiskManager) UpdateExposure(exposure float64) {
	rm.totalExposure = exposure
	rm.logger.Debugf("Total exposure updated: %.2f", rm.totalExposure)
}

// CalculatePositionSize calculates optimal position size based on risk parameters
func (rm *RiskManager) CalculatePositionSize(accountBalance, entryPrice, stopLoss float64) float64 {
	// Calculate risk amount per trade
	riskAmount := accountBalance * (rm.config.RiskPerTrade / 100.0)
	
	// Calculate stop loss distance as percentage
	stopLossDistance := math.Abs(entryPrice-stopLoss) / entryPrice
	
	// Calculate position size
	positionValue := riskAmount / stopLossDistance
	quantity := positionValue / entryPrice
	
	// Ensure position doesn't exceed maximum
	maxQuantity := rm.config.MaxPositionSize / entryPrice
	if quantity > maxQuantity {
		quantity = maxQuantity
	}
	
	return quantity
}

// CalculateStopLoss calculates stop loss price based on risk parameters
func (rm *RiskManager) CalculateStopLoss(entryPrice float64, side string) float64 {
	stopLossPercent := rm.config.StopLossPercent / 100.0
	
	if side == "BUY" || side == "LONG" {
		return entryPrice * (1.0 - stopLossPercent)
	} else {
		return entryPrice * (1.0 + stopLossPercent)
	}
}

// CalculateTakeProfit calculates take profit price based on risk parameters
func (rm *RiskManager) CalculateTakeProfit(entryPrice float64, side string) float64 {
	takeProfitPercent := rm.config.TakeProfitPercent / 100.0
	
	if side == "BUY" || side == "LONG" {
		return entryPrice * (1.0 + takeProfitPercent)
	} else {
		return entryPrice * (1.0 - takeProfitPercent)
	}
}

// GetRiskMetrics returns current risk metrics
func (rm *RiskManager) GetRiskMetrics() *RiskMetrics {
	rm.resetDailyCountersIfNeeded()
	
	return &RiskMetrics{
		DailyLoss:        rm.dailyLoss,
		DailyTrades:      rm.dailyTrades,
		TotalExposure:    rm.totalExposure,
		MaxExposure:      rm.maxExposure,
		ExposureRatio:    rm.totalExposure / rm.maxExposure,
		RemainingRisk:    math.Max(0, rm.config.MaxDailyLoss-rm.dailyLoss),
		TradingAllowed:   rm.isTradingAllowed(),
		LastResetDate:    rm.lastResetDate,
	}
}

// RiskMetrics represents current risk metrics
type RiskMetrics struct {
	DailyLoss      float64   `json:"daily_loss"`
	DailyTrades    int       `json:"daily_trades"`
	TotalExposure  float64   `json:"total_exposure"`
	MaxExposure    float64   `json:"max_exposure"`
	ExposureRatio  float64   `json:"exposure_ratio"`
	RemainingRisk  float64   `json:"remaining_risk"`
	TradingAllowed bool      `json:"trading_allowed"`
	LastResetDate  time.Time `json:"last_reset_date"`
}

// ValidatePortfolio validates the entire portfolio risk
func (rm *RiskManager) ValidatePortfolio(positions []PortfolioPosition) *PortfolioRisk {
	var totalValue, totalPnL, totalExposure float64
	var correlationRisk float64
	
	for _, pos := range positions {
		totalValue += pos.Value
		totalPnL += pos.UnrealizedPnL
		totalExposure += math.Abs(pos.Value)
	}
	
	// Calculate portfolio metrics
	portfolioReturn := 0.0
	if totalValue > 0 {
		portfolioReturn = totalPnL / totalValue * 100
	}
	
	// Simple VaR calculation (95% confidence)
	var returns []float64
	for _, pos := range positions {
		if pos.Value > 0 {
			returns = append(returns, pos.UnrealizedPnL/pos.Value)
		}
	}
	
	var95 := rm.calculateVaR95(returns)
	
	// Check risk limits
	isValid := true
	var violations []string
	
	if totalExposure > rm.maxExposure {
		isValid = false
		violations = append(violations, "Total exposure exceeds limit")
	}
	
	if var95 > rm.config.VaRLimit {
		isValid = false
		violations = append(violations, "VaR exceeds limit")
	}
	
	if math.Abs(portfolioReturn) > rm.config.MaxDrawdown {
		isValid = false
		violations = append(violations, "Drawdown exceeds limit")
	}
	
	return &PortfolioRisk{
		TotalValue:       totalValue,
		TotalPnL:         totalPnL,
		TotalExposure:    totalExposure,
		PortfolioReturn:  portfolioReturn,
		VaR95:           var95,
		CorrelationRisk: correlationRisk,
		IsValid:         isValid,
		Violations:      violations,
	}
}

// calculateVaR95 calculates 95% Value at Risk
func (rm *RiskManager) calculateVaR95(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	
	// Sort returns
	for i := 0; i < len(returns); i++ {
		for j := i + 1; j < len(returns); j++ {
			if returns[i] > returns[j] {
				returns[i], returns[j] = returns[j], returns[i]
			}
		}
	}
	
	// Get 5th percentile (95% VaR)
	index := int(float64(len(returns)) * 0.05)
	if index >= len(returns) {
		index = len(returns) - 1
	}
	
	return math.Abs(returns[index])
}

// PortfolioPosition represents a position in the portfolio
type PortfolioPosition struct {
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Size          float64 `json:"size"`
	EntryPrice    float64 `json:"entry_price"`
	CurrentPrice  float64 `json:"current_price"`
	Value         float64 `json:"value"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
	Leverage      int     `json:"leverage"`
}

// PortfolioRisk represents portfolio risk assessment
type PortfolioRisk struct {
	TotalValue       float64  `json:"total_value"`
	TotalPnL         float64  `json:"total_pnl"`
	TotalExposure    float64  `json:"total_exposure"`
	PortfolioReturn  float64  `json:"portfolio_return"`
	VaR95           float64  `json:"var_95"`
	CorrelationRisk float64  `json:"correlation_risk"`
	IsValid         bool     `json:"is_valid"`
	Violations      []string `json:"violations"`
}

// EmergencyStop implements emergency stop functionality
func (rm *RiskManager) EmergencyStop(reason string) {
	rm.logger.Errorf("EMERGENCY STOP TRIGGERED: %s", reason)
	
	// Set daily loss to maximum to prevent further trading
	rm.dailyLoss = rm.config.MaxDailyLoss
	
	// Additional emergency procedures could be implemented here
	// Such as closing all positions, sending alerts, etc.
}

// ShouldClosePosition determines if a position should be closed due to risk
func (rm *RiskManager) ShouldClosePosition(position PortfolioPosition) (bool, string) {
	// Check stop loss
	if position.Side == "LONG" && position.CurrentPrice <= rm.CalculateStopLoss(position.EntryPrice, "BUY") {
		return true, "Stop loss triggered"
	}
	
	if position.Side == "SHORT" && position.CurrentPrice >= rm.CalculateStopLoss(position.EntryPrice, "SELL") {
		return true, "Stop loss triggered"
	}
	
	// Check take profit
	if position.Side == "LONG" && position.CurrentPrice >= rm.CalculateTakeProfit(position.EntryPrice, "BUY") {
		return true, "Take profit triggered"
	}
	
	if position.Side == "SHORT" && position.CurrentPrice <= rm.CalculateTakeProfit(position.EntryPrice, "SELL") {
		return true, "Take profit triggered"
	}
	
	// Check maximum loss per position
	maxLossPercent := rm.config.StopLossPercent * 2 // Double stop loss as emergency exit
	currentLossPercent := math.Abs(position.UnrealizedPnL/position.Value) * 100
	
	if currentLossPercent > maxLossPercent {
		return true, "Maximum loss exceeded"
	}
	
	return false, ""
}
