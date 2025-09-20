package trading

import (
	"context"
	"contract_playground/internal/models"
)

type AIStrategy struct {
	name string
}

func NewAIStrategy() Strategy {
	return &AIStrategy{
		name: "AIStrategy",
	}
}

func (a *AIStrategy) Name() string {
	return a.name
}

func (a *AIStrategy) Initialize(config map[string]interface{}) error {
	return nil
}

func (a *AIStrategy) ShouldBuy(ctx context.Context, symbol string, data *MarketData) (*Signal, error) {
	// 在这里实现您的 AI 决策逻辑
	// 您可以使用 'data' 参数来获取市场数据
	return &Signal{
		Action: "BUY", // 这是一个示例，您需要替换为真实的决策
	}, nil
}

func (a *AIStrategy) ShouldSell(ctx context.Context, symbol string, data *MarketData, position *models.Position) (*Signal, error) {
	// 在这里实现您的 AI 决策逻辑
	// 您可以使用 'position' 参数来获取当前持仓信息
	return &Signal{
		Action: "SELL", // 这是一个示例，您需要替换为真实的决策
	}, nil
}
