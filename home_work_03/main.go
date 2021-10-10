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

func formCandles1m(ctx context.Context, pr <-chan domain.Price, wg *sync.WaitGroup) <-chan domain.Candle {
	candles := make(chan domain.Candle, 4)
	var candlessMap = make(map[string]domain.Candle)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				logger.Info("Generating 1m candles after SIGINT...")
				for _, out := range candlessMap {
					candles <- out
				}
				close(candles)
				return
			case j := <-pr:
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
		}
	}()
	return candles
}

func formCandles2m(ctx context.Context, pr <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	candles := make(chan domain.Candle, 4)
	go func() {
		defer wg.Done()
		var candlessMap = make(map[string]domain.Candle)
		for {
			select {
			case <-ctx.Done():
				logger.Info("Generating 2m candles after SIGINT...")
				for _, out := range candlessMap {
					candles <- out
				}
				close(candles)
				return
			case j := <-pr:
				candles <- j
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
		}
	}()
	return candles
}

func formCandles10m(ctx context.Context, pr <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	candles := make(chan domain.Candle, 4)
	go func() {
		defer wg.Done()
		var candlessMap = make(map[string]domain.Candle)
		for {
			select {
			case <-ctx.Done():
				logger.Info("Generating 10m candles after SIGINT...")
				for _, out := range candlessMap {
					candles <- out
				}
				close(candles)
				return
			case j := <-pr:
				candles <- j
				if j.Period == domain.CandlePeriod1m {
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
			}
		}
	}()
	return candles
}
func writeToFile(j domain.Candle, file, file2, file3 *os.File) {
	switch j.Period {
	case domain.CandlePeriod1m:
		fmt.Fprintf(file, "%s,%v,%.6f,%.6f,%.6f,%.6f\n", j.Ticker, j.TS.Format(time.RFC3339), j.Open, j.High, j.Low, j.Close)
	case domain.CandlePeriod2m:
		fmt.Fprintf(file2, "%s,%v,%.6f,%.6f,%.6f,%.6f\n", j.Ticker, j.TS.Format(time.RFC3339), j.Open, j.High, j.Low, j.Close)
	case domain.CandlePeriod10m:
		fmt.Fprintf(file3, "%s,%v,%.6f,%.6f,%.6f,%.6f\n", j.Ticker, j.TS.Format(time.RFC3339), j.Open, j.High, j.Low, j.Close)
	}
}
func makeOut(ctx context.Context, pr <-chan domain.Candle, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		file, err := os.OpenFile("candles_1m.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		file2, err2 := os.OpenFile("candles_2m.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		file3, err3 := os.OpenFile("candles_10m.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil || err2 != nil || err3 != nil {
			panic(err)
		}
		defer file.Close()
		defer file2.Close()
		defer file3.Close()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Saving data after SIGINT...")
				for j := range pr {
					writeToFile(j, file, file2, file3)
				}
				return
			case j := <-pr:
				writeToFile(j, file, file2, file3)
			}
		}
	}()
}
func main() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(ctx)
	ctx3, cancel3 := context.WithCancel(ctx2)
	ctx4, cancel4 := context.WithCancel(ctx3)

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})
	logger.Info("start prices generator...")
	prices := pg.Prices(ctx)

	wg.Add(4)
	logger.Info("start candles generator...")
	candles := formCandles1m(ctx, prices, &wg)
	candles2 := formCandles2m(ctx2, candles, &wg)
	candles10 := formCandles10m(ctx3, candles2, &wg)
	logger.Info("starting output to file...")
	makeOut(ctx4, candles10, &wg)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	logger.Info("SIGINT detected, saving data...")
	cancel()
	cancel2() // Не знаю, зачем это нужно, но попросил линтер так сделать. Беэ этих строк тоже работало
	cancel3()
	cancel4()
	wg.Wait()
	logger.Info("Finished")
}
