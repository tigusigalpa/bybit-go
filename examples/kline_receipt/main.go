// This example performs one public REST kline request when run explicitly.
// stdout contains only captured body bytes; diagnostic metadata goes to stderr.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
	client, err := bybit.NewClient(bybit.ClientConfig{ReceiptBodyLimit: 4 << 20})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Could not configure receipt client")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	receipt, captureErr := client.GetKlineReceipt(ctx, map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
		"interval": "1",
		"start":    int64(1700000000000),
		"end":      int64(1700000059999),
		"limit":    200,
	})
	if receipt != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s sha256=%x\n", receipt, receipt.ResponseBodySHA256())
		if _, err := os.Stdout.Write(receipt.ResponseBody()); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Could not write captured bytes")
			os.Exit(1)
		}
	}
	if captureErr != nil {
		// Do not print the whole error: HTTPError may include a response payload.
		_, _ = fmt.Fprintf(os.Stderr, "Capture failed (%T); preserve available evidence and inspect the error in your application\n", captureErr)
		os.Exit(1)
	}
	_, _ = fmt.Fprintln(os.Stderr, "Validate JSON and Bybit retCode separately before using these candles")
}
