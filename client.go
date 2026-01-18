package bybit

import (
	"bytes"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
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

type Client struct {
	apiKey        string
	apiSecret     string
	testnet       bool
	region        string
	recvWindow    int
	signature     string
	rsaPrivateKey *rsa.PrivateKey
	httpClient    *http.Client
	fees          map[string]map[string]map[string]float64
}

type ClientConfig struct {
	APIKey        string
	APISecret     string
	Testnet       bool
	Region        string
	RecvWindow    int
	Signature     string
	RSAPrivateKey string
	HTTPClient    *http.Client
}

func NewClient(config ClientConfig) (*Client, error) {
	if config.Region == "" {
		config.Region = "global"
	}
	if config.RecvWindow == 0 {
		config.RecvWindow = 5000
	}
	if config.Signature == "" {
		config.Signature = "hmac"
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	client := &Client{
		apiKey:     config.APIKey,
		apiSecret:  config.APISecret,
		testnet:    config.Testnet,
		region:     config.Region,
		recvWindow: config.RecvWindow,
		signature:  config.Signature,
		httpClient: config.HTTPClient,
		fees:       defaultFees(),
	}

	if config.Signature == "rsa" && config.RSAPrivateKey != "" {
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

func (c *Client) BaseURI() string {
	if c.testnet {
		return "https://api-testnet.bybit.com"
	}

	switch strings.ToLower(c.region) {
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
		signature, err := rsa.SignPKCS1v15(nil, c.rsaPrivateKey, 0, hash[:])
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(signature), nil
	}

	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	mac.Write([]byte(data))
	return strings.ToLower(fmt.Sprintf("%x", mac.Sum(nil))), nil
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
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return map[string]interface{}{"raw": string(bodyBytes)}, nil
	}

	return result, nil
}

func (c *Client) Endpoint() string {
	return c.BaseURI()
}

func (c *Client) GetServerTime() (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/time", nil)
}

func (c *Client) GetTickers(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/tickers", params)
}

func (c *Client) GetKline(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/kline", params)
}

func (c *Client) GetOrderbook(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/orderbook", params)
}

func (c *Client) GetRPIOrderbook(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/rpi-orderbook", params)
}

func (c *Client) GetOpenInterest(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/open-interest", params)
}

func (c *Client) GetRecentTrades(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/recent-trade", params)
}

func (c *Client) GetFundingRateHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/funding/history", params)
}

func (c *Client) GetHistoricalVolatility(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/historical-volatility", params)
}

func (c *Client) GetInsurance(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/insurance", params)
}

func (c *Client) GetRiskLimit(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/market/risk-limit", params)
}

func (c *Client) CreateOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/order/create", params)
}

func (c *Client) GetOpenOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/order/realtime", params)
}

func (c *Client) CancelOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/order/cancel", params)
}

func (c *Client) AmendOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/order/amend", params)
}

func (c *Client) CancelAllOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/order/cancel-all", params)
}

func (c *Client) GetHistoryOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/order/history", params)
}

func (c *Client) GetWalletBalance(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/wallet-balance", params)
}

func (c *Client) GetTransferableAmount(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/transferable-amount", params)
}

func (c *Client) GetTransactionLog(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/transaction-log", params)
}

func (c *Client) GetAccountInfo() (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/info", nil)
}

func (c *Client) GetAccountInstrumentsInfo(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/account/instruments", params)
}

func (c *Client) GetPositions(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/position/list", params)
}

func (c *Client) SwitchPositionMode(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/switch-mode", params)
}

func (c *Client) SetTradingStop(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/trading-stop", params)
}

func (c *Client) SetLeverage(category, symbol string, leverage float64, side *string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"category": category,
		"symbol":   symbol,
	}

	leverageStr := fmt.Sprintf("%.2f", leverage)

	if side != nil {
		if *side == "Buy" {
			payload["buyLeverage"] = leverageStr
		} else if *side == "Sell" {
			payload["sellLeverage"] = leverageStr
		}
	} else {
		payload["buyLeverage"] = leverageStr
		payload["sellLeverage"] = leverageStr
	}

	return c.Request("POST", "/v5/position/set-leverage", payload)
}

func (c *Client) SetAutoAddMargin(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/set-auto-add-margin", params)
}

func (c *Client) AddOrReduceMargin(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/add-margin", params)
}

func (c *Client) GetClosedPnL(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/position/closed-pnl", params)
}

func (c *Client) GetClosedOptionsPositions(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/position/close-position", params)
}

func (c *Client) MovePosition(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("POST", "/v5/position/move-positions", params)
}

func (c *Client) GetMovePositionHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return c.Request("GET", "/v5/position/move-position-history", params)
}

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
	return margin * leverage / price
}

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

type SlTpParams struct {
	Type       string
	TakeProfit *float64
	StopLoss   *float64
}

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
			c.SetLeverage(category, params.Symbol, leverage, &side)
		}

		if entryPrice < 0.0000001 {
			entryPrice = 0.0000001
		}
		qty := c.qtyFromMargin(params.Size, entryPrice, leverage)
		payload["qty"] = fmt.Sprintf("%.8f", qty)

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
