package bybit

import "testing"

func TestWebSocketURLs(t *testing.T) {
	tests := []struct {
		config WebSocketConfig
		want   string
	}{
		{WebSocketConfig{}, "wss://stream.bybit.com/v5/public/spot"},
		{WebSocketConfig{Region: "nl", IsPrivate: true}, "wss://stream.bybit.nl/v5/private"},
		{WebSocketConfig{Demo: true}, "wss://stream-demo.bybit.com/v5/public/spot"},
	}
	for _, tt := range tests {
		if got := NewWebSocket(tt.config).getWebSocketURL(); got != tt.want {
			t.Errorf("getWebSocketURL() = %q, want %q", got, tt.want)
		}
	}
}

func TestGetSubscriptionsReturnsCopy(t *testing.T) {
	ws := NewWebSocket(WebSocketConfig{})
	ws.subscriptions = []string{"tickers.BTCUSDT"}
	subs := ws.GetSubscriptions()
	subs[0] = "changed"
	if ws.GetSubscriptions()[0] != "tickers.BTCUSDT" {
		t.Fatal("GetSubscriptions returned the internal slice")
	}
}
