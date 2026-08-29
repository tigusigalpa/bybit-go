package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket manages a Bybit WebSocket connection and its subscriptions.
type WebSocket struct {
	apiKey          string
	apiSecret       string
	demo            bool
	region          string
	isPrivate       bool
	conn            *websocket.Conn
	subscriptions   []string
	messageCallback func(map[string]interface{})
	mu              sync.RWMutex
	writeMu         sync.Mutex
	connected       bool
}

// WebSocketConfig configures a WebSocket client.
type WebSocketConfig struct {
	APIKey    string
	APISecret string
	Demo      bool
	Region    string
	IsPrivate bool
}

// NewWebSocket creates a WebSocket client with the supplied configuration.
func NewWebSocket(config WebSocketConfig) *WebSocket {
	if config.Region == "" {
		config.Region = "global"
	}

	return &WebSocket{
		apiKey:        config.APIKey,
		apiSecret:     config.APISecret,
		demo:          config.Demo,
		region:        config.Region,
		isPrivate:     config.IsPrivate,
		subscriptions: make([]string, 0),
	}
}

func (ws *WebSocket) getWebSocketURL() string {
	if ws.demo {
		if ws.isPrivate {
			return "wss://stream-demo.bybit.com/v5/private"
		}
		return "wss://stream-demo.bybit.com/v5/public/spot"
	}

	switch strings.ToLower(ws.region) {
	case "nl":
		if ws.isPrivate {
			return "wss://stream.bybit.nl/v5/private"
		}
		return "wss://stream.bybit.nl/v5/public/spot"
	case "tr":
		if ws.isPrivate {
			return "wss://stream.bybit-tr.com/v5/private"
		}
		return "wss://stream.bybit-tr.com/v5/public/spot"
	case "kz":
		if ws.isPrivate {
			return "wss://stream.bybit.kz/v5/private"
		}
		return "wss://stream.bybit.kz/v5/public/spot"
	case "ge":
		if ws.isPrivate {
			return "wss://stream.bybitgeorgia.ge/v5/private"
		}
		return "wss://stream.bybitgeorgia.ge/v5/public/spot"
	case "ae":
		if ws.isPrivate {
			return "wss://stream.bybit.ae/v5/private"
		}
		return "wss://stream.bybit.ae/v5/public/spot"
	default:
		if ws.isPrivate {
			return "wss://stream.bybit.com/v5/private"
		}
		return "wss://stream.bybit.com/v5/public/spot"
	}
}

// Connect establishes the WebSocket connection and authenticates private clients.
func (ws *WebSocket) Connect() error {
	url := ws.getWebSocketURL()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	ws.mu.Lock()
	ws.conn = conn
	ws.connected = true
	ws.mu.Unlock()

	if ws.isPrivate && ws.apiKey != "" && ws.apiSecret != "" {
		if err := ws.authenticate(); err != nil {
			ws.Close()
			return err
		}
	}

	return nil
}

func (ws *WebSocket) authenticate() error {
	expires := time.Now().UnixMilli() + 10000
	message := "GET/realtime" + strconv.FormatInt(expires, 10)

	mac := hmac.New(sha256.New, []byte(ws.apiSecret))
	mac.Write([]byte(message))
	signature := fmt.Sprintf("%x", mac.Sum(nil))

	authMessage := map[string]interface{}{
		"op":   "auth",
		"args": []interface{}{ws.apiKey, expires, signature},
	}

	return ws.Send(authMessage)
}

// Send serializes and sends a WebSocket message, connecting first when necessary.
func (ws *WebSocket) Send(message map[string]interface{}) error {
	ws.mu.RLock()
	if !ws.connected || ws.conn == nil {
		ws.mu.RUnlock()
		if err := ws.Connect(); err != nil {
			return err
		}
		ws.mu.RLock()
	}
	conn := ws.conn
	ws.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// gorilla/websocket supports one concurrent writer per connection.
	ws.writeMu.Lock()
	defer ws.writeMu.Unlock()

	return conn.WriteMessage(websocket.TextMessage, data)
}

// Subscribe sends a subscription request for the supplied topics.
func (ws *WebSocket) Subscribe(topics []string) error {
	message := map[string]interface{}{
		"op":   "subscribe",
		"args": topics,
	}

	ws.mu.Lock()
	ws.subscriptions = append(ws.subscriptions, topics...)
	ws.mu.Unlock()

	return ws.Send(message)
}

