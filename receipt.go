package bybit

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// URL returns the exact URL, including the encoded query, sent by the client.
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

// Complete reports whether the entire body was read and closed without an error.
func (r *KlineReceipt) Complete() bool { return r.complete }

// ResponseBodySHA256 returns the SHA-256 digest of the exact captured response bytes.
func (r *KlineReceipt) ResponseBodySHA256() [sha256.Size]byte { return sha256.Sum256(r.responseBody) }

// GetKlineReceipt performs a public GET /v5/market/kline request and returns exact response evidence.
// It does not decode JSON or interpret Bybit retCode; application-level validation remains the caller's responsibility.
func (c *Client) GetKlineReceipt(ctx context.Context, params map[string]interface{}) (*KlineReceipt, error) {
	if ctx == nil {
		return nil, fmt.Errorf("kline receipt context must not be nil")
	}
	url := c.BaseURI() + "/v5/market/kline"
	if len(params) > 0 {
		url += "?" + c.buildQuery(params)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// Kline is a public endpoint. Do not add API credentials or signatures to this request.
	req.Header.Set("User-Agent", "bybit-go/1.0.0")
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	receipt := &KlineReceipt{
		method:      req.Method,
		url:         req.URL.String(),
		requestBody: []byte{},
		statusCode:  resp.StatusCode,
		status:      resp.Status,
		headers:     safeResponseHeaders(resp.Header),
		capturedAt:  time.Now().UTC(),
	}

	body, readErr := readReceiptBody(resp.Body, c.receiptBodyLimit)
	closeErr := resp.Body.Close()
	receipt.responseBody = append([]byte(nil), body...)
	receipt.completedAt = time.Now().UTC()
	receipt.complete = readErr == nil && closeErr == nil

	var errList []error
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		errList = append(errList, &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: string(body)})
	}
	if readErr != nil {
		errList = append(errList, readErr)
	}
	if closeErr != nil {
		errList = append(errList, closeErr)
	}
	if len(errList) > 0 {
		return receipt, errors.Join(errList...)
	}
	return receipt, nil
}

func readReceiptBody(body io.Reader, limit int64) ([]byte, error) {
	limited := io.LimitReader(body, limit+1)
	data, err := io.ReadAll(limited)
	if int64(len(data)) > limit {
		return data[:limit], &ReceiptBodyLimitError{Limit: limit}
	}
	return data, err
}

func safeResponseHeaders(headers http.Header) http.Header {
	safe := make(http.Header)
	for key, values := range headers {
		lower := strings.ToLower(key)
		if lower == "authorization" || lower == "cookie" || lower == "set-cookie" || strings.HasPrefix(lower, "x-bapi-") {
			continue
		}
		safe[key] = append([]string(nil), values...)
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
