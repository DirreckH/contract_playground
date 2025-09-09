package exchange

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"contract_playground/internal/config"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"
)

// Client defines the interface for exchange operations
type Client interface {
	// Account information
	GetAccountInfo(ctx context.Context) (*AccountInfo, error)
	GetPositions(ctx context.Context) ([]*PositionInfo, error)
	GetBalance(ctx context.Context) ([]*BalanceInfo, error)

	// Market data
	GetSymbolPrice(ctx context.Context, symbol string) (float64, error)
	GetSymbolInfo(ctx context.Context, symbol string) (*SymbolInfo, error)
	GetKlines(ctx context.Context, symbol string, interval string, limit int) ([]*KlineData, error)

	// Order operations
	PlaceOrder(ctx context.Context, order *OrderRequest) (*OrderResponse, error)
	CancelOrder(ctx context.Context, symbol string, orderID int64) error
	GetOrder(ctx context.Context, symbol string, orderID int64) (*OrderInfo, error)
	GetOpenOrders(ctx context.Context, symbol string) ([]*OrderInfo, error)

	// Real-time data streams
	StartUserDataStream(ctx context.Context, handler UserDataHandler) error
	StartMarketDataStream(ctx context.Context, symbols []string, handler MarketDataHandler) error

	// Exchange specific
	SetLeverage(ctx context.Context, symbol string, leverage int) error
	ChangeMarginType(ctx context.Context, symbol string, marginType string) error
	GetExchangeInfo(ctx context.Context) (*ExchangeInfo, error)
}

// Data structures
type AccountInfo struct {
	TotalWalletBalance      float64 `json:"total_wallet_balance"`
	TotalUnrealizedPnL      float64 `json:"total_unrealized_pnl"`
	TotalMarginBalance      float64 `json:"total_margin_balance"`
	TotalPositionIM         float64 `json:"total_position_im"`
	TotalOpenOrderIM        float64 `json:"total_open_order_im"`
	TotalCrossWalletBalance float64 `json:"total_cross_wallet_balance"`
	AvailableBalance        float64 `json:"available_balance"`
	MaxWithdrawAmount       float64 `json:"max_withdraw_amount"`
	CanTrade                bool    `json:"can_trade"`
	CanWithdraw             bool    `json:"can_withdraw"`
	CanDeposit              bool    `json:"can_deposit"`
	UpdateTime              int64   `json:"update_time"`
}

type PositionInfo struct {
	Symbol            string  `json:"symbol"`
	PositionSide      string  `json:"position_side"`
	PositionAmt       float64 `json:"position_amt"`
	EntryPrice        float64 `json:"entry_price"`
	MarkPrice         float64 `json:"mark_price"`
	UnrealizedPnL     float64 `json:"unrealized_pnl"`
	Percentage        float64 `json:"percentage"`
	Leverage          int     `json:"leverage"`
	Margin            float64 `json:"margin"`
	MaintenanceMargin float64 `json:"maintenance_margin"`
	UpdateTime        int64   `json:"update_time"`
}

type BalanceInfo struct {
	Asset              string  `json:"asset"`
	WalletBalance      float64 `json:"wallet_balance"`
	UnrealizedPnL      float64 `json:"unrealized_pnl"`
	MarginBalance      float64 `json:"margin_balance"`
	MaintMargin        float64 `json:"maint_margin"`
	InitialMargin      float64 `json:"initial_margin"`
	PositionIM         float64 `json:"position_im"`
	OpenOrderIM        float64 `json:"open_order_im"`
	CrossWalletBalance float64 `json:"cross_wallet_balance"`
	CrossUnPnL         float64 `json:"cross_un_pnl"`
	AvailableBalance   float64 `json:"available_balance"`
	MaxWithdrawAmount  float64 `json:"max_withdraw_amount"`
	MarginAvailable    bool    `json:"margin_available"`
	UpdateTime         int64   `json:"update_time"`
}

