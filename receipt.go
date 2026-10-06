package bybit

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ErrReceiptBodyLimitExceeded indicates that a raw receipt exceeded its configured body limit.
var ErrReceiptBodyLimitExceeded = errors.New("raw receipt body limit exceeded")

// ReceiptBodyLimitError describes a response that exceeded the configured receipt limit.
type ReceiptBodyLimitError struct {
	Limit int64
}

// Error implements error.
func (e *ReceiptBodyLimitError) Error() string {
	return fmt.Sprintf("%s: limit is %d bytes", ErrReceiptBodyLimitExceeded, e.Limit)
}

// Unwrap makes ReceiptBodyLimitError compatible with errors.Is.
func (e *ReceiptBodyLimitError) Unwrap() error { return ErrReceiptBodyLimitExceeded }

// KlineReceipt preserves an observed HTTP response for a public kline request.
// Its accessors return defensive copies. A complete receipt only means the HTTP
// body was fully read and closed; callers must still inspect Bybit's retCode.
type KlineReceipt struct {
	method       string
	url          string
	requestBody  []byte
	statusCode   int
	status       string
	responseBody []byte
	headers      http.Header
	capturedAt   time.Time
	completedAt  time.Time
	complete     bool
}

// Method returns the observed HTTP method.
func (r *KlineReceipt) Method() string { return r.method }

// URL returns the response's actual request URL, including the encoded query.
// If the injected HTTP client follows redirects, this is the final request URL.
func (r *KlineReceipt) URL() string { return r.url }

// RequestBody returns a copy of the request body. It is empty for GET requests.
func (r *KlineReceipt) RequestBody() []byte { return append([]byte(nil), r.requestBody...) }

// StatusCode returns the observed HTTP status code.
func (r *KlineReceipt) StatusCode() int { return r.statusCode }

// Status returns the observed HTTP status text.
func (r *KlineReceipt) Status() string { return r.status }

// ResponseBody returns a copy of the exact bytes read from the response body.
func (r *KlineReceipt) ResponseBody() []byte { return append([]byte(nil), r.responseBody...) }

// ResponseHeaders returns a copy of safe response headers. Sensitive headers are excluded.
func (r *KlineReceipt) ResponseHeaders() http.Header { return cloneHeader(r.headers) }

// CapturedAt returns when the HTTP response became available, in UTC.
func (r *KlineReceipt) CapturedAt() time.Time { return r.capturedAt }

// CompletedAt returns when reading and closing the original body completed, in UTC.
func (r *KlineReceipt) CompletedAt() time.Time { return r.completedAt }

// Complete reports whether EOF was observed within the limit and the body was
// closed without errors or context cancellation. It does not imply HTTP success.
func (r *KlineReceipt) Complete() bool { return r.complete }

// ResponseBodySHA256 returns the SHA-256 digest of the exact captured response bytes.
func (r *KlineReceipt) ResponseBodySHA256() [sha256.Size]byte { return sha256.Sum256(r.responseBody) }

// String returns a diagnostic summary without payloads, URLs, or header values.
func (r KlineReceipt) String() string {
	return fmt.Sprintf("KlineReceipt{method=%s status=%d bytes=%d complete=%t captured_at=%s completed_at=%s}",
		r.method, r.statusCode, len(r.responseBody), r.complete,
		r.capturedAt.Format(time.RFC3339Nano), r.completedAt.Format(time.RFC3339Nano))
}

// GoString uses the same safe summary for Go-syntax diagnostic formatting.
func (r KlineReceipt) GoString() string { return r.String() }

