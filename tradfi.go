package bybit

import (
	"fmt"
	"strings"
)

// TradFi asset classes available on Bybit
const (
	TradFiCategoryLinear  = "linear"
	TradFiCategoryInverse = "inverse"

	// TradFi asset class prefixes used in Bybit symbol naming
	TradFiAssetForex    = "forex"
	TradFiAssetMetal    = "metal"
	TradFiAssetStock    = "stock"
	TradFiAssetIndex    = "index"
	TradFiAssetCommodity = "commodity"
)

// Well-known TradFi symbols traded on Bybit (linear perpetuals)
var (
	// Metals
	TradFiMetals = []string{
		"XAUUSD", // Gold
		"XAGUSD", // Silver
		"XPTUSD", // Platinum
	}

	// Forex majors
	TradFiForexMajors = []string{
		"EURUSD",
		"GBPUSD",
		"USDJPY",
		"USDCHF",
		"AUDUSD",
		"NZDUSD",
		"USDCAD",
	}

	// Forex minors
	TradFiForexMinors = []string{
		"EURGBP",
		"EURJPY",
		"GBPJPY",
		"EURCHF",
		"AUDCAD",
		"AUDNZD",
		"CADJPY",
	}

	// US Stock CFDs
	TradFiUSStocks = []string{
		"AAPLUSDT",  // Apple
		"AMZNUSDT",  // Amazon
		"TSLAUSDT",  // Tesla
		"GOOGLSDT",  // Google
		"MSFTUSDT",  // Microsoft
		"METAUSDT",  // Meta
		"NVDAUSDT",  // NVIDIA
		"NFLXUSDT",  // Netflix
	}

	// Major indices
	TradFiIndices = []string{
		"US30USD",  // Dow Jones
		"US100USD", // NASDAQ 100
		"US500USD", // S&P 500
		"UK100USD", // FTSE 100
		"DE40USD",  // DAX 40
		"JP225USD", // Nikkei 225
	}
)

// GetTradFiInstruments returns all available TradFi instruments.
// Set assetClass to filter by "forex", "metal", "stock", "index", or "" for all.
func (c *Client) GetTradFiInstruments(assetClass string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"category": TradFiCategoryLinear,
	}
	result, err := c.Request("GET", "/v5/market/instruments-info", params)
	if err != nil {
		return nil, err
	}

	if assetClass == "" {
		return result, nil
	}

	// Filter by asset class based on symbol patterns
	filtered := filterInstrumentsByAssetClass(result, assetClass)
	return filtered, nil
}

// GetTradFiTickers returns ticker data for the given TradFi symbols.
// Pass nil or empty slice to get tickers for all linear instruments.
func (c *Client) GetTradFiTickers(symbols []string) (map[string]interface{}, error) {
	if len(symbols) == 1 {
		return c.Request("GET", "/v5/market/tickers", map[string]interface{}{
			"category": TradFiCategoryLinear,
			"symbol":   symbols[0],
		})
	}
	return c.Request("GET", "/v5/market/tickers", map[string]interface{}{
		"category": TradFiCategoryLinear,
	})
}

// GetMetalsTickers returns ticker data for gold, silver, and platinum.
func (c *Client) GetMetalsTickers() (map[string]interface{}, error) {
	return c.GetTradFiTickers(TradFiMetals)
}

// GetForexTickers returns ticker data for major forex pairs.
func (c *Client) GetForexTickers() (map[string]interface{}, error) {
	return c.GetTradFiTickers(TradFiForexMajors)
}

// GetStockTickers returns ticker data for US stock CFDs.
func (c *Client) GetStockTickers() (map[string]interface{}, error) {
	return c.GetTradFiTickers(TradFiUSStocks)
}

// GetIndexTickers returns ticker data for major indices.
func (c *Client) GetIndexTickers() (map[string]interface{}, error) {
	return c.GetTradFiTickers(TradFiIndices)
}

// GetTradFiTicker returns ticker data for a single TradFi symbol.
func (c *Client) GetTradFiTicker(symbol string) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/tickers", map[string]interface{}{
		"category": TradFiCategoryLinear,
		"symbol":   symbol,
	})
}

// GetTradFiKline returns kline/candlestick data for a TradFi symbol.
// interval: 1, 3, 5, 15, 30, 60, 120, 240, 360, 720, D, W, M
func (c *Client) GetTradFiKline(symbol, interval string, limit int) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"category": TradFiCategoryLinear,
		"symbol":   symbol,
		"interval": interval,
	}
	if limit > 0 {
		params["limit"] = limit
	}
	return c.Request("GET", "/v5/market/kline", params)
}