type SymbolInfo struct {
	Symbol                string  `json:"symbol"`
	Status                string  `json:"status"`
	BaseAsset             string  `json:"base_asset"`
	QuoteAsset            string  `json:"quote_asset"`
	PricePrecision        int     `json:"price_precision"`
	QuantityPrecision     int     `json:"quantity_precision"`
	MinQty                float64 `json:"min_qty"`
	MaxQty                float64 `json:"max_qty"`
	StepSize              float64 `json:"step_size"`
	MinPrice              float64 `json:"min_price"`
	MaxPrice              float64 `json:"max_price"`
	TickSize              float64 `json:"tick_size"`
	MinNotional           float64 `json:"min_notional"`
	MaintMarginPercent    float64 `json:"maint_margin_percent"`
	RequiredMarginPercent float64 `json:"required_margin_percent"`
}

type KlineData struct {
	OpenTime                 int64   `json:"open_time"`
	Open                     float64 `json:"open"`
	High                     float64 `json:"high"`
	Low                      float64 `json:"low"`
	Close                    float64 `json:"close"`
	Volume                   float64 `json:"volume"`
	CloseTime                int64   `json:"close_time"`
	QuoteAssetVolume         float64 `json:"quote_asset_volume"`
	TradeCount               int64   `json:"trade_count"`
	TakerBuyBaseAssetVolume  float64 `json:"taker_buy_base_asset_volume"`
	TakerBuyQuoteAssetVolume float64 `json:"taker_buy_quote_asset_volume"`
}

type OrderRequest struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	Type             string  `json:"type"`
	Quantity         float64 `json:"quantity"`
	Price            float64 `json:"price,omitempty"`
	StopPrice        float64 `json:"stop_price,omitempty"`
	TimeInForce      string  `json:"time_in_force,omitempty"`
	ReduceOnly       bool    `json:"reduce_only,omitempty"`
	ClosePosition    bool    `json:"close_position,omitempty"`
	PositionSide     string  `json:"position_side,omitempty"`
	WorkingType      string  `json:"working_type,omitempty"`
	PriceProtect     bool    `json:"price_protect,omitempty"`
	NewClientOrderID string  `json:"new_client_order_id,omitempty"`
}

type OrderResponse struct {
	OrderID       int64   `json:"order_id"`
	Symbol        string  `json:"symbol"`
	Status        string  `json:"status"`
	ClientOrderID string  `json:"client_order_id"`
	Price         float64 `json:"price"`
	AvgPrice      float64 `json:"avg_price"`
	OrigQty       float64 `json:"orig_qty"`
	ExecutedQty   float64 `json:"executed_qty"`
	CumQuote      float64 `json:"cum_quote"`
	TimeInForce   string  `json:"time_in_force"`
	Type          string  `json:"type"`
	ReduceOnly    bool    `json:"reduce_only"`
	ClosePosition bool    `json:"close_position"`
	Side          string  `json:"side"`
	PositionSide  string  `json:"position_side"`
	StopPrice     float64 `json:"stop_price"`
	WorkingType   string  `json:"working_type"`
	PriceProtect  bool    `json:"price_protect"`
	UpdateTime    int64   `json:"update_time"`
}

type OrderInfo struct {
	OrderID       int64   `json:"order_id"`
	Symbol        string  `json:"symbol"`
	Status        string  `json:"status"`
	ClientOrderID string  `json:"client_order_id"`
	Price         float64 `json:"price"`
	AvgPrice      float64 `json:"avg_price"`
	OrigQty       float64 `json:"orig_qty"`
	ExecutedQty   float64 `json:"executed_qty"`
	CumQuote      float64 `json:"cum_quote"`
	TimeInForce   string  `json:"time_in_force"`
	Type          string  `json:"type"`
	ReduceOnly    bool    `json:"reduce_only"`
	ClosePosition bool    `json:"close_position"`
	Side          string  `json:"side"`
	PositionSide  string  `json:"position_side"`
	StopPrice     float64 `json:"stop_price"`
	WorkingType   string  `json:"working_type"`
	PriceProtect  bool    `json:"price_protect"`
	Time          int64   `json:"time"`
	UpdateTime    int64   `json:"update_time"`
}

