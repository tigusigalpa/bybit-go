package bybit

import (
	"fmt"
	"math/big"
	"strings"
)

// XStockCategory is the V5 category used by xStocks. xStocks are Spot instruments.
const XStockCategory = "spot"

// XStockSymbolType identifies xStocks in the instruments-info response.
const XStockSymbolType = "xstocks"

// XStockInstrument is the string-preserving subset of an xStocks Spot instrument.
// Decimal exchange fields intentionally remain strings to avoid binary floating-point loss.
type XStockInstrument struct {
	Symbol           string
	BaseCoin         string
	QuoteCoin        string
	Status           string
	SymbolType       string
	XStockMultiplier string
	MarginTrading    string
	PriceFilter      XStockPriceFilter
	LotSizeFilter    XStockLotSizeFilter
}

// XStockPriceFilter contains the allowed price increment.
type XStockPriceFilter struct{ TickSize string }

// XStockLotSizeFilter contains xStocks quantity and notional constraints.
type XStockLotSizeFilter struct {
	BasePrecision     string
	QuotePrecision    string
	MinOrderQty       string
	MaxOrderQty       string
	MinOrderAmt       string
	MaxOrderAmt       string
	MaxLimitOrderQty  string
	MaxMarketOrderQty string
}

// XStockOrderParams describes a raw-token xStocks Spot order. Qty and Price are
// exchange values, not underlying share values.
type XStockOrderParams struct {
	Symbol      string
	Side        string // Buy or Sell
	OrderType   string // Limit or Market
	Qty         string
	Price       string // required for Limit orders
	TimeInForce string
	MarketUnit  string // baseCoin or quoteCoin; required for Market orders
	OrderLinkID string
}

// GetXStocks dynamically discovers the currently listed xStocks instruments.
func (c *Client) GetXStocks() ([]XStockInstrument, error) {
	response, err := c.Request("GET", "/v5/market/instruments-info", map[string]interface{}{
		"category": XStockCategory, "symbolType": XStockSymbolType,
	})
	if err != nil {
		return nil, err
	}
	return parseXStockList(response)
}

// GetXStock retrieves one instrument and rejects symbols that are not xStocks.
func (c *Client) GetXStock(symbol string) (*XStockInstrument, error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, fmt.Errorf("xstock symbol is required")
	}
	response, err := c.Request("GET", "/v5/market/instruments-info", map[string]interface{}{
		"category": XStockCategory, "symbol": symbol,
	})
	if err != nil {
		return nil, err
	}
	items, err := parseXStockList(response)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Symbol == symbol {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("%q is not a listed xstock", symbol)
}

// GetXStockAccountInstrument returns account-specific xStocks metadata.
func (c *Client) GetXStockAccountInstrument(symbol string) (*XStockInstrument, error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, fmt.Errorf("xstock symbol is required")
	}
	response, err := c.Request("GET", "/v5/account/instruments-info", map[string]interface{}{"category": XStockCategory, "symbol": symbol})
	if err != nil {
		return nil, err
	}
	items, err := parseXStockList(response)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, fmt.Errorf("%q is not available as an xstock for this account", symbol)
	}
	return &items[0], nil
}

// GetXStockTicker returns Spot ticker data for an xStocks symbol.
func (c *Client) GetXStockTicker(symbol string) (map[string]interface{}, error) {
	return c.GetTickers(map[string]interface{}{"category": XStockCategory, "symbol": symbol})
}

// GetXStockOrderbook returns Spot order-book data for an xStocks symbol.
func (c *Client) GetXStockOrderbook(symbol string, limit int) (map[string]interface{}, error) {
	p := map[string]interface{}{"category": XStockCategory, "symbol": symbol}
	if limit > 0 {
		p["limit"] = limit
	}
	return c.GetOrderbook(p)
}

// GetXStockKline returns Spot kline data for an xStocks symbol.
func (c *Client) GetXStockKline(symbol, interval string, limit int) (map[string]interface{}, error) {
	p := map[string]interface{}{"category": XStockCategory, "symbol": symbol, "interval": interval}
	if limit > 0 {
		p["limit"] = limit
	}
	return c.GetKline(p)
}

