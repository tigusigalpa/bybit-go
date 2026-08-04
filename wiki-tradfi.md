# TradFi — Gold, Forex, Stocks & Indices

Bybit Go SDK includes a dedicated TradFi module (`tradfi.go`) for trading traditional financial instruments — precious metals, forex pairs, stock CFDs, and equity indices. All instruments are available on Bybit as **linear perpetual contracts** and are accessed through the standard V5 API with `category=linear`.

---

## Table of Contents

- [Overview](#overview)
- [Supported Instruments](#supported-instruments)
- [Predefined Symbol Lists](#predefined-symbol-lists)
- [Market Data](#market-data)
  - [GetTradFiInstruments](#gettradfiinstruments)
  - [GetTradFiTicker](#gettradfitickerr)
  - [GetMetalsTickers / GetForexTickers / GetStockTickers / GetIndexTickers](#shortcut-tickers)
  - [GetTradFiKline](#gettradfikline)
  - [GetTradFiOrderbook](#gettradfiorderbook)
  - [GetTradFiSwapFee](#gettradfiswapfee)
- [Trading](#trading)
  - [TradFiOrderParams](#tradfiorderparams)
  - [PlaceTradFiOrder](#placetradfiorder)
  - [CloseTradFiPosition](#closetradfiposition)
  - [SetTradFiLeverage](#settradfileverage)
  - [CancelTradFiOrder](#canceltradfiorder)
  - [GetTradFiOpenOrders](#gettradfiOpenorders)
  - [GetTradFiPositions](#gettradfipositions)
  - [GetTradFiTradeHistory](#gettradfitradehistory)
  - [GetTradFiFeeRate](#gettradfifeerate)
- [Utilities](#utilities)
  - [IsTradFiSymbol](#istradfisymbol)
- [Constants](#constants)
- [Market Hours & Swap Fees](#market-hours--swap-fees)
- [Full Example](#full-example)

---

## Overview

TradFi methods are added directly to the `*Client` struct in `tradfi.go`. No separate client or constructor is needed — just use your existing client:

```go
import bybit "github.com/tigusigalpa/bybit-go"

client, err := bybit.NewClient(bybit.ClientConfig{
    APIKey:    os.Getenv("BYBIT_API_KEY"),
    APISecret: os.Getenv("BYBIT_API_SECRET"),
    Demo:      true,
})

// All TradFi methods are available on the same client
ticker, err := client.GetTradFiTicker("XAUUSD")
```

---

## Supported Instruments

| Asset Class | Examples | Symbol Format |
|---|---|---|
| **Metals** | Gold, Silver, Platinum | `XAUUSD`, `XAGUSD`, `XPTUSD` |
| **Forex Majors** | EUR/USD, GBP/USD, USD/JPY | `EURUSD`, `GBPUSD`, `USDJPY` |
| **Forex Minors** | EUR/GBP, GBP/JPY, EUR/JPY | `EURGBP`, `GBPJPY`, `EURJPY` |
| **US Stock CFDs** | Apple, Tesla, NVIDIA | `AAPLUSDT`, `TSLAUSDT`, `NVDAUSDT` |
| **Indices** | S&P 500, NASDAQ 100, DAX 40 | `US500USD`, `US100USD`, `DE40USD` |

> Use `GET /v5/market/instruments-info?category=linear` or `GetTradFiInstruments("")` to get the full up-to-date list from Bybit.

---

## Predefined Symbol Lists

The following exported slices are available for convenience:

```go
bybit.TradFiMetals       // ["XAUUSD", "XAGUSD", "XPTUSD"]
bybit.TradFiForexMajors  // ["EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "NZDUSD", "USDCAD"]
bybit.TradFiForexMinors  // ["EURGBP", "EURJPY", "GBPJPY", "EURCHF", "AUDCAD", "AUDNZD", "CADJPY"]
bybit.TradFiUSStocks     // ["AAPLUSDT", "AMZNUSDT", "TSLAUSDT", "GOOGLSDT", "MSFTUSDT", "METAUSDT", "NVDAUSDT", "NFLXUSDT"]
bybit.TradFiIndices      // ["US30USD", "US100USD", "US500USD", "UK100USD", "DE40USD", "JP225USD"]
```

---

## Market Data

### GetTradFiInstruments

Returns available TradFi instruments. Pass an asset class to filter results.

```go
func (c *Client) GetTradFiInstruments(assetClass string) (map[string]interface{}, error)
```

**Parameters:**

| Name | Type | Description |
|---|---|---|
| `assetClass` | `string` | Filter: `"metal"`, `"forex"`, `"stock"`, `"index"`, `"commodity"`, or `""` for all |

```go
// All TradFi instruments
all, err := client.GetTradFiInstruments("")

// Only forex instruments
forex, err := client.GetTradFiInstruments("forex")

// Only metals
metals, err := client.GetTradFiInstruments("metal")
```

---

### GetTradFiTicker

Returns ticker snapshot for a single symbol.

```go
func (c *Client) GetTradFiTicker(symbol string) (map[string]interface{}, error)
```

```go
ticker, err := client.GetTradFiTicker("XAUUSD")
if err != nil {
    log.Fatal(err)
}
list := ticker["result"].(map[string]interface{})["list"].([]interface{})
t := list[0].(map[string]interface{})
fmt.Println("Gold last price:", t["lastPrice"])
fmt.Println("Mark price:", t["markPrice"])
fmt.Println("24h change:", t["price24hPcnt"])
```

---

### Shortcut Tickers

Four convenience methods return tickers for predefined symbol groups:

```go
func (c *Client) GetMetalsTickers() (map[string]interface{}, error)
func (c *Client) GetForexTickers()  (map[string]interface{}, error)
func (c *Client) GetStockTickers()  (map[string]interface{}, error)
func (c *Client) GetIndexTickers()  (map[string]interface{}, error)
```

> Note: these methods call `GET /v5/market/tickers` without a specific symbol filter (Bybit does not support multi-symbol queries in one call for linear). All linear tickers are returned and you should filter by symbol client-side.

```go
// Get all linear tickers, then filter metals yourself
result, err := client.GetMetalsTickers()
```

---

### GetTradFiKline

Returns candlestick (OHLCV) data for a TradFi symbol.

```go
func (c *Client) GetTradFiKline(symbol, interval string, limit int) (map[string]interface{}, error)
```

**Parameters:**

| Name | Type | Description |
|---|---|---|
| `symbol` | `string` | e.g. `"XAUUSD"` |
| `interval` | `string` | `1`, `3`, `5`, `15`, `30`, `60`, `120`, `240`, `360`, `720`, `D`, `W`, `M` |
| `limit` | `int` | Number of candles (max 200); pass `0` to use API default |

```go
klines, err := client.GetTradFiKline("XAUUSD", "60", 10)
if err != nil {
    log.Fatal(err)
}
list := klines["result"].(map[string]interface{})["list"].([]interface{})
for _, c := range list {
    candle := c.([]interface{})
    // [timestamp, open, high, low, close, volume, turnover]
    fmt.Printf("O:%v H:%v L:%v C:%v\n", candle[1], candle[2], candle[3], candle[4])
}
```

---

### GetTradFiOrderbook

Returns order book depth for a TradFi symbol.

```go
func (c *Client) GetTradFiOrderbook(symbol string, depth int) (map[string]interface{}, error)
```

**Parameters:**

| Name | Type | Description |
|---|---|---|
| `symbol` | `string` | e.g. `"EURUSD"` |
| `depth` | `int` | `1`, `25`, `50`, `100`, `200`; pass `0` to use default (25) |

```go
ob, err := client.GetTradFiOrderbook("EURUSD", 5)
if err != nil {
    log.Fatal(err)
}
res := ob["result"].(map[string]interface{})
bids := res["b"].([]interface{})
asks := res["a"].([]interface{})
fmt.Println("Best bid:", bids[0].([]interface{})[0])
fmt.Println("Best ask:", asks[0].([]interface{})[0])
```

---

### GetTradFiSwapFee

Returns instrument info including swap/overnight financing fee fields for a TradFi symbol.

```go
func (c *Client) GetTradFiSwapFee(symbol string) (map[string]interface{}, error)
```

```go
info, err := client.GetTradFiSwapFee("XAUUSD")
```

> Swap fees are charged when holding a TradFi position past the daily market close. Check the `fundingRate` and related fields in the response before opening a position you plan to hold overnight.

---

## Trading

### TradFiOrderParams

Struct for placing TradFi orders:

```go
type TradFiOrderParams struct {
    Symbol      string // required — e.g. "XAUUSD"
    Side        string // required — "Buy" or "Sell"
    OrderType   string // required — "Market" or "Limit"
    Qty         string // required — quantity as string
    Price       string // required for Limit orders
    TimeInForce string // "GTC" (default), "IOC", "FOK", "PostOnly"
    TakeProfit  string // optional TP price
    StopLoss    string // optional SL price
    PositionIdx int    // 0=one-way (default), 1=hedge-long, 2=hedge-short
    ReduceOnly  bool   // true to reduce/close only
    OrderLinkID string // optional client order ID
}
```

---

### PlaceTradFiOrder

Places a limit or market order on a TradFi instrument.

```go
func (c *Client) PlaceTradFiOrder(p TradFiOrderParams) (map[string]interface{}, error)
```

```go
// Limit buy on gold with TP/SL
order, err := client.PlaceTradFiOrder(bybit.TradFiOrderParams{
    Symbol:      "XAUUSD",
    Side:        "Buy",
    OrderType:   "Limit",
    Qty:         "0.01",
    Price:       "3200",
    TakeProfit:  "3350",
    StopLoss:    "3100",
    TimeInForce: "GTC",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("Order ID:", order["result"].(map[string]interface{})["orderId"])
```

```go
// Market sell on EURUSD
order, err := client.PlaceTradFiOrder(bybit.TradFiOrderParams{
    Symbol:    "EURUSD",
    Side:      "Sell",
    OrderType: "Market",
    Qty:       "1",
})
```

---

### CloseTradFiPosition

Closes an open TradFi position at market price using `reduceOnly=true`.

```go
func (c *Client) CloseTradFiPosition(symbol, side string, qty string, positionIdx int) (map[string]interface{}, error)
```

**Parameters:**

| Name | Type | Description |
|---|---|---|
| `symbol` | `string` | e.g. `"XAUUSD"` |
| `side` | `string` | Current position side: `"Buy"` (long) or `"Sell"` (short) |
| `qty` | `string` | Quantity to close |
| `positionIdx` | `int` | `0` for one-way mode |

```go
// Close a long gold position
result, err := client.CloseTradFiPosition("XAUUSD", "Buy", "0.01", 0)
```

---

### SetTradFiLeverage

Sets buy and sell leverage for a TradFi symbol.

```go
func (c *Client) SetTradFiLeverage(symbol string, leverage float64) (map[string]interface{}, error)
```

> TradFi instruments typically support **1x–20x** leverage depending on the asset class. Metals tend to allow higher leverage than stock CFDs.

```go
result, err := client.SetTradFiLeverage("XAUUSD", 10)
result, err := client.SetTradFiLeverage("EURUSD", 20)
result, err := client.SetTradFiLeverage("US500USD", 5)
```

---

### CancelTradFiOrder

Cancels a TradFi order by `orderId` or `orderLinkId`.

```go
func (c *Client) CancelTradFiOrder(symbol, orderID, orderLinkID string) (map[string]interface{}, error)
```

```go
// Cancel by orderId
result, err := client.CancelTradFiOrder("XAUUSD", "abc123", "")

// Cancel by orderLinkId
result, err := client.CancelTradFiOrder("XAUUSD", "", "my-gold-order-1")
```

---

### GetTradFiOpenOrders

Returns open (untriggered) orders for TradFi instruments.

```go
func (c *Client) GetTradFiOpenOrders(symbol string) (map[string]interface{}, error)
```

```go
// Orders for a specific symbol
orders, err := client.GetTradFiOpenOrders("XAUUSD")

// All open TradFi orders
orders, err := client.GetTradFiOpenOrders("")
```

---

### GetTradFiPositions

Returns open TradFi positions for the account.

```go
func (c *Client) GetTradFiPositions(symbol string) (map[string]interface{}, error)
```

```go
// Positions for a specific symbol
positions, err := client.GetTradFiPositions("XAUUSD")

// All linear positions (filter by IsTradFiSymbol client-side)
positions, err := client.GetTradFiPositions("")
if err != nil {
    log.Fatal(err)
}
list := positions["result"].(map[string]interface{})["list"].([]interface{})
for _, item := range list {
    pos := item.(map[string]interface{})
    sym := pos["symbol"].(string)
    if bybit.IsTradFiSymbol(sym) {
        fmt.Printf("%s  side=%v  size=%v  unrealisedPnl=%v\n",
            sym, pos["side"], pos["size"], pos["unrealisedPnl"])
    }
}
```

---

### GetTradFiTradeHistory

Returns execution/fill history for TradFi instruments.

```go
func (c *Client) GetTradFiTradeHistory(symbol string, limit int) (map[string]interface{}, error)
```

```go
history, err := client.GetTradFiTradeHistory("XAUUSD", 50)
history, err := client.GetTradFiTradeHistory("", 100) // all TradFi fills
```

---

### GetTradFiFeeRate

Returns the trading fee rate for TradFi instruments.

```go
func (c *Client) GetTradFiFeeRate(symbol string) (map[string]interface{}, error)
```

```go
fee, err := client.GetTradFiFeeRate("EURUSD")
fee, err := client.GetTradFiFeeRate("") // all linear fee rates
```

---

## Utilities

### IsTradFiSymbol

Package-level function that returns `true` if a symbol is likely a TradFi instrument (metal, forex pair, or index) rather than a crypto perpetual.

```go
func IsTradFiSymbol(symbol string) bool
```

**Detection logic:**

| Rule | Examples |
|---|---|
| Starts with `XAU`, `XAG`, `XPT` | `XAUUSD` → true |
| 6 alpha chars, not a known crypto prefix | `EURUSD` → true, `BTCUSD` → false |
| Starts with known index prefix (`US30`, `US100`, `US500`, `UK100`, `DE40`, `JP225`, etc.) | `US500USD` → true |

```go
bybit.IsTradFiSymbol("XAUUSD")   // true  — gold
bybit.IsTradFiSymbol("EURUSD")   // true  — forex
bybit.IsTradFiSymbol("GBPJPY")   // true  — forex minor
bybit.IsTradFiSymbol("US500USD") // true  — S&P 500 index
bybit.IsTradFiSymbol("DE40USD")  // true  — DAX 40
bybit.IsTradFiSymbol("BTCUSDT")  // false — crypto
bybit.IsTradFiSymbol("ETHUSDT")  // false — crypto
bybit.IsTradFiSymbol("SOLUSDT")  // false — crypto
```

> For reliable stock CFD detection, query `GetTradFiInstruments("stock")` instead of relying on symbol heuristics.

---

## Constants

```go
// Category
bybit.TradFiCategoryLinear  // "linear"
bybit.TradFiCategoryInverse // "inverse"

// Asset class filter strings (for GetTradFiInstruments)
bybit.TradFiAssetForex     // "forex"
bybit.TradFiAssetMetal     // "metal"
bybit.TradFiAssetStock     // "stock"
bybit.TradFiAssetIndex     // "index"
bybit.TradFiAssetCommodity // "commodity"
```

---

## Market Hours & Swap Fees

### Market Hours

Unlike cryptocurrency markets (24/7), TradFi instruments follow standard market schedules:

| Asset Class | Trading Hours (UTC) |
|---|---|
| **Forex** | Sun 22:00 — Fri 22:00 (with brief daily breaks) |
| **Gold / Silver** | Sun 23:00 — Fri 22:00 |
| **US Stock CFDs** | Mon–Fri 13:30 — 20:00 |
| **US Indices** | Mon–Fri 13:30 — 21:00 (futures hours vary) |
| **EU Indices** | Mon–Fri 07:00 — 15:30 |

Outside trading sessions the Bybit API may return:
- Empty order books (`b: []`, `a: []`)
- Stale or zero prices in tickers
- `retCode: 0` with an empty `list`

Always check for empty results before processing TradFi market data.

### Swap Fees

Swap (overnight financing) fees are charged when a TradFi position is held past the daily rollover time. This is analogous to funding rates in crypto perpetuals but uses a fixed financing schedule.

Use `GetTradFiSwapFee` to retrieve current rates before opening a position you intend to hold overnight:

```go
info, err := client.GetTradFiSwapFee("XAUUSD")
if err != nil {
    log.Fatal(err)
}
list := info["result"].(map[string]interface{})["list"].([]interface{})
if len(list) > 0 {
    instrument := list[0].(map[string]interface{})
    fmt.Println("Funding interval:", instrument["fundingInterval"])
    fmt.Println("Funding rate:", instrument["fundingRate"])
}
```

---

## Full Example

The complete runnable example is available in `examples/tradfi.go`. To run it:

```bash
export BYBIT_API_KEY="your_key"
export BYBIT_API_SECRET="your_secret"

go run examples/tradfi.go
```

**What the example covers:**
1. Gold (XAUUSD) ticker
2. Forex (EURUSD) ticker
3. S&P 500 index (US500USD) ticker
4. XAUUSD kline (1h, 5 candles)
5. EURUSD order book (depth 5)
6. Open TradFi positions filtered with `IsTradFiSymbol`
7. Placing a limit buy on XAUUSD (far-from-market, demo-trading safe)
8. `IsTradFiSymbol` detection across various symbols

---

## See Also

- [Getting Started](Getting-Started) — client initialization and authentication
- [Market Data](Market-Data) — general market data methods
- [Order Management](Order-Management) — general order placement and management
- [Position Management](Position-Management) — leverage, margin, TP/SL
- [Bybit TradFi product page](https://www.bybit.com/future-activity/en/tradfi) — official Bybit TradFi information
- [Bybit V5 API — Instruments Info](https://bybit-exchange.github.io/docs/v5/market/instrument) — full instrument spec
