package bybit

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Client provides signed REST access to the Bybit V5 API.
type Client struct {
	apiKey        string
	apiSecret     string
	demo          bool
	region        string
	recvWindow    int
	signature     string
	rsaPrivateKey *rsa.PrivateKey
	httpClient    *http.Client
	fees          map[string]map[string]map[string]float64
}

// ClientConfig configures a Client instance.
type ClientConfig struct {
	APIKey        string
	APISecret     string
	Demo          bool
	Region        string
	RecvWindow    int
	Signature     string
	RSAPrivateKey string
	HTTPClient    *http.Client
}

// HTTPError describes a response that could not be completed at the HTTP layer.
// A successful Bybit HTTP response can still contain a non-zero retCode; callers
// receive that API response unchanged and can inspect it according to their flow.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("bybit API request failed: %s", e.Status)
	}
	return fmt.Sprintf("bybit API request failed: %s: %s", e.Status, e.Body)
}

// NewClient creates a REST client with the supplied configuration.
func NewClient(config ClientConfig) (*Client, error) {
	if config.Region == "" {
		config.Region = "global"
	}
	if config.RecvWindow == 0 {
		config.RecvWindow = 5000
	}
	config.Signature = strings.ToLower(strings.TrimSpace(config.Signature))
	if config.Signature == "" {
		config.Signature = "hmac"
	}
	if config.Signature != "hmac" && config.Signature != "rsa" {
		return nil, fmt.Errorf("unsupported signature type %q: use hmac or rsa", config.Signature)
	}
	if config.Signature == "rsa" && config.RSAPrivateKey == "" {
		return nil, fmt.Errorf("RSA signature requires RSAPrivateKey")
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	client := &Client{
		apiKey:     config.APIKey,
		apiSecret:  config.APISecret,
		demo:       config.Demo,
		region:     config.Region,
		recvWindow: config.RecvWindow,
		signature:  config.Signature,
		httpClient: config.HTTPClient,
		fees:       defaultFees(),
	}

	if config.Signature == "rsa" {
		key, err := parseRSAPrivateKey(config.RSAPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
		}
		client.rsaPrivateKey = key
	}

	return client, nil
}

func parseRSAPrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaKey, nil
}

func defaultFees() map[string]map[string]map[string]float64 {
	return map[string]map[string]map[string]float64{
		"spot": {
			"Non-VIP":     {"maker": 0.0010, "taker": 0.0010},
			"VIP1":        {"maker": 0.000675, "taker": 0.0010},
			"VIP2":        {"maker": 0.000650, "taker": 0.000775},
			"VIP3":        {"maker": 0.000625, "taker": 0.000750},
			"VIP4":        {"maker": 0.000500, "taker": 0.000600},
			"VIP5":        {"maker": 0.000400, "taker": 0.000500},
			"Supreme VIP": {"maker": 0.000300, "taker": 0.000450},
		},
		"derivatives": {
			"Non-VIP": {"maker": 0.000400, "taker": 0.001000},
		},
	}
}

// BaseURI returns the REST API base URL selected by the client's demo and region settings.
func (c *Client) BaseURI() string {
	if c.demo {
		return "https://api-demo.bybit.com"
	}

	switch strings.ToLower(c.region) {
	case "demo":
		return "https://api-demo.bybit.com"
	case "nl":
		return "https://api.bybit.nl"
	case "tr":
		return "https://api.bybit-tr.com"
	case "kz":
		return "https://api.bybit.kz"
	case "ge":
		return "https://api.bybitgeorgia.ge"
	case "ae":
		return "https://api.bybit.ae"
	default:
		return "https://api.bybit.com"
	}
}

func (c *Client) timestamp() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

