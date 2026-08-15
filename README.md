# bybit-go

[![CI](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/tigusigalpa/bybit-go.svg)](https://pkg.go.dev/github.com/tigusigalpa/bybit-go)
[![Go version](https://img.shields.io/github/go-mod/go-version/tigusigalpa/bybit-go)](go.mod)
[![License](https://img.shields.io/github/license/tigusigalpa/bybit-go)](LICENSE)

A small Go client for the [Bybit V5 API](https://bybit-exchange.github.io/docs/v5/intro). It makes it easier to get started with the REST API, demo trading, WebSocket streams, and TradFi instruments without imposing an application architecture on your project.

> Trading involves risk. Test both the integration and your strategy in the demo environment first, use narrowly scoped API-key permissions, and never commit keys to a repository.

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

## Placing an order

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

HTTP failures such as `429` and `401` are returned as `*bybit.HTTPError`. A Bybit response with `retCode != 0` will normally still use HTTP 200 and is returned as a map, so always inspect `retCode` before treating an operation as successful.

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
ws.OnMessage(func(message map[string]interface{}) {
	fmt.Printf("%+v\n", message)
})

if err := ws.SubscribeTicker("BTCUSDT"); err != nil {
	log.Fatal(err)
}
if err := ws.Listen(); err != nil {
	log.Fatal(err)
}
defer ws.Close()
```

The package's public WebSocket connects to the spot endpoint. For private streams, set `IsPrivate` together with `APIKey` and `APISecret`. Your application owns the connection lifecycle and should reconnect after an error.

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

## Examples and project checks

See the [examples directory](examples/README.md) for basic client, market data, orders, positions, demo trading, TradFi, and WebSocket programs. Examples may make network calls or require credentials, so read their source before running them.

```bash
go test ./...
go vet ./...
```

CI runs these checks on Go 1.21 and the current stable Go release. Pull requests are welcome: please add a test when changing behavior, and never include real API keys or personal data.

## Documentation and license

- [Official Bybit V5 documentation](https://bybit-exchange.github.io/docs/v5/intro)
- [Extended TradFi notes](wiki-tradfi.md)
- [Installation instructions](INSTALLATION.md)
- [MIT License](LICENSE)
