# Changelog

## Unreleased

### Fixed

- Hardened the existing `GetKlineReceipt` body lifecycle: context cancellation closes a blocked response body once, waits for cleanup, and preserves cancellation causes alongside read and close errors.
- Preserved both size-limit and original read errors, including non-2xx response evidence, and avoided integer overflow for large configured limits.
- Recorded the actual response request URL after redirects and retained response metadata when a redirect policy rejects a response already closed by `http.Client`.
- Restricted response headers to safe metadata and prevented automatic gzip decompression from changing captured body bytes.
- Shared request construction with legacy REST methods while preserving their signatures and decoding behavior.

### Added

- Added safe `KlineReceipt.String` and `GoString` diagnostic summaries.
- Added a runnable public no-key receipt example and offline acceptance tests for cancellation, deadlines, injected-client timeouts, body cleanup, redirects, exact payloads, and concurrent access.

## v1.5.0

### Added

- Added `GetKlineReceipt` for context-aware, byte-exact public REST kline receipts.
- Added bounded receipt capture, SHA-256 body digests, safe response headers, lifecycle timestamps, and evidence for HTTP/read/close failures.

### Notes

- This additive API does not interpret Bybit `retCode` and does not add retries or automatic provider activation.
