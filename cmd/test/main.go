package main

import (
	"fmt"
	"log"

	"contract_playground/internal/config"
)

func main() {
	fmt.Println("Testing configuration loading...")
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	fmt.Printf("Configuration loaded successfully!\n")
	fmt.Printf("Exchange: %s (testnet: %v)\n", cfg.Exchange.Name, cfg.Exchange.Testnet)
	fmt.Printf("Trading symbols: %v\n", cfg.Trading.Symbols)
	fmt.Printf("Max position size: %.2f\n", cfg.Trading.MaxPositionSize)
	fmt.Printf("Strategy: %s\n", cfg.Trading.Strategy.Type)
	fmt.Printf("Paper trading: %v\n", cfg.Trading.EnablePaperTrading)
	
	if cfg.Trading.EnablePaperTrading {
		fmt.Println("\n⚠️  Paper trading mode is ENABLED - no real trades will be executed")
	} else {
		fmt.Println("\n⚠️  Paper trading mode is DISABLED - REAL trades will be executed!")
	}
	
	fmt.Println("\nConfiguration test completed successfully!")
}
