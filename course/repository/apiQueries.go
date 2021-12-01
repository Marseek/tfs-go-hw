package repository

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Marseek/tfs-go-hw/course/domain"
)

func GenerateAuthent(postData, endpoint, apiSecret string) string {
	sha := sha256.New()
	sha.Write([]byte(postData + endpoint))

	apiDecode, _ := base64.StdEncoding.DecodeString(apiSecret)

	h := hmac.New(sha512.New, apiDecode)
	h.Write(sha.Sum(nil))

	res := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return res
}

func (r *Repo) SendOrder(symbol, side string, size int, addr string) (domain.APIResp, error) {
	v := url.Values{}
	v.Add("orderType", "mkt")
	v.Add("symbol", symbol)
	v.Add("side", side)
	v.Add("size", strconv.Itoa(size))
	queryString := v.Encode()

	req, err := http.NewRequest(http.MethodPost, addr+"?"+queryString, nil)
	if err != nil {
		return domain.APIResp{}, err
	}

	req.Header.Add("APIKey", r.secrets["public"])
	authent := GenerateAuthent(queryString, "/api/v3/sendorder", r.secrets["privat"])
	req.Header.Add("Authent", authent)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return domain.APIResp{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return domain.APIResp{}, err
	}
	resp.Body.Close()

	var respStruct domain.APIResp
	err = json.Unmarshal(b, &respStruct)
	if err != nil {
		return domain.APIResp{}, err
	}

	return respStruct, nil
}