type ExchangeInfo struct {
	Timezone   string        `json:"timezone"`
	ServerTime int64         `json:"server_time"`
	Symbols    []*SymbolInfo `json:"symbols"`
}

// Handler interfaces
type UserDataHandler interface {
	OnAccountUpdate(account *AccountInfo)
	OnOrderUpdate(order *OrderInfo)
	OnPositionUpdate(position *PositionInfo)
	OnTradeUpdate(trade *TradeInfo)
	OnError(err error)
}

type MarketDataHandler interface {
	OnPriceUpdate(symbol string, price float64)
	OnKlineUpdate(symbol string, kline *KlineData)
	OnError(err error)
}

type TradeInfo struct {
	Symbol          string  `json:"symbol"`
	ID              int64   `json:"id"`
	OrderID         int64   `json:"order_id"`
	Side            string  `json:"side"`
	Quantity        float64 `json:"quantity"`
	Price           float64 `json:"price"`
	Commission      float64 `json:"commission"`
	CommissionAsset string  `json:"commission_asset"`
	Time            int64   `json:"time"`
	IsMaker         bool    `json:"is_maker"`
	RealizedPnL     float64 `json:"realized_pnl"`
}

// BinanceClient implements Client interface for Binance futures
type BinanceClient struct {
	client *futures.Client
	config config.ExchangeConfig
	logger *logrus.Logger
}

// NewBinanceClient creates a new Binance futures client
func NewBinanceClient(cfg config.ExchangeConfig, logger *logrus.Logger) (Client, error) {
	if cfg.Testnet {
		futures.UseTestnet = true
	}

	client := futures.NewClient(cfg.APIKey, cfg.SecretKey)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Binance: %w", err)
	}

	logger.Info("Successfully connected to Binance futures API")

	return &BinanceClient{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// GetAccountInfo retrieves account information
func (b *BinanceClient) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
	account, err := b.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	totalWalletBalance, _ := strconv.ParseFloat(account.TotalWalletBalance, 64)
	totalUnrealizedPnL, _ := strconv.ParseFloat(account.TotalUnrealizedProfit, 64)
	totalMarginBalance, _ := strconv.ParseFloat(account.TotalMarginBalance, 64)
	totalPositionIM, _ := strconv.ParseFloat(account.TotalPositionInitialMargin, 64)
	totalOpenOrderIM, _ := strconv.ParseFloat(account.TotalOpenOrderInitialMargin, 64)
	totalCrossWalletBalance, _ := strconv.ParseFloat(account.TotalCrossWalletBalance, 64)
	availableBalance, _ := strconv.ParseFloat(account.AvailableBalance, 64)
	maxWithdrawAmount, _ := strconv.ParseFloat(account.MaxWithdrawAmount, 64)

	return &AccountInfo{
		TotalWalletBalance:      totalWalletBalance,
		TotalUnrealizedPnL:      totalUnrealizedPnL,
		TotalMarginBalance:      totalMarginBalance,
		TotalPositionIM:         totalPositionIM,
		TotalOpenOrderIM:        totalOpenOrderIM,
		TotalCrossWalletBalance: totalCrossWalletBalance,
		AvailableBalance:        availableBalance,
		MaxWithdrawAmount:       maxWithdrawAmount,
		CanTrade:                account.CanTrade,
		CanWithdraw:             account.CanWithdraw,
		CanDeposit:              account.CanDeposit,
		UpdateTime:              account.UpdateTime,
	}, nil
}

