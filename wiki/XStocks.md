# xStocks

xStocks are Bybit V5 **Spot** instruments. They are not TradFi CFDs and they are not RWA Earn subscription or redemption products. This package discovers them from the live instrument catalogue and identifies them exclusively through `symbolType: "xstocks"`.

## Discover instruments

```go
instruments, err := client.GetXStocks()
if err != nil {
	log.Fatal(err)
}

for _, instrument := range instruments {
	fmt.Printf("%s: multiplier=%s, tick=%s\n",
		instrument.Symbol,
		instrument.XStockMultiplier,
		instrument.PriceFilter.TickSize,
	)
}
```

Do not hard-code a list of symbols or derive xStocks from token names. Instrument availability, multipliers, status, and order limits can change. `GetXStock(symbol)` performs the same classification for one symbol, while `GetXStockAccountInstrument(symbol)` returns account-specific metadata.

## Decimal-safe conversions

All exchange decimals are strings. The conversion helpers use arbitrary-precision rational arithmetic and never `float64`.

```go
// Underlying shares -> exchange token quantity, rounded down to base precision.
tokenQty, err := bybit.ToXStockTokenQuantity("1", instrument.XStockMultiplier, instrument.LotSizeFilter.BasePrecision)

// Underlying limit price -> token price. Buy rounds down; Sell rounds up.
tokenPrice, err := bybit.ToXStockTokenPrice("210.50", instrument.XStockMultiplier, instrument.PriceFilter.TickSize, "Buy")
```

The conversion relationship is:

```text
stock_price = token_price / xstockMultiplier
stock_qty   = token_qty × xstockMultiplier
```

Use the multiplier returned with the instrument and preserve it with your order/execution records; do not retrospectively recalculate historical fills after a multiplier change.

## Submit an order

`PlaceXStockOrder` fetches fresh metadata, confirms that the symbol is an xStock in `Trading` state, validates the raw token values, then sends a normal V5 Spot order with `isLeverage: 0`.

```go
order, err := client.PlaceXStockOrder(bybit.XStockOrderParams{
	Symbol:      "AAPLXUSDT",
	Side:        "Buy",
	OrderType:   "Limit",
	Qty:         tokenQty,
	Price:       tokenPrice,
	TimeInForce: "GTC",
	OrderLinkID: "xs-aaplx-000001",
})
```

For a market order, specify `MarketUnit` explicitly: `baseCoin` when `Qty` is an xStock token quantity, or `quoteCoin` when it is the quote amount to spend.

```go
_, err := client.PlaceXStockOrder(bybit.XStockOrderParams{
	Symbol: "AAPLXUSDT", Side: "Buy", OrderType: "Market",
	Qty: "100", MarketUnit: "quoteCoin",
})
```

The exchange remains authoritative. An accepted create-order response is not a fill confirmation; reconcile through Spot order/execution REST endpoints and private WebSocket events.

## WebSocket

The ordinary public Spot helpers work for xStocks:

```go
ws.SubscribeTicker("AAPLXUSDT")
ws.SubscribeOrderbook("AAPLXUSDT", 50)
ws.SubscribeTrade("AAPLXUSDT")
ws.SubscribeKline("AAPLXUSDT", "1")
```

For a private Spot connection, use `SubscribeXStockOrder()` and `SubscribeXStockExecution()` for `order.spot` and `execution.spot`. Call `SubscribeWallet()` as well when balance updates are needed, and seed balances using REST because the wallet stream is incremental.

## Further reading

- [Bybit xStocks TradFi integration](https://bybit-exchange.github.io/docs/v5/tradfi-integration#instruments-info-for-xstock-stock--commodities)
- [Bybit V5 Get Instruments Info](https://bybit-exchange.github.io/docs/v5/market/instrument)
- [Bybit V5 Create Order](https://bybit-exchange.github.io/docs/v5/order/create-order)
