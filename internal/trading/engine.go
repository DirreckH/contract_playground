package trading

import (
	"context"
	"fmt"
	"sync"
	"time"

	"contract_playground/internal/config"
	"contract_playground/internal/database"
	"contract_playground/internal/exchange"
	"contract_playground/internal/models"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Engine represents the main trading engine
type Engine struct {
	config         config.TradingConfig
	db             *gorm.DB
	redis          *redis.Client
	repository     database.Repository
	exchangeClient exchange.Client
	logger         *logrus.Logger

	// Internal state
	isRunning bool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc

	// Strategy and risk management
	strategy    Strategy
	riskManager *RiskManager

	// Market data
	marketData   map[string]*exchange.KlineData
	marketDataMu sync.RWMutex

	// Performance tracking
	dailyPnL      float64
	totalTrades   int
	winningTrades int
	losingTrades  int
}

// EngineConfig holds the configuration for the trading engine
type EngineConfig struct {
	DB             *gorm.DB
	Redis          *redis.Client
	ExchangeClient exchange.Client
	Config         config.TradingConfig
	Logger         *logrus.Logger
}

// Strategy interface for trading strategies
type Strategy interface {
	Name() string
	ShouldBuy(ctx context.Context, symbol string, data *MarketData) (*Signal, error)
	ShouldSell(ctx context.Context, symbol string, data *MarketData, position *models.Position) (*Signal, error)
	Initialize(config map[string]interface{}) error
}

// Signal represents a trading signal
type Signal struct {
	Action       string // BUY, SELL, HOLD
	Quantity     float64
	Price        float64
	StopLoss     float64
	TakeProfit   float64
	Confidence   float64 // 0.0 to 1.0
	Reason       string
	PositionSide string // LONG, SHORT
}

// MarketData represents current market information
type MarketData struct {
	Symbol    string
	Price     float64
	Volume    float64
	Change    float64
	Timestamp time.Time
	Klines    []*exchange.KlineData
}

// NewEngine creates a new trading engine
func NewEngine(cfg *EngineConfig) *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	repository := database.NewMySQLRepository(cfg.DB)

	// Initialize strategy based on config
	var strategy Strategy
	switch cfg.Config.Strategy.Type {
	case "simple_moving_average":
		strategy = NewSMAStrategy()
	case "rsi":
		strategy = NewRSIStrategy()
	case "ai":
		strategy = NewAIStrategy()
	default:
		strategy = NewSMAStrategy() // Default strategy
	}

	// Initialize strategy with parameters
	if err := strategy.Initialize(cfg.Config.Strategy.Parameters); err != nil {
		cfg.Logger.Errorf("Failed to initialize strategy: %v", err)
	}

	// Initialize risk manager
	riskManager := NewRiskManager(&RiskConfig{
		MaxPositionSize:   cfg.Config.MaxPositionSize,
		StopLossPercent:   cfg.Config.StopLossPercent,
		TakeProfitPercent: cfg.Config.TakeProfitPercent,
		MaxDailyLoss:      cfg.Config.MaxDailyLoss,
		MaxLeverage:       cfg.Config.MaxLeverage,
		RiskPerTrade:      cfg.Config.RiskPerTrade,
	})

	return &Engine{
		config:         cfg.Config,
		db:             cfg.DB,
		redis:          cfg.Redis,
		repository:     repository,
		exchangeClient: cfg.ExchangeClient,
		logger:         cfg.Logger,
		ctx:            ctx,
		cancel:         cancel,
		strategy:       strategy,
		riskManager:    riskManager,
		marketData:     make(map[string]*exchange.KlineData),
		isRunning:      false,
	}
}

// Start starts the trading engine
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning {
		return fmt.Errorf("trading engine is already running")
	}

	e.isRunning = true
	e.logger.Info("Starting trading engine...")

	// Initialize symbols and leverage
	if err := e.initializeSymbols(ctx); err != nil {
		return fmt.Errorf("failed to initialize symbols: %w", err)
	}

	// Start market data collection
	go e.collectMarketData(ctx)

	// Start trading loop
	go e.tradingLoop(ctx)

	// Start risk monitoring
	go e.monitorRisk(ctx)

	// Start account monitoring
	go e.monitorAccount(ctx)

	e.logger.Info("Trading engine started successfully")
	return nil
}