// GetXStockRecentTrades returns public Spot trades for an xStocks symbol.
func (c *Client) GetXStockRecentTrades(symbol string) (map[string]interface{}, error) {
	return c.GetRecentTrades(map[string]interface{}{"category": XStockCategory, "symbol": symbol})
}

// ValidateXStockOrder validates a raw-token order against current instrument rules.
func ValidateXStockOrder(instrument XStockInstrument, order XStockOrderParams) error {
	if instrument.SymbolType != XStockSymbolType {
		return fmt.Errorf("%q is not an xstock instrument", instrument.Symbol)
	}
	if instrument.Status != "Trading" {
		return fmt.Errorf("xstock %q is not tradable (status %s)", instrument.Symbol, instrument.Status)
	}
	if order.Symbol != instrument.Symbol {
		return fmt.Errorf("order symbol %q does not match instrument %q", order.Symbol, instrument.Symbol)
	}
	if order.Side != "Buy" && order.Side != "Sell" {
		return fmt.Errorf("xstock side must be Buy or Sell")
	}
	if order.OrderType != "Limit" && order.OrderType != "Market" {
		return fmt.Errorf("xstock order type must be Limit or Market")
	}
	if _, err := positiveXStockDecimal(order.Qty); err != nil {
		return fmt.Errorf("xstock qty: %w", err)
	}
	if instrument.LotSizeFilter.BasePrecision != "" && !xStockMultiple(order.Qty, instrument.LotSizeFilter.BasePrecision) {
		return fmt.Errorf("xstock qty %s is not a multiple of basePrecision %s", order.Qty, instrument.LotSizeFilter.BasePrecision)
	}
	if instrument.LotSizeFilter.MinOrderQty != "" {
		comparison, err := xStockCompare(order.Qty, instrument.LotSizeFilter.MinOrderQty)
		if err != nil {
			return fmt.Errorf("xstock minOrderQty: %w", err)
		}
		if comparison < 0 {
			return fmt.Errorf("xstock qty %s is below minimum %s", order.Qty, instrument.LotSizeFilter.MinOrderQty)
		}
	}
	if instrument.LotSizeFilter.MaxOrderQty != "" {
		comparison, err := xStockCompare(order.Qty, instrument.LotSizeFilter.MaxOrderQty)
		if err != nil {
			return fmt.Errorf("xstock maxOrderQty: %w", err)
		}
		if comparison > 0 {
			return fmt.Errorf("xstock qty %s exceeds maximum %s", order.Qty, instrument.LotSizeFilter.MaxOrderQty)
		}
	}
	if order.OrderType == "Market" && order.MarketUnit == "" {
		return fmt.Errorf("xstock market orders require an explicit marketUnit")
	}
	if order.MarketUnit != "" && order.MarketUnit != "baseCoin" && order.MarketUnit != "quoteCoin" {
		return fmt.Errorf("xstock marketUnit must be baseCoin or quoteCoin")
	}
	if order.OrderType == "Limit" {
		if _, err := positiveXStockDecimal(order.Price); err != nil {
			return fmt.Errorf("xstock limit price: %w", err)
		}
		if instrument.PriceFilter.TickSize != "" && !xStockMultiple(order.Price, instrument.PriceFilter.TickSize) {
			return fmt.Errorf("xstock price %s is not a multiple of tickSize %s", order.Price, instrument.PriceFilter.TickSize)
		}
		if err := xStockCheckMinNotional(order.Price, order.Qty, instrument.LotSizeFilter.MinOrderAmt); err != nil {
			return err
		}
	}
	max := instrument.LotSizeFilter.MaxLimitOrderQty
	if order.OrderType == "Market" {
		max = instrument.LotSizeFilter.MaxMarketOrderQty
	}
	if max != "" {
		comparison, err := xStockCompare(order.Qty, max)
		if err != nil {
			return fmt.Errorf("xstock maximum quantity: %w", err)
		}
		if comparison > 0 {
			return fmt.Errorf("xstock qty %s exceeds maximum %s", order.Qty, max)
		}
	}
	if len(order.OrderLinkID) > 36 {
		return fmt.Errorf("xstock orderLinkId must not exceed 36 characters")
	}
	return nil
}

