package client

import (
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/fx"
)

const (
	retryCount = 3
	timeout    = 5 * time.Second
	retryDelay = 50 * time.Millisecond
)

func NewHTTPClient() *resty.Client {
	return resty.New().
		SetTimeout(timeout).
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryDelay).
		SetHeader("Accept", "application/json")
}

var Module = fx.Module("httpclient",
	fx.Provide(NewHTTPClient),
)
