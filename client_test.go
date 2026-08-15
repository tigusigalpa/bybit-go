package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func response(req *http.Request, statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}
}

func TestNewClientDefaultsAndValidation(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client.region != "global" || client.recvWindow != 5000 || client.signature != "hmac" {
		t.Fatalf("unexpected defaults: region=%q recvWindow=%d signature=%q", client.region, client.recvWindow, client.signature)
	}

	for _, config := range []ClientConfig{{Signature: "ed25519"}, {Signature: "rsa"}} {
		if _, err := NewClient(config); err == nil {
			t.Fatalf("NewClient(%+v) returned no validation error", config)
		}
	}
}

func TestBaseURIs(t *testing.T) {
	tests := map[string]string{
		"global": "https://api.bybit.com",
		"NL":     "https://api.bybit.nl",
		"tr":     "https://api.bybit-tr.com",
		"demo":   "https://api-demo.bybit.com",
	}
	for region, want := range tests {
		client, err := NewClient(ClientConfig{Region: region})
		if err != nil {
			t.Fatal(err)
		}
		if got := client.BaseURI(); got != want {
			t.Errorf("BaseURI(%q) = %q, want %q", region, got, want)
		}
	}
}

func TestRequestBuildsSignedGET(t *testing.T) {
	client, err := NewClient(ClientConfig{
		APIKey:    "key",
		APISecret: "secret",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet || req.URL.Path != "/v5/market/tickers" {
				t.Fatalf("unexpected request: %s %s", req.Method, req.URL)
			}
			if got, want := req.URL.RawQuery, "category=linear&symbol=BTCUSDT"; got != want {
				t.Fatalf("query = %q, want %q", got, want)
			}
			timestamp := req.Header.Get("X-BAPI-TIMESTAMP")
			mac := hmac.New(sha256.New, []byte("secret"))
			mac.Write([]byte(timestamp + "key" + "5000" + req.URL.RawQuery))
			if got, want := req.Header.Get("X-BAPI-SIGN"), hex.EncodeToString(mac.Sum(nil)); got != want {
				t.Fatalf("signature = %q, want %q", got, want)
			}
			if got := req.Header.Get("X-BAPI-SIGN-TYPE"); got != "2" {
				t.Fatalf("sign type = %q, want 2", got)
			}
			return response(req, http.StatusOK, `{"retCode":0,"result":{"list":[]}}`), nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.GetTickers(map[string]interface{}{"symbol": "BTCUSDT", "category": "linear"})
	if err != nil || result["retCode"].(float64) != 0 {
		t.Fatalf("GetTickers() = %#v, %v", result, err)
	}
}

func TestRequestReturnsUsefulHTTPAndDecodeErrors(t *testing.T) {
	tests := []struct {
		name string
		code int
		body string
		want string
	}{
		{"http error", http.StatusTooManyRequests, `{"retMsg":"rate limit"}`, "Too Many Requests"},
		{"invalid JSON", http.StatusOK, "not-json", "decode Bybit response"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(ClientConfig{HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return response(req, tt.code, tt.body), nil
			})}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.GetServerTime()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestSetLeverageRejectsInvalidValuesBeforeRequest(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.SetLeverage("linear", "BTCUSDT", 0, nil); err == nil {
		t.Fatal("SetLeverage accepted zero leverage")
	}
	side := "hold"
	if _, err := client.SetLeverage("linear", "BTCUSDT", 2, &side); err == nil {
		t.Fatal("SetLeverage accepted an invalid side")
	}
}
