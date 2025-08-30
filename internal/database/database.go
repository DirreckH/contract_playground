package database

import (
	"context"
	"fmt"
	"time"

	"contract_playground/internal/config"
	"contract_playground/internal/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitMySQL initializes MySQL database connection
func InitMySQL(cfg config.MySQLConfig) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	return db, nil
}

// InitRedis initializes Redis connection
func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	models := []interface{}{
		&models.TradingConfig{},
		&models.Order{},
		&models.Position{},
		&models.Trade{},
		&models.Account{},
		&models.Balance{},
		&models.Symbol{},
		&models.MarketData{},
		&models.Strategy{},
		&models.RiskMetric{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	return nil
}

// Repository interface for database operations
type Repository interface {
	// Order operations
	CreateOrder(order *models.Order) error
	UpdateOrder(order *models.Order) error
	GetOrder(id uint) (*models.Order, error)
	GetOrderByExchangeID(exchangeOrderID string) (*models.Order, error)
	GetOpenOrders(symbol string) ([]*models.Order, error)
	GetOrderHistory(symbol string, limit int) ([]*models.Order, error)

	// Position operations
	CreatePosition(position *models.Position) error
	UpdatePosition(position *models.Position) error
	GetPosition(symbol, side string) (*models.Position, error)
	GetAllPositions() ([]*models.Position, error)
	ClosePosition(id uint, closePrice float64, closedPnL float64) error

	// Trade operations
	CreateTrade(trade *models.Trade) error
	GetTradeHistory(symbol string, limit int) ([]*models.Trade, error)
	GetTradesByOrder(orderID uint) ([]*models.Trade, error)

	// Account operations
	UpdateAccount(account *models.Account) error
	GetLatestAccount() (*models.Account, error)
	UpdateBalance(balance *models.Balance) error
	GetBalances(accountID uint) ([]*models.Balance, error)

	// Symbol operations
	UpsertSymbol(symbol *models.Symbol) error
	GetSymbol(symbol string) (*models.Symbol, error)
	GetActiveSymbols() ([]*models.Symbol, error)

	// Market data operations
	SaveMarketData(data *models.MarketData) error
	GetLatestMarketData(symbol string) (*models.MarketData, error)

	// Strategy operations
	CreateStrategy(strategy *models.Strategy) error
	UpdateStrategy(strategy *models.Strategy) error
	GetStrategy(name string) (*models.Strategy, error)
	GetActiveStrategies() ([]*models.Strategy, error)

	// Risk metrics operations
	SaveRiskMetric(metric *models.RiskMetric) error
	GetRiskMetrics(days int) ([]*models.RiskMetric, error)
	GetLatestRiskMetric() (*models.RiskMetric, error)

	// Trading config operations
	CreateTradingConfig(config *models.TradingConfig) error
	UpdateTradingConfig(config *models.TradingConfig) error
	GetTradingConfig(name string) (*models.TradingConfig, error)
	GetActiveTradingConfigs() ([]*models.TradingConfig, error)
}

// MySQLRepository implements Repository interface
type MySQLRepository struct {
	db *gorm.DB
}

// NewMySQLRepository creates a new MySQL repository
func NewMySQLRepository(db *gorm.DB) Repository {
	return &MySQLRepository{db: db}
}

// Order operations
func (r *MySQLRepository) CreateOrder(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *MySQLRepository) UpdateOrder(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *MySQLRepository) GetOrder(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *MySQLRepository) GetOrderByExchangeID(exchangeOrderID string) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("exchange_order_id = ?", exchangeOrderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *MySQLRepository) GetOpenOrders(symbol string) ([]*models.Order, error) {
	var orders []*models.Order
	query := r.db.Where("status IN (?)", []string{"NEW", "PARTIALLY_FILLED"})
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	err := query.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *MySQLRepository) GetOrderHistory(symbol string, limit int) ([]*models.Order, error) {
	var orders []*models.Order
	query := r.db.Model(&models.Order{})
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// Position operations
func (r *MySQLRepository) CreatePosition(position *models.Position) error {
	return r.db.Create(position).Error
}

func (r *MySQLRepository) UpdatePosition(position *models.Position) error {
	return r.db.Save(position).Error
}

func (r *MySQLRepository) GetPosition(symbol, side string) (*models.Position, error) {
	var position models.Position
	err := r.db.Where("symbol = ? AND position_side = ? AND status = ?", symbol, side, "OPEN").First(&position).Error
	if err != nil {
		return nil, err
	}
	return &position, nil
}

func (r *MySQLRepository) GetAllPositions() ([]*models.Position, error) {
	var positions []*models.Position
	err := r.db.Where("status = ?", "OPEN").Find(&positions).Error
	return positions, err
}

func (r *MySQLRepository) ClosePosition(id uint, closePrice float64, closedPnL float64) error {
	now := time.Now()
	return r.db.Model(&models.Position{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "CLOSED",
		"close_time": &now,
		"closed_pnl": closedPnL,
	}).Error
}

// Trade operations
func (r *MySQLRepository) CreateTrade(trade *models.Trade) error {
	return r.db.Create(trade).Error
}

func (r *MySQLRepository) GetTradeHistory(symbol string, limit int) ([]*models.Trade, error) {
	var trades []*models.Trade
	query := r.db.Preload("Order")
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("trade_time DESC").Find(&trades).Error
	return trades, err
}

func (r *MySQLRepository) GetTradesByOrder(orderID uint) ([]*models.Trade, error) {
	var trades []*models.Trade
	err := r.db.Where("order_id = ?", orderID).Find(&trades).Error
	return trades, err
}

// Account operations
func (r *MySQLRepository) UpdateAccount(account *models.Account) error {
	return r.db.Save(account).Error
}

func (r *MySQLRepository) GetLatestAccount() (*models.Account, error) {
	var account models.Account
	err := r.db.Order("updated_at DESC").First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *MySQLRepository) UpdateBalance(balance *models.Balance) error {
	return r.db.Save(balance).Error
}

func (r *MySQLRepository) GetBalances(accountID uint) ([]*models.Balance, error) {
	var balances []*models.Balance
	err := r.db.Where("account_id = ?", accountID).Find(&balances).Error
	return balances, err
}

// Symbol operations
func (r *MySQLRepository) UpsertSymbol(symbol *models.Symbol) error {
	return r.db.Save(symbol).Error
}

func (r *MySQLRepository) GetSymbol(symbol string) (*models.Symbol, error) {
	var s models.Symbol
	err := r.db.Where("symbol = ?", symbol).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *MySQLRepository) GetActiveSymbols() ([]*models.Symbol, error) {
	var symbols []*models.Symbol
	err := r.db.Where("status = ?", "TRADING").Find(&symbols).Error
	return symbols, err
}

// Market data operations
func (r *MySQLRepository) SaveMarketData(data *models.MarketData) error {
	return r.db.Create(data).Error
}

func (r *MySQLRepository) GetLatestMarketData(symbol string) (*models.MarketData, error) {
	var data models.MarketData
	err := r.db.Where("symbol = ?", symbol).Order("timestamp DESC").First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// Strategy operations
func (r *MySQLRepository) CreateStrategy(strategy *models.Strategy) error {
	return r.db.Create(strategy).Error
}

func (r *MySQLRepository) UpdateStrategy(strategy *models.Strategy) error {
	return r.db.Save(strategy).Error
}

func (r *MySQLRepository) GetStrategy(name string) (*models.Strategy, error) {
	var strategy models.Strategy
	err := r.db.Where("name = ?", name).First(&strategy).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (r *MySQLRepository) GetActiveStrategies() ([]*models.Strategy, error) {
	var strategies []*models.Strategy
	err := r.db.Where("is_active = ?", true).Find(&strategies).Error
	return strategies, err
}

// Risk metrics operations
func (r *MySQLRepository) SaveRiskMetric(metric *models.RiskMetric) error {
	return r.db.Create(metric).Error
}

func (r *MySQLRepository) GetRiskMetrics(days int) ([]*models.RiskMetric, error) {
	var metrics []*models.RiskMetric
	since := time.Now().AddDate(0, 0, -days)
	err := r.db.Where("date >= ?", since).Order("date DESC").Find(&metrics).Error
	return metrics, err
}

func (r *MySQLRepository) GetLatestRiskMetric() (*models.RiskMetric, error) {
	var metric models.RiskMetric
	err := r.db.Order("date DESC").First(&metric).Error
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

// Trading config operations
func (r *MySQLRepository) CreateTradingConfig(config *models.TradingConfig) error {
	return r.db.Create(config).Error
}

func (r *MySQLRepository) UpdateTradingConfig(config *models.TradingConfig) error {
	return r.db.Save(config).Error
}

func (r *MySQLRepository) GetTradingConfig(name string) (*models.TradingConfig, error) {
	var config models.TradingConfig
	err := r.db.Where("name = ?", name).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *MySQLRepository) GetActiveTradingConfigs() ([]*models.TradingConfig, error) {
	var configs []*models.TradingConfig
	err := r.db.Where("is_active = ?", true).Find(&configs).Error
	return configs, err
}
