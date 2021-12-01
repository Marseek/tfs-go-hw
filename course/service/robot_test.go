package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Marseek/tfs-go-hw/course/domain"
	mock_service "github.com/Marseek/tfs-go-hw/course/service/mocks"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogin3(t *testing.T) {
	// Test Table
	type mockBehavior func(r *mock_service.MockrepoInterface, ch chan domain.WsResponse, opt domain.Options, resp domain.APIResp)
	type Test struct {
		Name         string
		Params       domain.Options
		APIResp      domain.APIResp
		mockBehavior mockBehavior
		ExpectStart  int
	}

	tests := [...]Test{
		{
			Name:    "All is OK",
			Params:  domain.Options{Start: 1, Side: "buy", Size: 1, Profit: 0.01, Ticker: "PI_XBTUSD"},
			APIResp: domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "placed", OrderEvents: []domain.OrderEvents{domain.OrderEvents{Price: 50000}}}},
			mockBehavior: func(r *mock_service.MockrepoInterface, ch chan domain.WsResponse, params domain.Options, resp domain.APIResp) {
				r.EXPECT().SetWSConnection("wss://demo-futures.kraken.com/ws/v1", params.Ticker).Return(ch, func() {}, nil)
				r.EXPECT().SendOrder(strings.ToLower(params.Ticker), params.Side, params.Size, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder").Return(resp, nil)
				price := resp.SendStatus.OrderEvents[0].Price
				message := fmt.Sprintf("Order had been opened.\nInstrument - %s, side - %s, size - %d, price - %.1f\nStoploss/takeprofit is %.1f/%.1f\n", params.Ticker, params.Side, params.Size, price, price*(1+params.Profit/100), price*(1-params.Profit/100))
				r.EXPECT().WriteToTelegramBot(message).Return()
				r.EXPECT().WriteOrderToDb(context.Background(), params.Ticker, params.Size, params.Side, price, "open", float32(0), params.Profit).Return(nil)
				r.EXPECT().SendOrder(strings.ToLower(params.Ticker), reverseSide(params.Side), params.Size, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder").Return(resp, nil)
				r.EXPECT().WriteOrderToDb(context.Background(), params.Ticker, params.Size, reverseSide(params.Side), price*1.1, "close", price*0.1, float32(0)).Return(nil)
				r.EXPECT().GetTotalProfitDb(context.Background()).Return(float32(50.0), nil)
				message = fmt.Sprintf("Order had been closed.\nInstrument - %s, side - %s, size - %d, open price - %.1f, close price - %.1f, profit is %.1f\nTotal profit is %.1f", params.Ticker, reverseSide(params.Side), params.Size, price, price*1.1, price*0.1*float32(params.Size), 50.0)
				r.EXPECT().WriteToTelegramBot(message).Return()
			},
		},
		{
			Name:    "Send order error",
			Params:  domain.Options{Start: 1, Side: "buy", Size: 1, Profit: 0.01, Ticker: "PI_XBTUSD"},
			APIResp: domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "placed", OrderEvents: []domain.OrderEvents{domain.OrderEvents{Price: 50000}}}},
			mockBehavior: func(r *mock_service.MockrepoInterface, ch chan domain.WsResponse, params domain.Options, resp domain.APIResp) {
				r.EXPECT().SetWSConnection("wss://demo-futures.kraken.com/ws/v1", params.Ticker).Return(ch, func() {}, nil)
				r.EXPECT().SendOrder(strings.ToLower(params.Ticker), params.Side, params.Size, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder").Return(domain.APIResp{}, errors.New("SendOrderError"))
			},
		},
		{
			Name:    "Kraken api response is not success",
			Params:  domain.Options{Start: 1, Side: "buy", Size: 1, Profit: 0.01, Ticker: "PI_XBTUSD"},
			APIResp: domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "placed", OrderEvents: []domain.OrderEvents{domain.OrderEvents{Price: 50000}}}},
			mockBehavior: func(r *mock_service.MockrepoInterface, ch chan domain.WsResponse, params domain.Options, resp domain.APIResp) {
				r.EXPECT().SetWSConnection("wss://demo-futures.kraken.com/ws/v1", params.Ticker).Return(ch, func() {}, nil)
				r.EXPECT().SendOrder(strings.ToLower(params.Ticker), params.Side, params.Size, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder").Return(resp, nil)
				price := resp.SendStatus.OrderEvents[0].Price
				message := fmt.Sprintf("Order had been opened.\nInstrument - %s, side - %s, size - %d, price - %.1f\nStoploss/takeprofit is %.1f/%.1f\n", params.Ticker, params.Side, params.Size, price, price*(1+params.Profit/100), price*(1-params.Profit/100))
				r.EXPECT().WriteToTelegramBot(message).Return()
				r.EXPECT().WriteOrderToDb(context.Background(), params.Ticker, params.Size, params.Side, price, "open", float32(0), params.Profit).Return(nil)
				resp.Result = "error"
				r.EXPECT().SendOrder(strings.ToLower(params.Ticker), reverseSide(params.Side), params.Size, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder").Return(resp, nil)
				r.EXPECT().WriteToTelegramBot("Order hadn't been placed:  error\n").Return()
			},
		},
		{
			Name:    "Response from kraken api to close position returns error",
			Params:  domain.Options{Start: 1, Side: "buy", Size: 1, Profit: 0.01, Ticker: "PI_XBTUSD"},
			APIResp: domain.APIResp{Result: "error", SendStatus: domain.SendStatus{Status: "placed", OrderEvents: []domain.OrderEvents{domain.OrderEvents{Price: 50000}}}},
			mockBehavior: func(r *mock_service.MockrepoInterface, ch chan domain.WsResponse, params domain.Options, resp domain.APIResp) {
				r.EXPECT().SetWSConnection("wss://demo-futures.kraken.com/ws/v1", params.Ticker).Return(ch, func() {}, nil)
				r.EXPECT().SendOrder(strings.ToLower(params.Ticker), params.Side, params.Size, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder").Return(resp, nil)
				r.EXPECT().WriteToTelegramBot("Order hadn't been placed:  error\n").Return()
			},
		},
		{
			Name:    "WS connection error",
			Params:  domain.Options{Start: 1, Side: "buy", Size: 1, Profit: 0.01, Ticker: "PI_XBTUSD"},
			APIResp: domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "placed", OrderEvents: []domain.OrderEvents{domain.OrderEvents{Price: 50000}}}},
			mockBehavior: func(r *mock_service.MockrepoInterface, ch chan domain.WsResponse, params domain.Options, resp domain.APIResp) {
				r.EXPECT().SetWSConnection("wss://demo-futures.kraken.com/ws/v1", "PI_XBTUSD").Return(ch, func() {}, errors.New("error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			logger := log.New()
			repo := mock_service.NewMockrepoInterface(c)
			ch := make(chan domain.WsResponse, 2)
			test.mockBehavior(repo, ch, test.Params, test.APIResp)

			serv := NewRobotService(repo, logger)
			serv.SetParams(1, test.Params.Size, test.Params.Profit, test.Params.Ticker, test.Params.Side)
			price := test.APIResp.SendStatus.OrderEvents[0].Price
			ch <- domain.WsResponse{Bid: price * 1.1, Ask: price * 1.1}
			ch <- domain.WsResponse{}
			time.Sleep(300 * time.Millisecond)

			assert.Equal(t, test.ExpectStart, serv.GetParams().Start)
		})
	}
}

func TestReverseSide(t *testing.T) {
	if expect, got := "sell", reverseSide("buy"); got != expect {
		t.Errorf(`Expect %v got %v\n`, expect, got)
	}
	assert.Equal(t, "buy", reverseSide("sell"))
}
