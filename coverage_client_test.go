package bybit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"testing"
)

func mockClient(t *testing.T) *Client {
	t.Helper()
	client, err := NewClient(ClientConfig{
		APIKey:    "key",
		APISecret: "secret",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return response(req, http.StatusOK, `{"retCode":0,"retMsg":"OK","result":{"list":[{"lastPrice":"100","markPrice":"99","bid1Price":"98"}]}}`), nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestClientEndpointWrappers(t *testing.T) {
	client := mockClient(t)
	params := map[string]interface{}{"category": "linear", "symbol": "BTCUSDT"}

	calls := []struct {
		name string
		call func() error
	}{
		{"GetServerTime", func() error { _, err := client.GetServerTime(); return err }},
		{"GetTickers", func() error { _, err := client.GetTickers(params); return err }},
		{"GetKline", func() error { _, err := client.GetKline(params); return err }},
		{"GetOrderbook", func() error { _, err := client.GetOrderbook(params); return err }},
		{"GetRPIOrderbook", func() error { _, err := client.GetRPIOrderbook(params); return err }},
		{"GetOpenInterest", func() error { _, err := client.GetOpenInterest(params); return err }},
		{"GetRecentTrades", func() error { _, err := client.GetRecentTrades(params); return err }},
		{"GetFundingRateHistory", func() error { _, err := client.GetFundingRateHistory(params); return err }},
		{"GetHistoricalVolatility", func() error { _, err := client.GetHistoricalVolatility(params); return err }},
		{"GetInsurance", func() error { _, err := client.GetInsurance(params); return err }},
		{"GetRiskLimit", func() error { _, err := client.GetRiskLimit(params); return err }},
		{"CreateOrder", func() error { _, err := client.CreateOrder(params); return err }},
		{"GetOpenOrders", func() error { _, err := client.GetOpenOrders(params); return err }},
		{"CancelOrder", func() error { _, err := client.CancelOrder(params); return err }},
		{"AmendOrder", func() error { _, err := client.AmendOrder(params); return err }},
		{"CancelAllOrders", func() error { _, err := client.CancelAllOrders(params); return err }},
		{"GetHistoryOrders", func() error { _, err := client.GetHistoryOrders(params); return err }},
		{"GetWalletBalance", func() error { _, err := client.GetWalletBalance(params); return err }},
		{"GetTransferableAmount", func() error { _, err := client.GetTransferableAmount(params); return err }},
		{"GetTransactionLog", func() error { _, err := client.GetTransactionLog(params); return err }},
		{"GetAccountInfo", func() error { _, err := client.GetAccountInfo(); return err }},
		{"GetAccountInstrumentsInfo", func() error { _, err := client.GetAccountInstrumentsInfo(params); return err }},
		{"GetPositions", func() error { _, err := client.GetPositions(params); return err }},
		{"SwitchPositionMode", func() error { _, err := client.SwitchPositionMode(params); return err }},
		{"SetTradingStop", func() error { _, err := client.SetTradingStop(params); return err }},
		{"SetLeverage", func() error { _, err := client.SetLeverage("linear", "BTCUSDT", 2, nil); return err }},
		{"SetAutoAddMargin", func() error { _, err := client.SetAutoAddMargin(params); return err }},
		{"AddOrReduceMargin", func() error { _, err := client.AddOrReduceMargin(params); return err }},
		{"GetClosedPnL", func() error { _, err := client.GetClosedPnL(params); return err }},
		{"GetClosedOptionsPositions", func() error { _, err := client.GetClosedOptionsPositions(params); return err }},
		{"MovePosition", func() error { _, err := client.MovePosition(params); return err }},
		{"GetMovePositionHistory", func() error { _, err := client.GetMovePositionHistory(params); return err }},
		{"ConfirmNewRiskLimit", func() error { _, err := client.ConfirmNewRiskLimit(params); return err }},
	}

	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestClientHelpersAndPlaceOrder(t *testing.T) {
	client := mockClient(t)
	price, err := client.lastPrice("BTCUSDT", "linear")
	if err != nil || price != 100 {
		t.Fatalf("lastPrice() = %v, %v", price, err)
	}
	if got := client.qtyFromMargin(100, 80, 2); got != 2.5 {
		t.Fatalf("qtyFromMargin() = %v", got)
	}
	if got := client.qtyFromMargin(100, 0, 2); got != 0 {
		t.Fatalf("qtyFromMargin with zero price = %v", got)
	}
	if got := client.ComputeFee("spot", 1000, "VIP1", "maker"); got <= 0 {
		t.Fatalf("ComputeFee() = %v", got)
	}
	if got := client.ComputeFee("linear", 1000, "unknown", "maker"); got != 0.4 {
		t.Fatalf("fallback fee = %v, want 0.4", got)
	}

	priceValue, tp, sl, leverage := 100.0, 0.1, 0.05, 2.0
	sell := "Sell"
	orders := []PlaceOrderParams{
		{Type: "spot", Symbol: "BTCUSDT", Execution: "limit", Price: &priceValue, Size: 1},
		{Type: "linear", Symbol: "BTCUSDT", Execution: "trigger", Price: &priceValue, Side: &sell, Leverage: &leverage, Size: 10, SlTp: &SlTpParams{Type: "percent", TakeProfit: &tp, StopLoss: &sl}, Extra: map[string]interface{}{"orderLinkId": "test-order"}},
	}
	for _, order := range orders {
		if _, err := client.PlaceOrder(order); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRSAClientAndQueryHelpers(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	client, err := NewClient(ClientConfig{Signature: " RSA ", RSAPrivateKey: string(pemKey)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.signString("payload"); err != nil {
		t.Fatal(err)
	}
	if got := client.buildQuery(map[string]interface{}{"z": "space here", "a": 2}); got != "a=2&z=space+here" {
		t.Fatalf("buildQuery() = %q", got)
	}
	if _, err := parseRSAPrivateKey("not a PEM key"); err == nil {
		t.Fatal("parseRSAPrivateKey accepted invalid data")
	}
}
