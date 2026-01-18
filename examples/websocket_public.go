package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
	fmt.Println("=== Bybit Go SDK - Public WebSocket Example ===\n")

	ws := bybit.NewWebSocket(bybit.WebSocketConfig{
		Testnet:   false,
		Region:    "global",
		IsPrivate: false,
	})

	fmt.Println("🔌 Connecting to Bybit WebSocket...")

	ws.SubscribeOrderbook("BTCUSDT", 50)
	ws.SubscribeTrade("BTCUSDT")
	ws.SubscribeTicker("BTCUSDT")
	ws.SubscribeKline("BTCUSDT", "1")

	fmt.Println("✅ Subscribed to:")
	fmt.Println("   - Orderbook (depth 50)")
	fmt.Println("   - Public Trades")
	fmt.Println("   - Ticker")
	fmt.Println("   - Kline (1m)")
	fmt.Println("\n📡 Listening for messages...\n")

	ws.OnMessage(func(data map[string]interface{}) {
		if errorMsg, ok := data["error"].(bool); ok && errorMsg {
			fmt.Printf("❌ Error: %v\n", data["message"])
			return
		}

		if topic, ok := data["topic"].(string); ok {
			switch {
			case contains(topic, "orderbook"):
				handleOrderbook(data)
			case contains(topic, "publicTrade"):
				handleTrade(data)
			case contains(topic, "tickers"):
				handleTicker(data)
			case contains(topic, "kline"):
				handleKline(data)
			}
		}

		if op, ok := data["op"].(string); ok {
			if op == "subscribe" {
				fmt.Printf("✅ Subscription confirmed: %v\n", data)
			}
		}
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Shutting down...")
		ws.Close()
		os.Exit(0)
	}()

	if err := ws.Listen(); err != nil {
		log.Fatal(err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func handleOrderbook(data map[string]interface{}) {
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		symbol := dataField["s"]
		fmt.Printf("📖 Orderbook Update - %v\n", symbol)
	}
}

func handleTrade(data map[string]interface{}) {
	if dataField, ok := data["data"].([]interface{}); ok && len(dataField) > 0 {
		if trade, ok := dataField[0].(map[string]interface{}); ok {
			fmt.Printf("💱 Trade - Symbol: %v, Price: %v, Qty: %v, Side: %v\n",
				trade["s"], trade["p"], trade["v"], trade["S"])
		}
	}
}

func handleTicker(data map[string]interface{}) {
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		fmt.Printf("📊 Ticker - Symbol: %v, Last: %v, 24h Change: %v%%\n",
			dataField["symbol"], dataField["lastPrice"], dataField["price24hPcnt"])
	}
}

func handleKline(data map[string]interface{}) {
	if dataField, ok := data["data"].([]interface{}); ok && len(dataField) > 0 {
		if kline, ok := dataField[0].(map[string]interface{}); ok {
			fmt.Printf("📈 Kline - Open: %v, High: %v, Low: %v, Close: %v, Volume: %v\n",
				kline["open"], kline["high"], kline["low"], kline["close"], kline["volume"])
		}
	}
}
