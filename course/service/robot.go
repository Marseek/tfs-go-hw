package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Marseek/tfs-go-hw/course/domain"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=robot.go -destination=mocks/mock.go

type repoInterface interface {
	SendOrder(symbol, side string, size int, addr string) (domain.APIResp, error)
	SetWSConnection(addr string, tick string) (chan domain.WsResponse, func(), error)
	GetTotalProfitDb(ctx context.Context) (float32, error)
	WriteOrderToDb(ctx context.Context, inst string, size int, side string, price float32, ordtype string, profit float32, stoploss float32) error
	WriteToTelegramBot(text string)
	GetUsersMap(string) map[string]string
}

type RobotInterface interface {
	SetParams(start, size int, profit float32, ticker, side string)
	SetStart(start int)
	SetParamsWithoutStart(size int, profit float32, ticker, side string)
	GetParams() domain.Options
	GetUsersMap(string) map[string]string
}

type RobotService struct {
	repo   repoInterface
	log    logrus.FieldLogger
	params domain.Options
	mu     sync.Mutex
}

func (r *RobotService) GetUsersMap(file string) map[string]string {
	return r.repo.GetUsersMap(file)
}

func (r *RobotService) SetStart(start int) {
	r.mu.Lock()
	r.params.Start = start
	r.mu.Unlock()
}

func (r *RobotService) SetParams(start, size int, profit float32, ticker, side string) {
	r.mu.Lock()
	r.params.Profit = profit
	r.params.Start = start
	r.params.Ticker = ticker
	r.params.Size = size
	r.params.Side = side
	r.mu.Unlock()
}

func (r *RobotService) SetParamsWithoutStart(size int, profit float32, ticker, side string) {
	r.mu.Lock()
	r.params.Profit = profit
	r.params.Ticker = ticker
	r.params.Size = size
	r.params.Side = side
	r.mu.Unlock()
}

func (r *RobotService) GetParams() domain.Options {
	var Opt domain.Options
	r.mu.Lock()
	Opt.Profit = r.params.Profit
	Opt.Start = r.params.Start
	Opt.Ticker = r.params.Ticker
	Opt.Size = r.params.Size
	Opt.Side = r.params.Side
	r.mu.Unlock()
	return Opt
}

func GetError(resp domain.APIResp) string {
	var s string
	if resp.Result == "success" {
		s = fmt.Sprintln("Order hadn't been placed: ", resp.SendStatus)
	} else {
		s = fmt.Sprintln("Order hadn't been placed: ", resp.Result)
	}
	return s
}

func (r *RobotService) GetStart() {
	for {
		time.Sleep(100 * time.Millisecond)
		start := r.GetParams().Start
		if start != 1 {
			continue
		}

		params := r.GetParams()
		r.log.Infoln("Start trading with params: ", params)

		priceChan, cancel, err := r.repo.SetWSConnection("wss://demo-futures.kraken.com/ws/v1", params.Ticker)
		if err != nil {
			r.log.Errorln("Bad request to WebSocket: ", err)
			r.SetStart(0)
			continue
		}

		// Небольшой анализ рынка, если направление сделки не задано вручную
		if params.Side == "" {
			maxAndMinPrice := map[string]float32{"max": 0, "min": 100000000000000, "mid": 0}
			var price0 float32
			for i, wsReturn := 0, <-priceChan; i < 7; i, wsReturn = i+1, <-priceChan {
				price0 = wsReturn.Ask
				r.log.Debugf("%+v\n", wsReturn)
				if price0 < maxAndMinPrice["min"] {
					maxAndMinPrice["min"] = price0
				}
				if price0 > maxAndMinPrice["max"] {
					maxAndMinPrice["max"] = price0
				}
			}
			maxAndMinPrice["mid"] = (maxAndMinPrice["max"] + maxAndMinPrice["min"]) / 2
			r.log.Debugf("%+v\n", maxAndMinPrice)
			// В зависимости от того, больше ли текущая цена средней цены или нет, устанавливаем направление сделки
			params.Side = "buy"
			if price0 > maxAndMinPrice["mid"] {
				params.Side = "sell"
			}
		}

		resp, err := r.repo.SendOrder(strings.ToLower(params.Ticker), params.Side, params.Size, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder")
		if err != nil {
			r.log.Errorln("Bad request to Api, while sending order: ", err)
			r.SetStart(0)
			cancel()
			continue
		}
		// Api запрос на открытие сделки вернул ошибку
		if resp.Result != "success" || resp.SendStatus.Status != "placed" {
			r.log.Infoln(GetError(resp))
			r.repo.WriteToTelegramBot(GetError(resp))
			r.SetStart(0)
			cancel()
			continue
		}

		// сообщение о покупке, запись в базу
		price := resp.SendStatus.OrderEvents[0].Price
		upperLimit := price * (1 + params.Profit/100)
		lowerLimit := price * (1 - params.Profit/100)
		message := fmt.Sprintf("Order had been opened.\nInstrument - %s, side - %s, size - %d, price - %.1f\nStoploss/takeprofit is %.1f/%.1f\n", params.Ticker, params.Side, params.Size, price, upperLimit, lowerLimit)
		r.repo.WriteToTelegramBot(message)
		err = r.repo.WriteOrderToDb(context.Background(), params.Ticker, params.Size, params.Side, price, "open", 0, params.Profit)
		if err != nil {
			r.log.Errorln("Can't write do Database: ", err)
		}

		// Слушаем канал и принимаем решение о закрытии
		for wsReturn := range priceChan {
			r.log.Debugf("%+v\n", wsReturn)
			closePrice := wsReturn.Ask
			if params.Side == "buy" {
				closePrice = wsReturn.Bid
			}
			if closePrice > upperLimit || closePrice < lowerLimit || r.GetParams().Start != 1 {
				params.Side = reverseSide(params.Side)
				resp, err = r.repo.SendOrder(strings.ToLower(params.Ticker), params.Side, params.Size, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder")
				if resp.Result != "success" || resp.SendStatus.Status != "placed" || err != nil {
					r.log.Errorln(GetError(resp))
					r.repo.WriteToTelegramBot(GetError(resp))
					r.SetStart(0)
					cancel() // closing WS connection
					break
				}
				cancel()
				r.SetStart(0)
				r.log.Infoln("The order had been closed")
				// Запись в базу и сообщение в телеграмм
				profit := closePrice - price
				if params.Side == "buy" {
					profit *= -1
				}
				err = r.repo.WriteOrderToDb(context.Background(), params.Ticker, params.Size, params.Side, closePrice, "close", profit*float32(params.Size), 0)
				if err != nil {
					r.log.Errorln("Can't write to DB: ", err)
				}
				total, _ := r.repo.GetTotalProfitDb(context.Background())
				message = fmt.Sprintf("Order had been closed.\nInstrument - %s, side - %s, size - %d, open price - %.1f, close price - %.1f, profit is %.1f\nTotal profit is %.1f", params.Ticker, params.Side, params.Size, price, closePrice, profit*float32(params.Size), total)
				r.repo.WriteToTelegramBot(message)
				break
			}
		}
	}
}

func NewRobotService(repo repoInterface, logger logrus.FieldLogger) RobotInterface {
	robot := RobotService{
		repo:   repo,
		log:    logger,
		params: domain.Options{},
		mu:     sync.Mutex{},
	}
	go robot.GetStart()

	return &robot
}

func reverseSide(s string) string {
	if s == "buy" {
		s = "sell"
	} else {
		s = "buy"
	}
	return s
}
