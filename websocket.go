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

type WebSocket struct {
	apiKey          string
	apiSecret       string
	testnet         bool
	region          string
	isPrivate       bool
	conn            *websocket.Conn
	subscriptions   []string
	messageCallback func(map[string]interface{})
	mu              sync.RWMutex
	connected       bool
}

type WebSocketConfig struct {
	APIKey    string
	APISecret string
	Testnet   bool
	Region    string
	IsPrivate bool
}

func NewWebSocket(config WebSocketConfig) *WebSocket {
	if config.Region == "" {
		config.Region = "global"
	}

	return &WebSocket{
		apiKey:        config.APIKey,
		apiSecret:     config.APISecret,
		testnet:       config.Testnet,
		region:        config.Region,
		isPrivate:     config.IsPrivate,
		subscriptions: make([]string, 0),
	}
}

func (ws *WebSocket) getWebSocketURL() string {
	if ws.testnet {
		if ws.isPrivate {
			return "wss://stream-testnet.bybit.com/v5/private"
		}
		return "wss://stream-testnet.bybit.com/v5/public/spot"
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

	return conn.WriteMessage(websocket.TextMessage, data)
}

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

func (ws *WebSocket) SubscribeOrderbook(symbol string, depth int) error {
	topic := fmt.Sprintf("orderbook.%d.%s", depth, symbol)
	return ws.Subscribe([]string{topic})
}

func (ws *WebSocket) SubscribeTrade(symbol string) error {
	topic := fmt.Sprintf("publicTrade.%s", symbol)
	return ws.Subscribe([]string{topic})
}

func (ws *WebSocket) SubscribeTicker(symbol string) error {
	topic := fmt.Sprintf("tickers.%s", symbol)
	return ws.Subscribe([]string{topic})
}

func (ws *WebSocket) SubscribeKline(symbol, interval string) error {
	topic := fmt.Sprintf("kline.%s.%s", interval, symbol)
	return ws.Subscribe([]string{topic})
}

func (ws *WebSocket) SubscribePosition() error {
	return ws.Subscribe([]string{"position"})
}

func (ws *WebSocket) SubscribeExecution() error {
	return ws.Subscribe([]string{"execution"})
}

func (ws *WebSocket) SubscribeOrder() error {
	return ws.Subscribe([]string{"order"})
}

func (ws *WebSocket) SubscribeWallet() error {
	return ws.Subscribe([]string{"wallet"})
}

func (ws *WebSocket) OnMessage(callback func(map[string]interface{})) {
	ws.mu.Lock()
	ws.messageCallback = callback
	ws.mu.Unlock()
}

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
			ws.Send(map[string]interface{}{"op": "pong"})
		}
	}

	return nil
}

func (ws *WebSocket) Ping() error {
	return ws.Send(map[string]interface{}{"op": "ping"})
}

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

func (ws *WebSocket) GetSubscriptions() []string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	subs := make([]string, len(ws.subscriptions))
	copy(subs, ws.subscriptions)
	return subs
}

func (ws *WebSocket) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.connected
}