// GetPositions retrieves current positions
func (b *BinanceClient) GetPositions(ctx context.Context) ([]*PositionInfo, error) {
	positions, err := b.client.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var result []*PositionInfo
	for _, pos := range positions {
		positionAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
		markPrice, _ := strconv.ParseFloat(pos.MarkPrice, 64)
		unrealizedPnL, _ := strconv.ParseFloat(pos.UnRealizedProfit, 64)
		leverage, _ := strconv.Atoi(pos.Leverage)

		// Only include positions with non-zero amounts
		if positionAmt != 0 {
			result = append(result, &PositionInfo{
				Symbol:        pos.Symbol,
				PositionSide:  pos.PositionSide,
				PositionAmt:   positionAmt,
				EntryPrice:    entryPrice,
				MarkPrice:     markPrice,
				UnrealizedPnL: unrealizedPnL,
				Percentage:    0, // Not available in PositionRisk
				Leverage:      leverage,
				UpdateTime:    0, // Not available in PositionRisk
			})
		}
	}

	return result, nil
}

// GetBalance retrieves account balance
func (b *BinanceClient) GetBalance(ctx context.Context) ([]*BalanceInfo, error) {
	account, err := b.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	var result []*BalanceInfo
	for _, asset := range account.Assets {
		walletBalance, _ := strconv.ParseFloat(asset.WalletBalance, 64)
		unrealizedPnL, _ := strconv.ParseFloat(asset.UnrealizedProfit, 64)
		marginBalance, _ := strconv.ParseFloat(asset.MarginBalance, 64)
		maintMargin, _ := strconv.ParseFloat(asset.MaintMargin, 64)
		initialMargin, _ := strconv.ParseFloat(asset.InitialMargin, 64)
		positionIM, _ := strconv.ParseFloat(asset.PositionInitialMargin, 64)
		openOrderIM, _ := strconv.ParseFloat(asset.OpenOrderInitialMargin, 64)
		crossWalletBalance, _ := strconv.ParseFloat(asset.CrossWalletBalance, 64)
		crossUnPnL := 0.0 // CrossUnPnL field not available
		availableBalance, _ := strconv.ParseFloat(asset.AvailableBalance, 64)
		maxWithdrawAmount, _ := strconv.ParseFloat(asset.MaxWithdrawAmount, 64)

		result = append(result, &BalanceInfo{
			Asset:              asset.Asset,
			WalletBalance:      walletBalance,
			UnrealizedPnL:      unrealizedPnL,
			MarginBalance:      marginBalance,
			MaintMargin:        maintMargin,
			InitialMargin:      initialMargin,
			PositionIM:         positionIM,
			OpenOrderIM:        openOrderIM,
			CrossWalletBalance: crossWalletBalance,
			CrossUnPnL:         crossUnPnL,
			AvailableBalance:   availableBalance,
			MaxWithdrawAmount:  maxWithdrawAmount,
			MarginAvailable:    asset.MarginAvailable,
			UpdateTime:         asset.UpdateTime,
		})
	}

	return result, nil
}

// GetSymbolPrice retrieves current price for a symbol
func (b *BinanceClient) GetSymbolPrice(ctx context.Context, symbol string) (float64, error) {
	price, err := b.client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get symbol price: %w", err)
	}

	if len(price) == 0 {
		return 0, fmt.Errorf("no price data for symbol %s", symbol)
	}

	priceFloat, err := strconv.ParseFloat(price[0].Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return priceFloat, nil
}

// GetSymbolInfo retrieves symbol information
func (b *BinanceClient) GetSymbolInfo(ctx context.Context, symbol string) (*SymbolInfo, error) {
	exchangeInfo, err := b.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %w", err)
	}

	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			return &SymbolInfo{
				Symbol:                s.Symbol,
				Status:                string(s.Status),
				BaseAsset:             s.BaseAsset,
				QuoteAsset:            s.QuoteAsset,
				PricePrecision:        s.PricePrecision,
				QuantityPrecision:     s.QuantityPrecision,
				MaintMarginPercent:    parseFloat(s.MaintMarginPercent),
				RequiredMarginPercent: parseFloat(s.RequiredMarginPercent),
			}, nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found", symbol)
}

