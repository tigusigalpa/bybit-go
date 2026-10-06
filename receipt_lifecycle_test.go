package bybit

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestKlineReceiptCancellationClosesBlockedBody(t *testing.T) {
	cause := errors.New("caller stopped backfill")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	readErr := &receiptReadError{message: "read interrupted by close"}
	closeErr := errors.New("close diagnostic")
	body := &blockingReceiptBody{
		prefix: []byte("partial"), started: make(chan struct{}), closed: make(chan struct{}),
		readErr: readErr, closeErr: closeErr,
	}
	client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
		return receiptResponse(req, http.StatusOK, body), nil
	})
	result := startReceiptCall(client, ctx)
	select {
	case <-body.started:
	case <-time.After(3 * time.Second):
		t.Fatal("body did not start reading")
	}
	cancel(cause)
	outcome := awaitReceiptCall(t, result)
	var original *receiptReadError
	if !errors.Is(outcome.err, context.Canceled) || !errors.Is(outcome.err, cause) || !errors.As(outcome.err, &original) || original != readErr || !errors.Is(outcome.err, closeErr) {
		t.Fatalf("lost error causes: %v", outcome.err)
	}
	if outcome.receipt == nil || outcome.receipt.Complete() || string(outcome.receipt.ResponseBody()) != "partial" || body.closes.Load() != 1 {
		t.Fatalf("receipt=%v closes=%d", outcome.receipt, body.closes.Load())
	}
}

func TestKlineReceiptWaitsForCloseAndMarksCancellationAfterEOF(t *testing.T) {
	t.Run("completion after close", func(t *testing.T) {
		body := &gatedCloseBody{Reader: strings.NewReader("exact"), started: make(chan struct{}), release: make(chan struct{})}
		client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
			return receiptResponse(req, http.StatusOK, body), nil
		})
		result := startReceiptCall(client, context.Background())
		select {
		case <-body.started:
		case <-time.After(3 * time.Second):
			t.Fatal("body did not start closing")
		}
		select {
		case <-result:
			t.Fatal("receipt was returned before Close completed")
		default:
		}
		releasedAt := time.Now().UTC()
		close(body.release)
		outcome := awaitReceiptCall(t, result)
		if outcome.err != nil || !outcome.receipt.Complete() || outcome.receipt.CompletedAt().Before(releasedAt) || body.closes.Load() != 1 {
			t.Fatalf("receipt=%v err=%v closes=%d", outcome.receipt, outcome.err, body.closes.Load())
		}
	})
	t.Run("cancellation in close after EOF", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		body := &cancelOnCloseBody{Reader: strings.NewReader("complete bytes"), cancel: cancel}
		client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
			return receiptResponse(req, http.StatusOK, body), nil
		})
		receipt, err := client.GetKlineReceipt(ctx, nil)
		if receipt == nil || receipt.Complete() || !errors.Is(err, context.Canceled) || string(receipt.ResponseBody()) != "complete bytes" || body.closes != 1 {
			t.Fatalf("receipt=%v err=%v closes=%d", receipt, err, body.closes)
		}
	})
}

func TestKlineReceiptDeadlineAndHTTPClientTimeout(t *testing.T) {
	for _, mode := range []string{"context deadline", "injected client timeout"} {
		t.Run(mode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/v5/market/kline" {
					t.Errorf("path=%s", req.URL.Path)
				}
				if _, err := io.WriteString(w, "partial"); err != nil {
					return
				}
				w.(http.Flusher).Flush()
				<-req.Context().Done()
			}))
			defer server.Close()
			ctx := context.Background()
			httpClient := &http.Client{}
			if mode == "context deadline" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
				defer cancel()
			} else {
				httpClient.Timeout = 200 * time.Millisecond
			}
			client := localReceiptClient(t, httpClient, server.URL)
			receipt, err := client.GetKlineReceipt(ctx, klineReceiptParams())
			if receipt == nil || receipt.Complete() || string(receipt.ResponseBody()) != "partial" || !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("receipt=%v err=%v", receipt, err)
			}
			if mode == "injected client timeout" {
				var timeout net.Error
				if !errors.As(err, &timeout) || !timeout.Timeout() {
					t.Fatalf("client timeout cause lost: %v", err)
				}
			}
		})
	}
}

