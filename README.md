# Bybit Golang SDK

![Bybit Golang SDK](https://i.postimg.cc/2yjbGXVh/bybit-go-github-hero.jpg)

[![CI](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml)
[![Tests](https://img.shields.io/badge/tests-go%20test%20--race-brightgreen)](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml)
[![Go vet](https://img.shields.io/badge/code%20analysis-go%20vet-brightgreen)](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml)
[![CodeQL](https://github.com/tigusigalpa/bybit-go/actions/workflows/codeql.yml/badge.svg?branch=main)](https://github.com/tigusigalpa/bybit-go/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/tigusigalpa/bybit-go/graph/badge.svg)](https://codecov.io/gh/tigusigalpa/bybit-go)
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
- [xStocks](#xstocks)
- [TradFi](#tradfi)
- [Examples, testing, and contributing](#examples-testing-and-contributing)

## Features

- REST methods for market data, orders, accounts, and positions;
- HMAC-SHA256 and RSA-SHA256 signing for REST requests;
- demo mode plus regional REST and WebSocket endpoints;
- public and private WebSocket subscriptions;
- typed discovery, validation, and decimal-safe conversion helpers for xStocks Spot tokens;
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
| `ReceiptBodyLimit` | 4 MiB | Maximum response bytes retained by raw REST receipt methods. |

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

Most REST methods return the decoded Bybit response as `map[string]interface{}`. Pass the fields documented by Bybit in the `params` map. `GetKlineReceipt` additionally exposes the HTTP response body without JSON decoding.

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

### Exact kline receipts

`GetKlineReceipt(ctx, params)` captures the public HTTP response body for archival, repair, or backfill workflows. It uses the same request builder, endpoint settings, and injected `HTTPClient` as the decoded methods. No API key is required, and the SDK does not attach authentication headers. The only accepted request fields are `category`, `symbol`, `interval`, `start`, `end`, and `limit`, as documented for [Bybit Get Kline](https://bybit-exchange.github.io/docs/v5/market/kline).

The method does not parse JSON, so whitespace, key order, unknown fields, numeric/string lexemes, and the original ordering of candle arrays are preserved. Bybit returns candles in reverse start-time order and may include an unfinished candle; filtering or normalization belongs to the caller.

```go
client, err := bybit.NewClient(bybit.ClientConfig{
	ReceiptBodyLimit: 4 << 20, // 4 MiB; zero selects this default
})
if err != nil {
	log.Fatal(err)
}
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

receipt, captureErr := client.GetKlineReceipt(ctx, map[string]interface{}{
	"category": "linear",
	"symbol":   "BTCUSDT",
	"interval": "1",
	"start":    "1700000000000",
	"end":      "1700000059999",
	"limit":    200,
})
if receipt != nil {
	// Archive evidence before handling captureErr, including partial bodies.
	raw := receipt.ResponseBody()
	digest := receipt.ResponseBodySHA256()
	fmt.Printf("%s hash=%x\n", receipt, digest) // summary excludes payload/header values
	_ = raw // pass raw and receipt metadata to your application's archive
}
if captureErr != nil {
	// errors.Is / errors.As retain context, transport, read, close, limit,
	// and *bybit.HTTPError causes. A nil receipt means no response was observed.
	return
}
// Decode raw separately and validate retCode before using any market data.
```

See the runnable [no-key receipt example](examples/kline_receipt/main.go), which writes captured body bytes to stdout and a safe summary to stderr.

`ClientConfig.ReceiptBodyLimit` defaults to 4 MiB. The SDK retains at most that many bytes and probes at most one additional byte to detect overflow. Exactly the limit followed by EOF is accepted. On overflow, `errors.Is(err, bybit.ErrReceiptBodyLimitExceeded)` succeeds; the retained bytes and their SHA-256 describe the captured prefix, not the full response. The limit bounds body capture, not HTTP header memory or every temporary allocation.

Receipt lifecycle and errors:

| Result | Receipt evidence | `Complete()` |
|---|---|---|
| Body read to EOF and closed successfully | Exact body, including an empty body | `true` |
| Non-2xx with a fully read body | Exact status/body plus `*HTTPError` | `true` (HTTP capture only) |
| Read failure, overflow, close failure, or cancellation | Available bytes/status plus all error causes | `false` |
| Transport failure without a response | `nil` receipt; original transport/context error | No receipt |
| Injected client's redirect policy rejects a redirect | Observed status/headers; body already closed by `http.Client`, so no captured bytes | `false` |

`CapturedAt()` records when the response is available from the HTTP client. `CompletedAt()` is recorded after body reading and the original body's single `Close` call have finished. Both are local UTC times, independent of candle timestamps. Cancellation closes the body to unblock reading; injected transports must obey `net/http`'s concurrent `Read`/`Close` contract. The injected client's timeout and redirect policy remain active. `URL()` records the actual response request URL, including the final query after redirects.

`ResponseBody()`, `RequestBody()`, and `ResponseHeaders()` return defensive copies and are safe to access concurrently after return. The GET request body is empty. Response headers are restricted to content metadata (`Content-Type`, `Content-Length`, `Content-Encoding`, `Date`, `ETag`, `Last-Modified`), `Retry-After`, Bybit rate-limit headers (`X-Bapi-Limit`, `X-Bapi-Limit-Status`, `X-Bapi-Limit-Reset-Timestamp`), and correlation headers (`X-Request-Id`, `X-Trace`). Authentication/cookie headers and unlisted headers are excluded. Formatting the receipt with `%v`, `%+v`, or `%#v` shows a summary without URLs, payloads, or header values; the SDK does not log receipts automatically.

Requests use `Accept-Encoding: identity` to prevent standard `net/http` gzip decompression from changing captured bytes. Receipts preserve response **body** bytes exposed by the transport, not HTTP framing or TLS packets. An injected transport must preserve body bytes if wire-level body fidelity is required.

A complete HTTP receipt does **not** validate JSON or Bybit's application-level `retCode`. Callers must inspect the body before treating it as a successful market-data result. The SDK adds no retries, and a receipt represents one observed response rather than an attempt history. Redirects or retries performed inside an injected client/transport are not a full history exposed by this API. Existing `GetKline` and `Request` retain their decoding and cleanup behavior; the receipt body limit applies only to receipt capture.

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

The package's public WebSocket connects to the Spot endpoint by default. For private streams, set `IsPrivate` together with `APIKey` and `APISecret`; the client authenticates after connecting.

For a linear perpetual market-data collector, select `WebSocketCategoryLinear`. Public streams do not require an API key. `OnRawMessage` receives the exact wire JSON and a receive timestamp before decoding, which is useful when persisting source events. In the kline payload, `confirm: true` means the candle is closed.

```go
ws := bybit.NewWebSocket(bybit.WebSocketConfig{
	PublicCategory: bybit.WebSocketCategoryLinear,
})
defer ws.Close()

ws.OnRawMessage(func(raw []byte, receivedAt time.Time) {
	// Persist raw and receivedAt before normalizing the message.
})

if err := ws.SubscribeKline("BTCUSDT", "1"); err != nil {
	log.Fatal(err)
}
if err := ws.ListenContext(context.Background()); err != nil {
	log.Fatal(err) // reconnecting is owned by the application
}
```

`ListenContext` returns context cancellation and unexpected network errors. The legacy `Listen` method remains available for existing integrations. The linear public endpoint is currently exposed only for the documented global environment; the SDK rejects regional or demo configurations rather than guessing a URL.

| Helper | Topic |
|---|---|
| `SubscribeOrderbook("BTCUSDT", 50)` | `orderbook.50.BTCUSDT` |
| `SubscribeTrade("BTCUSDT")` | `publicTrade.BTCUSDT` |
| `SubscribeTicker("BTCUSDT")` | `tickers.BTCUSDT` |
| `SubscribeKline("BTCUSDT", "1")` | `kline.1.BTCUSDT` |
| `SubscribePosition`, `SubscribeOrder`, `SubscribeExecution`, `SubscribeWallet` | Private account topics |

Call `Unsubscribe` with the exact topics when they are no longer needed. Your application owns the connection lifecycle and should reconnect after an error; on reconnect, subscribe again using `GetSubscriptions` as your source of truth.

## xStocks

xStocks are regular V5 Spot instruments, not TradFi CFDs. Discover the live catalogue with `GetXStocks()`—the client accepts an instrument as xStock only when Bybit returns the exact `symbolType: "xstocks"` value.

```go
instruments, err := client.GetXStocks()
if err != nil {
	log.Fatal(err)
}

order, err := client.PlaceXStockOrder(bybit.XStockOrderParams{
	Symbol: "AAPLXUSDT", Side: "Buy", OrderType: "Market",
	Qty: "100", MarketUnit: "quoteCoin",
})
```

`PlaceXStockOrder` refreshes metadata, checks the status, precision, tick size, limits, and submits a standard Spot request with margin disabled. Decimal helper functions such as `ToXStockTokenQuantity` and `ToXStockTokenPrice` retain exact arithmetic and apply the required rounding policy. Read the [xStocks Wiki guide](wiki/XStocks.md) before placing production orders.

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
- [Installation instructions](INSTALLATION.md)
- [MIT License](LICENSE)
