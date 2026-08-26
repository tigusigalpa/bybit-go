package bybit

// DemoClient provides REST helpers for the Bybit demo trading environment.
type DemoClient struct {
	*Client
}

// NewDemoClient creates a client configured for the demo trading environment.
func NewDemoClient(config ClientConfig) (*DemoClient, error) {
	config.Demo = true
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	return &DemoClient{Client: client}, nil
}

// GetWalletBalance returns demo-account wallet balances and defaults accountType to UNIFIED.
func (dc *DemoClient) GetWalletBalance(params map[string]interface{}) (map[string]interface{}, error) {
	if params == nil {
		params = map[string]interface{}{}
	}
	if _, ok := params["accountType"]; !ok {
		params["accountType"] = "UNIFIED"
	}
	return dc.Request("GET", "/v5/account/wallet-balance", params)
}

// CreateOrder submits a demo order using the supplied V5 fields.
func (dc *DemoClient) CreateOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/create", params)
}

// AmendOrder updates an existing demo order.
func (dc *DemoClient) AmendOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/amend", params)
}

// CancelOrder cancels a demo order.
func (dc *DemoClient) CancelOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/cancel", params)
}

// CancelAllOrders cancels demo orders matching the supplied parameters.
func (dc *DemoClient) CancelAllOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/cancel-all", params)
}

// GetOpenOrders returns current demo orders for the supplied parameters.
func (dc *DemoClient) GetOpenOrders(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/order/realtime", params)
}

// GetOrderHistory returns historical demo orders for the supplied parameters.
func (dc *DemoClient) GetOrderHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/order/history", params)
}

// GetTradeHistory returns demo execution history for the supplied parameters.
func (dc *DemoClient) GetTradeHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/execution/list", params)
}

// BatchPlaceOrder submits a batch of demo orders.
func (dc *DemoClient) BatchPlaceOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/create-batch", params)
}

// BatchAmendOrder updates a batch of demo orders.
func (dc *DemoClient) BatchAmendOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/amend-batch", params)
}

// BatchCancelOrder cancels a batch of demo orders.
func (dc *DemoClient) BatchCancelOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/cancel-batch", params)
}

// GetPositions returns demo positions for the supplied parameters.
func (dc *DemoClient) GetPositions(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/position/list", params)
}

// SetLeverage configures demo position leverage using the supplied parameters.
func (dc *DemoClient) SetLeverage(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/set-leverage", params)
}

// SwitchPositionMode changes the demo account's position mode.
func (dc *DemoClient) SwitchPositionMode(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/switch-mode", params)
}

// SetTradingStop configures demo position stop settings.
func (dc *DemoClient) SetTradingStop(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/trading-stop", params)
}

// SetAutoAddMargin configures automatic margin additions for a demo position.
func (dc *DemoClient) SetAutoAddMargin(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/set-auto-add-margin", params)
}

// AddOrReduceMargin adjusts a demo position's margin.
func (dc *DemoClient) AddOrReduceMargin(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/position/add-margin", params)
}

// GetClosedPnL returns closed PnL records from the demo account.
func (dc *DemoClient) GetClosedPnL(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/position/closed-pnl", params)
}

// GetBorrowHistory returns demo account borrow history.
func (dc *DemoClient) GetBorrowHistory(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/borrow-history", params)
}

// SetCollateralCoin updates collateral settings for a demo account.
func (dc *DemoClient) SetCollateralCoin(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/account/set-collateral-switch", params)
}

// GetCollateralInfo returns demo account collateral information.
func (dc *DemoClient) GetCollateralInfo(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/collateral-info", params)
}

// GetCoinGreeks returns option Greeks for the supplied demo-account parameters.
func (dc *DemoClient) GetCoinGreeks(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/asset/coin-greeks", params)
}

// GetAccountInfo returns metadata for the authenticated demo account.
func (dc *DemoClient) GetAccountInfo() (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/info", nil)
}

// GetTransactionLog returns demo account transactions.
func (dc *DemoClient) GetTransactionLog(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/transaction-log", params)
}

// SetMarginMode changes the demo account margin mode.
func (dc *DemoClient) SetMarginMode(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/account/set-margin-mode", params)
}

// SetSpotHedging configures spot hedging for a demo account.
func (dc *DemoClient) SetSpotHedging(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/account/set-hedging-mode", params)
}

// GetDeliveryRecord returns demo asset delivery records.
func (dc *DemoClient) GetDeliveryRecord(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/asset/delivery-record", params)
}

// GetUSDCSettlement returns demo USDC settlement records.
func (dc *DemoClient) GetUSDCSettlement(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/asset/settlement-record", params)
}

// ToggleMarginTrade changes spot margin trading mode for a demo account.
func (dc *DemoClient) ToggleMarginTrade(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/spot-margin-trade/switch-mode", params)
}

// SetSpotMarginLeverage sets demo spot-margin leverage.
func (dc *DemoClient) SetSpotMarginLeverage(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/spot-margin-trade/set-leverage", params)
}

// GetSpotMarginStatus returns demo spot-margin trading status.
func (dc *DemoClient) GetSpotMarginStatus() (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/spot-margin-trade/state", nil)
}

// DemoFundRequest describes a single demo-fund adjustment.
type DemoFundRequest struct {
	Coin      string `json:"coin"`
	AmountStr string `json:"amountStr"`
}

// ApplyForDemoFunds requests an adjustment to one or more demo account balances.
func (dc *DemoClient) ApplyForDemoFunds(adjustType int, funds []DemoFundRequest) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"adjustType":        adjustType,
		"utaDemoApplyMoney": funds,
	}
	return dc.Request("POST", "/v5/account/demo-apply-money", params)
}

// ApplyForDemoFundsSimple requests a standard demo-fund adjustment for one coin.
func (dc *DemoClient) ApplyForDemoFundsSimple(coin string, amount string) (map[string]interface{}, error) {
	return dc.ApplyForDemoFunds(0, []DemoFundRequest{
		{Coin: coin, AmountStr: amount},
	})
}

// CreateDemoAccount creates a demo member through an authorized mainnet client.
func (dc *DemoClient) CreateDemoAccount(mainnetClient *Client) (map[string]interface{}, error) {
	return mainnetClient.Request("POST", "/v5/user/create-demo-member", map[string]interface{}{})
}

// CreateDemoAPIKey creates a demo sub-account API key through a mainnet client.
func (dc *DemoClient) CreateDemoAPIKey(mainnetClient *Client, demoUID string, params map[string]interface{}) (map[string]interface{}, error) {
	if params == nil {
		params = map[string]interface{}{}
	}
	params["subuid"] = demoUID
	return mainnetClient.Request("POST", "/v5/user/create-sub-api", params)
}

// UpdateDemoAPIKey updates a demo sub-account API key through a mainnet client.
func (dc *DemoClient) UpdateDemoAPIKey(mainnetClient *Client, params map[string]interface{}) (map[string]interface{}, error) {
	return mainnetClient.Request("POST", "/v5/user/update-sub-api", params)
}

// GetAPIKeyInfo returns information about the authenticated demo API key.
func (dc *DemoClient) GetAPIKeyInfo() (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/user/query-api", nil)
}

// DeleteDemoAPIKey removes a demo sub-account API key through a mainnet client.
func (dc *DemoClient) DeleteDemoAPIKey(mainnetClient *Client, params map[string]interface{}) (map[string]interface{}, error) {
	return mainnetClient.Request("POST", "/v5/user/delete-sub-api", params)
}
