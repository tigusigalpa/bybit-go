package bybit

import "testing"

func TestIsTradFiSymbol(t *testing.T) {
	for _, symbol := range []string{"XAUUSD", "EURUSD", "US500USD", "JP225USD"} {
		if !IsTradFiSymbol(symbol) {
			t.Errorf("IsTradFiSymbol(%q) = false, want true", symbol)
		}
	}
	for _, symbol := range []string{"BTCUSDT", "ETHUSD", "SOLUSD"} {
		if IsTradFiSymbol(symbol) {
			t.Errorf("IsTradFiSymbol(%q) = true, want false", symbol)
		}
	}
}

func TestFilterInstrumentsByAssetClass(t *testing.T) {
	response := map[string]interface{}{
		"retCode": float64(0),
		"result": map[string]interface{}{"category": "linear", "list": []interface{}{
			map[string]interface{}{"symbol": "XAUUSD"},
			map[string]interface{}{"symbol": "EURUSD"},
			map[string]interface{}{"symbol": "BTCUSDT"},
		}},
	}
	filtered := filterInstrumentsByAssetClass(response, TradFiAssetMetal)
	list := filtered["result"].(map[string]interface{})["list"].([]interface{})
	if len(list) != 1 || list[0].(map[string]interface{})["symbol"] != "XAUUSD" {
		t.Fatalf("metal filter returned %#v", list)
	}
}
