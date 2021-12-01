package repository

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Marseek/tfs-go-hw/course/domain"
	"github.com/stretchr/testify/assert"
)

func TestSendOrder_mock(t *testing.T) {
	// Test Table
	type Test struct {
		Name   string
		In     domain.Options
		Expect domain.APIResp
	}
	tests := [...]Test{
		{Name: "All is OK", In: domain.Options{Start: 1, Ticker: "PI_XBTUSD", Size: 2, Profit: 0.1, Side: "buy"}, Expect: domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "placed"}}},
		{Name: "Ticker is invalid", In: domain.Options{Start: 1, Ticker: "invalid_ticker", Size: 2, Profit: 0.1, Side: "buy"}, Expect: domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "error"}}},
		{Name: "Side is invalid", In: domain.Options{Start: 1, Ticker: "PI_XBTUSD", Size: 2, Profit: 0.1, Side: "side_invalid"}, Expect: domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "error"}}},
		{Name: "Size is invalid", In: domain.Options{Start: 1, Ticker: "PI_XBTUSD", Size: -1, Profit: 0.1, Side: "buy"}, Expect: domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "error"}}},
	}

	var publicAPIKey = flag.String("public", "", "public key from Kraken")
	var privatAPIKey = flag.String("privat", "", "Privat key from Kraken")
	flag.Parse()
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		answer := domain.APIResp{Result: "success", SendStatus: domain.SendStatus{Status: "placed"}}
		mapQuery := req.URL.Query()
		symArr, ok1 := mapQuery["symbol"]
		sideArr, ok2 := mapQuery["side"]
		sizeArr, ok3 := mapQuery["size"]
		orderArr, ok4 := mapQuery["orderType"]

		params := domain.Options{}
		var err1 error
		params.Size, err1 = strconv.Atoi(sizeArr[0])
		if !ok1 || !ok2 || !ok3 || !ok4 || err1 != nil {
			answer.SendStatus.Status = "error"
			str, _ := json.Marshal(answer)
			_, _ = resp.Write(str)
			return
		}
		params.Ticker = symArr[0]
		params.Side = sideArr[0]
		order := orderArr[0]
		err := checkInput(params, order)
		if err != nil {
			answer.SendStatus.Status = "error"
			str, _ := json.Marshal(answer)
			_, _ = resp.Write(str)
			return
		}

		APIKey := req.Header.Get("APIKey")
		Authent := req.Header.Get("Authent")

		if APIKey != *publicAPIKey || Authent != GenerateAuthent2(req.URL.RawQuery, "/api/v3/sendorder", *privatAPIKey) {
			answer.Result = "error"
			str, _ := json.Marshal(answer)
			_, _ = resp.Write(str)
			return
		}

		str, _ := json.Marshal(answer)
		_, _ = resp.Write(str)
	}))
	defer func() { server.Close() }()

	r := Repo{}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := r.SendOrder(test.In.Ticker, test.In.Side, test.In.Size, server.URL)
			assert.NoError(t, err)
			assert.Equal(t, got.Result, test.Expect.Result)
			assert.Equal(t, got.SendStatus.Status, test.Expect.SendStatus.Status)
		})
	}
}

func checkInput(opt domain.Options, order string) error {
	if opt.Side != "buy" && opt.Side != "sell" {
		return errors.New(`'side' option must be 'buy' or 'sell'`)
	}
	if opt.Size < 1 {
		return errors.New(`'size' option must be more than 0`)
	}
	if opt.Ticker != "PI_XBTUSD" && opt.Ticker != "PI_ETHUSD" && opt.Ticker != "PI_LTCUSD" && opt.Ticker != "PI_XRPUSD" && opt.Ticker != "PI_BCHUSD" {
		return errors.New(`'ticker' option must be 'PI_XBTUSD' or 'PI_ETHUSD' or 'PI_LTCUSD' or 'PI_XRPUSD' or 'PI_BCHUSD'`)
	}
	if order != "mkt" {
		return errors.New(`'order should be mkt`)
	}
	return nil
}

func GenerateAuthent2(postData, endpoint, apiSecret string) string {
	sha := sha256.New()
	sha.Write([]byte(postData + endpoint))

	apiDecode, _ := base64.StdEncoding.DecodeString(apiSecret)

	h := hmac.New(sha512.New, apiDecode)
	h.Write(sha.Sum(nil))

	res := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return res
}
