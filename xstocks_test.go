package bybit

import (
	"encoding/json"
	"net/http"
	"testing"
)

func xStockResponse(symbolType string) string {
	return `{"retCode":0,"result":{"list":[{"symbol":"AAPLXUSDT","baseCoin":"AAPLX","quoteCoin":"USDT","status":"Trading","symbolType":"` + symbolType + `","xstockMultiplier":"0.5","priceFilter":{"tickSize":"0.01"},"lotSizeFilter":{"basePrecision":"0.001","minOrderAmt":"5","maxLimitOrderQty":"100","maxMarketOrderQty":"50"}}]}}`
}

func TestGetXStocksUsesExactSymbolType(t *testing.T) {
	client, err := NewClient(ClientConfig{HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.URL.Query().Get("category"); got != XStockCategory {
			t.Fatalf("category = %q", got)
		}
		if got := req.URL.Query().Get("symbolType"); got != XStockSymbolType {
			t.Fatalf("symbolType = %q", got)
		}
		return response(req, http.StatusOK, xStockResponse(XStockSymbolType)), nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	items, err := client.GetXStocks()
	if err != nil || len(items) != 1 {
		t.Fatalf("GetXStocks() = %#v, %v", items, err)
	}
	if items[0].XStockMultiplier != "0.5" || items[0].PriceFilter.TickSize != "0.01" {
		t.Fatalf("metadata was not preserved: %#v", items[0])
	}
}

func TestGetXStockRejectsNonXStock(t *testing.T) {
	client, _ := NewClient(ClientConfig{HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return response(req, http.StatusOK, xStockResponse("innovation")), nil
	})}})
	if _, err := client.GetXStock("AAPLXUSDT"); err == nil {
		t.Fatal("GetXStock accepted a non-xstocks symbolType")
	}
}

func TestValidateAndPlaceXStockOrder(t *testing.T) {
	calls := 0
	client, _ := NewClient(ClientConfig{HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return response(req, http.StatusOK, xStockResponse(XStockSymbolType)), nil
		}
		if req.URL.Path != "/v5/order/create" {
			t.Fatalf("path = %s", req.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["category"] != "spot" || body["isLeverage"].(float64) != 0 || body["marketUnit"] != "baseCoin" {
			t.Fatalf("unexpected request: %#v", body)
		}
		return response(req, http.StatusOK, `{"retCode":0,"result":{"orderId":"1"}}`), nil
	})}})
	_, err := client.PlaceXStockOrder(XStockOrderParams{Symbol: "AAPLXUSDT", Side: "Buy", OrderType: "Market", Qty: "1.000", MarketUnit: "baseCoin"})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestXStockDecimalConversionsAndValidation(t *testing.T) {
	if got, err := ToXStockTokenQuantity("1", "3", "0.01"); err != nil || got != "0.33" {
		t.Fatalf("token quantity = %q, %v", got, err)
	}
	if got, err := ToXStockTokenPrice("10.01", "0.5", "0.1", "Sell"); err != nil || got != "5.1" {
		t.Fatalf("sell token price = %q, %v", got, err)
	}
	instrument := XStockInstrument{Symbol: "AAPLXUSDT", SymbolType: XStockSymbolType, Status: "Trading", PriceFilter: XStockPriceFilter{TickSize: "0.01"}, LotSizeFilter: XStockLotSizeFilter{BasePrecision: "0.001", MinOrderAmt: "5"}}
	if err := ValidateXStockOrder(instrument, XStockOrderParams{Symbol: "AAPLXUSDT", Side: "Buy", OrderType: "Limit", Qty: "1.0005", Price: "5.00"}); err == nil {
		t.Fatal("accepted quantity outside base precision")
	}
	if err := ValidateXStockOrder(instrument, XStockOrderParams{Symbol: "AAPLXUSDT", Side: "Buy", OrderType: "Limit", Qty: "1.000", Price: "5.005"}); err == nil {
		t.Fatal("accepted price outside tick size")
	}
}
