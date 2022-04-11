package task

import (
	"fmt"
	"strconv"
	"time"
)

// PriceStream is TickerPrice and error channels wrapper with iterator functionality.
type PriceStream interface {
	// Next tries to get next TickerPrice from prices channel.
	// Returns true if successfully, false in other cases.
	Next() bool
	// TickerPrice not nil after successful Next call or nil if Next wasn't called or stream was closed.
	TickerPrice() *TickerPrice
	// Err returns not nil error if stream was closed.
	Err() error
}

// NewPriceStream creates new PriceStream.
func NewPriceStream(prices chan TickerPrice, errors chan error) PriceStream {
	return &priceStream{
		prices: prices,
		errors: errors,
	}
}

// priceStream implements PriceStream.
type priceStream struct {
	prices chan TickerPrice
	errors chan error

	tickerPrice *TickerPrice
	error       error
}

func (s *priceStream) TickerPrice() *TickerPrice {
	return s.tickerPrice
}

func (s *priceStream) Next() bool {
	if s.error != nil {
		return false
	}
	select {
	case price := <-s.prices:
		s.tickerPrice = &price
		return true
	case err := <-s.errors:
		s.error = err
		return false
	default:
		return false
	}
}

func (s *priceStream) Err() error {
	return s.error
}

// MostRelevantStreamTickerPrice get from stream the most relevant TickerPrice to t before or equal t or the first after t.
func MostRelevantStreamTickerPrice(t time.Time, s PriceStream) (*TickerPrice, bool) {
	if s.TickerPrice() == nil && !s.Next() {
		return nil, false
	}
	tickerPrice := s.TickerPrice()
	for tickerPrice.Time.Before(t) || tickerPrice.Time.Equal(t) {
		tickerPrice = s.TickerPrice()
		if !s.Next() {
			break
		}
	}
	return tickerPrice, true
}

// CalculateIndexPrice calculates index price based on most relevant TickerPrice from open streams
// which TickerPrice.Time is <= t and TickerPrice.Time in window [t - relevantOffset:t].
func CalculateIndexPrice(t time.Time, ss []PriceStream, relevantOffset time.Duration, indexFunc func(prices []string) (float64, error)) (float64, error) {
	prices := make([]string, 0, len(ss))
	for _, stream := range ss {
		tp, ok := MostRelevantStreamTickerPrice(t, stream)
		if ok && (tp.Time.Before(t) || tp.Time.Equal(t)) && tp.Time.Sub(t) <= relevantOffset {
			// ok: Time in [t - relevantOffset:t]
			prices = append(prices, tp.Price)
		}
	}
	return indexFunc(prices)
}

// MeanPrice calculates mean from prices.
func MeanPrice(prices []string) (float64, error) {
	if len(prices) == 0 {
		return 0.0, fmt.Errorf("no prices for calculatation")
	}
	n := float64(len(prices))
	sum := 0.0
	for i, p := range prices {
		price, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return 0.0, fmt.Errorf("prices[%v]='%v' isn't a number: %w", price, i, err)
		}
		sum += price
	}
	return sum / n, nil
}
