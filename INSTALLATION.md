# Installation Guide - Bybit Go SDK

This guide will help you install and set up the Bybit Go SDK in your project.

## Requirements

- **Go 1.21 or higher**
- Active Bybit account (for API credentials)
- Basic understanding of Go programming

## Installation Methods

### Method 1: Using `go get` (Recommended)

```bash
go get github.com/tigusigalpa/bybit-go
```

### Method 2: Manual Installation

1. Clone or download the repository:

```bash
git clone https://github.com/tigusigalpa/bybit-go.git
```

2. Add to your `go.mod`:

```go
require github.com/tigusigalpa/bybit-go v1.0.0
```

3. Run:

```bash
go mod tidy
```

## Getting API Credentials

1. **Sign up** for a Bybit account at [https://www.bybit.com](https://www.bybit.com)

2. **Enable API access:**
   - Go to Account & Security → API Management
   - Create a new API key
   - Set appropriate permissions (Read, Trade, Wallet as needed)
   - Save your API Key and Secret securely

3. **For testing (recommended):**
   - Use Bybit Testnet: [https://testnet.bybit.com](https://testnet.bybit.com)
   - Create a testnet account (separate from mainnet)
   - Generate testnet API credentials
   - Fund your testnet account with virtual funds

## Configuration

### Environment Variables

Create a `.env` file or export environment variables:

```bash
export BYBIT_API_KEY="your_api_key_here"
export BYBIT_API_SECRET="your_api_secret_here"
export BYBIT_TESTNET="true"
export BYBIT_REGION="global"
```

### In Your Code

```go
package main

import (
    "log"
    "os"
    
    bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
    client, err := bybit.NewClient(bybit.ClientConfig{
        APIKey:     os.Getenv("BYBIT_API_KEY"),
        APISecret:  os.Getenv("BYBIT_API_SECRET"),
        Testnet:    true,
        Region:     "global",
        RecvWindow: 5000,
        Signature:  "hmac",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the client
    serverTime, err := client.GetServerTime()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Server time: %v", serverTime)
}
```

## Verifying Installation

Run this simple test to verify everything is working:

```go
package main

import (
    "fmt"
    "log"
    
    bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
    client, err := bybit.NewClient(bybit.ClientConfig{
        APIKey:     "test",
        APISecret:  "test",
        Testnet:    true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    time, err := client.GetServerTime()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("✅ Installation successful! Server time: %v\n", time)
}
```

Save as `test.go` and run:

```bash
go run test.go
```

## Troubleshooting

### "package github.com/tigusigalpa/bybit-go is not in GOROOT"

**Solution:** Run `go mod tidy` or `go get github.com/tigusigalpa/bybit-go`

### "cannot find module providing package github.com/gorilla/websocket"

**Solution:** Run `go mod tidy` to download all dependencies

### "API signature verification failed"

**Solution:** 
- Verify your API key and secret are correct
- Check that you're using the right environment (testnet vs mainnet)
- Ensure your system time is synchronized

### "Rate limit exceeded"

**Solution:**
- Reduce request frequency
- Implement exponential backoff
- Check Bybit's rate limit documentation

## Next Steps

1. ✅ Read the [README.md](README.md) for feature overview
2. ✅ Check the [examples/](examples/) directory for usage examples
3. ✅ Review [Bybit API Documentation](https://bybit-exchange.github.io/docs/v5/guide)
4. ✅ Start building your trading application!

## Security Best Practices

- ⚠️ **Never commit API keys to version control**
- ⚠️ Use environment variables or secure vaults for credentials
- ⚠️ Test on testnet before using mainnet
- ⚠️ Set appropriate API key permissions (principle of least privilege)
- ⚠️ Regularly rotate API keys
- ⚠️ Monitor API usage and set up alerts

## Support

- **Issues:** [GitHub Issues](https://github.com/tigusigalpa/bybit-go/issues)
- **Email:** sovletig@gmail.com
- **Bybit API Docs:** [https://bybit-exchange.github.io/docs/v5/guide](https://bybit-exchange.github.io/docs/v5/guide)
