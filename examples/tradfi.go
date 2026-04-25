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

	fmt.Println("=== Bybit Go SDK - TradFi Example ===\n")

	// --- Metals ---
	fmt.Println("🥇 Metals Tickers (XAUUSD, XAGUSD):")
	metalsTicker, err := client.GetTradFiTicker("XAUUSD")
	if err != nil {
		log.Printf("Error getting metals ticker: %v", err)
	} else {
		printTicker(metalsTicker, "XAUUSD")
	}

	// --- Forex ---
	fmt.Println("\n💱 Forex Ticker (EURUSD):")
	forexTicker, err := client.GetTradFiTicker("EURUSD")
	if err != nil {
		log.Printf("Error getting forex ticker: %v", err)
	} else {
		printTicker(forexTicker, "EURUSD")
	}

	// --- Index ---
	fmt.Println("\n📈 Index Ticker (US500USD - S&P 500):")
	indexTicker, err := client.GetTradFiTicker("US500USD")
	if err != nil {
		log.Printf("Error getting index ticker: %v", err)
	} else {
		printTicker(indexTicker, "US500USD")
	}

	// --- Kline ---
	fmt.Println("\n📊 XAUUSD Kline (1h, last 5 candles):")
	kline, err := client.GetTradFiKline("XAUUSD", "60", 5)
	if err != nil {
		log.Printf("Error getting kline: %v", err)
	} else {
		if result, ok := kline["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				for _, candle := range list {
					if c, ok := candle.([]interface{}); ok && len(c) >= 5 {
						fmt.Printf("   time=%v open=%v high=%v low=%v close=%v\n",
							c[0], c[1], c[2], c[3], c[4])
					}
				}
			}
		}
	}

	// --- Orderbook ---
	fmt.Println("\n📒 EURUSD Orderbook (depth 5):")
	ob, err := client.GetTradFiOrderbook("EURUSD", 5)
	if err != nil {
		log.Printf("Error getting orderbook: %v", err)
	} else {
		if result, ok := ob["result"].(map[string]interface{}); ok {
			if bids, ok := result["b"].([]interface{}); ok && len(bids) > 0 {
				fmt.Printf("   Best bid: %v @ %v\n", bids[0].([]interface{})[1], bids[0].([]interface{})[0])
			}
			if asks, ok := result["a"].([]interface{}); ok && len(asks) > 0 {
				fmt.Printf("   Best ask: %v @ %v\n", asks[0].([]interface{})[1], asks[0].([]interface{})[0])
			}
		}
	}

	// --- Positions ---
	if apiKey != "" {
		fmt.Println("\n📂 Open TradFi positions:")
		positions, err := client.GetTradFiPositions("")
		if err != nil {
			log.Printf("Error getting positions: %v", err)
		} else {
			if result, ok := positions["result"].(map[string]interface{}); ok {
				if list, ok := result["list"].([]interface{}); ok {
					tradfiCount := 0
					for _, item := range list {
						if pos, ok := item.(map[string]interface{}); ok {
							symbol, _ := pos["symbol"].(string)
							if bybit.IsTradFiSymbol(symbol) {
								tradfiCount++
								fmt.Printf("   %s side=%v size=%v unrealisedPnl=%v\n",
									symbol, pos["side"], pos["size"], pos["unrealisedPnl"])
							}
						}
					}
					if tradfiCount == 0 {
						fmt.Println("   No open TradFi positions")
					}
				}
			}
		}

		// --- Place a demo limit order on XAUUSD ---
		fmt.Println("\n📝 Placing limit buy order on XAUUSD (demo, far from market):")
		order, err := client.PlaceTradFiOrder(bybit.TradFiOrderParams{
			Symbol:      "XAUUSD",
			Side:        "Buy",
			OrderType:   "Limit",
			Qty:         "0.01",
			Price:       "1000",
			TimeInForce: "GTC",
		})
		if err != nil {
			log.Printf("Error placing order: %v", err)
		} else {
			if result, ok := order["result"].(map[string]interface{}); ok {
				fmt.Printf("   ✅ Order placed! ID: %v\n", result["orderId"])
			} else {
				fmt.Printf("   Response: retCode=%v retMsg=%v\n", order["retCode"], order["retMsg"])
			}
		}
	}

	// --- IsTradFiSymbol helper ---
	fmt.Println("\n🔍 IsTradFiSymbol detection:")
	testSymbols := []string{"XAUUSD", "EURUSD", "US500USD", "BTCUSDT", "ETHUSDT", "XAGUSD", "GBPUSD"}
	for _, sym := range testSymbols {
		fmt.Printf("   %-12s → TradFi: %v\n", sym, bybit.IsTradFiSymbol(sym))
	}

	fmt.Println("\n✅ TradFi example completed!")
}

func printTicker(data map[string]interface{}, symbol string) {
	if result, ok := data["result"].(map[string]interface{}); ok {
		if list, ok := result["list"].([]interface{}); ok && len(list) > 0 {
			if ticker, ok := list[0].(map[string]interface{}); ok {
				fmt.Printf("   %s: lastPrice=%v markPrice=%v 24hChange=%v%%\n",
					symbol,
					ticker["lastPrice"],
					ticker["markPrice"],
					ticker["price24hPcnt"],
				)
			}
		}
	}
}
