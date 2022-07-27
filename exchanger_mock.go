package pricer

import (
	"errors"
	"strconv"
	"time"
)

type Mock struct {
	intervalProduce time.Duration
	startValue      float64
	stepValue       float64
	timeFinish      time.Time
	price           chan TickerPrice
	err             chan error
	ticker          Ticker
}

func NewMock(ticker Ticker, intervalProduce time.Duration, startValue float64, stepValue float64, timeFinish time.Time) *Mock {
	return &Mock{
		intervalProduce: intervalProduce,
		startValue:      startValue,
		stepValue:       stepValue,
		timeFinish:      timeFinish,
		price:           make(chan TickerPrice, 4*1024),
		err:             make(chan error),
		ticker:          ticker,
	}
}

func (m *Mock) SubscribePriceStream(tick Ticker) (chan TickerPrice, chan error) {
	return m.price, m.err
}

func (m *Mock) Run() {
	for time.Now().Before(m.timeFinish) {
		m.price <- TickerPrice{
			Ticker: m.ticker,
			Time:   time.Now(),
			Price:  strconv.FormatFloat(m.startValue, 'g', -1, 64),
		}
		m.startValue += m.stepValue
		time.Sleep(m.intervalProduce)
	}
	m.err <- errors.New("finish")
}