// GetTradFiOrderbook returns order book depth for a TradFi symbol.
// depth: 1, 25, 50, 100, 200
func (c *Client) GetTradFiOrderbook(symbol string, depth int) (map[string]interface{}, error) {
	if depth == 0 {
		depth = 25
	}
	return c.Request("GET", "/v5/market/orderbook", map[string]interface{}{
		"category": TradFiCategoryLinear,
		"symbol":   symbol,
		"limit":    depth,
	})
}

// GetTradFiSwapFee returns the swap (overnight financing) fee info for a TradFi symbol.
// Swap fees apply when holding TradFi positions past market close.
func (c *Client) GetTradFiSwapFee(symbol string) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/instruments-info", map[string]interface{}{
		"category": TradFiCategoryLinear,
		"symbol":   symbol,
	})
}

// GetTradFiPositions returns open TradFi positions for the account.
// Pass symbol="" to get all TradFi positions.
func (c *Client) GetTradFiPositions(symbol string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"category": TradFiCategoryLinear,
	}
	if symbol != "" {
		params["symbol"] = symbol
	}
	return c.Request("GET", "/v5/position/list", params)
}

// TradFiOrderParams holds parameters for placing a TradFi order.
type TradFiOrderParams struct {
	Symbol      string
	Side        string  // "Buy" or "Sell"
	OrderType   string  // "Market" or "Limit"
	Qty         string  // quantity as string
	Price       string  // required for Limit orders
	TimeInForce string  // "GTC", "IOC", "FOK", "PostOnly"
	TakeProfit  string  // optional TP price
	StopLoss    string  // optional SL price
	PositionIdx int     // 0=one-way, 1=hedge-long, 2=hedge-short
	ReduceOnly  bool
	OrderLinkID string
}

// PlaceTradFiOrder places an order for a TradFi instrument (forex, metals, stocks, indices).
func (c *Client) PlaceTradFiOrder(p TradFiOrderParams) (map[string]interface{}, error) {
	if p.TimeInForce == "" {
		p.TimeInForce = "GTC"
	}

	payload := map[string]interface{}{
		"category":    TradFiCategoryLinear,
		"symbol":      p.Symbol,
		"side":        p.Side,
		"orderType":   p.OrderType,
		"qty":         p.Qty,
		"timeInForce": p.TimeInForce,
		"positionIdx": p.PositionIdx,
	}

	if p.OrderType == "Limit" && p.Price != "" {
		payload["price"] = p.Price
	}
	if p.TakeProfit != "" {
		payload["takeProfit"] = p.TakeProfit
	}
	if p.StopLoss != "" {
		payload["stopLoss"] = p.StopLoss
	}
	if p.ReduceOnly {
		payload["reduceOnly"] = true
	}
	if p.OrderLinkID != "" {
		payload["orderLinkId"] = p.OrderLinkID
	}

	return c.Request("POST", "/v5/order/create", payload)
}

// CloseTradFiPosition closes an open TradFi position at market price.
func (c *Client) CloseTradFiPosition(symbol, side string, qty string, positionIdx int) (map[string]interface{}, error) {
	closeSide := "Sell"
	if strings.ToUpper(side) == "SELL" || strings.ToUpper(side) == "SHORT" {
		closeSide = "Buy"
	}

	return c.Request("POST", "/v5/order/create", map[string]interface{}{
		"category":    TradFiCategoryLinear,
		"symbol":      symbol,
		"side":        closeSide,
		"orderType":   "Market",
		"qty":         qty,
		"timeInForce": "IOC",
		"positionIdx": positionIdx,
		"reduceOnly":  true,
	})
}

// SetTradFiLeverage sets leverage for a TradFi symbol.
// TradFi instruments typically support 1x–20x leverage depending on the instrument.
func (c *Client) SetTradFiLeverage(symbol string, leverage float64) (map[string]interface{}, error) {
	leverageStr := fmt.Sprintf("%.2f", leverage)
	return c.Request("POST", "/v5/position/set-leverage", map[string]interface{}{
		"category":     TradFiCategoryLinear,
		"symbol":       symbol,
		"buyLeverage":  leverageStr,
		"sellLeverage": leverageStr,
	})
}

// GetTradFiTradeHistory returns execution/trade history for TradFi symbols.
func (c *Client) GetTradFiTradeHistory(symbol string, limit int) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"category": TradFiCategoryLinear,
	}
	if symbol != "" {
		params["symbol"] = symbol
	}
	if limit > 0 {
		params["limit"] = limit
	}
	return c.Request("GET", "/v5/execution/list", params)
}

