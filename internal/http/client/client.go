package client

import (
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	maxConnsPerHost = 40
	retryCount      = 3
	timeout         = 5 * time.Second
	retryDelay      = 5 * time.Millisecond
)

func NewHttpClient() *resty.Client {
	client := resty.New().
		SetTimeout(timeout).
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryDelay)

	return client
}
