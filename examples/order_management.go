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

	fmt.Println("=== Bybit Go SDK - Order Management Example ===\n")

	fmt.Println("📊 Getting current BTC price...")
	tickers, err := client.GetTickers(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Fatal(err)
	}

	var currentPrice string
	if result, ok := tickers["result"].(map[string]interface{}); ok {
		if list, ok := result["list"].([]interface{}); ok && len(list) > 0 {
			if ticker, ok := list[0].(map[string]interface{}); ok {
				currentPrice = ticker["lastPrice"].(string)
				fmt.Printf("   Current BTC Price: $%s\n\n", currentPrice)
			}
		}
	}

	fmt.Println("📝 Placing a limit buy order...")
	order, err := client.CreateOrder(map[string]interface{}{
		"category":    "linear",
		"symbol":      "BTCUSDT",
		"side":        "Buy",
		"orderType":   "Limit",
		"qty":         "0.001",
		"price":       "30000",
		"timeInForce": "GTC",
	})
	if err != nil {
		log.Printf("Error placing order: %v\n", err)
	} else {
		if result, ok := order["result"].(map[string]interface{}); ok {
			fmt.Printf("   ✅ Order placed successfully!\n")
			fmt.Printf("   Order ID: %v\n", result["orderId"])
			fmt.Printf("   Order Link ID: %v\n\n", result["orderLinkId"])
		}
	}

	fmt.Println("📋 Getting open orders...")
	openOrders, err := client.GetOpenOrders(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Printf("Error getting open orders: %v\n", err)
	} else {
		if result, ok := openOrders["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d open order(s)\n\n", len(list))
				for i, orderData := range list {
					if order, ok := orderData.(map[string]interface{}); ok {
						fmt.Printf("   Order #%d:\n", i+1)
						fmt.Printf("      ID: %v\n", order["orderId"])
						fmt.Printf("      Symbol: %v\n", order["symbol"])
						fmt.Printf("      Side: %v\n", order["side"])
						fmt.Printf("      Price: %v\n", order["price"])
						fmt.Printf("      Qty: %v\n", order["qty"])
						fmt.Printf("      Status: %v\n\n", order["orderStatus"])
					}
				}
			}
		}
	}

	fmt.Println("📜 Getting order history...")
	history, err := client.GetHistoryOrders(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
		"limit":    10,
	})
	if err != nil {
		log.Printf("Error getting order history: %v\n", err)
	} else {
		if result, ok := history["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d historical order(s)\n", len(list))
			}
		}
	}

	fmt.Println("\n✅ Order management example completed successfully!")
}
