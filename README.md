# bybit-go

[![CI](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tigusigalpa/bybit-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/tigusigalpa/bybit-go.svg)](https://pkg.go.dev/github.com/tigusigalpa/bybit-go)
[![Go version](https://img.shields.io/github/go-mod/go-version/tigusigalpa/bybit-go)](go.mod)
[![License](https://img.shields.io/github/license/tigusigalpa/bybit-go)](LICENSE)

Небольшой Go-клиент для [Bybit V5 API](https://bybit-exchange.github.io/docs/v5/intro). Он помогает быстрее начать работу с REST API, demo trading, WebSocket и TradFi-инструментами, не навязывая собственную архитектуру вашему приложению.

> Торговля связана с риском. Сначала проверяйте интеграцию и стратегию в demo-окружении, ограничивайте права API-ключа и никогда не добавляйте ключи в репозиторий.

## Возможности

- REST-методы для рыночных данных, ордеров, аккаунта и позиций;
- HMAC-SHA256 и RSA-SHA256 подписи для REST-запросов;
- demo-режим и региональные REST/WebSocket endpoints;
- публичные и приватные WebSocket-подписки;
- удобные обёртки для TradFi (forex, metals, stocks и indices);
- готовые, отдельные примеры в [`examples/`](examples/README.md).

## Установка

Требуется Go 1.21 или новее.

```bash
go get github.com/tigusigalpa/bybit-go
```

```go
import bybit "github.com/tigusigalpa/bybit-go"
```

## Быстрый старт

Публичные запросы тоже проходят через общий клиент; API-ключ для них не обязателен.

```go
package main

import (
	"fmt"
	"log"

	bybit "github.com/tigusigalpa/bybit-go"
)

func main() {
	client, err := bybit.NewClient(bybit.ClientConfig{Demo: true})
	if err != nil {
		log.Fatal(err)
	}

	tickers, err := client.GetTickers(map[string]interface{}{
		"category": "linear",
		"symbol":   "BTCUSDT",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", tickers["result"])
}
```

Для операций с аккаунтом передавайте ключи через переменные окружения:

```go
client, err := bybit.NewClient(bybit.ClientConfig{
	APIKey:    os.Getenv("BYBIT_API_KEY"),
	APISecret: os.Getenv("BYBIT_API_SECRET"),
	Demo:      true,
	Region:    "global", // также: nl, tr, kz, ge, ae
})
```

## Ордер

Параметры API передаются напрямую, поэтому новые поля Bybit можно использовать без ожидания новой версии SDK.

```go
order, err := client.CreateOrder(map[string]interface{}{
	"category":    "linear",
	"symbol":      "BTCUSDT",
	"side":        "Buy",
	"orderType":   "Limit",
	"qty":         "0.001",
	"price":       "30000",
	"timeInForce": "GTC",
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(order)
```

HTTP-ошибки (например, 429 или 401) возвращаются как `*bybit.HTTPError`; ответ Bybit с `retCode != 0` обычно приходит с HTTP 200 и остаётся в возвращаемой map. Всегда проверяйте `retCode` перед тем, как считать операцию успешной.

## RSA-подпись

Для RSA API-ключа укажите PEM приватного ключа. Если выбран `Signature: "rsa"`, ключ обязателен; неподдерживаемые типы подписи отклоняются при создании клиента.

```go
client, err := bybit.NewClient(bybit.ClientConfig{
	APIKey:        os.Getenv("BYBIT_API_KEY"),
	Signature:     "rsa",
	RSAPrivateKey: os.Getenv("BYBIT_RSA_PRIVATE_KEY"),
})
```

## WebSocket

```go
ws := bybit.NewWebSocket(bybit.WebSocketConfig{Demo: true})
ws.OnMessage(func(message map[string]interface{}) {
	fmt.Printf("%+v\n", message)
})

if err := ws.SubscribeTicker("BTCUSDT"); err != nil {
	log.Fatal(err)
}
if err := ws.Listen(); err != nil {
	log.Fatal(err)
}
defer ws.Close()
```

Публичный WebSocket этого пакета подключается к spot endpoint. Для приватного потока укажите `IsPrivate`, `APIKey` и `APISecret`. Приложение отвечает за жизненный цикл соединения и повторное подключение после ошибки.

## TradFi

Для популярных инструментов есть списки и вспомогательные методы:

```go
tickers, err := client.GetTradFiTicker("XAUUSD")
positions, err := client.GetTradFiPositions("XAUUSD")
order, err := client.PlaceTradFiOrder(bybit.TradFiOrderParams{
	Symbol: "XAUUSD", Side: "Buy", OrderType: "Market", Qty: "1",
})
```

Доступность символов и торговых условий зависит от региона и аккаунта. Получайте актуальный список через `GetTradFiInstruments` перед размещением ордера.

## Примеры и проверка проекта

Смотрите [каталог примеров](examples/README.md): basic client, market data, orders, positions, demo trading, TradFi и WebSocket. Примеры могут обращаться к сети или требовать ключи; перед запуском прочитайте их исходный код.

```bash
go test ./...
go vet ./...
```

CI запускает эти проверки на Go 1.21 и текущей стабильной версии Go. Pull request приветствуются: пожалуйста, добавляйте тест для изменения поведения и не включайте реальные ключи или персональные данные.

## Документация и лицензия

- [Официальная документация Bybit V5](https://bybit-exchange.github.io/docs/v5/intro)
- [Расширенная заметка по TradFi](wiki-tradfi.md)
- [Инструкции по установке](INSTALLATION.md)
- [MIT License](LICENSE)