// PlaceXStockOrder validates against freshly fetched instrument metadata and submits a Spot order.
func (c *Client) PlaceXStockOrder(order XStockOrderParams) (map[string]interface{}, error) {
	instrument, err := c.GetXStock(order.Symbol)
	if err != nil {
		return nil, err
	}
	if err := ValidateXStockOrder(*instrument, order); err != nil {
		return nil, err
	}
	if order.TimeInForce == "" {
		order.TimeInForce = "GTC"
	}
	payload := map[string]interface{}{"category": XStockCategory, "symbol": order.Symbol, "side": order.Side, "orderType": order.OrderType, "qty": order.Qty, "timeInForce": order.TimeInForce, "isLeverage": 0}
	if order.Price != "" {
		payload["price"] = order.Price
	}
	if order.MarketUnit != "" {
		payload["marketUnit"] = order.MarketUnit
	}
	if order.OrderLinkID != "" {
		payload["orderLinkId"] = order.OrderLinkID
	}
	return c.CreateOrder(payload)
}

// CancelXStockOrder cancels an xStocks Spot order by order ID or order-link ID.
func (c *Client) CancelXStockOrder(symbol, orderID, orderLinkID string) (map[string]interface{}, error) {
	if orderID == "" && orderLinkID == "" {
		return nil, fmt.Errorf("xstock orderId or orderLinkId is required")
	}
	p := map[string]interface{}{"category": XStockCategory, "symbol": symbol}
	if orderID != "" {
		p["orderId"] = orderID
	}
	if orderLinkID != "" {
		p["orderLinkId"] = orderLinkID
	}
	return c.CancelOrder(p)
}

// ToXStockTokenQuantity converts underlying-stock quantity to a rounded-down token quantity.
func ToXStockTokenQuantity(stockQty, multiplier, basePrecision string) (string, error) {
	return xStockConvertAndRound(stockQty, multiplier, basePrecision, true, false)
}

// ToXStockStockQuantity converts token quantity to an underlying-stock quantity.
func ToXStockStockQuantity(tokenQty, multiplier string) (string, error) {
	return xStockConvert(tokenQty, multiplier, false)
}

// ToXStockTokenPrice converts underlying-stock price to token price. Buy prices round down; Sell prices round up.
func ToXStockTokenPrice(stockPrice, multiplier, tickSize, side string) (string, error) {
	return xStockConvertAndRound(stockPrice, multiplier, tickSize, false, side == "Sell")
}

// ToXStockStockPrice converts token price to underlying-stock price.
func ToXStockStockPrice(tokenPrice, multiplier string) (string, error) {
	return xStockConvert(tokenPrice, multiplier, true)
}

