package trading

import (
	"context"
	"fmt"
	"math"

	"contract_playground/internal/models"
)

// SMAStrategy implements Simple Moving Average strategy
type SMAStrategy struct {
	name            string
	shortPeriod     int
	longPeriod      int
	minConfidence   float64
	priceHistory    map[string][]float64
}

// NewSMAStrategy creates a new SMA strategy
func NewSMAStrategy() Strategy {
	return &SMAStrategy{
		name:          "Simple Moving Average",
		shortPeriod:   10,
		longPeriod:    20,
		minConfidence: 0.7,
		priceHistory:  make(map[string][]float64),
	}
}

// Name returns the strategy name
func (s *SMAStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with parameters
func (s *SMAStrategy) Initialize(config map[string]interface{}) error {
	if val, ok := config["short_period"]; ok {
		if period, ok := val.(float64); ok {
			s.shortPeriod = int(period)
		}
	}
	
	if val, ok := config["long_period"]; ok {
		if period, ok := val.(float64); ok {
			s.longPeriod = int(period)
		}
	}
	
	if val, ok := config["min_confidence"]; ok {
		if conf, ok := val.(float64); ok {
			s.minConfidence = conf
		}
	}
	
	if s.shortPeriod >= s.longPeriod {
		return fmt.Errorf("short period must be less than long period")
	}
	
	return nil
}

// ShouldBuy determines if we should buy
func (s *SMAStrategy) ShouldBuy(ctx context.Context, symbol string, data *MarketData) (*Signal, error) {
	s.updatePriceHistory(symbol, data.Price)
	
	prices := s.priceHistory[symbol]
	if len(prices) < s.longPeriod {
		return &Signal{Action: "HOLD", Reason: "Insufficient data"}, nil
	}
	
	shortSMA := s.calculateSMA(prices, s.shortPeriod)
	longSMA := s.calculateSMA(prices, s.longPeriod)
	
	// Buy signal: short SMA crosses above long SMA
	if shortSMA > longSMA {
		// Calculate crossover strength for confidence
		crossoverStrength := (shortSMA - longSMA) / longSMA
		confidence := math.Min(crossoverStrength*10, 1.0) // Scale to 0-1
		
		if confidence >= s.minConfidence {
			quantity := s.calculateQuantity(data.Price, 1000) // $1000 position
			
			return &Signal{
				Action:       "BUY",
				Quantity:     quantity,
				Price:        data.Price,
				Confidence:   confidence,
				Reason:       fmt.Sprintf("SMA crossover: short=%.2f, long=%.2f", shortSMA, longSMA),
				PositionSide: "LONG",
			}, nil
		}
	}
	
	return &Signal{Action: "HOLD", Reason: "No buy signal"}, nil
}

// ShouldSell determines if we should sell
func (s *SMAStrategy) ShouldSell(ctx context.Context, symbol string, data *MarketData, position *models.Position) (*Signal, error) {
	s.updatePriceHistory(symbol, data.Price)
	
	prices := s.priceHistory[symbol]
	if len(prices) < s.longPeriod {
		return &Signal{Action: "HOLD", Reason: "Insufficient data"}, nil
	}
	
	shortSMA := s.calculateSMA(prices, s.shortPeriod)
	longSMA := s.calculateSMA(prices, s.longPeriod)
	
	// Sell signal: short SMA crosses below long SMA
	if shortSMA < longSMA {
		crossoverStrength := (longSMA - shortSMA) / longSMA
		confidence := math.Min(crossoverStrength*10, 1.0)
		
		if confidence >= s.minConfidence {
			return &Signal{
				Action:     "SELL",
				Quantity:   position.Size,
				Price:      data.Price,
				Confidence: confidence,
				Reason:     fmt.Sprintf("SMA crossover: short=%.2f, long=%.2f", shortSMA, longSMA),
			}, nil
		}
	}
	
	// Also check for stop loss or take profit
	pnlPercent := (data.Price - position.EntryPrice) / position.EntryPrice * 100
	
	if pnlPercent <= -2.0 { // 2% stop loss
		return &Signal{
			Action:     "SELL",
			Quantity:   position.Size,
			Price:      data.Price,
			Confidence: 1.0,
			Reason:     fmt.Sprintf("Stop loss triggered: %.2f%%", pnlPercent),
		}, nil
	}
	
	if pnlPercent >= 5.0 { // 5% take profit
		return &Signal{
			Action:     "SELL",
			Quantity:   position.Size,
			Price:      data.Price,
			Confidence: 1.0,
			Reason:     fmt.Sprintf("Take profit triggered: %.2f%%", pnlPercent),
		}, nil
	}
	
	return &Signal{Action: "HOLD", Reason: "No sell signal"}, nil
}

// updatePriceHistory updates the price history for a symbol
func (s *SMAStrategy) updatePriceHistory(symbol string, price float64) {
	if s.priceHistory[symbol] == nil {
		s.priceHistory[symbol] = make([]float64, 0)
	}
	
	s.priceHistory[symbol] = append(s.priceHistory[symbol], price)
	
	// Keep only the data we need
	maxLength := s.longPeriod + 10
	if len(s.priceHistory[symbol]) > maxLength {
		s.priceHistory[symbol] = s.priceHistory[symbol][len(s.priceHistory[symbol])-maxLength:]
	}
}

// calculateSMA calculates Simple Moving Average
func (s *SMAStrategy) calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	
	return sum / float64(period)
}