func TestKlineReceiptPreservesOverflowErrorAndHugeLimit(t *testing.T) {
	t.Run("overflow with original error and HTTP status", func(t *testing.T) {
		readErr := &receiptReadError{message: "error on overflow byte"}
		body := &probeErrorBody{readErr: readErr}
		client := receiptClient(t, 3, func(req *http.Request) (*http.Response, error) {
			return receiptResponse(req, http.StatusBadGateway, body), nil
		})
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		var limitErr *ReceiptBodyLimitError
		var httpErr *HTTPError
		if receipt == nil || receipt.Complete() || string(receipt.ResponseBody()) != "abc" || !errors.Is(err, readErr) || !errors.Is(err, ErrReceiptBodyLimitExceeded) || !errors.As(err, &limitErr) || limitErr.Limit != 3 || !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusBadGateway || httpErr.Body != "abc" || body.closes != 1 {
			t.Fatalf("receipt=%v err=%v closes=%d", receipt, err, body.closes)
		}
	})
	t.Run("maximum int64 limit", func(t *testing.T) {
		client := receiptClient(t, int64(^uint64(0)>>1), func(req *http.Request) (*http.Response, error) {
			return response(req, http.StatusOK, "actual bytes"), nil
		})
		receipt, err := client.GetKlineReceipt(context.Background(), nil)
		if err != nil || !receipt.Complete() || string(receipt.ResponseBody()) != "actual bytes" {
			t.Fatalf("receipt=%v err=%v", receipt, err)
		}
	})
}

func TestKlineReceiptActualURLAndRedirectErrors(t *testing.T) {
	redirectErr := errors.New("redirect policy rejected request")
	for _, mode := range []string{"follow", "return response", "reject"} {
		t.Run(mode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Query().Get("symbol") == "BTCUSDT" {
					http.Redirect(w, req, "/v5/market/kline?interval=1&symbol=ETHUSDT", http.StatusTemporaryRedirect)
					return
				}
				_, _ = io.WriteString(w, `{"retCode":0,"result":{"list":[]}}`)
			}))
			defer server.Close()
			httpClient := &http.Client{}
			switch mode {
			case "return response":
				httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
			case "reject":
				httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return redirectErr }
			}
			client := localReceiptClient(t, httpClient, server.URL)
			var closeCount atomic.Int32
			transport := httpClient.Transport
			httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				resp, err := transport.RoundTrip(req)
				if resp != nil {
					resp.Body = &countReceiptCloseBody{ReadCloser: resp.Body, closes: &closeCount}
				}
				return resp, err
			})
			receipt, err := client.GetKlineReceipt(context.Background(), klineReceiptParams())
			if receipt == nil {
				t.Fatalf("missing actual response evidence: %v", err)
			}
			switch mode {
			case "follow":
				if err != nil || receipt.URL() != server.URL+"/v5/market/kline?interval=1&symbol=ETHUSDT" || !receipt.Complete() {
					t.Fatalf("receipt=%v url=%s err=%v", receipt, receipt.URL(), err)
				}
			case "return response":
				var httpErr *HTTPError
				if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusTemporaryRedirect || !receipt.Complete() || len(receipt.ResponseBody()) == 0 {
					t.Fatalf("receipt=%v err=%v", receipt, err)
				}
			case "reject":
				var urlErr *url.Error
				if !errors.Is(err, redirectErr) || !errors.As(err, &urlErr) || receipt.Complete() || receipt.StatusCode() != http.StatusTemporaryRedirect || len(receipt.ResponseBody()) != 0 {
					t.Fatalf("receipt=%v err=%v", receipt, err)
				}
				if closeCount.Load() != 1 {
					t.Fatalf("redirect response body closed %d times, want one", closeCount.Load())
				}
			}
		})
	}
}

func TestKlineReceiptNoTransparentDecompression(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write([]byte(`{"retCode":0}`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Accept-Encoding") != "identity" {
			t.Errorf("Accept-Encoding=%q", req.Header.Get("Accept-Encoding"))
		}
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(compressed.Bytes())
	}))
	defer server.Close()
	client := localReceiptClient(t, &http.Client{}, server.URL)
	receipt, err := client.GetKlineReceipt(context.Background(), klineReceiptParams())
	if err != nil || !bytes.Equal(receipt.ResponseBody(), compressed.Bytes()) || receipt.ResponseHeaders().Get("Content-Encoding") != "gzip" || receipt.ResponseBodySHA256() != sha256.Sum256(compressed.Bytes()) {
		t.Fatalf("compressed response changed: receipt=%v err=%v", receipt, err)
	}
}

