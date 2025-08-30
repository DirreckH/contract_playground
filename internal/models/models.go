package models

import (
	"time"
)

// TradingConfig stores trading configuration parameters
type TradingConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Symbol      string    `gorm:"not null" json:"symbol"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	MaxPosition float64   `gorm:"not null" json:"max_position"`
	StopLoss    float64   `gorm:"not null" json:"stop_loss"`
	TakeProfit  float64   `gorm:"not null" json:"take_profit"`
	Leverage    int       `gorm:"default:1" json:"leverage"`
	RiskPercent float64   `gorm:"not null" json:"risk_percent"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Order represents a trading order
type Order struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ExchangeOrderID string    `gorm:"uniqueIndex;not null" json:"exchange_order_id"`
	Symbol          string    `gorm:"not null;index" json:"symbol"`
	Side            string    `gorm:"not null" json:"side"` // BUY, SELL
	Type            string    `gorm:"not null" json:"type"` // MARKET, LIMIT, STOP_MARKET
	Status          string    `gorm:"not null;index" json:"status"` // NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED
	Quantity        float64   `gorm:"not null" json:"quantity"`
	Price           float64   `json:"price"`
	StopPrice       float64   `json:"stop_price"`
	ExecutedQty     float64   `gorm:"default:0" json:"executed_qty"`
	CumulativeQuote float64   `gorm:"default:0" json:"cumulative_quote"`
	Commission      float64   `gorm:"default:0" json:"commission"`
	CommissionAsset string    `json:"commission_asset"`
	TimeInForce     string    `json:"time_in_force"` // GTC, IOC, FOK
	ReduceOnly      bool      `gorm:"default:false" json:"reduce_only"`
	ClosePosition   bool      `gorm:"default:false" json:"close_position"`
	PositionSide    string    `json:"position_side"` // BOTH, LONG, SHORT
	Strategy        string    `json:"strategy"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Position represents a trading position
type Position struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Symbol         string    `gorm:"not null;index" json:"symbol"`
	PositionSide   string    `gorm:"not null" json:"position_side"` // LONG, SHORT
	Size           float64   `gorm:"not null" json:"size"`
	EntryPrice     float64   `gorm:"not null" json:"entry_price"`
	MarkPrice      float64   `json:"mark_price"`
	UnrealizedPnL  float64   `gorm:"default:0" json:"unrealized_pnl"`
	Percentage     float64   `gorm:"default:0" json:"percentage"`
	Leverage       int       `gorm:"default:1" json:"leverage"`
	Margin         float64   `gorm:"default:0" json:"margin"`
	MaintenanceMargin float64 `gorm:"default:0" json:"maintenance_margin"`
	Status         string    `gorm:"not null;default:'OPEN'" json:"status"` // OPEN, CLOSED
	OpenTime       time.Time `gorm:"not null" json:"open_time"`
	CloseTime      *time.Time `json:"close_time"`
	ClosedPnL      float64   `gorm:"default:0" json:"closed_pnl"`
	Strategy       string    `json:"strategy"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Trade represents an executed trade
type Trade struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ExchangeTradeID string    `gorm:"uniqueIndex;not null" json:"exchange_trade_id"`
	OrderID         uint      `gorm:"not null;index" json:"order_id"`
	Symbol          string    `gorm:"not null;index" json:"symbol"`
	Side            string    `gorm:"not null" json:"side"` // BUY, SELL
	Quantity        float64   `gorm:"not null" json:"quantity"`
	Price           float64   `gorm:"not null" json:"price"`
	QuoteQty        float64   `gorm:"not null" json:"quote_qty"`
	Commission      float64   `gorm:"default:0" json:"commission"`
	CommissionAsset string    `json:"commission_asset"`
	RealizedPnL     float64   `gorm:"default:0" json:"realized_pnl"`
	IsMaker         bool      `gorm:"default:false" json:"is_maker"`
	PositionSide    string    `json:"position_side"`
	Strategy        string    `json:"strategy"`
	TradeTime       time.Time `gorm:"not null" json:"trade_time"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationship
	Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// Account represents account information
type Account struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	TotalWalletBalance float64  `gorm:"default:0" json:"total_wallet_balance"`
	TotalUnrealizedPnL float64  `gorm:"default:0" json:"total_unrealized_pnl"`
	TotalMarginBalance float64  `gorm:"default:0" json:"total_margin_balance"`
	TotalPositionIM    float64  `gorm:"default:0" json:"total_position_im"`
	TotalOpenOrderIM   float64  `gorm:"default:0" json:"total_open_order_im"`
	TotalCrossWalletBalance float64 `gorm:"default:0" json:"total_cross_wallet_balance"`
	AvailableBalance   float64  `gorm:"default:0" json:"available_balance"`
	MaxWithdrawAmount  float64  `gorm:"default:0" json:"max_withdraw_amount"`
	CanTrade           bool     `gorm:"default:true" json:"can_trade"`
	CanWithdraw        bool     `gorm:"default:true" json:"can_withdraw"`
	CanDeposit         bool     `gorm:"default:true" json:"can_deposit"`
	UpdateTime         int64    `json:"update_time"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Balance represents asset balance
