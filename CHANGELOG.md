# Changelog

## Unreleased

### Added

- Added `GetKlineReceipt` for context-aware, byte-exact public REST kline receipts.
- Added bounded receipt capture, SHA-256 body digests, safe response headers, lifecycle timestamps, and evidence for HTTP/read/close failures.

### Notes

- This additive API does not interpret Bybit `retCode` and does not add retries or automatic provider activation.