func TestKlineReceiptSafeDiagnosticsAndOwnedMetadata(t *testing.T) {
	secret := "synthetic-private-marker"
	headers := http.Header{
		"Content-Type": {"application/json"}, "X-Trace": {"trace-id"},
		"X-Bapi-Limit": {"600"}, "Authorization": {secret}, "Proxy-Authorization": {secret},
		"Set-Cookie": {secret}, "X-Api-Key": {secret}, "X-Bapi-Sign": {secret},
		"X-Secret-Custom": {secret},
	}
	var sent *http.Request
	client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
		sent = req
		resp := response(req, http.StatusOK, secret)
		resp.Header = headers
		return resp, nil
	})
	receipt, err := client.GetKlineReceipt(context.Background(), klineReceiptParams())
	if err != nil {
		t.Fatal(err)
	}
	actualURL := receipt.URL()
	headers["Content-Type"][0] = "changed"
	sent.URL.RawQuery = "api_key=" + secret
	if receipt.URL() != actualURL || receipt.ResponseHeaders().Get("Content-Type") != "application/json" || len(receipt.ResponseHeaders()) != 3 {
		t.Fatal("receipt retained mutable or unsafe metadata")
	}
	for _, summary := range []string{fmt.Sprint(receipt), fmt.Sprintf("%+v", receipt), fmt.Sprintf("%#v", receipt), fmt.Sprintf("%+v", *receipt), fmt.Sprintf("%#v", *receipt)} {
		if strings.Contains(summary, secret) || strings.Contains(summary, "api_key") || strings.Contains(summary, "Authorization") {
			t.Fatalf("unsafe diagnostic summary: %s", summary)
		}
	}
}

func TestKlineReceiptRejectsUndocumentedFieldsWithoutSending(t *testing.T) {
	client := receiptClient(t, 100, func(*http.Request) (*http.Response, error) {
		t.Fatal("transport used for invalid request")
		return nil, nil
	})
	if receipt, err := client.GetKlineReceipt(context.Background(), map[string]interface{}{"api_key": "synthetic-secret"}); receipt != nil || err == nil || strings.Contains(err.Error(), "synthetic-secret") {
		t.Fatalf("receipt=%v err=%v", receipt, err)
	}
	if receipt, err := client.GetKlineReceipt(nil, nil); receipt != nil || err == nil {
		t.Fatalf("receipt=%v err=%v", receipt, err)
	}
}

func TestKlineReceiptPreSendDeadlineCauseAndConfiguration(t *testing.T) {
	cause := errors.New("backfill deadline")
	ctx, cancel := context.WithDeadlineCause(context.Background(), time.Now().Add(-time.Second), cause)
	defer cancel()
	client := receiptClient(t, 100, func(*http.Request) (*http.Response, error) {
		t.Fatal("sent a request after deadline")
		return nil, nil
	})
	receipt, err := client.GetKlineReceipt(ctx, klineReceiptParams())
	if receipt != nil || !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, cause) {
		t.Fatalf("receipt=%v err=%v", receipt, err)
	}
	if _, err := NewClient(ClientConfig{ReceiptBodyLimit: -1}); err == nil {
		t.Fatal("negative receipt limit accepted")
	}
	defaultClient, err := NewClient(ClientConfig{})
	if err != nil || defaultClient.receiptBodyLimit != DefaultReceiptBodyLimit {
		t.Fatalf("default body limit changed: %v", err)
	}
}

func TestKlineReceiptLegacyBehaviorAndProviderValidation(t *testing.T) {
	// Legacy callers still decode maps and ignore close errors, while the raw
	// method returns bytes even for malformed JSON or non-zero provider retCode.
	for _, raw := range []string{`{"retCode":10001,"unknown":42}`, ` {"broken": `} {
		t.Run(raw, func(t *testing.T) {
			client := receiptClient(t, 100, func(req *http.Request) (*http.Response, error) {
				return response(req, http.StatusOK, raw), nil
			})
			receipt, err := client.GetKlineReceipt(context.Background(), klineReceiptParams())
			if err != nil || !receipt.Complete() || string(receipt.ResponseBody()) != raw {
				t.Fatalf("receipt=%v err=%v", receipt, err)
			}
			legacy, err := client.GetKline(klineReceiptParams())
			if json.Valid([]byte(raw)) {
				if err != nil || legacy["retCode"] != float64(10001) || legacy["unknown"] != float64(42) {
					t.Fatalf("legacy decoding changed: %v %v", legacy, err)
				}
			} else if err == nil {
				t.Fatal("legacy method did not report JSON decoding error")
			}
		})
	}
	client := receiptClient(t, 1, func(req *http.Request) (*http.Response, error) {
		return receiptResponse(req, http.StatusOK, &receiptTestBody{reader: strings.NewReader(`{"retCode":0}`), closeErr: errors.New("ignored legacy cleanup")}), nil
	})
	if _, err := client.Request("get", "/v5/market/kline", klineReceiptParams()); err != nil {
		t.Fatalf("legacy request applied receipt bounds or close semantics: %v", err)
	}
}