func (c *Client) signString(data string) (string, error) {
	if c.signature == "rsa" && c.rsaPrivateKey != nil {
		hash := sha256.Sum256([]byte(data))
		signature, err := rsa.SignPKCS1v15(nil, c.rsaPrivateKey, crypto.SHA256, hash[:])
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(signature), nil
	}

	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (c *Client) buildQuery(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := url.Values{}
	for _, k := range keys {
		values.Add(k, fmt.Sprintf("%v", params[k]))
	}

	return values.Encode()
}

func (c *Client) headers(method, path string, params map[string]interface{}) (map[string]string, error) {
	ts := c.timestamp()
	recv := strconv.Itoa(c.recvWindow)

	var toSign string
	if strings.ToUpper(method) == "GET" {
		query := c.buildQuery(params)
		toSign = ts + c.apiKey + recv + query
	} else {
		body := "{}"
		if len(params) > 0 {
			jsonBody, _ := json.Marshal(params)
			body = string(jsonBody)
		}
		toSign = ts + c.apiKey + recv + body
	}

	sign, err := c.signString(toSign)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"X-BAPI-API-KEY":     c.apiKey,
		"X-BAPI-TIMESTAMP":   ts,
		"X-BAPI-RECV-WINDOW": recv,
		"X-BAPI-SIGN":        sign,
		"User-Agent":         "bybit-go/1.0.0",
		"X-Referer":          "bybit-go",
	}

	if c.signature == "hmac" {
		headers["X-BAPI-SIGN-TYPE"] = "2"
	}

	if strings.ToUpper(method) != "GET" {
		headers["Content-Type"] = "application/json"
		headers["Accept"] = "application/json"
	}

	return headers, nil
}

// Request performs a signed Bybit REST request and decodes its JSON response.
func (c *Client) Request(method, path string, params map[string]interface{}) (map[string]interface{}, error) {
	method = strings.ToUpper(method)
	fullURL := c.BaseURI() + path

	var req *http.Request
	var err error

	if method == "GET" {
		if len(params) > 0 {
			fullURL += "?" + c.buildQuery(params)
		}
		req, err = http.NewRequest(method, fullURL, nil)
	} else {
		var body []byte
		if len(params) > 0 {
			body, err = json.Marshal(params)
			if err != nil {
				return nil, err
			}
		} else {
			body = []byte("{}")
		}
		req, err = http.NewRequest(method, fullURL, bytes.NewBuffer(body))
	}

	if err != nil {
		return nil, err
	}

	headers, err := c.headers(method, path, params)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: string(bodyBytes)}
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("decode Bybit response: %w", err)
	}

	return result, nil
}

// Endpoint returns the active REST API endpoint.
func (c *Client) Endpoint() string {
	return c.BaseURI()
}

// GetServerTime returns the current Bybit server time.
func (c *Client) GetServerTime() (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/time", nil)
}

// GetTickers returns ticker data for the supplied market parameters.
func (c *Client) GetTickers(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/tickers", params)
}

// GetKline returns candlestick data for the supplied market parameters.
func (c *Client) GetKline(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/kline", params)
}

// GetOrderbook returns an order book snapshot for the supplied market parameters.
func (c *Client) GetOrderbook(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/orderbook", params)
}

// GetRPIOrderbook returns an RPI order book snapshot for the supplied parameters.
func (c *Client) GetRPIOrderbook(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/rpi-orderbook", params)
}

// GetOpenInterest returns open-interest data for the supplied parameters.
func (c *Client) GetOpenInterest(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/open-interest", params)
}

// GetRecentTrades returns recent public trades for the supplied parameters.
func (c *Client) GetRecentTrades(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/recent-trade", params)
}

// GetFundingRateHistory returns funding-rate history for the supplied parameters.
func (c *Client) GetFundingRateHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/funding/history", params)
}

// GetHistoricalVolatility returns historical volatility data for the supplied parameters.
func (c *Client) GetHistoricalVolatility(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/historical-volatility", params)
}

// GetInsurance returns insurance-pool data for the supplied parameters.
func (c *Client) GetInsurance(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/insurance", params)
}

// GetRiskLimit returns risk-limit data for the supplied parameters.
func (c *Client) GetRiskLimit(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/risk-limit", params)
}

// CreateOrder submits an order using the supplied Bybit V5 fields.
func (c *Client) CreateOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/order/create", params)
}

// GetOpenOrders returns current open orders for the supplied parameters.
func (c *Client) GetOpenOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/order/realtime", params)
}

// CancelOrder cancels an order identified by the supplied parameters.
func (c *Client) CancelOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/order/cancel", params)
}

// AmendOrder updates an existing order using the supplied parameters.
func (c *Client) AmendOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/order/amend", params)
}

// CancelAllOrders cancels all orders matching the supplied parameters.
func (c *Client) CancelAllOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/order/cancel-all", params)
}

// GetHistoryOrders returns historical orders for the supplied parameters.
func (c *Client) GetHistoryOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/order/history", params)
}

// GetWalletBalance returns wallet balances for the supplied account parameters.
func (c *Client) GetWalletBalance(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/wallet-balance", params)
}

// GetTransferableAmount returns the transferable amount for the supplied parameters.
func (c *Client) GetTransferableAmount(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/transferable-amount", params)
}

