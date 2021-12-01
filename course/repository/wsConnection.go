package repository

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Marseek/tfs-go-hw/course/domain"
	"github.com/gorilla/websocket"
)

func (r *Repo) EstablishWsConnection(addr string, tick string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	for err != nil { // redialling. Это когда сразу не получается подключиться.
		time.Sleep(time.Second)
		c, _, err = websocket.DefaultDialer.Dial(addr, nil)
	}

	wsRequest, err := json.Marshal(domain.SubscribeWS{Event: "subscribe", Feed: "ticker_lite", Prod: []string{tick}})
	if err != nil {
		r.logger.Fatalln("Error, while unmarshalling WS request. ", err)
	}
	err = c.WriteMessage(websocket.TextMessage, wsRequest)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Repo) SetWSConnection(addr string, tick string) (chan domain.WsResponse, func(), error) {
	ch := make(chan domain.WsResponse)

	c, err := r.EstablishWsConnection(addr, tick)
	if err != nil {
		return nil, nil, err
	}

	cancel := make(chan struct{})
	go func() {
		for {
			var resp domain.WsResponse
			_, message, err := c.ReadMessage()
			if err != nil {
				r.logger.Debugln("WS connection failed. Establishing new connection")
				c, _ = r.EstablishWsConnection(addr, tick)
				continue
			}
			err = json.Unmarshal(message, &resp)
			if err != nil {
				log.Println(err)
				r.logger.Debugln("Unmarshall error: ", err)
				continue
			}
			if resp.ProductID == "" { // Игнорируем всякие странные сообщения
				continue
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
		<-ch // нужно вычитать одно сообщение из канала, чтобы запись в него не блокировалась на 57 строчке
		cancel <- struct{}{}
	}, nil
}
