package main

import (
	"context"
	"fmt"
	"hw-async/domain"
	"hw-async/generator"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}
var logger = logrus.New()

func formCandles1m(pr <-chan domain.Price, wg *sync.WaitGroup) <-chan domain.Candle {
	candles := make(chan domain.Candle)
	var candlessMap = make(map[string]domain.Candle)
	go func() {
		defer func() {
			logger.Info("Generating 1m candles after SIGINT...")
			for _, out := range candlessMap {
				candles <- out
			}
			close(candles)
			wg.Done()
		}()
		for j := range pr {
			t, _ := domain.PeriodTS(domain.CandlePeriod1m, j.TS)
			if candlessMap[j.Ticker].TS == t {
				outCandle := candlessMap[j.Ticker]
				if outCandle.High < j.Value {
					outCandle.High = j.Value
				}
				if outCandle.Low > j.Value {
					outCandle.Low = j.Value
				}
				outCandle.Close = j.Value
				candlessMap[j.Ticker] = outCandle
			} else {
				if _, ok := candlessMap[j.Ticker]; ok {
					candles <- candlessMap[j.Ticker]
				}
				candlessMap[j.Ticker] = domain.Candle{Ticker: j.Ticker, Period: domain.CandlePeriod1m, Open: j.Value, High: j.Value, Low: j.Value, Close: j.Value, TS: t}
			}
		}
	}()
	return candles
}

func formCandles2m(pr <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	candles := make(chan domain.Candle)
	go func() {
		defer wg.Done()
		var candlessMap = make(map[string]domain.Candle)
		for j := range pr {
			t, _ := domain.PeriodTS(domain.CandlePeriod2m, j.TS)
			if candlessMap[j.Ticker].TS == t {
				outCandle := candlessMap[j.Ticker]
				if outCandle.High < j.High {
					outCandle.High = j.High
				}
				if outCandle.Low > j.Low {
					outCandle.Low = j.Low
				}
				outCandle.Close = j.Close
				candlessMap[j.Ticker] = outCandle
			} else {
				if _, ok := candlessMap[j.Ticker]; ok {
					candles <- candlessMap[j.Ticker]
				}
				candlessMap[j.Ticker] = domain.Candle{Ticker: j.Ticker, Period: domain.CandlePeriod2m, Open: j.Open, High: j.High, Low: j.Low, Close: j.Close, TS: t}
			}
		}
		logger.Info("Generating 2m candles after SIGINT...")
		for _, out := range candlessMap {
			candles <- out
		}
		close(candles)
	}()
	return candles
}

func formCandles10m(pr <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	candles := make(chan domain.Candle)
	go func() {
		defer wg.Done()
		var candlessMap = make(map[string]domain.Candle)
		for j := range pr {
			t, _ := domain.PeriodTS(domain.CandlePeriod10m, j.TS)
			if candlessMap[j.Ticker].TS == t {
				outCandle := candlessMap[j.Ticker]
				if outCandle.High < j.High {
					outCandle.High = j.High
				}
				if outCandle.Low > j.Low {
					outCandle.Low = j.Low
				}
				outCandle.Close = j.Close
				candlessMap[j.Ticker] = outCandle
			} else {
				if _, ok := candlessMap[j.Ticker]; ok {
					candles <- candlessMap[j.Ticker]
				}
				candlessMap[j.Ticker] = domain.Candle{Ticker: j.Ticker, Period: domain.CandlePeriod10m, Open: j.Open, High: j.High, Low: j.Low, Close: j.Close, TS: t}
			}
		}
		logger.Info("Generating 10m candles after SIGINT...")
		for _, out := range candlessMap {
			candles <- out
		}
		close(candles)
	}()
	return candles
}

func makeOut1(pr <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	candles := make(chan domain.Candle)
	go func() {
		file, err := os.OpenFile("candles_1m.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		defer func() {
			logger.Info("Closing channel, which saves 1m candles after SIGINT...")
			wg.Done()
			close(candles)
			file.Close()
		}()
		if err != nil {
			panic(err)
		}
		for j := range pr {
			fmt.Fprintf(file, "%s,%v,%.6f,%.6f,%.6f,%.6f\n", j.Ticker, j.TS.Format(time.RFC3339), j.Open, j.High, j.Low, j.Close)
			candles <- j
		}
	}()
	return candles
}

func makeOut2(pr <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	candles := make(chan domain.Candle)
	go func() {
		file, err := os.OpenFile("candles_2m.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		defer func() {
			logger.Info("Closing channel, which saves 2m candles after SIGINT...")
			wg.Done()
			close(candles)
			file.Close()
		}()
		if err != nil {
			panic(err)
		}
		for j := range pr {
			fmt.Fprintf(file, "%s,%v,%.6f,%.6f,%.6f,%.6f\n", j.Ticker, j.TS.Format(time.RFC3339), j.Open, j.High, j.Low, j.Close)
			candles <- j
		}
	}()
	return candles
}

func makeOut10(pr <-chan domain.Candle, wg *sync.WaitGroup) {
	go func() {
		file, err := os.OpenFile("candles_10m.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		defer func() {
			logger.Info("Closing channel, which saves 10m candles after SIGINT...")
			wg.Done()
			file.Close()
		}()
		if err != nil {
			panic(err)
		}
		for j := range pr {
			fmt.Fprintf(file, "%s,%v,%.6f,%.6f,%.6f,%.6f\n", j.Ticker, j.TS.Format(time.RFC3339), j.Open, j.High, j.Low, j.Close)
		}
	}()
}

func main() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})
	logger.Info("start prices generator...")
	prices := pg.Prices(ctx)

	wg.Add(6)
	logger.Info("start candles generator...")
	candles := formCandles1m(prices, &wg)
	candles1 := makeOut1(candles, &wg)
	candles2out := formCandles2m(candles1, &wg)
	candles2 := makeOut2(candles2out, &wg)
	candles10 := formCandles10m(candles2, &wg)
	makeOut10(candles10, &wg)
	logger.Info("starting output to file...")

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	logger.Info("SIGINT detected, saving data...")
	cancel()

	wg.Wait()
	logger.Info("Finished")
}