// calculateQuantity calculates position quantity based on position value
func (s *SMAStrategy) calculateQuantity(price, positionValue float64) float64 {
	return positionValue / price
}

// RSIStrategy implements RSI strategy
type RSIStrategy struct {
	name          string
	period        int
	oversold      float64
	overbought    float64
	minConfidence float64
	priceHistory  map[string][]float64
}

// NewRSIStrategy creates a new RSI strategy
func NewRSIStrategy() Strategy {
	return &RSIStrategy{
		name:          "RSI Strategy",
		period:        14,
		oversold:      30,
		overbought:    70,
		minConfidence: 0.6,
		priceHistory:  make(map[string][]float64),
	}
}

// Name returns the strategy name
func (r *RSIStrategy) Name() string {
	return r.name
}

// Initialize initializes the strategy with parameters
func (r *RSIStrategy) Initialize(config map[string]interface{}) error {
	if val, ok := config["period"]; ok {
		if period, ok := val.(float64); ok {
			r.period = int(period)
		}
	}
	
	if val, ok := config["oversold"]; ok {
		if oversold, ok := val.(float64); ok {
			r.oversold = oversold
		}
	}
	
	if val, ok := config["overbought"]; ok {
		if overbought, ok := val.(float64); ok {
			r.overbought = overbought
		}
	}
	
	if val, ok := config["min_confidence"]; ok {
		if conf, ok := val.(float64); ok {
			r.minConfidence = conf
		}
	}
	
	return nil
}

// ShouldBuy determines if we should buy based on RSI
func (r *RSIStrategy) ShouldBuy(ctx context.Context, symbol string, data *MarketData) (*Signal, error) {
	r.updatePriceHistory(symbol, data.Price)
	
	prices := r.priceHistory[symbol]
	if len(prices) < r.period+1 {
		return &Signal{Action: "HOLD", Reason: "Insufficient data for RSI"}, nil
	}
	
	rsi := r.calculateRSI(prices)
	
	// Buy signal: RSI is oversold
	if rsi < r.oversold {
		confidence := (r.oversold - rsi) / r.oversold
		
		if confidence >= r.minConfidence {
			quantity := r.calculateQuantity(data.Price, 1000)
			
			return &Signal{
				Action:       "BUY",
				Quantity:     quantity,
				Price:        data.Price,
				Confidence:   confidence,
				Reason:       fmt.Sprintf("RSI oversold: %.2f", rsi),
				PositionSide: "LONG",
			}, nil
		}
	}
	
	return &Signal{Action: "HOLD", Reason: fmt.Sprintf("RSI: %.2f", rsi)}, nil
}