// GetKlines retrieves kline/candlestick data
func (b *BinanceClient) GetKlines(ctx context.Context, symbol string, interval string, limit int) ([]*KlineData, error) {
	klines, err := b.client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %w", err)
	}

	var result []*KlineData
	for _, k := range klines {
		result = append(result, &KlineData{
			OpenTime:                 k.OpenTime,
			Open:                     parseFloat(k.Open),
			High:                     parseFloat(k.High),
			Low:                      parseFloat(k.Low),
			Close:                    parseFloat(k.Close),
			Volume:                   parseFloat(k.Volume),
			CloseTime:                k.CloseTime,
			QuoteAssetVolume:         parseFloat(k.QuoteAssetVolume),
			TradeCount:               0, // TradeCount field not available
			TakerBuyBaseAssetVolume:  parseFloat(k.TakerBuyBaseAssetVolume),
			TakerBuyQuoteAssetVolume: parseFloat(k.TakerBuyQuoteAssetVolume),
		})
	}

	return result, nil
}

// PlaceOrder places a new order
func (b *BinanceClient) PlaceOrder(ctx context.Context, order *OrderRequest) (*OrderResponse, error) {
	service := b.client.NewCreateOrderService().
		Symbol(order.Symbol).
		Side(futures.SideType(order.Side)).
		Type(futures.OrderType(order.Type)).
		Quantity(fmt.Sprintf("%.8f", order.Quantity))

	if order.Price > 0 {
		service = service.Price(fmt.Sprintf("%.8f", order.Price))
	}

	if order.StopPrice > 0 {
		service = service.StopPrice(fmt.Sprintf("%.8f", order.StopPrice))
	}

	if order.TimeInForce != "" {
		service = service.TimeInForce(futures.TimeInForceType(order.TimeInForce))
	}

	if order.ReduceOnly {
		service = service.ReduceOnly(order.ReduceOnly)
	}

	if order.ClosePosition {
		service = service.ClosePosition(order.ClosePosition)
	}

	if order.PositionSide != "" {
		service = service.PositionSide(futures.PositionSideType(order.PositionSide))
	}

	if order.NewClientOrderID != "" {
		service = service.NewClientOrderID(order.NewClientOrderID)
	}

	response, err := service.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	return &OrderResponse{
		OrderID:       response.OrderID,
		Symbol:        response.Symbol,
		Status:        string(response.Status),
		ClientOrderID: response.ClientOrderID,
		Price:         parseFloat(response.Price),
		AvgPrice:      parseFloat(response.AvgPrice),
		OrigQty:       parseFloat(response.OrigQuantity),
		ExecutedQty:   parseFloat(response.ExecutedQuantity),
		CumQuote:      parseFloat(response.CumQuote),
		TimeInForce:   string(response.TimeInForce),
		Type:          string(response.Type),
		ReduceOnly:    response.ReduceOnly,
		ClosePosition: response.ClosePosition,
		Side:          string(response.Side),
		PositionSide:  string(response.PositionSide),
		StopPrice:     parseFloat(response.StopPrice),
		WorkingType:   string(response.WorkingType),
		PriceProtect:  response.PriceProtect,
		UpdateTime:    response.UpdateTime,
	}, nil
}

