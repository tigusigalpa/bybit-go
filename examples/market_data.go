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

	fmt.Println("=== Bybit Go SDK - Market Data Example ===\n")

	symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}

	for _, symbol := range symbols {
		tickers, err := client.GetTickers(map[string]interface{}{
			"category": "linear",
			"symbol":   symbol,
		})
		if err != nil {
			log.Printf("Error getting ticker for %s: %v", symbol, err)
			continue
		}

		if result, ok := tickers["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok && len(list) > 0 {
				if ticker, ok := list[0].(map[string]interface{}); ok {
					fmt.Printf("📊 %s:\n", symbol)
					fmt.Printf("   Last Price: %v\n", ticker["lastPrice"])
					fmt.Printf("   Mark Price: %v\n", ticker["markPrice"])
					fmt.Printf("   Index Price: %v\n", ticker["indexPrice"])
					fmt.Printf("   24h Volume: %v\n", ticker["volume24h"])
					fmt.Printf("   24h Turnover: %v\n", ticker["turnover24h"])
					fmt.Printf("   24h Change: %v%%\n", ticker["price24hPcnt"])
					fmt.Printf("   Open Interest: %v\n", ticker["openInterest"])
					fmt.Printf("   Funding Rate: %v\n\n", ticker["fundingRate"])
				}
			}
		}
	}

	fmt.Println("✅ Market data example completed successfully!")
}
