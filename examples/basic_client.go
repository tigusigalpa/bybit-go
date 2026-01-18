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

	fmt.Println("=== Bybit Go SDK - Basic Client Example ===\n")

	serverTime, err := client.GetServerTime()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Server Time: %v\n\n", serverTime)

	fmt.Printf("📍 Endpoint: %s\n", client.Endpoint())
	fmt.Printf("🌍 Region: global\n")
	fmt.Printf("🧪 Testnet: true\n\n")

	tickers, err := client.GetTickers(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Fatal(err)
	}

	if result, ok := tickers["result"].(map[string]interface{}); ok {
		if list, ok := result["list"].([]interface{}); ok && len(list) > 0 {
			if ticker, ok := list[0].(map[string]interface{}); ok {
				fmt.Printf("📊 BTC/USDT Ticker:\n")
				fmt.Printf("   Last Price: %v\n", ticker["lastPrice"])
				fmt.Printf("   24h Volume: %v\n", ticker["volume24h"])
				fmt.Printf("   24h Change: %v%%\n", ticker["price24hPcnt"])
			}
		}
	}

	fmt.Println("\n✅ Basic client example completed successfully!")
}