// GetKlineReceipt performs a public GET /v5/market/kline request and returns exact response evidence.
// Only category, symbol, interval, start, end, and limit request fields are accepted.
// It does not decode JSON or interpret Bybit retCode; application-level validation remains the caller's responsibility.
func (c *Client) GetKlineReceipt(ctx context.Context, params map[string]interface{}) (*KlineReceipt, error) {
	if ctx == nil {
		return nil, fmt.Errorf("kline receipt context must not be nil")
	}
	if err := receiptContextError(ctx); err != nil {
		return nil, err
	}
	for key := range params {
		switch key {
		case "category", "symbol", "interval", "start", "end", "limit":
		default:
			return nil, fmt.Errorf("unsupported kline request field %q", key)
		}
	}
	req, err := c.newRequest(ctx, http.MethodGet, "/v5/market/kline", params, false)
	if err != nil {
		return nil, err
	}
	// Suppress net/http's transparent gzip decompression so hashes describe the
	// original response body. An injected transport must also preserve these bytes.
	req.Header.Set("Accept-Encoding", "identity")

	resp, doErr := c.httpClient.Do(req)
	capturedAt := time.Now().UTC()
	if resp == nil {
		return nil, errors.Join(doErr, receiptContextError(ctx))
	}
	actualRequest := req
	if resp.Request != nil {
		actualRequest = resp.Request
	}
	receipt := &KlineReceipt{
		method:     actualRequest.Method,
		url:        actualRequest.URL.String(),
		statusCode: resp.StatusCode,
		status:     resp.Status,
		headers:    safeResponseHeaders(resp.Header),
		capturedAt: capturedAt,
	}
	if doErr != nil {
		// http.Client.Do returns a response with an error only when CheckRedirect
		// fails. It has already closed that body; do not read or close it again.
		receipt.completedAt = time.Now().UTC()
		return receipt, errors.Join(doErr, receiptContextError(ctx), receiptHTTPError(resp, nil))
	}

	body, lifecycleErr := captureReceiptBody(ctx, resp.Body, c.receiptBodyLimit)
	receipt.completedAt = time.Now().UTC()
	receipt.responseBody = body
	receipt.complete = lifecycleErr == nil
	return receipt, errors.Join(lifecycleErr, receiptHTTPError(resp, body))
}

func receiptHTTPError(resp *http.Response, body []byte) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	return &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: string(body)}
}

func receiptContextError(ctx context.Context) error {
	if ctx.Err() == nil {
		return nil
	}
	cause := context.Cause(ctx)
	if errors.Is(cause, ctx.Err()) {
		return cause
	}
	return errors.Join(ctx.Err(), cause)
}

// captureReceiptBody closes the original body once, including on cancellation.
// Waiting for the cancellation callback ensures no close or receipt mutation
// remains in progress when a completed receipt becomes visible to the caller.
func captureReceiptBody(ctx context.Context, body io.ReadCloser, limit int64) ([]byte, error) {
	var once sync.Once
	var closeErr error
	closeBody := func() {
		once.Do(func() { closeErr = body.Close() })
	}
	closed := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		closeBody()
		close(closed)
	})
	data, readErr := readReceiptBody(body, limit)
	if !stop() {
		<-closed
	}
	closeBody()
	return data, errors.Join(readErr, closeErr, receiptContextError(ctx))
}

func readReceiptBody(body io.Reader, limit int64) ([]byte, error) {
	// Read at most limit bytes, then probe one extra byte to distinguish an exact
	// limit followed by EOF from overflow. This also avoids limit+1 int overflow.
	limited := &io.LimitedReader{R: body, N: limit}
	data, err := io.ReadAll(limited)
	if err != nil || limited.N > 0 {
		return data, err
	}
	var probe [1]byte
	for attempts := 0; attempts < 100; attempts++ {
		n, err := body.Read(probe[:])
		if n > 0 {
			if err == io.EOF {
				err = nil
			}
			return data, errors.Join(&ReceiptBodyLimitError{Limit: limit}, err)
		}
		if err == io.EOF {
			return data, nil
		}
		if err != nil {
			return data, err
		}
	}
	return data, io.ErrNoProgress
}

func safeResponseHeaders(headers http.Header) http.Header {
	safe := make(http.Header)
	for key, values := range headers {
		// Allow only transport metadata and documented rate-limit / correlation
		// identifiers; a denylist cannot cover arbitrary secret header names.
		switch strings.ToLower(key) {
		case "content-type", "content-length", "content-encoding", "date", "etag", "last-modified", "retry-after",
			"x-bapi-limit", "x-bapi-limit-status", "x-bapi-limit-reset-timestamp", "x-request-id", "x-trace":
			safe[http.CanonicalHeaderKey(key)] = append([]string(nil), values...)
		}
	}
	return safe
}

func cloneHeader(headers http.Header) http.Header {
	clone := make(http.Header, len(headers))
	for key, values := range headers {
		clone[key] = append([]string(nil), values...)
	}
	return clone
}
