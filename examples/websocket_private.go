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
	apiKey := os.Getenv("BYBIT_API_KEY")
	apiSecret := os.Getenv("BYBIT_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("Please set BYBIT_API_KEY and BYBIT_API_SECRET environment variables")
	}

	fmt.Println("=== Bybit Go SDK - Private WebSocket Example ===\n")

	ws := bybit.NewWebSocket(bybit.WebSocketConfig{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Testnet:   true,
		Region:    "global",
		IsPrivate: true,
	})

	fmt.Println("🔌 Connecting to Bybit Private WebSocket...")

	ws.SubscribePosition()
	ws.SubscribeOrder()
	ws.SubscribeExecution()
	ws.SubscribeWallet()

	fmt.Println("✅ Subscribed to:")
	fmt.Println("   - Position updates")
	fmt.Println("   - Order updates")
	fmt.Println("   - Execution updates")
	fmt.Println("   - Wallet updates")
	fmt.Println("\n📡 Listening for messages...\n")

	ws.OnMessage(func(data map[string]interface{}) {
		if errorMsg, ok := data["error"].(bool); ok && errorMsg {
			fmt.Printf("❌ Error: %v\n", data["message"])
			return
		}

		if topic, ok := data["topic"].(string); ok {
			switch topic {
			case "position":
				handlePositionUpdate(data)
			case "order":
				handleOrderUpdate(data)
			case "execution":
				handleExecutionUpdate(data)
			case "wallet":
				handleWalletUpdate(data)
			}
		}

		if op, ok := data["op"].(string); ok {
			if op == "subscribe" {
				fmt.Printf("✅ Subscription confirmed: %v\n", data)
			} else if op == "auth" {
				if success, ok := data["success"].(bool); ok && success {
					fmt.Println("✅ Authentication successful!\n")
				}
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

func handlePositionUpdate(data map[string]interface{}) {
	if dataField, ok := data["data"].([]interface{}); ok && len(dataField) > 0 {
		if position, ok := dataField[0].(map[string]interface{}); ok {
			fmt.Printf("📍 Position Update:\n")
			fmt.Printf("   Symbol: %v\n", position["symbol"])
			fmt.Printf("   Side: %v\n", position["side"])
			fmt.Printf("   Size: %v\n", position["size"])
			fmt.Printf("   Entry Price: %v\n", position["entryPrice"])
			fmt.Printf("   Unrealized PnL: %v\n\n", position["unrealisedPnl"])
		}
	}
}

func handleOrderUpdate(data map[string]interface{}) {
	if dataField, ok := data["data"].([]interface{}); ok && len(dataField) > 0 {
		if order, ok := dataField[0].(map[string]interface{}); ok {
			fmt.Printf("📝 Order Update:\n")
			fmt.Printf("   Order ID: %v\n", order["orderId"])
			fmt.Printf("   Symbol: %v\n", order["symbol"])
			fmt.Printf("   Side: %v\n", order["side"])
			fmt.Printf("   Price: %v\n", order["price"])
			fmt.Printf("   Qty: %v\n", order["qty"])
			fmt.Printf("   Status: %v\n\n", order["orderStatus"])
		}
	}
}

func handleExecutionUpdate(data map[string]interface{}) {
	if dataField, ok := data["data"].([]interface{}); ok && len(dataField) > 0 {
		if execution, ok := dataField[0].(map[string]interface{}); ok {
			fmt.Printf("⚡ Execution Update:\n")
			fmt.Printf("   Symbol: %v\n", execution["symbol"])
			fmt.Printf("   Side: %v\n", execution["side"])
			fmt.Printf("   Exec Price: %v\n", execution["execPrice"])
			fmt.Printf("   Exec Qty: %v\n", execution["execQty"])
			fmt.Printf("   Exec Fee: %v\n\n", execution["execFee"])
		}
	}
}

func handleWalletUpdate(data map[string]interface{}) {
	if dataField, ok := data["data"].([]interface{}); ok && len(dataField) > 0 {
		if wallet, ok := dataField[0].(map[string]interface{}); ok {
			fmt.Printf("💰 Wallet Update:\n")
			fmt.Printf("   Account Type: %v\n", wallet["accountType"])
			if coins, ok := wallet["coin"].([]interface{}); ok && len(coins) > 0 {
				for _, coinData := range coins {
					if coin, ok := coinData.(map[string]interface{}); ok {
						fmt.Printf("   Coin: %v\n", coin["coin"])
						fmt.Printf("   Wallet Balance: %v\n", coin["walletBalance"])
						fmt.Printf("   Available: %v\n\n", coin["availableToWithdraw"])
					}
				}
			}
		}
	}
}