// GetTransactionLog returns account transaction records for the supplied parameters.
func (c *Client) GetTransactionLog(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/transaction-log", params)
}

// GetAccountInfo returns metadata for the authenticated account.
func (c *Client) GetAccountInfo() (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/info", nil)
}

// GetAccountInstrumentsInfo returns account-specific instrument information.
func (c *Client) GetAccountInstrumentsInfo(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/instruments", params)
}

// GetPositions returns positions matching the supplied parameters.
func (c *Client) GetPositions(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/position/list", params)
}

// SwitchPositionMode changes the position mode using the supplied parameters.
func (c *Client) SwitchPositionMode(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/switch-mode", params)
}

// SetTradingStop configures take-profit, stop-loss, or trailing-stop settings.
func (c *Client) SetTradingStop(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/trading-stop", params)
}

// SetLeverage sets leverage for a symbol and optional Buy or Sell side.
func (c *Client) SetLeverage(category, symbol string, leverage float64, side *string) (map[string]interface{}, error) {
	if leverage <= 0 {
		return nil, fmt.Errorf("leverage must be greater than zero")
	}
	payload := map[string]interface{}{
		"category": category,
		"symbol":   symbol,
	}

	leverageStr := fmt.Sprintf("%.2f", leverage)

	if side != nil {
		switch strings.ToLower(*side) {
		case "buy":
			payload["buyLeverage"] = leverageStr
		case "sell":
			payload["sellLeverage"] = leverageStr
		default:
			return nil, fmt.Errorf("side must be Buy or Sell")
		}
	} else {
		payload["buyLeverage"] = leverageStr
		payload["sellLeverage"] = leverageStr
	}

	return c.Request("POST", "/v5/position/set-leverage", payload)
}

// SetAutoAddMargin enables or disables automatic margin additions.
func (c *Client) SetAutoAddMargin(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/set-auto-add-margin", params)
}

// AddOrReduceMargin adjusts isolated-position margin using the supplied parameters.
func (c *Client) AddOrReduceMargin(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/add-margin", params)
}

// GetClosedPnL returns closed profit-and-loss records for the supplied parameters.
func (c *Client) GetClosedPnL(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/position/closed-pnl", params)
}

// GetClosedOptionsPositions returns closed options-position records.
func (c *Client) GetClosedOptionsPositions(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/position/close-position", params)
}

// MovePosition transfers a position using the supplied parameters.
func (c *Client) MovePosition(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/move-positions", params)
}

// GetMovePositionHistory returns position-transfer history for the supplied parameters.
func (c *Client) GetMovePositionHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/position/move-position-history", params)
}

// ConfirmNewRiskLimit confirms a pending maintenance-margin requirement change.
func (c *Client) ConfirmNewRiskLimit(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/confirm-pending-mmr", params)
}

func (c *Client) lastPrice(symbol, category string) (float64, error) {
	res, err := c.GetTickers(map[string]interface{}{
		"category": category,
		"symbol":   symbol,
	})
	if err != nil {
		return 0, err
	}

	result, ok := res["result"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response format")
	}

	list, ok := result["list"].([]interface{})
	if !ok || len(list) == 0 {
		return 0, fmt.Errorf("no ticker data found")
	}

	ticker, ok := list[0].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid ticker format")
	}

	if lastPrice, ok := ticker["lastPrice"].(string); ok {
		return strconv.ParseFloat(lastPrice, 64)
	}
	if markPrice, ok := ticker["markPrice"].(string); ok {
		return strconv.ParseFloat(markPrice, 64)
	}
	if bid1Price, ok := ticker["bid1Price"].(string); ok {
		return strconv.ParseFloat(bid1Price, 64)
	}

	return 0, fmt.Errorf("no price data found")
}

func (c *Client) qtyFromMargin(margin, price, leverage float64) float64 {
	if price == 0 {
		return 0
	}
	qty := margin * leverage / price
	// Round to 3 decimal places for most symbols
	return float64(int(qty*1000)) / 1000
}

// PlaceOrderParams describes the higher-level order helper input.
type PlaceOrderParams struct {
	Type      string
	Symbol    string
	Execution string
	Price     *float64
	Side      *string
	Leverage  *float64
	Size      float64
	SlTp      *SlTpParams
	Extra     map[string]interface{}
}

// SlTpParams describes optional take-profit and stop-loss settings for PlaceOrder.
type SlTpParams struct {
	Type       string
	TakeProfit *float64
	StopLoss   *float64
}

