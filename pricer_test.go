package pricer

import (
	"testing"
	"time"
)

func TestFairPricer_Run(t *testing.T) {
	priceTicker1 := NewMock(BTCUSDTicker, time.Millisecond, 100.0, 0.01, time.Now().Add(50*time.Second))
	exchanger1 := Exchanger{
		Name:                  "exchanger1",
		Weight:                1.0,
		PriceStreamSubscriber: priceTicker1,
	}
	priceTicker2 := NewMock(BTCUSDTicker, time.Millisecond, 102.0, 0.02, time.Now().Add(30*time.Second))
	exchanger2 := Exchanger{
		Name:                  "exchanger2",
		Weight:                1.0,
		PriceStreamSubscriber: priceTicker2,
	}

	fairPricer := New(10*time.Second, BTCUSDTicker, exchanger1, exchanger2)
	go priceTicker1.Run()
	go priceTicker2.Run()
	fairPricer.Run()
}
