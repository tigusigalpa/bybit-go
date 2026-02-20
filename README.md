# 🚀 Bybit Go SDK: Your Ticket to the World of Lightning-Fast Trading

<div align="center">

### ⚡️ A Golang Client for the V5 API, as Fast as a Rocket

![Bybit Golang SDK](https://github.com/user-attachments/assets/3463194b-a042-4f84-adfb-708ec7ade75a)

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](LICENSE)
[![WebSocket](https://img.shields.io/badge/WebSocket-Real--Time-brightgreen?style=for-the-badge&logo=socketdotio)](https://bybit-exchange.github.io/docs/v5/ws/connect)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-ff69b4?style=for-the-badge)](https://github.com/tigusigalpa/bybit-go/pulls)

</div>

Turn your boldest trading ideas into reality with the most powerful Go SDK for the Bybit V5 API. Forget wrestling with documentation and focus on what truly matters—building profitable strategies.

---

## 🎯 Why Do Traders and Developers Choose the Bybit Go SDK?

This SDK is more than just an API wrapper. It's your trusted partner in the world of algorithmic trading, built with a love for Golang and a deep understanding of traders' needs.

<br>

<details>
<summary><b>🤖 For Algorithmic Traders</b></summary>

No more battling with latency and inefficient code. Our SDK allows you to execute complex strategies with millisecond precision. Get real-time market data via WebSocket, use advanced order types, and manage risk with automatic controls. The built-in fee calculator ensures crystal-clear tracking of your profits.

</details>

<details>
<summary><b>👨‍💻 For Developers</b></summary>

We've taken care of your comfort. Type-safe operations eliminate runtime errors, and Testnet support lets you experiment without risk. The SDK has zero unnecessary dependencies, is easily configurable, and comes with comprehensive documentation and examples. It's the tool you'll love to use.

</details>

<details>
<summary><b>🏢 For Production Systems</b></summary>

Reliability is our middle name. Thread-safe operations, smart reconnection with exponential backoff, and detailed error messages make this SDK the perfect choice for 24/7 systems. Built-in rate limit handling and multi-region support guarantee your application's stability under any conditions.

</details>

---

## ✨ Key Features

We've packed everything you need for successful trading into one powerful package.

| Category | Features |
|---|---|
| 📊 **Market Data** | Get real-time tickers, order books, K-lines, and trade history. |
| 💰 **Trading** | Spot, derivatives, limit, and market orders, TP/SL—all at your fingertips. |
| 🔐 **Security** | HMAC-SHA256 and RSA-SHA256 support to protect your data. |
| 🌐 **WebSocket** | Real-time data streaming with automatic reconnection. |
| ⚙️ **Management** | Control positions, leverage, risk, and your wallet. |
| 🌍 **Global Access** | Multi-region and testnet support for low-latency trading. |

---

## 📦 Installation & Configuration

Getting started is a breeze. You'll be up and running in seconds.

**Step 1:** Install the package using Go.
```bash
go get github.com/tigusigalpa/bybit-go
```

**Step 2:** Import it into your project.
```go
import bybit "github.com/tigusigalpa/bybit-go"
```

**Step 3:** Configure the client using environment variables or programmatic setup.

*Configuration Example:*
```go
client, err := bybit.NewClient(
    bybit.ClientConfig{
        APIKey:     os.Getenv("BYBIT_API_KEY"),
        APISecret:  os.Getenv("BYBIT_API_SECRET"),
        Testnet:    true, // Set to true for safe testing
        Region:     "global",
    },
)
```

---

## 🚀 Quick Start: From Zero to Your First Order

Let's take your first step into the world of automated trading together.

```go
package main

import (
    "fmt"
    "log"
    bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
    // Step 1: Initialize the client
    client, err := bybit.NewClient(bybit.ClientConfig{
        APIKey:     "your_api_key",
        APISecret:  "your_api_secret",
        Testnet:    true, // A safe environment for testing
    })
    if err != nil {
        log.Fatal(err)
    }

    // Step 2: Verify the connection to the server
    serverTime, err := client.GetServerTime()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ Connected! Server Time: %v\n", serverTime)

    // Step 3: Get real-time market data
    tickers, err := client.GetTickers(map[string]interface{}{
        "category": "linear",
        "symbol":   "BTCUSDT",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("📊 BTC Price: %v\n", tickers)

    // Step 4: Place your first order
    order, err := client.CreateOrder(map[string]interface{}{
        "category":    "linear",
        "symbol":      "BTCUSDT",
        "side":        "Buy",
        "orderType":   "Limit",
        "qty":         "0.01",
        "price":       "30000",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("🎉 Order placed successfully! ID: %v\n", order["result"].(map[string]interface{})["orderId"])
}
```

**Congratulations! You've just made your first trade via the API!**

---

## 📚 Examples & Documentation

To dive deeper into all the features of the SDK, check out the `examples/` directory. There you'll find working examples for all major functions, from fetching market data to managing WebSocket streams.

---

## 🤝 Make the Project Better

We always welcome community contributions! If you have ideas on how to improve the SDK or have found a bug, please create an issue or submit a pull request. Let's build the best tool for traders together!

---

## 📄 License

The project is distributed under the MIT License, giving you complete freedom to use and modify it.

<br>

*<sub>⚠️ **Disclaimer:** Trading cryptocurrencies involves high risk. This SDK is provided "as is" without any warranty. Always test your strategies on the Testnet and never invest more than you can afford to lose.</sub>*