// CancelOrder cancels an order
func (b *BinanceClient) CancelOrder(ctx context.Context, symbol string, orderID int64) error {
	_, err := b.client.NewCancelOrderService().
		Symbol(symbol).
		OrderID(orderID).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

// GetOrder retrieves order information
func (b *BinanceClient) GetOrder(ctx context.Context, symbol string, orderID int64) (*OrderInfo, error) {
	order, err := b.client.NewGetOrderService().
		Symbol(symbol).
		OrderID(orderID).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &OrderInfo{
		OrderID:       order.OrderID,
		Symbol:        order.Symbol,
		Status:        string(order.Status),
		ClientOrderID: order.ClientOrderID,
		Price:         parseFloat(order.Price),
		AvgPrice:      parseFloat(order.AvgPrice),
		OrigQty:       parseFloat(order.OrigQuantity),
		ExecutedQty:   parseFloat(order.ExecutedQuantity),
		CumQuote:      parseFloat(order.CumQuote),
		TimeInForce:   string(order.TimeInForce),
		Type:          string(order.Type),
		ReduceOnly:    order.ReduceOnly,
		ClosePosition: order.ClosePosition,
		Side:          string(order.Side),
		PositionSide:  string(order.PositionSide),
		StopPrice:     parseFloat(order.StopPrice),
		WorkingType:   string(order.WorkingType),
		PriceProtect:  order.PriceProtect,
		Time:          order.Time,
		UpdateTime:    order.UpdateTime,
	}, nil
}

// GetOpenOrders retrieves open orders
func (b *BinanceClient) GetOpenOrders(ctx context.Context, symbol string) ([]*OrderInfo, error) {
	service := b.client.NewListOpenOrdersService()
	if symbol != "" {
		service = service.Symbol(symbol)
	}

	orders, err := service.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	var result []*OrderInfo
	for _, order := range orders {
		result = append(result, &OrderInfo{
			OrderID:       order.OrderID,
			Symbol:        order.Symbol,
			Status:        string(order.Status),
			ClientOrderID: order.ClientOrderID,
			Price:         parseFloat(order.Price),
			AvgPrice:      parseFloat(order.AvgPrice),
			OrigQty:       parseFloat(order.OrigQuantity),
			ExecutedQty:   parseFloat(order.ExecutedQuantity),
			CumQuote:      parseFloat(order.CumQuote),
			TimeInForce:   string(order.TimeInForce),
			Type:          string(order.Type),
			ReduceOnly:    order.ReduceOnly,
			ClosePosition: order.ClosePosition,
			Side:          string(order.Side),
			PositionSide:  string(order.PositionSide),
			StopPrice:     parseFloat(order.StopPrice),
			WorkingType:   string(order.WorkingType),
			PriceProtect:  order.PriceProtect,
			Time:          order.Time,
			UpdateTime:    order.UpdateTime,
		})
	}

	return result, nil
}

// SetLeverage sets leverage for a symbol
func (b *BinanceClient) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	_, err := b.client.NewChangeLeverageService().
		Symbol(symbol).
		Leverage(leverage).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to set leverage: %w", err)
	}

	b.logger.Infof("Set leverage for %s to %d", symbol, leverage)
	return nil
}

// ChangeMarginType changes margin type for a symbol
func (b *BinanceClient) ChangeMarginType(ctx context.Context, symbol string, marginType string) error {
	err := b.client.NewChangeMarginTypeService().
		Symbol(symbol).
		MarginType(futures.MarginType(marginType)).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to change margin type: %w", err)
	}

	b.logger.Infof("Changed margin type for %s to %s", symbol, marginType)
	return nil
}

// GetExchangeInfo retrieves exchange information
func (b *BinanceClient) GetExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	info, err := b.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %w", err)
	}

	var symbols []*SymbolInfo
	for _, s := range info.Symbols {
		symbols = append(symbols, &SymbolInfo{
			Symbol:                s.Symbol,
			Status:                string(s.Status),
			BaseAsset:             s.BaseAsset,
			QuoteAsset:            s.QuoteAsset,
			PricePrecision:        s.PricePrecision,
			QuantityPrecision:     s.QuantityPrecision,
			MaintMarginPercent:    parseFloat(s.MaintMarginPercent),
			RequiredMarginPercent: parseFloat(s.RequiredMarginPercent),
		})
	}

	return &ExchangeInfo{
		Timezone:   info.Timezone,
		ServerTime: info.ServerTime,
		Symbols:    symbols,
	}, nil
}

// StartUserDataStream starts user data stream (placeholder implementation)
func (b *BinanceClient) StartUserDataStream(ctx context.Context, handler UserDataHandler) error {
	// This would implement WebSocket user data stream
	// For now, it's a placeholder
	b.logger.Info("User data stream would be started here")
	return nil
}

// StartMarketDataStream starts market data stream (placeholder implementation)
func (b *BinanceClient) StartMarketDataStream(ctx context.Context, symbols []string, handler MarketDataHandler) error {
	// This would implement WebSocket market data stream
	// For now, it's a placeholder
	b.logger.Infof("Market data stream would be started for symbols: %v", symbols)
	return nil
}

// Helper function to parse float strings
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