// GetTradFiOpenOrders returns open orders for TradFi instruments.
func (c *Client) GetTradFiOpenOrders(symbol string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"category": TradFiCategoryLinear,
	}
	if symbol != "" {
		params["symbol"] = symbol
	}
	return c.Request("GET", "/v5/order/realtime", params)
}

// CancelTradFiOrder cancels a specific TradFi order by orderId or orderLinkId.
func (c *Client) CancelTradFiOrder(symbol, orderID, orderLinkID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"category": TradFiCategoryLinear,
		"symbol":   symbol,
	}
	if orderID != "" {
		payload["orderId"] = orderID
	}
	if orderLinkID != "" {
		payload["orderLinkId"] = orderLinkID
	}
	return c.Request("POST", "/v5/order/cancel", payload)
}

// GetTradFiFeeRate returns the trading fee rate for TradFi instruments.
func (c *Client) GetTradFiFeeRate(symbol string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"category": TradFiCategoryLinear,
	}
	if symbol != "" {
		params["symbol"] = symbol
	}
	return c.Request("GET", "/v5/account/fee-rate", params)
}

// IsTradFiSymbol returns true if the given symbol is likely a TradFi instrument
// (forex, metals, stock CFD, or index) rather than a crypto perpetual.
func IsTradFiSymbol(symbol string) bool {
	s := strings.ToUpper(symbol)

	// Metals
	if strings.HasPrefix(s, "XAU") || strings.HasPrefix(s, "XAG") || strings.HasPrefix(s, "XPT") {
		return true
	}

	// Forex: 6-char all-alpha, no "BTC"/"ETH"/"SOL"/"BNB" etc.
	cryptoPrefixes := []string{"BTC", "ETH", "SOL", "BNB", "XRP", "ADA", "DOT", "MATIC", "AVAX", "LINK", "LTC", "DOGE", "SHIB"}
	if len(s) == 6 && isAlpha(s) {
		for _, prefix := range cryptoPrefixes {
			if strings.HasPrefix(s, prefix) {
				return false
			}
		}
		return true
	}

	// Indices: US30, US100, US500, UK100, DE40, JP225, etc.
	indexPrefixes := []string{"US30", "US100", "US500", "UK100", "DE40", "JP225", "AU200", "EU50", "HK50", "CN50"}
	for _, idx := range indexPrefixes {
		if strings.HasPrefix(s, idx) {
			return true
		}
	}

	// Stock CFDs typically end in "USDT" and have non-crypto-like names
	// (simple heuristic — real detection requires instruments-info lookup)
	return false
}

// isAlpha checks if all characters in s are ASCII letters.
func isAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// filterInstrumentsByAssetClass filters the instruments-info response by asset class keyword.
func filterInstrumentsByAssetClass(response map[string]interface{}, assetClass string) map[string]interface{} {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return response
	}
	list, ok := result["list"].([]interface{})
	if !ok {
		return response
	}

	assetClass = strings.ToLower(assetClass)
	var filtered []interface{}

	for _, item := range list {
		instrument, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		symbol, _ := instrument["symbol"].(string)
		symbol = strings.ToUpper(symbol)

		match := false
		switch assetClass {
		case TradFiAssetMetal:
			match = strings.HasPrefix(symbol, "XAU") ||
				strings.HasPrefix(symbol, "XAG") ||
				strings.HasPrefix(symbol, "XPT")
		case TradFiAssetForex:
			match = len(symbol) == 6 && isAlpha(symbol) && IsTradFiSymbol(symbol)
		case TradFiAssetIndex:
			for _, idx := range TradFiIndices {
				if strings.HasPrefix(symbol, idx[:4]) {
					match = true
					break
				}
			}
		case TradFiAssetStock:
			match = strings.HasSuffix(symbol, "USDT") && !IsTradFiSymbol(symbol[:len(symbol)-4]) && len(symbol) > 8
		case TradFiAssetCommodity:
			commodityPrefixes := []string{"OIL", "NATGAS", "COPPER", "WHEAT", "CORN"}
			for _, prefix := range commodityPrefixes {
				if strings.HasPrefix(symbol, prefix) {
					match = true
					break
				}
			}
		}

		if match {
			filtered = append(filtered, item)
		}
	}

	if filtered == nil {
		filtered = []interface{}{}
	}

	filteredResult := map[string]interface{}{
		"category":       result["category"],
		"list":           filtered,
		"nextPageCursor": result["nextPageCursor"],
	}

	return map[string]interface{}{
		"retCode": response["retCode"],
		"retMsg":  response["retMsg"],
		"result":  filteredResult,
		"time":    response["time"],
	}
}