// ShouldSell determines if we should sell based on RSI
func (r *RSIStrategy) ShouldSell(ctx context.Context, symbol string, data *MarketData, position *models.Position) (*Signal, error) {
	r.updatePriceHistory(symbol, data.Price)
	
	prices := r.priceHistory[symbol]
	if len(prices) < r.period+1 {
		return &Signal{Action: "HOLD", Reason: "Insufficient data for RSI"}, nil
	}
	
	rsi := r.calculateRSI(prices)
	
	// Sell signal: RSI is overbought
	if rsi > r.overbought {
		confidence := (rsi - r.overbought) / (100 - r.overbought)
		
		if confidence >= r.minConfidence {
			return &Signal{
				Action:     "SELL",
				Quantity:   position.Size,
				Price:      data.Price,
				Confidence: confidence,
				Reason:     fmt.Sprintf("RSI overbought: %.2f", rsi),
			}, nil
		}
	}
	
	// Check stop loss and take profit
	pnlPercent := (data.Price - position.EntryPrice) / position.EntryPrice * 100
	
	if pnlPercent <= -2.0 {
		return &Signal{
			Action:     "SELL",
			Quantity:   position.Size,
			Price:      data.Price,
			Confidence: 1.0,
			Reason:     fmt.Sprintf("Stop loss: %.2f%%", pnlPercent),
		}, nil
	}
	
	if pnlPercent >= 5.0 {
		return &Signal{
			Action:     "SELL",
			Quantity:   position.Size,
			Price:      data.Price,
			Confidence: 1.0,
			Reason:     fmt.Sprintf("Take profit: %.2f%%", pnlPercent),
		}, nil
	}
	
	return &Signal{Action: "HOLD", Reason: fmt.Sprintf("RSI: %.2f", rsi)}, nil
}

// updatePriceHistory updates the price history for RSI calculation
func (r *RSIStrategy) updatePriceHistory(symbol string, price float64) {
	if r.priceHistory[symbol] == nil {
		r.priceHistory[symbol] = make([]float64, 0)
	}
	
	r.priceHistory[symbol] = append(r.priceHistory[symbol], price)
	
	// Keep only the data we need
	maxLength := r.period + 20
	if len(r.priceHistory[symbol]) > maxLength {
		r.priceHistory[symbol] = r.priceHistory[symbol][len(r.priceHistory[symbol])-maxLength:]
	}
}