// Stop stops the trading engine
func (e *Engine) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning {
		return nil
	}

	e.logger.Info("Stopping trading engine...")

	// Cancel context to stop all goroutines
	e.cancel()

	// Close all positions if needed (optional)
	if err := e.closeAllPositions(ctx); err != nil {
		e.logger.Errorf("Error closing positions during shutdown: %v", err)
	}

	e.isRunning = false
	e.logger.Info("Trading engine stopped")
	return nil
}

// initializeSymbols sets up trading symbols with leverage and margin type
func (e *Engine) initializeSymbols(ctx context.Context) error {
	for _, symbol := range e.config.Symbols {
		// Set leverage
		if err := e.exchangeClient.SetLeverage(ctx, symbol, e.config.MaxLeverage); err != nil {
			e.logger.Warnf("Failed to set leverage for %s: %v", symbol, err)
		}

		// Set margin type to CROSSED (default for most strategies)
		if err := e.exchangeClient.ChangeMarginType(ctx, symbol, "CROSSED"); err != nil {
			e.logger.Warnf("Failed to set margin type for %s: %v", symbol, err)
		}

		e.logger.Infof("Initialized symbol %s with leverage %d", symbol, e.config.MaxLeverage)
	}

	return nil
}

// collectMarketData continuously collects market data
func (e *Engine) collectMarketData(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(e.config.TradingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, symbol := range e.config.Symbols {
				if err := e.updateMarketData(ctx, symbol); err != nil {
					e.logger.Errorf("Failed to update market data for %s: %v", symbol, err)
				}
			}
		}
	}
}

// updateMarketData updates market data for a symbol
func (e *Engine) updateMarketData(ctx context.Context, symbol string) error {
	// Get current price
	price, err := e.exchangeClient.GetSymbolPrice(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get price for %s: %w", symbol, err)
	}

	// Get kline data for strategy analysis
	klines, err := e.exchangeClient.GetKlines(ctx, symbol, "1m", 100)
	if err != nil {
		return fmt.Errorf("failed to get klines for %s: %w", symbol, err)
	}

	if len(klines) > 0 {
		e.marketDataMu.Lock()
		e.marketData[symbol] = klines[len(klines)-1] // Store latest kline
		e.marketDataMu.Unlock()

		// Save to database
		marketData := &models.MarketData{
			Symbol:    symbol,
			Price:     price,
			Volume:    klines[len(klines)-1].Volume,
			High:      klines[len(klines)-1].High,
			Low:       klines[len(klines)-1].Low,
			Open:      klines[len(klines)-1].Open,
			Close:     klines[len(klines)-1].Close,
			Timestamp: time.Now().Unix(),
		}

		if err := e.repository.SaveMarketData(marketData); err != nil {
			e.logger.Errorf("Failed to save market data: %v", err)
		}
	}

	return nil
}

// tradingLoop is the main trading logic loop
func (e *Engine) tradingLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(e.config.TradingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.processTradingSignals(ctx); err != nil {
				e.logger.Errorf("Error processing trading signals: %v", err)
			}
		}
	}
}

// processTradingSignals processes trading signals for all symbols
func (e *Engine) processTradingSignals(ctx context.Context) error {
	// Check if paper trading mode
	if e.config.EnablePaperTrading {
		e.logger.Debug("Paper trading mode enabled - not executing real trades")
		return nil
	}

	for _, symbol := range e.config.Symbols {
		if err := e.processSymbolSignals(ctx, symbol); err != nil {
			e.logger.Errorf("Error processing signals for %s: %v", symbol, err)
		}
	}

	return nil
}

// processSymbolSignals processes trading signals for a specific symbol
func (e *Engine) processSymbolSignals(ctx context.Context, symbol string) error {
	// Get current market data
	marketData, err := e.getMarketData(symbol)
	if err != nil {
		return fmt.Errorf("failed to get market data for %s: %w", symbol, err)
	}

	// Get current position
	position, err := e.repository.GetPosition(symbol, "LONG")
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to get position for %s: %w", symbol, err)
	}

	// Check for sell signals if we have a position
	if position != nil && position.Status == "OPEN" {
		sellSignal, err := e.strategy.ShouldSell(ctx, symbol, marketData, position)
		if err != nil {
			return fmt.Errorf("failed to get sell signal: %w", err)
		}

		if sellSignal != nil && sellSignal.Action == "SELL" {
			if err := e.executeSellOrder(ctx, symbol, sellSignal, position); err != nil {
				e.logger.Errorf("Failed to execute sell order: %v", err)
			}
		}
	}

	// Check for buy signals if we don't have a position
	if position == nil || position.Status != "OPEN" {
		buySignal, err := e.strategy.ShouldBuy(ctx, symbol, marketData)
		if err != nil {
			return fmt.Errorf("failed to get buy signal: %w", err)
		}

		if buySignal != nil && buySignal.Action == "BUY" {
			// Validate with risk manager
			if !e.riskManager.ValidateOrder(ctx, &OrderInfo{
				Symbol:   symbol,
				Side:     "BUY",
				Quantity: buySignal.Quantity,
				Price:    buySignal.Price,
			}) {
				e.logger.Warnf("Order rejected by risk manager for %s", symbol)
				return nil
			}

			if err := e.executeBuyOrder(ctx, symbol, buySignal); err != nil {
				e.logger.Errorf("Failed to execute buy order: %v", err)
			}
		}
	}

	return nil
}

