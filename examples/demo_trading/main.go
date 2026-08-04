package main

import (
	"fmt"
	"log"
	"os"

	bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
	apiKey := os.Getenv("BYBIT_DEMO_API_KEY")
	apiSecret := os.Getenv("BYBIT_DEMO_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("Please set BYBIT_DEMO_API_KEY and BYBIT_DEMO_API_SECRET environment variables")
	}

	fmt.Println("=== Bybit Go SDK - Demo Trading Example ===\n")

	demoClient, err := bybit.NewDemoClient(bybit.ClientConfig{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		RecvWindow: 5000,
		Signature:  "hmac",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("🧪 Connected to Demo Trading Environment\n")
	fmt.Printf("   Endpoint: %s\n\n", demoClient.BaseURI())

	fmt.Println("💰 Getting Demo Trading Balance...")
	balance, err := demoClient.GetWalletBalance(nil)
	if err != nil {
		log.Printf("Error getting demo balance: %v\n", err)
	} else {
		if result, ok := balance["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok && len(list) > 0 {
				if account, ok := list[0].(map[string]interface{}); ok {
					fmt.Printf("   Total Balance: %v\n", account["totalWalletBalance"])
					fmt.Printf("   Available Balance: %v\n", account["totalAvailableBalance"])

					if coins, ok := account["coin"].([]interface{}); ok {
						fmt.Printf("\n   Demo Balances:\n")
						for _, coinData := range coins {
							if coin, ok := coinData.(map[string]interface{}); ok {
								fmt.Printf("      %v: %v\n", coin["coin"], coin["walletBalance"])
							}
						}
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("📊 Getting Current Market Price...")
	tickers, err := demoClient.GetTickers(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Printf("Error getting tickers: %v\n", err)
	} else {
		if result, ok := tickers["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok && len(list) > 0 {
				if ticker, ok := list[0].(map[string]interface{}); ok {
					fmt.Printf("   BTC/USDT Price: $%v\n\n", ticker["lastPrice"])
				}
			}
		}
	}

	fmt.Println("📝 Placing Demo Order...")
	order, err := demoClient.CreateOrder(map[string]interface{}{
		"category":    "linear",
		"symbol":      "BTCUSDT",
		"side":        "Buy",
		"orderType":   "Limit",
		"qty":         "0.01",
		"price":       "30000",
		"timeInForce": "GTC",
	})
	if err != nil {
		log.Printf("Error placing demo order: %v\n", err)
	} else {
		if result, ok := order["result"].(map[string]interface{}); ok {
			fmt.Printf("   ✅ Demo Order Placed!\n")
			fmt.Printf("   Order ID: %v\n", result["orderId"])
			fmt.Printf("   Order Link ID: %v\n\n", result["orderLinkId"])
		}
	}

	fmt.Println("📊 Getting Demo Positions...")
	positions, err := demoClient.GetPositions(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Printf("Error getting demo positions: %v\n", err)
	} else {
		if result, ok := positions["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d demo position(s)\n", len(list))
				for i, pos := range list {
					if p, ok := pos.(map[string]interface{}); ok {
						fmt.Printf("   Position #%d:\n", i+1)
						fmt.Printf("      Symbol: %v\n", p["symbol"])
						fmt.Printf("      Side: %v\n", p["side"])
						fmt.Printf("      Size: %v\n", p["size"])
						fmt.Printf("      Unrealized PnL: %v\n", p["unrealisedPnl"])
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("💵 Applying for Demo Funds (if needed)...")
	fundResult, err := demoClient.ApplyForDemoFundsSimple("USDT", "10000")
	if err != nil {
		log.Printf("Note: %v\n", err)
	} else {
		fmt.Printf("   Demo Funds Applied: %v\n\n", fundResult)
	}

	fmt.Println("✅ Demo trading example completed successfully!")
	fmt.Println("\n⚠️  Remember: This is demo trading with virtual funds.")
	fmt.Println("   No real money is involved. Perfect for testing strategies!")
}
