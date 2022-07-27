package pricer

import (
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Ticker string

const (
	BTCUSDTicker Ticker = "BTC_USD"
)

type TickerPrice struct {
	Ticker Ticker
	Time   time.Time
	Price  string // decimal value. example: "0", "10", "12.2", "13.2345122"
}

type PriceStreamSubscriber interface {
	SubscribePriceStream(Ticker) (chan TickerPrice, chan error)
}

type Exchanger struct {
	PriceStreamSubscriber
	Name   string
	Weight float64
}

type FairPricer struct {
	interval   time.Duration
	exchangers []Exchanger
	ticker     Ticker
	epsilon    float64
}

func New(interval time.Duration, ticker Ticker, exchangers ...Exchanger) *FairPricer {
	return &FairPricer{
		interval:   interval,
		exchangers: exchangers,
		ticker:     ticker,
	}
}

type outPrice struct {
	name  string
	time  time.Time
	price float64
}

func (fp *FairPricer) Run() {
	outResult := make(chan outPrice, len(fp.exchangers)*1024*8)
	errResult := make(chan error, len(fp.exchangers))
	wg := &sync.WaitGroup{}
	for _, exchanger := range fp.exchangers {
		wg.Add(1)
		exchange, err := exchanger.SubscribePriceStream(fp.ticker)
		go run(wg, exchanger.Name, exchange, err, outResult, errResult)
	}

	closedExchangersCount := 0
	tickInterval := time.NewTicker(fp.interval)
	now := time.Now()
	sumPricesByExchanger := make(map[string][]float64)
	fmt.Println("Timestamp, IndexPrice")
	for {
		select {
		case <-errResult:
			closedExchangersCount++
			if closedExchangersCount == len(fp.exchangers) {
				tickInterval.Stop()
				wg.Wait()
				fmt.Println("finish")

				return
			}
		case out := <-outResult:
			if out.time.After(now) && out.time.Before(now.Add(fp.interval)) {
				sumPricesByExchanger[out.name] = append(sumPricesByExchanger[out.name], out.price)
			}
		case <-tickInterval.C:
			sumMedians := 0.0
			countMedians := 0.0
			for _, exchanger := range fp.exchangers {
				if _, ok := sumPricesByExchanger[exchanger.Name]; ok {
					sumMedians += sumPricesByExchanger[exchanger.Name][len(sumPricesByExchanger[exchanger.Name])/2] * exchanger.Weight
					countMedians++
				}
			}
			fmt.Println(now.Add(fp.interval).Unix(), " ", sumMedians/countMedians)

			sumPricesByExchanger = make(map[string][]float64)
			now = time.Now()
		}
	}

}

func run(wg *sync.WaitGroup, name string, exchange chan TickerPrice, err chan error, outResult chan outPrice, errResult chan error) {
	defer wg.Done()
	for {
		select {
		case errCh := <-err:
			errResult <- errCh
			return
		case exchangePrice := <-exchange:
			res, errDecimal := decimal.NewFromString(exchangePrice.Price)
			if errDecimal != nil {
				continue
			}
			outResult <- outPrice{
				name:  name,
				time:  exchangePrice.Time,
				price: res.InexactFloat64(),
			}
		}
	}
}