// getMarketData gets market data for analysis
func (e *Engine) getMarketData(symbol string) (*MarketData, error) {
	e.marketDataMu.RLock()
	kline, exists := e.marketData[symbol]
	e.marketDataMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no market data available for %s", symbol)
	}

	return &MarketData{
		Symbol:    symbol,
		Price:     kline.Close,
		Volume:    kline.Volume,
		Timestamp: time.Unix(kline.CloseTime/1000, 0),
		Klines:    []*exchange.KlineData{kline},
	}, nil
}

// executeBuyOrder executes a buy order
func (e *Engine) executeBuyOrder(ctx context.Context, symbol string, signal *Signal) error {
	e.logger.Infof("Executing BUY order for %s: quantity=%.6f, price=%.6f",
		symbol, signal.Quantity, signal.Price)

	orderRequest := &exchange.OrderRequest{
		Symbol:           symbol,
		Side:             "BUY",
		Type:             "MARKET",
		Quantity:         signal.Quantity,
		PositionSide:     "BOTH",
		NewClientOrderID: fmt.Sprintf("buy_%s_%d", symbol, time.Now().Unix()),
	}

	response, err := e.exchangeClient.PlaceOrder(ctx, orderRequest)
	if err != nil {
		return fmt.Errorf("failed to place buy order: %w", err)
	}

	// Save order to database
	order := &models.Order{
		ExchangeOrderID: fmt.Sprintf("%d", response.OrderID),
		Symbol:          response.Symbol,
		Side:            response.Side,
		Type:            response.Type,
		Status:          response.Status,
		Quantity:        response.OrigQty,
		Price:           response.Price,
		ExecutedQty:     response.ExecutedQty,
		CumulativeQuote: response.CumQuote,
		TimeInForce:     response.TimeInForce,
		ReduceOnly:      response.ReduceOnly,
		ClosePosition:   response.ClosePosition,
		PositionSide:    response.PositionSide,
		Strategy:        e.strategy.Name(),
		Notes:           signal.Reason,
	}

	if err := e.repository.CreateOrder(order); err != nil {
		e.logger.Errorf("Failed to save order to database: %v", err)
	}

	// Create position if order is filled
	if response.Status == "FILLED" {
		position := &models.Position{
			Symbol:       symbol,
			PositionSide: "LONG",
			Size:         response.ExecutedQty,
			EntryPrice:   response.AvgPrice,
			Leverage:     e.config.MaxLeverage,
			Status:       "OPEN",
			OpenTime:     time.Now(),
			Strategy:     e.strategy.Name(),
		}

		if err := e.repository.CreatePosition(position); err != nil {
			e.logger.Errorf("Failed to save position to database: %v", err)
		}
	}

	e.totalTrades++
	e.logger.Infof("Buy order executed successfully: %s", response.ClientOrderID)

	return nil
}

