package bybit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebSocketURLsAndCategories(t *testing.T) {
	tests := []struct {
		name   string
		config WebSocketConfig
		want   string
	}{
		{"spot default", WebSocketConfig{}, "wss://stream.bybit.com/v5/public/spot"},
		{"linear global", WebSocketConfig{PublicCategory: WebSocketCategoryLinear}, "wss://stream.bybit.com/v5/public/linear"},
		{"spot regional", WebSocketConfig{Region: "nl"}, "wss://stream.bybit.nl/v5/public/spot"},
		{"private ignores category", WebSocketConfig{IsPrivate: true, PublicCategory: WebSocketCategoryLinear}, "wss://stream.bybit.com/v5/private"},
		{"demo spot", WebSocketConfig{Demo: true}, "wss://stream-demo.bybit.com/v5/public/spot"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := NewWebSocket(tt.config)
			if got := ws.getWebSocketURL(); got != tt.want {
				t.Errorf("getWebSocketURL() = %q, want %q", got, tt.want)
			}
		})
	}
	for _, config := range []WebSocketConfig{{PublicCategory: "inverse"}, {PublicCategory: WebSocketCategoryLinear, Region: "nl"}, {PublicCategory: WebSocketCategoryLinear, Demo: true}} {
		ws := NewWebSocket(config)
		if ws.configErr == nil {
			t.Fatalf("NewWebSocket(%+v) did not reject unsupported configuration", config)
		}
		if err := ws.Connect(); err == nil {
			t.Fatalf("Connect() accepted unsupported configuration %+v", config)
		}
		if err := ws.Listen(); err == nil {
			t.Fatalf("Listen() returned nil for unsupported configuration %+v", config)
		}
	}
}

func TestSubscribeKlineTopic(t *testing.T) {
	messages := make(chan []byte, 1)
	server := newWebSocketServer(t, func(conn *websocket.Conn) {
		_, message, err := conn.ReadMessage()
		if err == nil {
			messages <- message
		}
	})
	defer server.Close()
	ws := connectedWebSocket(t, server.URL)
	defer func() { _ = ws.Close() }()
	if err := ws.SubscribeKline("BTCUSDT", "1"); err != nil {
		t.Fatal(err)
	}
	select {
	case message := <-messages:
		if !bytes.Contains(message, []byte(`"kline.1.BTCUSDT"`)) {
			t.Fatalf("subscription = %s", message)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive subscription")
	}
}

func TestWebSocketHeartbeatRepliesDoNotSendJSONPong(t *testing.T) {
	tests := []struct {
		name   string
		config WebSocketConfig
		reply  string
		op     string
		topic  string
	}{
		{
			name:  "public spot",
			reply: `{"success":true,"ret_msg":"pong","conn_id":"spot-test","op":"ping"}`,
			op:    "ping",
			topic: "tickers.BTCUSDT",
		},
		{
			name:   "public linear",
			config: WebSocketConfig{PublicCategory: WebSocketCategoryLinear},
			reply:  `{"success":true,"ret_msg":"pong","conn_id":"linear-test","req_id":"","op":"ping"}`,
			op:     "ping",
			topic:  "kline.1.BTCUSDT",
		},
		{
			name:   "private",
			config: WebSocketConfig{IsPrivate: true},
			reply:  `{"op":"pong","args":["1675418560633"],"conn_id":"private-test"}`,
			op:     "pong",
			topic:  "wallet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := []byte(fmt.Sprintf(`{"topic":%q,"data":[]}`, tt.topic))
			serverDone := make(chan error, 1)
			server := newWebSocketServer(t, func(conn *websocket.Conn) {
				defer func() { _ = conn.Close() }()
				serverDone <- func() error {
					messageType, message, err := conn.ReadMessage()
					if err != nil {
						return err
					}
					if messageType != websocket.TextMessage || string(message) != `{"op":"ping"}` {
						return fmt.Errorf("client heartbeat = %q, type %d; want JSON ping", message, messageType)
					}

					// Protocol control frames must still be handled by gorilla/websocket.
					var controlPong string
					conn.SetPongHandler(func(payload string) error { controlPong = payload; return nil })
					if err := conn.WriteMessage(websocket.PingMessage, []byte("control-heartbeat")); err != nil {
						return err
					}
					for _, payload := range [][]byte{[]byte(tt.reply), update} {
						if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
							return err
						}
					}

					// The update callback sends a second ping after the reply has been
					// processed. An erroneous JSON pong would arrive before that ping.
					messageType, message, err = conn.ReadMessage()
					if err != nil {
						return err
					}
					if messageType != websocket.TextMessage || string(message) != `{"op":"ping"}` {
						return fmt.Errorf("unexpected message after heartbeat reply: type=%d payload=%q", messageType, message)
					}
					if controlPong != "control-heartbeat" {
						return fmt.Errorf("protocol control pong = %q", controlPong)
					}
					return nil
				}()
			})
			defer server.Close()
			conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
			if err != nil {
				t.Fatal(err)
			}
			ws := NewWebSocket(tt.config)
			ws.conn, ws.connected = conn, true
			defer func() { _ = ws.Close() }()
			var rawMessages [][]byte
			var decodedMessages []map[string]interface{}
			ws.OnRawMessage(func(message []byte, receivedAt time.Time) {
				rawMessages = append(rawMessages, message)
				if receivedAt.IsZero() {
					t.Error("raw heartbeat callback received a zero timestamp")
				}
			})
			ws.OnMessage(func(message map[string]interface{}) {
				decodedMessages = append(decodedMessages, message)
				if message["topic"] == tt.topic {
					if err := ws.Ping(); err != nil {
						t.Errorf("follow-up ping: %v", err)
					}
				}
			})
			if err := ws.Ping(); err != nil {
				t.Fatal(err)
			}
			if err := ws.ListenContext(context.Background()); err == nil {
				t.Error("ListenContext returned nil after the server closed")
			}
			if err := <-serverDone; err != nil {
				t.Fatal(err)
			}
			if len(rawMessages) != 2 || !bytes.Equal(rawMessages[0], []byte(tt.reply)) || !bytes.Equal(rawMessages[1], update) {
				t.Fatalf("raw callbacks did not preserve heartbeat reply and update: %q", rawMessages)
			}
			if len(decodedMessages) != 2 || decodedMessages[0]["op"] != tt.op || decodedMessages[1]["topic"] != tt.topic {
				t.Fatalf("decoded callbacks = %#v", decodedMessages)
			}
		})
	}
}

