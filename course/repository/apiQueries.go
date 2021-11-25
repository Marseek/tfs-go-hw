package repository

import (
	"course/domain"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
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

func (r *repo) SendOrder(symbol, side string, size int) (domain.APIResp, error) {
	v := url.Values{}
	v.Add("orderType", "mkt")
	v.Add("symbol", symbol)
	v.Add("side", side)
	v.Add("size", strconv.Itoa(size))
	queryString := v.Encode()

	req, err := http.NewRequest(http.MethodPost, "http://demo-futures.kraken.com/derivatives/api/v3/sendorder"+"?"+queryString, nil)
	if err != nil {
		return domain.APIResp{}, err
	}

	req.Header.Add("APIKey", "r7Fw1MM/nMIKLO+GKk47eTs1HoGWnEEs94VpIfPqgAZ5t75a/yrdPm7u")
	authent := GenerateAuthent(queryString, "/api/v3/sendorder", "XiFDWJwfH70H65EBSlzq5N5HmhWA/Ce1ZR8HU/kSotdN2qSMCAj7aVXwf1sQpVBcc3YDllecDVNSrt/na9ARjupB")
	req.Header.Add("Authent", authent)

	c := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := c.Do(req)
	if err != nil {
		return domain.APIResp{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return domain.APIResp{}, err
	}

	var respStruct domain.APIResp
	err = json.Unmarshal(b, &respStruct)
	if err != nil {
		return domain.APIResp{}, err
	}

	return respStruct, nil
}