// Unsubscribe sends an unsubscribe request for the supplied topics.
func (ws *WebSocket) Unsubscribe(topics []string) error {
	message := map[string]interface{}{
		"op":   "unsubscribe",
		"args": topics,
	}

	ws.mu.Lock()
	for _, topic := range topics {
		for i, sub := range ws.subscriptions {
			if sub == topic {
				ws.subscriptions = append(ws.subscriptions[:i], ws.subscriptions[i+1:]...)
				break
			}
		}
	}
	ws.mu.Unlock()

	return ws.Send(message)
}

// SubscribeOrderbook subscribes to an order book topic for a symbol and depth.
func (ws *WebSocket) SubscribeOrderbook(symbol string, depth int) error {
	topic := fmt.Sprintf("orderbook.%d.%s", depth, symbol)
	return ws.Subscribe([]string{topic})
}

// SubscribeTrade subscribes to public trade updates for a symbol.
func (ws *WebSocket) SubscribeTrade(symbol string) error {
	topic := fmt.Sprintf("publicTrade.%s", symbol)
	return ws.Subscribe([]string{topic})
}

// SubscribeTicker subscribes to ticker updates for a symbol.
func (ws *WebSocket) SubscribeTicker(symbol string) error {
	topic := fmt.Sprintf("tickers.%s", symbol)
	return ws.Subscribe([]string{topic})
}

// SubscribeKline subscribes to candlestick updates for a symbol and interval.
func (ws *WebSocket) SubscribeKline(symbol, interval string) error {
	topic := fmt.Sprintf("kline.%s.%s", interval, symbol)
	return ws.Subscribe([]string{topic})
}

// SubscribePosition subscribes to private position updates.
func (ws *WebSocket) SubscribePosition() error {
	return ws.Subscribe([]string{"position"})
}

// SubscribeExecution subscribes to private execution updates.
func (ws *WebSocket) SubscribeExecution() error {
	return ws.Subscribe([]string{"execution"})
}

// SubscribeOrder subscribes to private order updates.
func (ws *WebSocket) SubscribeOrder() error {
	return ws.Subscribe([]string{"order"})
}

// SubscribeWallet subscribes to private wallet updates.
func (ws *WebSocket) SubscribeWallet() error {
	return ws.Subscribe([]string{"wallet"})
}

// OnMessage registers the callback invoked for each decoded message or read error.
func (ws *WebSocket) OnMessage(callback func(map[string]interface{})) {
	ws.mu.Lock()
	ws.messageCallback = callback
	ws.mu.Unlock()
}

// Listen reads messages until the connection closes or a read error occurs.
func (ws *WebSocket) Listen() error {
	ws.mu.RLock()
	if !ws.connected || ws.conn == nil {
		ws.mu.RUnlock()
		if err := ws.Connect(); err != nil {
			return err
		}
	} else {
		ws.mu.RUnlock()
	}

	for {
		ws.mu.RLock()
		conn := ws.conn
		ws.mu.RUnlock()

		if conn == nil {
			break
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			ws.mu.RLock()
			callback := ws.messageCallback
			ws.mu.RUnlock()

			if callback != nil {
				callback(map[string]interface{}{
					"error":   true,
					"message": err.Error(),
				})
			}
			break
		}

		var data map[string]interface{}
		if err := json.Unmarshal(message, &data); err != nil {
			continue
		}

		ws.mu.RLock()
		callback := ws.messageCallback
		ws.mu.RUnlock()

		if callback != nil {
			callback(data)
		}

		if op, ok := data["op"].(string); ok && op == "ping" {
			if err := ws.Send(map[string]interface{}{"op": "pong"}); err != nil {
				return err
			}
		}
	}

	return nil
}

// Ping sends a ping operation to the WebSocket server.
func (ws *WebSocket) Ping() error {
	return ws.Send(map[string]interface{}{"op": "ping"})
}

// Close closes the active WebSocket connection.
func (ws *WebSocket) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.conn != nil {
		err := ws.conn.Close()
		ws.conn = nil
		ws.connected = false
		return err
	}

	return nil
}

// GetSubscriptions returns a copy of locally tracked subscriptions.
func (ws *WebSocket) GetSubscriptions() []string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	subs := make([]string, len(ws.subscriptions))
	copy(subs, ws.subscriptions)
	return subs
}

// IsConnected reports whether the client currently has an active local connection.
func (ws *WebSocket) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.connected
}
