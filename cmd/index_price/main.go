package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ivanrybin/price_ticker/internal/task"
	"github.com/ivanrybin/price_ticker/mocks"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// mock subscribers for prices generation
	subscribers := []task.PriceStreamSubscriber{
		mocks.ConstantPriceStreamSubscriber(ctx, time.Second*1, 100),
		mocks.ConstantPriceStreamSubscriber(ctx, time.Second*1, 101),
		mocks.RandomPriceStreamSubscriber(ctx, time.Second*1, 98, 102),
	}

	// streams (wrappers on TickerPrices and errors channels)
	var streams []task.PriceStream
	for _, s := range subscribers {
		streams = append(streams, task.NewPriceStream(s.SubscribePriceStream(task.BTCUSDTicker)))
	}

	// time offset from current timestamp
	// when TickerPrice with TickerPrice.Time in window [timestamp - relevantOffset, timestamp] is valid
	relevantOffset := time.Minute

	// next minute timestamp
	now := time.Now()
	timestamp := time.Unix(now.Unix()+60-(now.Unix()%60), 0)

	// delay until next minute
	<-time.After(timestamp.Sub(now))

	// first index price
	if indexPrice, err := task.CalculateIndexPrice(timestamp, streams, relevantOffset, task.MeanPrice); err != nil {
		fmt.Printf("%d err: %v\n", timestamp.Unix(), err)
	} else {
		fmt.Printf("%d %.2f\n", timestamp.Unix(), indexPrice)
	}
	timestamp = timestamp.Add(time.Minute)

	// index price loop
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute):
			if indexPrice, err := task.CalculateIndexPrice(timestamp, streams, relevantOffset, task.MeanPrice); err != nil {
				fmt.Printf("%d err: %v\n", timestamp.Unix(), err)
			} else {
				fmt.Printf("%d %.2f\n", timestamp.Unix(), indexPrice)
			}
			timestamp = timestamp.Add(time.Minute)
		}
	}
}