type Balance struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	AccountID         uint      `gorm:"not null;index" json:"account_id"`
	Asset             string    `gorm:"not null;index" json:"asset"`
	WalletBalance     float64   `gorm:"default:0" json:"wallet_balance"`
	UnrealizedPnL     float64   `gorm:"default:0" json:"unrealized_pnl"`
	MarginBalance     float64   `gorm:"default:0" json:"margin_balance"`
	MaintMargin       float64   `gorm:"default:0" json:"maint_margin"`
	InitialMargin     float64   `gorm:"default:0" json:"initial_margin"`
	PositionIM        float64   `gorm:"default:0" json:"position_im"`
	OpenOrderIM       float64   `gorm:"default:0" json:"open_order_im"`
	CrossWalletBalance float64  `gorm:"default:0" json:"cross_wallet_balance"`
	CrossUnPnL        float64   `gorm:"default:0" json:"cross_un_pnl"`
	AvailableBalance  float64   `gorm:"default:0" json:"available_balance"`
	MaxWithdrawAmount float64   `gorm:"default:0" json:"max_withdraw_amount"`
	MarginAvailable   bool      `gorm:"default:true" json:"margin_available"`
	UpdateTime        int64     `json:"update_time"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// Relationship
	Account Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

// Symbol represents trading symbol information
type Symbol struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Symbol             string    `gorm:"uniqueIndex;not null" json:"symbol"`
	Pair               string    `gorm:"not null" json:"pair"`
	ContractType       string    `json:"contract_type"`
	DeliveryDate       int64     `json:"delivery_date"`
	OnboardDate        int64     `json:"onboard_date"`
	Status             string    `gorm:"not null" json:"status"`
	MaintMarginPercent float64   `json:"maint_margin_percent"`
	RequiredMarginPercent float64 `json:"required_margin_percent"`
	BaseAsset          string    `gorm:"not null" json:"base_asset"`
	QuoteAsset         string    `gorm:"not null" json:"quote_asset"`
	MarginAsset        string    `json:"margin_asset"`
	PricePrecision     int       `json:"price_precision"`
	QuantityPrecision  int       `json:"quantity_precision"`
	BaseAssetPrecision int       `json:"base_asset_precision"`
	QuotePrecision     int       `json:"quote_precision"`
	UnderlyingType     string    `json:"underlying_type"`
	TriggerProtect     float64   `json:"trigger_protect"`
	LiquidationFee     float64   `json:"liquidation_fee"`
	MarketTakeBound    float64   `json:"market_take_bound"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// MarketData represents market data cache
type MarketData struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Symbol    string    `gorm:"not null;index" json:"symbol"`
	Price     float64   `gorm:"not null" json:"price"`
	Volume    float64   `gorm:"not null" json:"volume"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Open      float64   `json:"open"`
	Close     float64   `json:"close"`
	Change    float64   `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Timestamp int64     `gorm:"not null;index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

// Strategy represents trading strategy information
type Strategy struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Type        string    `gorm:"not null" json:"type"`
	Description string    `json:"description"`
	Parameters  string    `gorm:"type:json" json:"parameters"` // JSON string
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	Performance string    `gorm:"type:json" json:"performance"` // JSON string for performance metrics
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RiskMetric represents risk management metrics
type RiskMetric struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Date            time.Time `gorm:"not null;index" json:"date"`
	TotalPnL        float64   `gorm:"default:0" json:"total_pnl"`
	DailyPnL        float64   `gorm:"default:0" json:"daily_pnl"`
	MaxDrawdown     float64   `gorm:"default:0" json:"max_drawdown"`
	TotalTrades     int       `gorm:"default:0" json:"total_trades"`
	WinningTrades   int       `gorm:"default:0" json:"winning_trades"`
	LosingTrades    int       `gorm:"default:0" json:"losing_trades"`
	WinRate         float64   `gorm:"default:0" json:"win_rate"`
	AvgWin          float64   `gorm:"default:0" json:"avg_win"`
	AvgLoss         float64   `gorm:"default:0" json:"avg_loss"`
	ProfitFactor    float64   `gorm:"default:0" json:"profit_factor"`
	SharpeRatio     float64   `gorm:"default:0" json:"sharpe_ratio"`
	VaR95           float64   `gorm:"default:0" json:"var_95"` // Value at Risk 95%
	MaxLeverage     float64   `gorm:"default:0" json:"max_leverage"`
	TotalExposure   float64   `gorm:"default:0" json:"total_exposure"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName methods for custom table names
func (TradingConfig) TableName() string {
	return "trading_configs"
}

func (Order) TableName() string {
	return "orders"
}

func (Position) TableName() string {
	return "positions"
}

func (Trade) TableName() string {
	return "trades"
}

func (Account) TableName() string {
	return "accounts"
}

func (Balance) TableName() string {
	return "balances"
}

func (Symbol) TableName() string {
	return "symbols"
}

func (MarketData) TableName() string {
	return "market_data"
}

func (Strategy) TableName() string {
	return "strategies"
}

func (RiskMetric) TableName() string {
	return "risk_metrics"
}
