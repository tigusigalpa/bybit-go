package bybit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const (
	DemoBaseURL      = "https://api-demo.bybit.com"
	DemoWebSocketURL = "wss://stream-demo.bybit.com"
)

type DemoClient struct {
	*Client
}

func NewDemoClient(config ClientConfig) (*DemoClient, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	return &DemoClient{Client: client}, nil
}

func (dc *DemoClient) BaseURI() string {
	return DemoBaseURL
}

func (dc *DemoClient) WebSocketURL() string {
	return DemoWebSocketURL
}

func (dc *DemoClient) Request(method, path string, params map[string]interface{}) (map[string]interface{}, error) {
	method = strings.ToUpper(method)
	fullURL := dc.BaseURI() + path

	var req *http.Request
	var err error

	if method == "GET" {
		if len(params) > 0 {
			fullURL += "?" + dc.buildQuery(params)
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

	headers, err := dc.headers(method, path, params)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := dc.httpClient.Do(req)
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

func (dc *DemoClient) GetWalletBalance(params map[string]interface{}) (map[string]interface{}, error) {
	if params == nil {
		params = map[string]interface{}{}
	}
	if _, ok := params["accountType"]; !ok {
		params["accountType"] = "UNIFIED"
	}
	return dc.Request("GET", "/v5/account/wallet-balance", params)
}

func (dc *DemoClient) CreateOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/create", params)
}

func (dc *DemoClient) AmendOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/amend", params)
}

func (dc *DemoClient) CancelOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/cancel", params)
}

func (dc *DemoClient) CancelAllOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/cancel-all", params)
}

func (dc *DemoClient) GetOpenOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/order/realtime", params)
}

func (dc *DemoClient) GetOrderHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/order/history", params)
}

func (dc *DemoClient) GetTradeHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/execution/list", params)
}

func (dc *DemoClient) BatchPlaceOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/create-batch", params)
}

func (dc *DemoClient) BatchAmendOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/amend-batch", params)
}

func (dc *DemoClient) BatchCancelOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/cancel-batch", params)
}

func (dc *DemoClient) GetPositions(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/position/list", params)
}

func (dc *DemoClient) SetLeverage(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/set-leverage", params)
}

func (dc *DemoClient) SwitchPositionMode(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/switch-mode", params)
}

func (dc *DemoClient) SetTradingStop(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/trading-stop", params)
}

func (dc *DemoClient) SetAutoAddMargin(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/set-auto-add-margin", params)
}

func (dc *DemoClient) AddOrReduceMargin(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/add-margin", params)
}

func (dc *DemoClient) GetClosedPnL(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/position/closed-pnl", params)
}

func (dc *DemoClient) GetBorrowHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/borrow-history", params)
}

func (dc *DemoClient) SetCollateralCoin(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/account/set-collateral-switch", params)
}

func (dc *DemoClient) GetCollateralInfo(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/collateral-info", params)
}

func (dc *DemoClient) GetCoinGreeks(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/asset/coin-greeks", params)
}

func (dc *DemoClient) GetAccountInfo() (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/info", nil)
}

func (dc *DemoClient) GetTransactionLog(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/transaction-log", params)
}

func (dc *DemoClient) SetMarginMode(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/account/set-margin-mode", params)
}

func (dc *DemoClient) SetSpotHedging(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/account/set-hedging-mode", params)
}

func (dc *DemoClient) GetDeliveryRecord(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/asset/delivery-record", params)
}

func (dc *DemoClient) GetUSDCSettlement(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/asset/settlement-record", params)
}

func (dc *DemoClient) ToggleMarginTrade(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/spot-margin-trade/switch-mode", params)
}

func (dc *DemoClient) SetSpotMarginLeverage(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/spot-margin-trade/set-leverage", params)
}

func (dc *DemoClient) GetSpotMarginStatus() (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/spot-margin-trade/state", nil)
}

type DemoFundRequest struct {
	Coin      string `json:"coin"`
	AmountStr string `json:"amountStr"`
}

func (dc *DemoClient) ApplyForDemoFunds(adjustType int, funds []DemoFundRequest) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"adjustType":        adjustType,
		"utaDemoApplyMoney": funds,
	}
	return dc.Request("POST", "/v5/account/demo-apply-money", params)
}

func (dc *DemoClient) ApplyForDemoFundsSimple(coin string, amount string) (map[string]interface{}, error) {
	return dc.ApplyForDemoFunds(0, []DemoFundRequest{
		{Coin: coin, AmountStr: amount},
	})
}

func (dc *DemoClient) CreateDemoAccount(mainnetClient *Client) (map[string]interface{}, error) {
	return mainnetClient.Request("POST", "/v5/user/create-demo-member", map[string]interface{}{})
}

func (dc *DemoClient) CreateDemoAPIKey(mainnetClient *Client, demoUID string, params map[string]interface{}) (map[string]interface{}, error) {
	if params == nil {
		params = map[string]interface{}{}
	}
	params["subuid"] = demoUID
	return mainnetClient.Request("POST", "/v5/user/create-sub-api", params)
}

func (dc *DemoClient) UpdateDemoAPIKey(mainnetClient *Client, params map[string]interface{}) (map[string]interface{}, error) {
	return mainnetClient.Request("POST", "/v5/user/update-sub-api", params)
}

func (dc *DemoClient) GetAPIKeyInfo() (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/user/query-api", nil)
}

func (dc *DemoClient) DeleteDemoAPIKey(mainnetClient *Client, params map[string]interface{}) (map[string]interface{}, error) {
	return mainnetClient.Request("POST", "/v5/user/delete-sub-api", params)
}
