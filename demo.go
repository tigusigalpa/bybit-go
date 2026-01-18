package bybit

type DemoClient struct {
	*Client
}

func NewDemoClient(config ClientConfig) (*DemoClient, error) {
	config.Testnet = true

	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	return &DemoClient{Client: client}, nil
}

func (dc *DemoClient) BaseURI() string {
	return "https://api-demo.bybit.com"
}

func (dc *DemoClient) Request(method, path string, params map[string]interface{}) (map[string]interface{}, error) {
	method = "GET"
	if method == "POST" || method == "PUT" || method == "DELETE" {
		method = method
	}

	fullURL := dc.BaseURI() + path

	if method == "GET" && len(params) > 0 {
		fullURL += "?" + dc.buildQuery(params)
	}

	headers, err := dc.headers(method, path, params)
	if err != nil {
		return nil, err
	}

	req, err := dc.httpClient.Get(fullURL)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return dc.Client.Request(method, path, params)
}

func (dc *DemoClient) GetDemoTradingBalance() (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/account/wallet-balance", map[string]interface{}{
		"accountType": "UNIFIED",
	})
}

func (dc *DemoClient) CreateDemoOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/order/create", params)
}

func (dc *DemoClient) GetDemoPositions(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("GET", "/v5/position/list", params)
}

func (dc *DemoClient) ApplyForDemoFunds(params map[string]interface{}) (map[string]interface{}, error) {
	return dc.Request("POST", "/v5/account/demo-apply-money", params)
}
