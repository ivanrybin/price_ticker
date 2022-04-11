package mocks

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ivanrybin/price_ticker/internal/task"
)

type priceStreamSubscriber struct {
	prices chan task.TickerPrice
	errors chan error
}

func (s *priceStreamSubscriber) SubscribePriceStream(task.Ticker) (chan task.TickerPrice, chan error) {
	return s.prices, s.errors
}

func ConstantPriceStreamSubscriber(ctx context.Context, tick time.Duration, price float64) task.PriceStreamSubscriber {
	s := &priceStreamSubscriber{
		prices: make(chan task.TickerPrice, 120),
		errors: make(chan error, 1),
	}
	s.prices <- task.TickerPrice{
		Ticker: task.BTCUSDTicker,
		Time:   time.Now(),
		Price:  fmt.Sprintf("%.6f", price),
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.errors <- fmt.Errorf("stream was closed")
				return
			case <-time.After(tick):
				s.prices <- task.TickerPrice{
					Ticker: task.BTCUSDTicker,
					Time:   time.Now(),
					Price:  fmt.Sprintf("%.6f", price),
				}
			}
		}
	}()
	return s
}

func RandomPriceStreamSubscriber(ctx context.Context, tick time.Duration, min, max float64) task.PriceStreamSubscriber {
	s := &priceStreamSubscriber{
		prices: make(chan task.TickerPrice, 120),
		errors: make(chan error, 1),
	}
	s.prices <- task.TickerPrice{
		Ticker: task.BTCUSDTicker,
		Time:   time.Now(),
		Price:  fmt.Sprintf("%.6f", min+rand.Float64()*(max-min)),
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.errors <- fmt.Errorf("stream was closed")
				return
			case <-time.After(tick):
				s.prices <- task.TickerPrice{
					Ticker: task.BTCUSDTicker,
					Time:   time.Now(),
					Price:  fmt.Sprintf("%.6f", min+rand.Float64()*(max-min)),
				}
			}
		}
	}()
	return s
}