func TestKlineReceiptEndpointConfiguration(t *testing.T) {
	for _, config := range []ClientConfig{{}, {Demo: true}, {Region: "nl"}, {Region: "demo"}, {Region: "ae"}} {
		config.HTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return response(req, http.StatusOK, ""), nil
		})}
		client, err := NewClient(config)
		if err != nil {
			t.Fatal(err)
		}
		receipt, err := client.GetKlineReceipt(context.Background(), klineReceiptParams())
		if err != nil || !strings.HasPrefix(receipt.URL(), client.Endpoint()+"/v5/market/kline?") {
			t.Fatalf("endpoint=%s receipt=%v err=%v", client.Endpoint(), receipt, err)
		}
	}
}

type receiptOutcome struct {
	receipt *KlineReceipt
	err     error
}

func startReceiptCall(client *Client, ctx context.Context) <-chan receiptOutcome {
	result := make(chan receiptOutcome, 1)
	go func() {
		receipt, err := client.GetKlineReceipt(ctx, klineReceiptParams())
		result <- receiptOutcome{receipt: receipt, err: err}
	}()
	return result
}

func awaitReceiptCall(t *testing.T, result <-chan receiptOutcome) receiptOutcome {
	t.Helper()
	select {
	case outcome := <-result:
		return outcome
	case <-time.After(3 * time.Second):
		t.Fatal("receipt call did not finish")
		return receiptOutcome{}
	}
}

func klineReceiptParams() map[string]interface{} {
	return map[string]interface{}{"category": "linear", "symbol": "BTCUSDT", "interval": "1"}
}

func localReceiptClient(t *testing.T, httpClient *http.Client, serverURL string) *Client {
	t.Helper()
	target, err := url.Parse(serverURL)
	if err != nil {
		t.Fatal(err)
	}
	httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		local := req.Clone(req.Context())
		local.URL.Scheme = target.Scheme
		local.URL.Host = target.Host
		return http.DefaultTransport.RoundTrip(local)
	})
	client, err := NewClient(ClientConfig{HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func receiptResponse(req *http.Request, status int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: status, Status: fmt.Sprintf("%d %s", status, http.StatusText(status)), Body: body, Header: make(http.Header), Request: req}
}

type receiptReadError struct{ message string }

func (e *receiptReadError) Error() string { return e.message }

type blockingReceiptBody struct {
	prefix   []byte
	started  chan struct{}
	closed   chan struct{}
	start    sync.Once
	closes   atomic.Int32
	readErr  error
	closeErr error
}

func (b *blockingReceiptBody) Read(p []byte) (int, error) {
	if len(b.prefix) > 0 {
		n := copy(p, b.prefix)
		b.prefix = b.prefix[n:]
		return n, nil
	}
	b.start.Do(func() { close(b.started) })
	<-b.closed
	return 0, b.readErr
}

func (b *blockingReceiptBody) Close() error {
	if b.closes.Add(1) == 1 {
		close(b.closed)
	}
	return b.closeErr
}

type gatedCloseBody struct {
	io.Reader
	started chan struct{}
	release chan struct{}
	closes  atomic.Int32
}

func (b *gatedCloseBody) Close() error {
	b.closes.Add(1)
	close(b.started)
	<-b.release
	return nil
}

type cancelOnCloseBody struct {
	io.Reader
	cancel context.CancelFunc
	closes int
}

func (b *cancelOnCloseBody) Close() error {
	b.closes++
	b.cancel()
	return nil
}

type probeErrorBody struct {
	readErr error
	reads   int
	closes  int
}

func (b *probeErrorBody) Read(p []byte) (int, error) {
	b.reads++
	if b.reads == 1 {
		return copy(p, "abc"), nil
	}
	return copy(p, "d"), b.readErr
}

func (b *probeErrorBody) Close() error { b.closes++; return nil }

type countReceiptCloseBody struct {
	io.ReadCloser
	closes *atomic.Int32
}

func (b *countReceiptCloseBody) Close() error {
	b.closes.Add(1)
	return b.ReadCloser.Close()
}