func parseXStockList(response map[string]interface{}) ([]XStockInstrument, error) {
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("xstock instruments response has no result")
	}
	list, ok := result["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("xstock instruments response has no list")
	}
	items := make([]XStockInstrument, 0, len(list))
	for _, raw := range list {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if stringXStock(m, "symbolType") != XStockSymbolType {
			continue
		}
		items = append(items, XStockInstrument{Symbol: stringXStock(m, "symbol"), BaseCoin: stringXStock(m, "baseCoin"), QuoteCoin: stringXStock(m, "quoteCoin"), Status: stringXStock(m, "status"), SymbolType: stringXStock(m, "symbolType"), XStockMultiplier: stringXStock(m, "xstockMultiplier"), MarginTrading: stringXStock(m, "marginTrading"), PriceFilter: XStockPriceFilter{TickSize: stringXStockMap(m, "priceFilter", "tickSize")}, LotSizeFilter: XStockLotSizeFilter{BasePrecision: stringXStockMap(m, "lotSizeFilter", "basePrecision"), QuotePrecision: stringXStockMap(m, "lotSizeFilter", "quotePrecision"), MinOrderQty: stringXStockMap(m, "lotSizeFilter", "minOrderQty"), MaxOrderQty: stringXStockMap(m, "lotSizeFilter", "maxOrderQty"), MinOrderAmt: stringXStockMap(m, "lotSizeFilter", "minOrderAmt"), MaxOrderAmt: stringXStockMap(m, "lotSizeFilter", "maxOrderAmt"), MaxLimitOrderQty: stringXStockMap(m, "lotSizeFilter", "maxLimitOrderQty"), MaxMarketOrderQty: stringXStockMap(m, "lotSizeFilter", "maxMarketOrderQty")}})
	}
	return items, nil
}
func stringXStock(m map[string]interface{}, key string) string { v, _ := m[key].(string); return v }
func stringXStockMap(m map[string]interface{}, parent, key string) string {
	nested, _ := m[parent].(map[string]interface{})
	return stringXStock(nested, key)
}
func positiveXStockDecimal(v string) (*big.Rat, error) {
	r, ok := new(big.Rat).SetString(v)
	if !ok || r.Sign() <= 0 {
		return nil, fmt.Errorf("must be a positive decimal")
	}
	return r, nil
}
func xStockMultiple(value, step string) bool {
	v, e1 := positiveXStockDecimal(value)
	s, e2 := positiveXStockDecimal(step)
	if e1 != nil || e2 != nil {
		return false
	}
	q := new(big.Rat).Quo(v, s)
	return q.Denom().Cmp(big.NewInt(1)) == 0
}
func xStockCompare(a, b string) (int, error) {
	ar, err := positiveXStockDecimal(a)
	if err != nil {
		return 0, err
	}
	br, err := positiveXStockDecimal(b)
	if err != nil {
		return 0, err
	}
	return ar.Cmp(br), nil
}
func xStockCheckMinNotional(price, qty, min string) error {
	if min == "" {
		return nil
	}
	p, e := positiveXStockDecimal(price)
	if e != nil {
		return e
	}
	q, e := positiveXStockDecimal(qty)
	if e != nil {
		return e
	}
	m, e := positiveXStockDecimal(min)
	if e != nil {
		return e
	}
	if new(big.Rat).Mul(p, q).Cmp(m) < 0 {
		return fmt.Errorf("xstock order notional is below minOrderAmt %s", min)
	}
	return nil
}
func xStockConvert(value, multiplier string, divide bool) (string, error) {
	v, e := positiveXStockDecimal(value)
	if e != nil {
		return "", e
	}
	m, e := positiveXStockDecimal(multiplier)
	if e != nil {
		return "", e
	}
	if divide {
		v.Quo(v, m)
	} else {
		v.Mul(v, m)
	}
	return xStockDecimal(v)
}
func xStockConvertAndRound(value, multiplier, step string, divide, roundUp bool) (string, error) {
	v, e := positiveXStockDecimal(value)
	if e != nil {
		return "", e
	}
	m, e := positiveXStockDecimal(multiplier)
	if e != nil {
		return "", e
	}
	if divide {
		v.Quo(v, m)
	} else {
		v.Mul(v, m)
	}
	s, e := positiveXStockDecimal(step)
	if e != nil {
		return "", fmt.Errorf("precision: %w", e)
	}
	q := new(big.Rat).Quo(v, s)
	n, rem := new(big.Int), new(big.Int)
	n.QuoRem(q.Num(), q.Denom(), rem)
	if roundUp && rem.Sign() != 0 {
		n.Add(n, big.NewInt(1))
	}
	return xStockDecimal(new(big.Rat).Mul(new(big.Rat).SetInt(n), s))
}
func xStockDecimal(r *big.Rat) (string, error) {
	d := new(big.Int).Set(r.Denom())
	two, five := 0, 0
	z := big.NewInt(0)
	for new(big.Int).Mod(d, big.NewInt(2)).Cmp(z) == 0 {
		d.Div(d, big.NewInt(2))
		two++
	}
	for new(big.Int).Mod(d, big.NewInt(5)).Cmp(z) == 0 {
		d.Div(d, big.NewInt(5))
		five++
	}
	if d.Cmp(big.NewInt(1)) != 0 {
		return "", fmt.Errorf("decimal cannot be represented exactly")
	}
	scale := two
	if five > scale {
		scale = five
	}
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	n := new(big.Int).Mul(r.Num(), factor)
	n.Quo(n, r.Denom())
	sign := ""
	if n.Sign() < 0 {
		sign = "-"
		n.Abs(n)
	}
	digits := n.String()
	if scale == 0 {
		return sign + digits, nil
	}
	for len(digits) <= scale {
		digits = "0" + digits
	}
	out := sign + digits[:len(digits)-scale] + "." + digits[len(digits)-scale:]
	out = strings.TrimRight(out, "0")
	return strings.TrimRight(out, "."), nil
}
