package bybit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketSubscriptionAndListenerFlow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	var received []string
	var receivedMu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer func() {
			_ = conn.Close()
		}()

		for i := 0; i < 3; i++ {
			_, message, err := conn.ReadMessage()
			if err != nil {
				t.Error(err)
				return
			}
			receivedMu.Lock()
			received = append(received, string(message))
			receivedMu.Unlock()
		}
		if err := conn.WriteJSON(map[string]interface{}{"topic": "tickers.BTCUSDT"}); err != nil {
			t.Error(err)
			return
		}
		if err := conn.WriteJSON(map[string]interface{}{"success": true, "ret_msg": "pong", "op": "ping"}); err != nil {
			t.Error(err)
			return
		}
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	ws := &WebSocket{conn: conn, connected: true, subscriptions: []string{}}
	var callbacks []map[string]interface{}
	ws.OnMessage(func(message map[string]interface{}) { callbacks = append(callbacks, message) })

	if err := ws.SubscribeTicker("BTCUSDT"); err != nil {
		t.Fatal(err)
	}
	if err := ws.Unsubscribe([]string{"tickers.BTCUSDT"}); err != nil {
		t.Fatal(err)
	}
	if err := ws.Ping(); err != nil {
		t.Fatal(err)
	}
	if err := ws.Listen(); err != nil {
		t.Fatal(err)
	}
	if len(callbacks) != 3 || callbacks[0]["topic"] != "tickers.BTCUSDT" || callbacks[1]["op"] != "ping" || callbacks[1]["ret_msg"] != "pong" || callbacks[1]["success"] != true || callbacks[2]["error"] != true {
		t.Fatalf("callbacks = %#v", callbacks)
	}
	if len(ws.GetSubscriptions()) != 0 {
		t.Fatal("Unsubscribe did not remove the topic")
	}
	if err := ws.Close(); err != nil {
		t.Fatal(err)
	}
	if ws.IsConnected() {
		t.Fatal("WebSocket should be disconnected after Close")
	}

	receivedMu.Lock()
	defer receivedMu.Unlock()
	if len(received) != 3 {
		t.Fatalf("received %d client messages, want 3", len(received))
	}
}

func TestWebSocketURLRegionsAndClosedConnection(t *testing.T) {
	for _, region := range []string{"nl", "tr", "kz", "ge", "ae"} {
		ws := NewWebSocket(WebSocketConfig{Region: region, IsPrivate: true})
		if !strings.Contains(ws.getWebSocketURL(), "/v5/private") {
			t.Fatalf("private URL for %s is invalid", region)
		}
	}
	ws := NewWebSocket(WebSocketConfig{})
	if err := ws.Close(); err != nil {
		t.Fatal(err)
	}
}
