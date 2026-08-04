# Bybit Go SDK - Examples

This directory contains working examples demonstrating how to use the Bybit Go SDK.

## Prerequisites

1. Set up your environment variables:

```bash
export BYBIT_API_KEY="your_api_key"
export BYBIT_API_SECRET="your_api_secret"
```

2. Install dependencies:

```bash
cd ..
go mod tidy
```

## Running Examples

Each example is a standalone Go program. Run them individually:

### Basic Client Example

Demonstrates basic client initialization and server time retrieval.

```bash
go run basic_client/main.go
```

### Market Data Example

Shows how to fetch market data for multiple symbols.

```bash
go run market_data/main.go
```

### Order Management Example

Demonstrates order placement, retrieval, and management.

```bash
go run order_management/main.go
```

**Note:** This example requires valid API credentials with trading permissions.

### WebSocket Public Streams

Shows how to subscribe to public WebSocket streams (orderbook, trades, ticker, klines).

```bash
go run websocket_public/main.go
```

Press `Ctrl+C` to stop the WebSocket listener.

### WebSocket Private Streams

Demonstrates authenticated WebSocket connections for account updates.

```bash
go run websocket_private/main.go
```

**Note:** Requires valid API credentials.

Press `Ctrl+C` to stop the WebSocket listener.

## Important Notes

- **Demo Trading vs Mainnet**: Most examples use demo trading by default. Change `Demo: false` to use mainnet.
- **API Permissions**: Some examples require specific API key permissions (trading, wallet access).
- **Rate Limits**: Be mindful of Bybit's API rate limits when running examples repeatedly.
- **WebSocket Examples**: These run indefinitely until you press `Ctrl+C`.

## Example Structure

Each example follows this pattern:

1. Load environment variables
2. Initialize client/websocket
3. Perform operations
4. Display results
5. Handle errors gracefully

## Customization

Feel free to modify these examples to suit your needs. They serve as templates for building your own trading applications.
