# Bybit Golang SDK

![Bybit Golang SDK](https://i.postimg.cc/2yjbGXVh/bybit-go-github-hero.jpg)

[![CI](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml)
[![Tests](https://img.shields.io/badge/tests-go%20test%20--race-brightgreen)](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml)
[![Go vet](https://img.shields.io/badge/code%20analysis-go%20vet-brightgreen)](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tigusigalpa/bybit-go)](https://goreportcard.com/report/github.com/tigusigalpa/bybit-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/tigusigalpa/bybit-go.svg)](https://pkg.go.dev/github.com/tigusigalpa/bybit-go)
[![Go version](https://img.shields.io/github/go-mod/go-version/tigusigalpa/bybit-go)](go.mod)
[![License](https://img.shields.io/github/license/tigusigalpa/bybit-go)](LICENSE)

A small Go client for the [Bybit V5 API](https://bybit-exchange.github.io/docs/v5/intro). It makes it easier to get started with the REST API, demo trading, WebSocket streams, and TradFi instruments without imposing an application architecture on your project.

> 📚 **Looking for the deeper dive?** Explore the [bybit-go Wiki](https://github.com/tigusigalpa/bybit-go/wiki) for guides, endpoint notes, and extended documentation.

> Trading involves risk. Test both the integration and your strategy in the demo environment first, use narrowly scoped API-key permissions, and never commit keys to a repository.

## Contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Quick start](#quick-start)
- [REST API](#rest-api)
- [Orders and positions](#orders-and-positions)
- [Errors and response handling](#errors-and-response-handling)
- [RSA signatures](#rsa-signatures)
- [WebSocket](#websocket)
- [TradFi](#tradfi)
- [Examples, testing, and contributing](#examples-testing-and-contributing)

## Features

- REST methods for market data, orders, accounts, and positions;
- HMAC-SHA256 and RSA-SHA256 signing for REST requests;
- demo mode plus regional REST and WebSocket endpoints;
- public and private WebSocket subscriptions;
- convenience helpers for TradFi instruments (forex, metals, stocks, and indices);
- standalone working examples in [`examples/`](examples/README.md).

## Installation

Go 1.21 or later is required.

```bash
go get github.com/tigusigalpa/bybit-go
```

```go
import bybit "github.com/tigusigalpa/bybit-go"
```

## Configuration

Create one client and reuse it for the lifetime of your application. The default HTTP client has a 30-second timeout. You may provide your own `*http.Client` when you need a proxy, custom transport, observability, or different timeout settings.

| `ClientConfig` field | Default | Description |
|---|---:|---|
| `APIKey` | — | API key for signed endpoints. |
| `APISecret` | — | API secret for HMAC signing. |
| `Demo` | `false` | Routes REST requests to the Bybit demo environment. |
| `Region` | `global` | Endpoint region: `global`, `nl`, `tr`, `kz`, `ge`, or `ae`. `demo` also selects the demo REST endpoint. |
| `RecvWindow` | `5000` | Bybit receive window in milliseconds. |
| `Signature` | `hmac` | Signature algorithm: `hmac` or `rsa`. |
| `RSAPrivateKey` | — | PEM private key, required when `Signature` is `rsa`. |
| `HTTPClient` | 30 s timeout | Optional custom HTTP client. |

```go
httpClient := &http.Client{Timeout: 10 * time.Second}
client, err := bybit.NewClient(bybit.ClientConfig{
	APIKey:     os.Getenv("BYBIT_API_KEY"),
	APISecret:  os.Getenv("BYBIT_API_SECRET"),
	Demo:       true,
	RecvWindow: 5_000,
	HTTPClient: httpClient,
})
```

`Demo: true` takes precedence over `Region`. Use demo credentials that belong to the demo environment; do not expect production credentials or balances to work there.

## Quick start

Public requests use the same client and do not require an API key.

```go
package main

import (
	"fmt"
	"log"

	bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
	client, err := bybit.NewClient(bybit.ClientConfig{Demo: true})
	if err != nil {
		log.Fatal(err)
	}

	tickers, err := client.GetTickers(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", tickers["result"])
}
```

For account operations, pass credentials through environment variables:

```go
client, err := bybit.NewClient(bybit.ClientConfig{
	APIKey:    os.Getenv("BYBIT_API_KEY"),
	APISecret: os.Getenv("BYBIT_API_SECRET"),
	Demo:      true,
	Region:    "global", // also: nl, tr, kz, ge, ae
})
```

## REST API

Methods return the decoded Bybit response as `map[string]interface{}`. This deliberately leaves the full V5 request surface available: pass the exact fields documented by Bybit in the `params` map.

### Market data

```go
orderbook, err := client.GetOrderbook(map[string]interface{}{
	"category": "spot",
	"symbol":   "BTCUSDT",
	"limit":    50,
})

klines, err := client.GetKline(map[string]interface{}{
	"category": "linear",
	"symbol":   "BTCUSDT",
	"interval": "60",
	"limit":    200,
})

trades, err := client.GetRecentTrades(map[string]interface{}{
	"category": "linear",
	"symbol":   "BTCUSDT",
})
```

Available market helpers include `GetServerTime`, `GetTickers`, `GetKline`, `GetOrderbook`, `GetRPIOrderbook`, `GetOpenInterest`, `GetRecentTrades`, `GetFundingRateHistory`, `GetHistoricalVolatility`, `GetInsurance`, and `GetRiskLimit`.

### Account and positions

```go
wallet, err := client.GetWalletBalance(map[string]interface{}{
	"accountType": "UNIFIED",
})
positions, err := client.GetPositions(map[string]interface{}{
	"category": "linear",
	"symbol":   "BTCUSDT",
})
```

The client also provides helpers for account info, transaction logs, open and closed positions, trading stops, margin, leverage, and risk-limit operations. Refer to the [official V5 documentation](https://bybit-exchange.github.io/docs/v5/intro) for endpoint-specific required fields and account-mode rules.

## Orders and positions

API parameters are passed through directly, so you can use newly added Bybit fields without waiting for an SDK release.

```go
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
	log.Fatal(err)
}
fmt.Println(order)
```

For a higher-level order helper, `PlaceOrder` can calculate a derivatives quantity from margin, price, and leverage. Use it only when that calculation matches your instrument's quantity rules; for precise production order sizing, retrieve instrument constraints and submit `CreateOrder` yourself.

`SetLeverage` rejects a non-positive leverage value and accepts `Buy` or `Sell` when changing one side only. Passing no side changes both buy and sell leverage.

### Demo trading

`NewDemoClient` creates a `DemoClient` with `Demo` enabled. It exposes the usual order and account helpers as well as demo-specific operations such as funding requests.

```go
demo, err := bybit.NewDemoClient(bybit.ClientConfig{
	APIKey:    os.Getenv("BYBIT_DEMO_API_KEY"),
	APISecret: os.Getenv("BYBIT_DEMO_API_SECRET"),
})
if err != nil {
	log.Fatal(err)
}

result, err := demo.ApplyForDemoFundsSimple("USDT", "10000")
```

## Errors and response handling

There are two error layers to handle:

1. Transport, request-construction, HTTP-status, signature, and JSON decoding failures are returned as Go errors.
2. A Bybit business error, represented by a non-zero `retCode`, usually arrives with HTTP 200. It is returned in the response map and must be checked by the caller.

```go
response, err := client.CreateOrder(params)
if err != nil {
	var httpErr *bybit.HTTPError
	if errors.As(err, &httpErr) {
		log.Printf("Bybit HTTP failure: status=%d body=%s", httpErr.StatusCode, httpErr.Body)
	}
	log.Fatal(err)
}

if code, ok := response["retCode"].(float64); !ok || code != 0 {
	log.Fatalf("Bybit rejected request: code=%v message=%v", response["retCode"], response["retMsg"])
}
```

For public calls, an empty API key produces an empty HMAC signature. Bybit may ignore these headers for public endpoints, but authenticated operations require valid credentials. Treat API responses as untrusted input: check types before using nested values from the decoded map.

## RSA signatures

For an RSA API key, provide its PEM-encoded private key. `Signature: "rsa"` requires a private key, and unsupported signature types are rejected when creating the client.

```go
client, err := bybit.NewClient(bybit.ClientConfig{
	APIKey:        os.Getenv("BYBIT_API_KEY"),
	Signature:     "rsa",
	RSAPrivateKey: os.Getenv("BYBIT_RSA_PRIVATE_KEY"),
})
```

## WebSocket

```go
ws := bybit.NewWebSocket(bybit.WebSocketConfig{Demo: true})
defer ws.Close()

ws.OnMessage(func(message map[string]interface{}) {
	fmt.Printf("%+v\n", message)
})

if err := ws.SubscribeTicker("BTCUSDT"); err != nil {
	log.Fatal(err)
}
if err := ws.Listen(); err != nil {
	log.Fatal(err)
}
```

The package's public WebSocket connects to the spot endpoint. For private streams, set `IsPrivate` together with `APIKey` and `APISecret`; the client authenticates after connecting.

| Helper | Topic |
|---|---|
| `SubscribeOrderbook("BTCUSDT", 50)` | `orderbook.50.BTCUSDT` |
| `SubscribeTrade("BTCUSDT")` | `publicTrade.BTCUSDT` |
| `SubscribeTicker("BTCUSDT")` | `tickers.BTCUSDT` |
| `SubscribeKline("BTCUSDT", "1")` | `kline.1.BTCUSDT` |
| `SubscribePosition`, `SubscribeOrder`, `SubscribeExecution`, `SubscribeWallet` | Private account topics |

Call `Unsubscribe` with the exact topics when they are no longer needed. Your application owns the connection lifecycle and should reconnect after an error; on reconnect, subscribe again using `GetSubscriptions` as your source of truth.

## TradFi

The package includes lists of popular instruments and focused helper methods:

```go
tickers, err := client.GetTradFiTicker("XAUUSD")
positions, err := client.GetTradFiPositions("XAUUSD")
order, err := client.PlaceTradFiOrder(bybit.TradFiOrderParams{
	Symbol: "XAUUSD", Side: "Buy", OrderType: "Market", Qty: "1",
})
```

Instrument availability and trading conditions vary by account and region. Call `GetTradFiInstruments` to retrieve the current list before placing an order.

## Examples, testing, and contributing

See the [examples directory](examples/README.md) for basic client, market data, orders, positions, demo trading, TradFi, and WebSocket programs. Examples may make network calls or require credentials, so read their source before running them.

```bash
go test ./...
go vet ./...
```

CI runs these checks on Go 1.21 and the current stable Go release. Pull requests are welcome: please add a test when changing behavior, and never include real API keys or personal data.

The repository's GitHub Actions workflow additionally runs the test suite with Go's race detector. Before opening a pull request, format changed Go files with `gofmt` and keep `go.mod` and `go.sum` tidy.

## Documentation and license

- [Official Bybit V5 documentation](https://bybit-exchange.github.io/docs/v5/intro)
- [Extended TradFi notes](wiki-tradfi.md)
- [Installation instructions](INSTALLATION.md)
- [MIT License](LICENSE)
