package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Marseek/tfs-go-hw/course/domain"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var upgrader = websocket.Upgrader{}

func MockWsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		var req domain.SubscribeWS
		_, message, err := c.ReadMessage()
		if err != nil {
			continue
		}
		err = json.Unmarshal(message, &req)
		if err != nil {
			continue
		}
		if req.Event != "subscribe" || req.Feed != "ticker_lite" {
			wsRequest, _ := json.Marshal(domain.WsResponse{ProductID: "error"})
			_ = c.WriteMessage(websocket.TextMessage, wsRequest)
			continue
		}
		wsRequest, _ := json.Marshal(domain.WsResponse{ProductID: "success"})
		err = c.WriteMessage(websocket.TextMessage, wsRequest)
		if err != nil {
			continue
		}
	}
}

func TestEstablishWsConnection(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(MockWsHandler))
	defer s.Close()
	const expext = "success"
	addr := "ws" + strings.TrimPrefix(s.URL, "http")

	r := Repo{}

	priceChan, _, err := r.SetWSConnection(addr, "Ticker")
	resp := <-priceChan

	assert.NoError(t, err)
	assert.Equal(t, resp.ProductID, expext)
}
