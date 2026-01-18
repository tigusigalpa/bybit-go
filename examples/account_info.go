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

	fmt.Println("=== Bybit Go SDK - Account Information Example ===\n")

	fmt.Println("👤 Getting Account Info...")
	accountInfo, err := client.GetAccountInfo()
	if err != nil {
		log.Printf("Error getting account info: %v\n", err)
	} else {
		if result, ok := accountInfo["result"].(map[string]interface{}); ok {
			fmt.Printf("   Account Type: %v\n", result["unifiedMarginStatus"])
			fmt.Printf("   Account Status: %v\n", result["status"])
			if marginMode, ok := result["marginMode"].(string); ok {
				fmt.Printf("   Margin Mode: %v\n", marginMode)
			}
		}
		fmt.Println()
	}

	fmt.Println("💰 Getting Wallet Balance...")
	balance, err := client.GetWalletBalance(map[string]interface{}{
		"accountType": "UNIFIED",
	})
	if err != nil {
		log.Printf("Error getting wallet balance: %v\n", err)
	} else {
		if result, ok := balance["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok && len(list) > 0 {
				if account, ok := list[0].(map[string]interface{}); ok {
					fmt.Printf("   Total Wallet Balance: %v\n", account["totalWalletBalance"])
					fmt.Printf("   Total Available Balance: %v\n", account["totalAvailableBalance"])
					fmt.Printf("   Total Equity: %v\n", account["totalEquity"])

					if coins, ok := account["coin"].([]interface{}); ok {
						fmt.Printf("\n   Coin Balances:\n")
						for _, coinData := range coins {
							if coin, ok := coinData.(map[string]interface{}); ok {
								fmt.Printf("      %v: %v (Available: %v)\n",
									coin["coin"], coin["walletBalance"], coin["availableToWithdraw"])
							}
						}
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("💸 Getting Transferable Amount...")
	transferable, err := client.GetTransferableAmount(map[string]interface{}{
		"accountType": "UNIFIED",
		"coin":        "USDT",
	})
	if err != nil {
		log.Printf("Error getting transferable amount: %v\n", err)
	} else {
		if result, ok := transferable["result"].(map[string]interface{}); ok {
			fmt.Printf("   Transferable USDT: %v\n", result["transferableAmount"])
		}
		fmt.Println()
	}

	fmt.Println("📜 Getting Transaction Log...")
	transactions, err := client.GetTransactionLog(map[string]interface{}{
		"accountType": "UNIFIED",
		"limit":       10,
	})
	if err != nil {
		log.Printf("Error getting transaction log: %v\n", err)
	} else {
		if result, ok := transactions["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok {
				fmt.Printf("   Found %d transaction(s)\n", len(list))
				for i, tx := range list {
					if t, ok := tx.(map[string]interface{}); ok {
						fmt.Printf("   Transaction #%d:\n", i+1)
						fmt.Printf("      Type: %v\n", t["type"])
						fmt.Printf("      Coin: %v\n", t["coin"])
						fmt.Printf("      Amount: %v\n", t["cashFlow"])
						fmt.Printf("      Time: %v\n", t["transactionTime"])
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("🔧 Getting Account Instruments Info...")
	instruments, err := client.GetAccountInstrumentsInfo(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Printf("Error getting instruments info: %v\n", err)
	} else {
		if result, ok := instruments["result"].(map[string]interface{}); ok {
			if list, ok := result["list"].([]interface{}); ok && len(list) > 0 {
				if instrument, ok := list[0].(map[string]interface{}); ok {
					fmt.Printf("   Symbol: %v\n", instrument["symbol"])
					fmt.Printf("   Leverage: %v\n", instrument["leverage"])
					fmt.Printf("   Margin Mode: %v\n", instrument["marginMode"])
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("✅ Account information example completed successfully!")
}
