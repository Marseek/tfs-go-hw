package service

import (
	"context"
	"course/domain"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var mu sync.Mutex

type repoInterface interface {
	SendOrder(symbol, side string, size int) (domain.APIResp, error)
	SetWSConnection(tick string) (chan domain.WsResponse, func(), error)
	GetTotalProfitDb(ctx context.Context) (float32, error)
	WriteOrderToDb(ctx context.Context, inst string, size int, side string, price float32, ordtype string, profit float32, stoploss float32) error
	WriteToTelegramBot(text string)
	GetUsersMap() map[string]string
}

type RobotInterface interface {
	SetParams(start, size int, profit float32, ticker, side string)
	SetStart(start int)
	SetParamsWithoutStart(size int, profit float32, ticker, side string)
	GetParams() domain.Options
	GetUsersMap() map[string]string
}

type RobotService struct {
	repo   repoInterface
	log    logrus.FieldLogger
	params domain.Options
}

func (r *RobotService) GetUsersMap() map[string]string {
	return r.repo.GetUsersMap()
}

func (r *RobotService) SetStart(start int) {
	mu.Lock()
	r.params.Start = start
	mu.Unlock()
}

func (r *RobotService) SetParams(start, size int, profit float32, ticker, side string) {
	mu.Lock()
	r.params.Profit = profit
	r.params.Start = start
	r.params.Ticker = ticker
	r.params.Size = size
	r.params.Side = side
	mu.Unlock()
}

func (r *RobotService) SetParamsWithoutStart(size int, profit float32, ticker, side string) {
	mu.Lock()
	r.params.Profit = profit
	r.params.Ticker = ticker
	r.params.Size = size
	r.params.Side = side
	mu.Unlock()
}

func (r *RobotService) GetParams() domain.Options {
	var Opt domain.Options
	mu.Lock()
	Opt.Profit = r.params.Profit
	Opt.Start = r.params.Start
	Opt.Ticker = r.params.Ticker
	Opt.Size = r.params.Size
	Opt.Side = r.params.Side
	mu.Unlock()
	return Opt
}

func GetErrror(resp domain.APIResp) string {
	var s string
	if resp.Result == "success" {
		s = fmt.Sprintln("Order hadn't been placed: ", resp.SendStatus)
	} else {
		s = fmt.Sprintln("Order hadn't been placed: ", resp.Result)
	}
	return s
}

func NewRobotService(repo repoInterface, logger logrus.FieldLogger) RobotInterface {
	robot := RobotService{
		repo:   repo,
		log:    logger,
		params: domain.Options{},
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			mu.Lock()
			start := robot.params.Start
			mu.Unlock()
			if start != 1 {
				continue
			}

			robot.log.Infoln("Start trading with params: ", robot.params)
			params := robot.GetParams()

			priceChan, cancel, err := robot.repo.SetWSConnection(params.Ticker)
			if err != nil {
				robot.log.Errorln("Bad request to WebSocket: ", err)
				robot.params.Start = 0
				continue
			}
			resp, err := robot.repo.SendOrder(strings.ToLower(params.Ticker), params.Side, params.Size)
			if err != nil {
				robot.log.Errorln("Bad request to Api, while sending order: ", err)
				robot.params.Start = 0
				cancel()
				continue
			}
			// Api запрос на открытие сделки вернул ошибку
			if resp.Result != "success" || resp.SendStatus.Status != "placed" {
				robot.log.Infoln(GetErrror(resp))
				robot.repo.WriteToTelegramBot(GetErrror(resp))
				robot.params.Start = 0
				cancel()
				continue
			}

			// сообщение о покупке, запись в базу
			price := resp.SendStatus.OrderEvents[0].Price
			upperLimit := price * (1 + params.Profit/100)
			lowerLimit := price * (1 - params.Profit/100)
			message := fmt.Sprintf("Order had been opened.\nInstrument - %s, side - %s, size - %d, price - %.1f\nStoploss/takeprofit is %.1f/%.1f\n", params.Ticker, params.Side, params.Size, price, upperLimit, lowerLimit)
			robot.repo.WriteToTelegramBot(message)
			err = robot.repo.WriteOrderToDb(context.Background(), params.Ticker, params.Size, params.Side, price, "open", 0, params.Profit)
			if err != nil {
				robot.log.Errorln("Can't write do Database: ", err)
			}

			// Слушаем канал и принимаем решение о закрытии
			for wsReturn := range priceChan {
				robot.log.Debugf("%+v\n", wsReturn)
				closePrice := wsReturn.Ask
				if params.Side == "buy" {
					closePrice = wsReturn.Bid
				}
				if closePrice > upperLimit || closePrice < lowerLimit || robot.params.Start != 1 {
					params.Side = reverseSide(params.Side)
					resp, err = robot.repo.SendOrder(strings.ToLower(params.Ticker), params.Side, params.Size)
					if resp.Result != "success" || resp.SendStatus.Status != "placed" || err != nil {
						robot.log.Errorln(GetErrror(resp))
						robot.repo.WriteToTelegramBot(GetErrror(resp))
						robot.params.Start = 0
						<-priceChan
						cancel() // closing WS connection
						break
					}
					<-priceChan
					cancel()
					robot.params.Start = 0
					robot.log.Infoln("The order had been closed")
					// Запись в базу и сообщение в телеграмм
					profit := closePrice - price
					if params.Side == "buy" {
						profit *= -1
					}
					err = robot.repo.WriteOrderToDb(context.Background(), params.Ticker, params.Size, params.Side, closePrice, "close", profit*float32(params.Size), 0)
					if err != nil {
						robot.log.Errorln("Can't write to DB: ", err)
					}
					total, _ := robot.repo.GetTotalProfitDb(context.Background())
					message = fmt.Sprintf("Order had been closed.\nInstrument - %s, side - %s, size - %d, open price - %.1f, close price - %.1f, profit is %.1f\nTotal profit is %.1f", params.Ticker, params.Side, params.Size, price, closePrice, profit*float32(params.Size), total)
					robot.repo.WriteToTelegramBot(message)
				}
			}
		}
	}()
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
