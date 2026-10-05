package bybit

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGetKlineReceiptPreservesExactPublicResponse(t *testing.T) {
	raw := []byte(" {\n \"unknown\": 1e-999, \"result\":{\"list\":[[\"1700000000000\",\"0.00000000000000000001\",1E+9]]}, \"retCode\":0 } \n")
	client, err := NewClient(ClientConfig{APIKey: "must-not-send", APISecret: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet || req.URL.Path != "/v5/market/kline" {
			t.Fatalf("request = %s %s", req.Method, req.URL.Path)
		}
		if got, want := req.URL.RawQuery, "category=linear&end=1700000059999&interval=1&limit=200&start=1700000000000&symbol=BTCUSDT"; got != want {
			t.Fatalf("query = %q, want %q", got, want)
		}
		if req.Header.Get("X-BAPI-API-KEY") != "" || req.Header.Get("X-BAPI-SIGN") != "" {
			t.Fatal("public receipt request sent credentials")
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{"X-Trace": {"abc"}, "Set-Cookie": {"secret"}}, Body: io.NopCloser(bytes.NewReader(raw)), Request: req}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := client.GetKlineReceipt(context.Background(), map[string]interface{}{"symbol": "BTCUSDT", "category": "linear", "interval": "1", "start": "1700000000000", "end": "1700000059999", "limit": 200})
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Complete() || receipt.Method() != http.MethodGet || len(receipt.RequestBody()) != 0 {
		t.Fatalf("unexpected receipt state: complete=%t method=%s", receipt.Complete(), receipt.Method())
	}
	if !bytes.Equal(receipt.ResponseBody(), raw) {
		t.Fatal("response bytes were not preserved")
	}
	if got, want := receipt.ResponseBodySHA256(), sha256.Sum256(raw); got != want {
		t.Fatal("response hash differs from exact bytes")
	}
	if receipt.ResponseHeaders().Get("Set-Cookie") != "" || receipt.ResponseHeaders().Get("X-Trace") != "abc" {
		t.Fatal("response header filtering is incorrect")
	}
	if receipt.CapturedAt().Location() != time.UTC || receipt.CompletedAt().Location() != time.UTC || receipt.CompletedAt().Before(receipt.CapturedAt()) {
		t.Fatal("receipt timestamps are not honest UTC lifecycle times")
	}
	body := receipt.ResponseBody()
	body[0] = 'x'
	headers := receipt.ResponseHeaders()
	headers.Set("X-Trace", "changed")
	if !bytes.Equal(receipt.ResponseBody(), raw) || receipt.ResponseHeaders().Get("X-Trace") != "abc" {
		t.Fatal("receipt accessors leaked mutable state")
	}
}

func TestGetKlineReceiptFailureEvidence(t *testing.T) {
	t.Run("http status", func(t *testing.T) {
		client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
			return response(req, http.StatusTooManyRequests, " {\"retCode\":10006} "), nil
		})
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		var httpErr *HTTPError
		if receipt == nil || !receipt.Complete() || !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("receipt=%#v err=%v", receipt, err)
		}
	})
	t.Run("exact limit plus EOF", func(t *testing.T) {
		client := receiptClient(t, 3, func(req *http.Request) (*http.Response, error) { return response(req, http.StatusOK, "abc"), nil })
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		if err != nil || !receipt.Complete() || string(receipt.ResponseBody()) != "abc" {
			t.Fatalf("receipt=%#v err=%v", receipt, err)
		}
	})
	t.Run("limit overflow", func(t *testing.T) {
		client := receiptClient(t, 3, func(req *http.Request) (*http.Response, error) { return response(req, http.StatusOK, "abcd"), nil })
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		if receipt == nil || receipt.Complete() || !errors.Is(err, ErrReceiptBodyLimitExceeded) || string(receipt.ResponseBody()) != "abc" {
			t.Fatalf("receipt=%#v err=%v", receipt, err)
		}
	})
	t.Run("empty response", func(t *testing.T) {
		client := receiptClient(t, 3, func(req *http.Request) (*http.Response, error) { return response(req, http.StatusOK, ""), nil })
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		if receipt == nil || err != nil || !receipt.Complete() || len(receipt.ResponseBody()) != 0 {
			t.Fatalf("receipt=%#v err=%v", receipt, err)
		}
	})
	t.Run("no response", func(t *testing.T) {
		transportErr := errors.New("network down")
		client := receiptClient(t, 3, func(*http.Request) (*http.Response, error) { return nil, transportErr })
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		if receipt != nil || !errors.Is(err, transportErr) {
			t.Fatalf("receipt=%#v err=%v", receipt, err)
		}
	})
}

