package bybit

import "testing"

func TestDemoClientEndpointWrappers(t *testing.T) {
	client := mockClient(t)
	demo := &DemoClient{Client: client}
	params := map[string]interface{}{"category": "linear", "symbol": "BTCUSDT"}

	calls := []struct {
		name string
		call func() error
	}{
		{"GetWalletBalance", func() error { _, err := demo.GetWalletBalance(nil); return err }},
		{"CreateOrder", func() error { _, err := demo.CreateOrder(params); return err }},
		{"AmendOrder", func() error { _, err := demo.AmendOrder(params); return err }},
		{"CancelOrder", func() error { _, err := demo.CancelOrder(params); return err }},
		{"CancelAllOrders", func() error { _, err := demo.CancelAllOrders(params); return err }},
		{"GetOpenOrders", func() error { _, err := demo.GetOpenOrders(params); return err }},
		{"GetOrderHistory", func() error { _, err := demo.GetOrderHistory(params); return err }},
		{"GetTradeHistory", func() error { _, err := demo.GetTradeHistory(params); return err }},
		{"BatchPlaceOrder", func() error { _, err := demo.BatchPlaceOrder(params); return err }},
		{"BatchAmendOrder", func() error { _, err := demo.BatchAmendOrder(params); return err }},
		{"BatchCancelOrder", func() error { _, err := demo.BatchCancelOrder(params); return err }},
		{"GetPositions", func() error { _, err := demo.GetPositions(params); return err }},
		{"SetLeverage", func() error { _, err := demo.SetLeverage(params); return err }},
		{"SwitchPositionMode", func() error { _, err := demo.SwitchPositionMode(params); return err }},
		{"SetTradingStop", func() error { _, err := demo.SetTradingStop(params); return err }},
		{"SetAutoAddMargin", func() error { _, err := demo.SetAutoAddMargin(params); return err }},
		{"AddOrReduceMargin", func() error { _, err := demo.AddOrReduceMargin(params); return err }},
		{"GetClosedPnL", func() error { _, err := demo.GetClosedPnL(params); return err }},
		{"GetBorrowHistory", func() error { _, err := demo.GetBorrowHistory(params); return err }},
		{"SetCollateralCoin", func() error { _, err := demo.SetCollateralCoin(params); return err }},
		{"GetCollateralInfo", func() error { _, err := demo.GetCollateralInfo(params); return err }},
		{"GetCoinGreeks", func() error { _, err := demo.GetCoinGreeks(params); return err }},
		{"GetAccountInfo", func() error { _, err := demo.GetAccountInfo(); return err }},
		{"GetTransactionLog", func() error { _, err := demo.GetTransactionLog(params); return err }},
		{"SetMarginMode", func() error { _, err := demo.SetMarginMode(params); return err }},
		{"SetSpotHedging", func() error { _, err := demo.SetSpotHedging(params); return err }},
		{"GetDeliveryRecord", func() error { _, err := demo.GetDeliveryRecord(params); return err }},
		{"GetUSDCSettlement", func() error { _, err := demo.GetUSDCSettlement(params); return err }},
		{"ToggleMarginTrade", func() error { _, err := demo.ToggleMarginTrade(params); return err }},
		{"SetSpotMarginLeverage", func() error { _, err := demo.SetSpotMarginLeverage(params); return err }},
		{"GetSpotMarginStatus", func() error { _, err := demo.GetSpotMarginStatus(); return err }},
		{"ApplyForDemoFunds", func() error {
			_, err := demo.ApplyForDemoFunds(0, []DemoFundRequest{{Coin: "USDT", AmountStr: "10"}})
			return err
		}},
		{"ApplyForDemoFundsSimple", func() error { _, err := demo.ApplyForDemoFundsSimple("USDT", "10"); return err }},
		{"CreateDemoAccount", func() error { _, err := demo.CreateDemoAccount(client); return err }},
		{"CreateDemoAPIKey", func() error { _, err := demo.CreateDemoAPIKey(client, "1", nil); return err }},
		{"UpdateDemoAPIKey", func() error { _, err := demo.UpdateDemoAPIKey(client, params); return err }},
		{"GetAPIKeyInfo", func() error { _, err := demo.GetAPIKeyInfo(); return err }},
		{"DeleteDemoAPIKey", func() error { _, err := demo.DeleteDemoAPIKey(client, params); return err }},
	}
	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTradFiEndpointWrappers(t *testing.T) {
	client := mockClient(t)
	order := TradFiOrderParams{Symbol: "XAUUSD", Side: "Buy", OrderType: "Limit", Qty: "1", Price: "2000", TakeProfit: "2100", StopLoss: "1950", ReduceOnly: true, OrderLinkID: "tradfi-test"}
	calls := []struct {
		name string
		call func() error
	}{
		{"GetTradFiInstruments", func() error { _, err := client.GetTradFiInstruments(TradFiAssetMetal); return err }},
		{"GetTradFiTickers", func() error { _, err := client.GetTradFiTickers([]string{"XAUUSD"}); return err }},
		{"GetMetalsTickers", func() error { _, err := client.GetMetalsTickers(); return err }},
		{"GetForexTickers", func() error { _, err := client.GetForexTickers(); return err }},
		{"GetStockTickers", func() error { _, err := client.GetStockTickers(); return err }},
		{"GetIndexTickers", func() error { _, err := client.GetIndexTickers(); return err }},
		{"GetTradFiTicker", func() error { _, err := client.GetTradFiTicker("XAUUSD"); return err }},
		{"GetTradFiKline", func() error { _, err := client.GetTradFiKline("XAUUSD", "60", 10); return err }},
		{"GetTradFiOrderbook", func() error { _, err := client.GetTradFiOrderbook("XAUUSD", 0); return err }},
		{"GetTradFiSwapFee", func() error { _, err := client.GetTradFiSwapFee("XAUUSD"); return err }},
		{"GetTradFiPositions", func() error { _, err := client.GetTradFiPositions("XAUUSD"); return err }},
		{"PlaceTradFiOrder", func() error { _, err := client.PlaceTradFiOrder(order); return err }},
		{"CloseTradFiPosition", func() error { _, err := client.CloseTradFiPosition("XAUUSD", "SHORT", "1", 0); return err }},
		{"SetTradFiLeverage", func() error { _, err := client.SetTradFiLeverage("XAUUSD", 2); return err }},
		{"GetTradFiTradeHistory", func() error { _, err := client.GetTradFiTradeHistory("XAUUSD", 10); return err }},
		{"GetTradFiOpenOrders", func() error { _, err := client.GetTradFiOpenOrders("XAUUSD"); return err }},
		{"CancelTradFiOrder", func() error { _, err := client.CancelTradFiOrder("XAUUSD", "id", "link"); return err }},
		{"GetTradFiFeeRate", func() error { _, err := client.GetTradFiFeeRate("XAUUSD"); return err }},
	}
	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