// executeSellOrder executes a sell order
func (e *Engine) executeSellOrder(ctx context.Context, symbol string, signal *Signal, position *models.Position) error {
	e.logger.Infof("Executing SELL order for %s: quantity=%.6f", symbol, position.Size)

	orderRequest := &exchange.OrderRequest{
		Symbol:           symbol,
		Side:             "SELL",
		Type:             "MARKET",
		Quantity:         position.Size,
		PositionSide:     "BOTH",
		NewClientOrderID: fmt.Sprintf("sell_%s_%d", symbol, time.Now().Unix()),
	}

	response, err := e.exchangeClient.PlaceOrder(ctx, orderRequest)
	if err != nil {
		return fmt.Errorf("failed to place sell order: %w", err)
	}

	// Save order to database
	order := &models.Order{
		ExchangeOrderID: fmt.Sprintf("%d", response.OrderID),
		Symbol:          response.Symbol,
		Side:            response.Side,
		Type:            response.Type,
		Status:          response.Status,
		Quantity:        response.OrigQty,
		Price:           response.Price,
		ExecutedQty:     response.ExecutedQty,
		CumulativeQuote: response.CumQuote,
		TimeInForce:     response.TimeInForce,
		ReduceOnly:      response.ReduceOnly,
		ClosePosition:   response.ClosePosition,
		PositionSide:    response.PositionSide,
		Strategy:        e.strategy.Name(),
		Notes:           signal.Reason,
	}

	if err := e.repository.CreateOrder(order); err != nil {
		e.logger.Errorf("Failed to save order to database: %v", err)
	}

	// Close position if order is filled
	if response.Status == "FILLED" {
		pnl := (response.AvgPrice - position.EntryPrice) * position.Size

		if err := e.repository.ClosePosition(position.ID, response.AvgPrice, pnl); err != nil {
			e.logger.Errorf("Failed to close position in database: %v", err)
		}

		// Update statistics
		e.dailyPnL += pnl
		if pnl > 0 {
			e.winningTrades++
		} else {
			e.losingTrades++
		}
	}

	e.logger.Infof("Sell order executed successfully: %s", response.ClientOrderID)

	return nil
}

// monitorRisk monitors risk metrics
func (e *Engine) monitorRisk(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.updateRiskMetrics(ctx); err != nil {
				e.logger.Errorf("Failed to update risk metrics: %v", err)
			}
		}
	}
}

// updateRiskMetrics updates risk metrics
func (e *Engine) updateRiskMetrics(ctx context.Context) error {
	// Calculate win rate
	winRate := 0.0
	if e.totalTrades > 0 {
		winRate = float64(e.winningTrades) / float64(e.totalTrades) * 100
	}

	metric := &models.RiskMetric{
		Date:          time.Now(),
		DailyPnL:      e.dailyPnL,
		TotalTrades:   e.totalTrades,
		WinningTrades: e.winningTrades,
		LosingTrades:  e.losingTrades,
		WinRate:       winRate,
	}

	return e.repository.SaveRiskMetric(metric)
}

// monitorAccount monitors account information
func (e *Engine) monitorAccount(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.updateAccountInfo(ctx); err != nil {
				e.logger.Errorf("Failed to update account info: %v", err)
			}
		}
	}
}

// updateAccountInfo updates account information
func (e *Engine) updateAccountInfo(ctx context.Context) error {
	accountInfo, err := e.exchangeClient.GetAccountInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get account info: %w", err)
	}

	account := &models.Account{
		TotalWalletBalance:      accountInfo.TotalWalletBalance,
		TotalUnrealizedPnL:      accountInfo.TotalUnrealizedPnL,
		TotalMarginBalance:      accountInfo.TotalMarginBalance,
		TotalPositionIM:         accountInfo.TotalPositionIM,
		TotalOpenOrderIM:        accountInfo.TotalOpenOrderIM,
		TotalCrossWalletBalance: accountInfo.TotalCrossWalletBalance,
		AvailableBalance:        accountInfo.AvailableBalance,
		MaxWithdrawAmount:       accountInfo.MaxWithdrawAmount,
		CanTrade:                accountInfo.CanTrade,
		CanWithdraw:             accountInfo.CanWithdraw,
		CanDeposit:              accountInfo.CanDeposit,
		UpdateTime:              accountInfo.UpdateTime,
	}

	return e.repository.UpdateAccount(account)
}

// closeAllPositions closes all open positions
func (e *Engine) closeAllPositions(ctx context.Context) error {
	positions, err := e.repository.GetAllPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	for _, position := range positions {
		orderRequest := &exchange.OrderRequest{
			Symbol:        position.Symbol,
			Side:          "SELL",
			Type:          "MARKET",
			Quantity:      position.Size,
			PositionSide:  "BOTH",
			ClosePosition: true,
		}

		_, err := e.exchangeClient.PlaceOrder(ctx, orderRequest)
		if err != nil {
			e.logger.Errorf("Failed to close position for %s: %v", position.Symbol, err)
			continue
		}

		e.logger.Infof("Closed position for %s", position.Symbol)
	}

	return nil
}

// OrderInfo represents order information for risk validation
type OrderInfo struct {
	Symbol   string
	Side     string
	Quantity float64
	Price    float64
}