// calculateRSI calculates the Relative Strength Index
func (r *RSIStrategy) calculateRSI(prices []float64) float64 {
	if len(prices) < r.period+1 {
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
	
	if len(gains) < r.period {
		return 50
	}
	
	// Calculate average gain and loss over the period
	avgGain := 0.0
	avgLoss := 0.0
	
	// Initial averages
	for i := 0; i < r.period; i++ {
		avgGain += gains[len(gains)-r.period+i]
		avgLoss += losses[len(losses)-r.period+i]
	}
	
	avgGain /= float64(r.period)
	avgLoss /= float64(r.period)
	
	if avgLoss == 0 {
		return 100
	}
	
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	
	return rsi
}

// calculateQuantity calculates position quantity for RSI strategy
func (r *RSIStrategy) calculateQuantity(price, positionValue float64) float64 {
	return positionValue / price
}

// GridStrategy implements a grid trading strategy
type GridStrategy struct {
	name          string
	gridSize      float64
	numGrids      int
	basePrice     float64
	positions     map[string][]GridPosition
	minConfidence float64
}

// GridPosition represents a position in the grid
type GridPosition struct {
	Price    float64
	Quantity float64
	Active   bool
}

// NewGridStrategy creates a new grid strategy
func NewGridStrategy() Strategy {
	return &GridStrategy{
		name:          "Grid Strategy",
		gridSize:      0.01, // 1% grid size
		numGrids:      10,
		positions:     make(map[string][]GridPosition),
		minConfidence: 0.8,
	}
}

// Name returns the strategy name
func (g *GridStrategy) Name() string {
	return g.name
}

// Initialize initializes the grid strategy
func (g *GridStrategy) Initialize(config map[string]interface{}) error {
	if val, ok := config["grid_size"]; ok {
		if size, ok := val.(float64); ok {
			g.gridSize = size
		}
	}
	
	if val, ok := config["num_grids"]; ok {
		if num, ok := val.(float64); ok {
			g.numGrids = int(num)
		}
	}
	
	if val, ok := config["min_confidence"]; ok {
		if conf, ok := val.(float64); ok {
			g.minConfidence = conf
		}
	}
	
	return nil
}

// ShouldBuy determines if we should buy based on grid strategy
func (g *GridStrategy) ShouldBuy(ctx context.Context, symbol string, data *MarketData) (*Signal, error) {
	if g.basePrice == 0 {
		g.basePrice = data.Price
		g.initializeGrid(symbol, data.Price)
	}
	
	// Find the appropriate grid level
	gridLevel := g.findGridLevel(data.Price)
	if gridLevel < 0 || gridLevel >= len(g.positions[symbol]) {
		return &Signal{Action: "HOLD", Reason: "Price outside grid range"}, nil
	}
	
	// Buy at support levels (lower grid levels)
	if data.Price <= g.positions[symbol][gridLevel].Price && !g.positions[symbol][gridLevel].Active {
		quantity := g.calculateGridQuantity(data.Price)
		
		return &Signal{
			Action:       "BUY",
			Quantity:     quantity,
			Price:        data.Price,
			Confidence:   g.minConfidence,
			Reason:       fmt.Sprintf("Grid buy at level %d", gridLevel),
			PositionSide: "LONG",
		}, nil
	}
	
	return &Signal{Action: "HOLD", Reason: "No grid buy signal"}, nil
}

// ShouldSell determines if we should sell based on grid strategy
func (g *GridStrategy) ShouldSell(ctx context.Context, symbol string, data *MarketData, position *models.Position) (*Signal, error) {
	if g.basePrice == 0 {
		return &Signal{Action: "HOLD", Reason: "Grid not initialized"}, nil
	}
	
	// Sell at resistance levels (higher grid levels)
	profitTarget := position.EntryPrice * (1 + g.gridSize)
	
	if data.Price >= profitTarget {
		return &Signal{
			Action:     "SELL",
			Quantity:   position.Size,
			Price:      data.Price,
			Confidence: g.minConfidence,
			Reason:     fmt.Sprintf("Grid sell target reached: %.2f", profitTarget),
		}, nil
	}
	
	// Stop loss
	stopLoss := position.EntryPrice * (1 - g.gridSize*2)
	if data.Price <= stopLoss {
		return &Signal{
			Action:     "SELL",
			Quantity:   position.Size,
			Price:      data.Price,
			Confidence: 1.0,
			Reason:     fmt.Sprintf("Grid stop loss: %.2f", stopLoss),
		}, nil
	}
	
	return &Signal{Action: "HOLD", Reason: "No grid sell signal"}, nil
}

// initializeGrid initializes the trading grid
func (g *GridStrategy) initializeGrid(symbol string, basePrice float64) {
	g.positions[symbol] = make([]GridPosition, g.numGrids)
	
	for i := 0; i < g.numGrids; i++ {
		offset := float64(i-g.numGrids/2) * g.gridSize
		price := basePrice * (1 + offset)
		
		g.positions[symbol][i] = GridPosition{
			Price:    price,
			Quantity: 0,
			Active:   false,
		}
	}
}

// findGridLevel finds the grid level for a given price
func (g *GridStrategy) findGridLevel(price float64) int {
	if g.basePrice == 0 {
		return -1
	}
	
	offset := (price - g.basePrice) / g.basePrice / g.gridSize
	level := int(offset) + g.numGrids/2
	
	return level
}

// calculateGridQuantity calculates quantity for grid trading
func (g *GridStrategy) calculateGridQuantity(price float64) float64 {
	return 100 / price // Fixed $100 per grid level
}
