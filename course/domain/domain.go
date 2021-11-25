package domain

type SubscribeWS struct {
	Event string   `json:"event"`
	Feed  string   `json:"feed"`
	Prod  []string `json:"product_ids"`
}

type Options struct {
	Start  int     `json:"start"`
	Ticker string  `json:"ticker"`
	Size   int     `json:"size"`
	Profit float32 `json:"profit"`
	Side   string  `json:"side"`
}

type WsResponse struct {
	ProductID string  `json:"product_id"`
	Bid       float32 `json:"bid"`
	Ask       float32 `json:"ask"`
}

type SendStatus struct {
	OrderID     string        `json:"order_id"`
	Status      string        `json:"status"`
	OrderEvents []OrderEvents `json:"orderEvents"`
}

type APIResp struct {
	Result     string     `json:"result"`
	SendStatus SendStatus `json:"sendStatus"`
	Error      string     `json:"error"`
}

type OrderEvents struct {
	Price float32 `json:"price"`
}
