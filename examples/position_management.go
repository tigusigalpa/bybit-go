package main

import (
	"fmt"
	"log"
	"os"

	bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
	apiKey := os.Getenv("BYBIT_API_KEY")
	apiSecret := os.Getenv("BYBIT_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("Please set BYBIT_API_KEY and BYBIT_API_SECRET environment variables")
	}

	client, err := bybit.NewClient(bybit.ClientConfig{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		Testnet:    true,
		Region:     "global",
		RecvWindow: 5000,
		Signature:  "hmac",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Bybit Go SDK - Position Management Example ===\n")

	fmt.Println("📊 Getting Current Positions...")
	positions, err := client.GetPositions(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Printf("Error getting positions: %v\n", err)
	} else {
		if result, ok := positions["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d position(s)\n", len(list))
				for i, pos := range list {
					if p, ok := pos.(map[string]interface{}); ok {
						fmt.Printf("   Position #%d:\n", i+1)
						fmt.Printf("      Symbol: %v\n", p["symbol"])
						fmt.Printf("      Side: %v\n", p["side"])
						fmt.Printf("      Size: %v\n", p["size"])
						fmt.Printf("      Entry Price: %v\n", p["avgPrice"])
						fmt.Printf("      Unrealized PnL: %v\n", p["unrealisedPnl"])
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("⚙️ Setting Auto Add Margin...")
	autoMargin, err := client.SetAutoAddMargin(map[string]interface{}{
		"category":      "linear",
		"symbol":        "BTCUSDT",
		"autoAddMargin": 1,
		"positionIdx":   0,
	})
	if err != nil {
		log.Printf("Error setting auto add margin: %v\n", err)
	} else {
		fmt.Printf("   Auto Add Margin set: %v\n\n", autoMargin)
	}

	fmt.Println("💰 Getting Closed PnL...")
	closedPnL, err := client.GetClosedPnL(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
		"limit":    10,
	})
	if err != nil {
		log.Printf("Error getting closed PnL: %v\n", err)
	} else {
		if result, ok := closedPnL["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d closed position(s)\n", len(list))
				totalPnL := 0.0
				for i, pnl := range list {
					if p, ok := pnl.(map[string]interface{}); ok {
						fmt.Printf("   PnL #%d: %v (Symbol: %v)\n",
							i+1, p["closedPnl"], p["symbol"])
						if closedPnl, ok := p["closedPnl"].(string); ok {
							var val float64
							fmt.Sscanf(closedPnl, "%f", &val)
							totalPnL += val
						}
					}
				}
				fmt.Printf("   Total Closed PnL: %.2f\n", totalPnL)
			}
		}
		fmt.Println()
	}

	fmt.Println("📜 Getting Move Position History...")
	moveHistory, err := client.GetMovePositionHistory(map[string]interface{}{
		"category": "linear",
		"limit":    5,
	})
	if err != nil {
		log.Printf("Error getting move position history: %v\n", err)
	} else {
		if result, ok := moveHistory["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d move position record(s)\n", len(list))
			}
		}
		fmt.Println()
	}

	fmt.Println("⚠️ Getting Risk Limit Info...")
	riskLimit, err := client.GetRiskLimit(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Printf("Error getting risk limit: %v\n", err)
	} else {
		fmt.Printf("   Risk Limit: %v\n\n", riskLimit)
	}

	fmt.Println("✅ Position management example completed successfully!")
}