func TestGetKlineReceiptReadCloseAndContextErrors(t *testing.T) {
	t.Run("partial read and close once", func(t *testing.T) {
		readErr := errors.New("read failed")
		body := &receiptTestBody{reader: strings.NewReader("partial"), readErr: readErr}
		client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: make(http.Header), Body: body, Request: req}, nil
		})
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		if receipt == nil || receipt.Complete() || !errors.Is(err, readErr) || body.closes != 1 || string(receipt.ResponseBody()) != "partial" {
			t.Fatalf("receipt=%#v err=%v closes=%d", receipt, err, body.closes)
		}
	})
	t.Run("close error", func(t *testing.T) {
		closeErr := errors.New("close failed")
		body := &receiptTestBody{reader: strings.NewReader("ok"), closeErr: closeErr}
		client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: make(http.Header), Body: body, Request: req}, nil
		})
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		if receipt == nil || receipt.Complete() || !errors.Is(err, closeErr) || body.closes != 1 {
			t.Fatalf("receipt=%#v err=%v closes=%d", receipt, err, body.closes)
		}
	})
	t.Run("cancelled before send", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		client := receiptClient(t, 100, func(*http.Request) (*http.Response, error) {
			t.Fatal("transport called after cancellation")
			return nil, nil
		})
		receipt, err := client.GetKlineReceipt(ctx, nil)
		if receipt != nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("receipt=%#v err=%v", receipt, err)
		}
	})
	t.Run("cancelled during read", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		body := &contextReceiptBody{ctx: ctx, started: make(chan struct{})}
		client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: make(http.Header), Body: body, Request: req}, nil
		})
		result := make(chan struct {
			receipt *KlineReceipt
			err     error
		}, 1)
		go func() {
			receipt, err := client.GetKlineReceipt(ctx, nil)
			result <- struct {
				receipt *KlineReceipt
				err     error
			}{receipt, err}
		}()
		<-body.started
		cancel()
		outcome := <-result
		if outcome.receipt == nil || outcome.receipt.Complete() || !errors.Is(outcome.err, context.Canceled) || body.closes != 1 {
			t.Fatalf("receipt=%#v err=%v closes=%d", outcome.receipt, outcome.err, body.closes)
		}
	})
}

func TestKlineReceiptAccessorsAreConcurrentSafe(t *testing.T) {
	client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) { return response(req, http.StatusOK, "payload"), nil })
	receipt, err := client.GetKlineReceipt(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := receipt.ResponseBody()
			if len(body) > 0 {
				body[0] = 'x'
			}
			receipt.ResponseHeaders().Set("X-Test", "x")
			_ = receipt.ResponseBodySHA256()
		}()
	}
	wg.Wait()
	if string(receipt.ResponseBody()) != "payload" {
		t.Fatal("concurrent accessor mutation leaked into receipt")
	}
}

func receiptClient(t *testing.T, limit int64, transport roundTripFunc) *Client {
	t.Helper()
	client, err := NewClient(ClientConfig{ReceiptBodyLimit: limit, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

type receiptTestBody struct {
	reader            io.Reader
	readErr, closeErr error
	closes            int
}

func (b *receiptTestBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	if err == io.EOF && b.readErr != nil {
		return n, b.readErr
	}
	return n, err
}
func (b *receiptTestBody) Close() error { b.closes++; return b.closeErr }

type contextReceiptBody struct {
	ctx     context.Context
	started chan struct{}
	closes  int
}

func (b *contextReceiptBody) Read([]byte) (int, error) {
	close(b.started)
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}
func (b *contextReceiptBody) Close() error { b.closes++; return nil }
