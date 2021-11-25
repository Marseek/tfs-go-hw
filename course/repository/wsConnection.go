package repository

import (
	"course/domain"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func (r *repo) EstablishWsConnection(tick string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial("wss://demo-futures.kraken.com/ws/v1", nil)
	for err != nil { // redialling. Это когда сразу не получается подключиться.
		c, _, err = websocket.DefaultDialer.Dial("wss://demo-futures.kraken.com/ws/v1", nil)
	}

	wsRequest, err := json.Marshal(domain.SubscribeWS{Event: "subscribe", Feed: "ticker_lite", Prod: []string{tick}})
	if err != nil {
		r.logger.Fatalln("Error, while unmarshalling WS request. ", err)
	}
	err = c.WriteMessage(websocket.TextMessage, wsRequest)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 2; i++ { // вычитывает ненужные сообщения из ВебСокета
		_, _, _ = c.ReadMessage()
	}
	return c, nil
}

func (r *repo) SetWSConnection(tick string) (chan domain.WsResponse, func(), error) {
	ch := make(chan domain.WsResponse)
	var resp domain.WsResponse

	c, err := r.EstablishWsConnection(tick)
	if err != nil {
		return nil, nil, err
	}

	cancel := make(chan struct{})
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				c, _ = r.EstablishWsConnection(tick)
				r.logger.Debugln("WS connection failed. Establishing new connection")
				continue
			}
			err = json.Unmarshal(message, &resp)
			if err != nil {
				log.Println(err)
			}
			ch <- resp
			select {
			case <-cancel:
				_ = c.Close()
				close(ch)
				return
			case <-time.After(time.Millisecond * 100):
			}
		}
	}()
	return ch, func() {
		cancel <- struct{}{}
	}, nil
}
