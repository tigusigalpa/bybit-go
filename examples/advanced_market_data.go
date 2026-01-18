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

	fmt.Println("=== Bybit Go SDK - Advanced Market Data Example ===\n")

	fmt.Println("📊 Getting Open Interest...")
	openInterest, err := client.GetOpenInterest(map[string]interface{}{
		"category":     "linear",
		"symbol":       "BTCUSDT",
		"intervalTime": "5min",
	})
	if err != nil {
		log.Printf("Error getting open interest: %v\n", err)
	} else {
		fmt.Printf("   Open Interest: %v\n\n", openInterest)
	}

	fmt.Println("💱 Getting Recent Public Trades...")
	recentTrades, err := client.GetRecentTrades(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
		"limit":    10,
	})
	if err != nil {
		log.Printf("Error getting recent trades: %v\n", err)
	} else {
		if result, ok := recentTrades["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d recent trades\n", len(list))
				for i, trade := range list {
					if t, ok := trade.(map[string]interface{}); ok {
						fmt.Printf("   Trade #%d: Price=%v, Qty=%v, Side=%v\n",
							i+1, t["price"], t["size"], t["side"])
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("💰 Getting Funding Rate History...")
	fundingRate, err := client.GetFundingRateHistory(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
		"limit":    5,
	})
	if err != nil {
		log.Printf("Error getting funding rate: %v\n", err)
	} else {
		if result, ok := fundingRate["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d funding rate records\n", len(list))
				for i, rate := range list {
					if r, ok := rate.(map[string]interface{}); ok {
						fmt.Printf("   Rate #%d: %v at %v\n",
							i+1, r["fundingRate"], r["fundingRateTimestamp"])
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("📈 Getting Historical Volatility...")
	volatility, err := client.GetHistoricalVolatility(map[string]interface{}{
		"category": "option",
	})
	if err != nil {
		log.Printf("Error getting volatility: %v\n", err)
	} else {
		fmt.Printf("   Volatility: %v\n\n", volatility)
	}

	fmt.Println("🏦 Getting Insurance Pool...")
	insurance, err := client.GetInsurance(map[string]interface{}{
		"coin": "USDT",
	})
	if err != nil {
		log.Printf("Error getting insurance: %v\n", err)
	} else {
		fmt.Printf("   Insurance Pool: %v\n\n", insurance)
	}

	fmt.Println("⚠️ Getting Risk Limit...")
	riskLimit, err := client.GetRiskLimit(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Printf("Error getting risk limit: %v\n", err)
	} else {
		fmt.Printf("   Risk Limit: %v\n\n", riskLimit)
	}

	fmt.Println("📊 Getting Kline Data...")
	klines, err := client.GetKline(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
		"interval": "1",
		"limit":    5,
	})
	if err != nil {
		log.Printf("Error getting klines: %v\n", err)
	} else {
		if result, ok := klines["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d klines\n", len(list))
				for i, kline := range list {
					if k, ok := kline.([]interface{}); ok && len(k) >= 5 {
						fmt.Printf("   Kline #%d: Open=%v, High=%v, Low=%v, Close=%v\n",
							i+1, k[1], k[2], k[3], k[4])
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("📖 Getting Orderbook...")
	orderbook, err := client.GetOrderbook(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
		"limit":    5,
	})
	if err != nil {
		log.Printf("Error getting orderbook: %v\n", err)
	} else {
		if result, ok := orderbook["result"].(map[string]interface{}); ok {
			fmt.Printf("   Orderbook for BTCUSDT:\n")
			if bids, ok := result["b"].([]interface{}); ok {
				fmt.Printf("   Bids: %d levels\n", len(bids))
			}
			if asks, ok := result["a"].([]interface{}); ok {
				fmt.Printf("   Asks: %d levels\n", len(asks))
			}
		}
		fmt.Println()
	}

	fmt.Println("✅ Advanced market data example completed successfully!")
}