func TestListenContextRawAndDecodedCallbacks(t *testing.T) {
	valid := []byte(`{"topic":"kline.1.BTCUSDT","data":[{"confirm":true,"price":1.2}]}`)
	invalid := []byte(`{"topic":`)
	server := newWebSocketServer(t, func(conn *websocket.Conn) {
		_ = conn.WriteMessage(websocket.TextMessage, valid)
		_ = conn.WriteMessage(websocket.TextMessage, invalid)
		_ = conn.Close()
	})
	defer server.Close()
	ws := connectedWebSocket(t, server.URL)
	defer func() { _ = ws.Close() }()
	var mu sync.Mutex
	var rawMessages [][]byte
	var receivedAt []time.Time
	decoded := 0
	ws.OnRawMessage(func(message []byte, at time.Time) {
		mu.Lock()
		rawMessages = append(rawMessages, append([]byte(nil), message...))
		receivedAt = append(receivedAt, at)
		mu.Unlock()
		message[0] = 'X'
	})
	ws.OnMessage(func(message map[string]interface{}) {
		mu.Lock()
		decoded++
		mu.Unlock()
		if message["topic"] != "kline.1.BTCUSDT" {
			t.Errorf("topic = %v", message["topic"])
		}
	})
	if err := ws.ListenContext(context.Background()); err == nil {
		t.Fatal("ListenContext() returned nil after server close")
	}
	mu.Lock()
	defer mu.Unlock()
	if decoded != 1 {
		t.Fatalf("decoded callbacks = %d, want 1", decoded)
	}
	if len(rawMessages) != 2 {
		t.Fatalf("raw callbacks = %d, want 2", len(rawMessages))
	}
	if !bytes.Equal(rawMessages[0], valid) || !bytes.Equal(rawMessages[1], invalid) {
		t.Fatalf("raw payloads changed: %q / %q", rawMessages[0], rawMessages[1])
	}
	if receivedAt[0].IsZero() || receivedAt[1].IsZero() {
		t.Fatal("raw callback received a zero timestamp")
	}
}

func TestListenContextCancellationAndUnexpectedClose(t *testing.T) {
	t.Run("cancellation", func(t *testing.T) {
		connected := make(chan struct{})
		server := newWebSocketServer(t, func(conn *websocket.Conn) { close(connected); time.Sleep(time.Second) })
		defer server.Close()
		ws := connectedWebSocket(t, server.URL)
		defer func() { _ = ws.Close() }()
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan error, 1)
		go func() { result <- ws.ListenContext(ctx) }()
		<-connected
		cancel()
		select {
		case err := <-result:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("ListenContext() error = %v, want context.Canceled", err)
			}
		case <-time.After(time.Second):
			t.Fatal("ListenContext did not stop after cancellation")
		}
	})
	t.Run("unexpected close", func(t *testing.T) {
		server := newWebSocketServer(t, func(conn *websocket.Conn) { _ = conn.Close() })
		defer server.Close()
		ws := connectedWebSocket(t, server.URL)
		defer func() { _ = ws.Close() }()
		called := false
		ws.OnMessage(func(map[string]interface{}) { called = true })
		if err := ws.ListenContext(context.Background()); err == nil {
			t.Fatal("ListenContext() returned nil after unexpected close")
		}
		if called {
			t.Fatal("ListenContext sent a synthetic error to OnMessage")
		}
	})
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

func TestCloseWhileListening(t *testing.T) {
	server := newWebSocketServer(t, func(conn *websocket.Conn) { time.Sleep(time.Second) })
	defer server.Close()
	ws := connectedWebSocket(t, server.URL)
	result := make(chan error, 1)
	go func() { result <- ws.ListenContext(context.Background()) }()
	time.Sleep(20 * time.Millisecond)
	if err := ws.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("ListenContext() returned nil after Close")
		}
	case <-time.After(time.Second):
		t.Fatal("ListenContext did not stop after Close")
	}
}

func newWebSocketServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err == nil {
			handler(conn)
		}
	}))
}

func connectedWebSocket(t *testing.T, serverURL string) *WebSocket {
	t.Helper()
	url := "ws" + strings.TrimPrefix(serverURL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	ws := NewWebSocket(WebSocketConfig{})
	ws.conn = conn
	ws.connected = true
	return ws
}