// PlaceOrder builds and submits an order from the higher-level PlaceOrderParams structure.
func (c *Client) PlaceOrder(params PlaceOrderParams) (map[string]interface{}, error) {
	isSpot := strings.ToLower(params.Type) == "spot"
	category := "linear"
	if isSpot {
		category = "spot"
	}

	orderType := "Limit"
	if strings.ToLower(params.Execution) == "market" {
		orderType = "Market"
	}

	payload := map[string]interface{}{
		"category": category,
		"symbol":   params.Symbol,
	}

	side := "Buy"
	if params.Side != nil {
		side = *params.Side
	}

	if isSpot {
		payload["side"] = side
		payload["orderType"] = orderType
		if orderType == "Limit" && params.Price != nil {
			payload["price"] = fmt.Sprintf("%.8f", *params.Price)
		}
		payload["qty"] = fmt.Sprintf("%.8f", params.Size)
	} else {
		payload["side"] = side
		payload["orderType"] = orderType

		entryPrice := 0.0
		if orderType == "Limit" && params.Price != nil {
			entryPrice = *params.Price
		} else {
			price, err := c.lastPrice(params.Symbol, category)
			if err == nil {
				entryPrice = price
			} else if params.Price != nil {
				entryPrice = *params.Price
			}
		}

		leverage := 1.0
		if params.Leverage != nil && *params.Leverage > 0 {
			leverage = *params.Leverage
			if _, err := c.SetLeverage(category, params.Symbol, leverage, &side); err != nil {
				return nil, fmt.Errorf("set leverage: %w", err)
			}
		}

		if entryPrice < 0.0000001 {
			entryPrice = 0.0000001
		}
		qty := c.qtyFromMargin(params.Size, entryPrice, leverage)
		payload["qty"] = fmt.Sprintf("%.3f", qty)

		if orderType == "Limit" && params.Price != nil {
			payload["price"] = fmt.Sprintf("%.8f", *params.Price)
		}
		payload["positionIdx"] = 0
	}

	if strings.ToLower(params.Execution) == "trigger" {
		payload["orderType"] = "Market"
		if params.Price != nil {
			payload["triggerPrice"] = fmt.Sprintf("%.8f", *params.Price)
		}
		if side == "Buy" {
			payload["triggerDirection"] = 1
		} else {
			payload["triggerDirection"] = 2
		}
	}

	if params.SlTp != nil && !isSpot {
		mode := "absolute"
		if params.SlTp.Type != "" {
			mode = params.SlTp.Type
		}

		tp := params.SlTp.TakeProfit
		sl := params.SlTp.StopLoss

		if mode == "percent" {
			entryPrice := 0.0
			if priceStr, ok := payload["price"].(string); ok {
				entryPrice, _ = strconv.ParseFloat(priceStr, 64)
			} else {
				price, _ := c.lastPrice(params.Symbol, category)
				entryPrice = price
			}

			if tp != nil {
				if side == "Buy" {
					tpVal := entryPrice * (1 + *tp)
					tp = &tpVal
				} else {
					tpVal := entryPrice * (1 - *tp)
					tp = &tpVal
				}
			}

			if sl != nil {
				if side == "Buy" {
					slVal := entryPrice * (1 - *sl)
					sl = &slVal
				} else {
					slVal := entryPrice * (1 + *sl)
					sl = &slVal
				}
			}
		}

		if tp != nil {
			payload["takeProfit"] = fmt.Sprintf("%.8f", *tp)
		}
		if sl != nil {
			payload["stopLoss"] = fmt.Sprintf("%.8f", *sl)
		}
	}

	if params.Extra != nil {
		for k, v := range params.Extra {
			payload[k] = v
		}
	}

	return c.Request("POST", "/v5/order/create", payload)
}

// ComputeFee estimates a fee from the client's built-in fee table.
func (c *Client) ComputeFee(tradeType string, volume float64, level, liquidity string) float64 {
	typeKey := "derivatives"
	if strings.ToLower(tradeType) == "spot" {
		typeKey = "spot"
	}

	rate := 0.0
	if typeMap, ok := c.fees[typeKey]; ok {
		if levelMap, ok := typeMap[level]; ok {
			if r, ok := levelMap[liquidity]; ok {
				rate = r
			}
		} else if levelMap, ok := typeMap["Non-VIP"]; ok {
			if r, ok := levelMap[liquidity]; ok {
				rate = r
			}
		}
	}

	return volume * rate
}
