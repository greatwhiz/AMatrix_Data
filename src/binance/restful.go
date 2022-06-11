package binance

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func GetAPI(api string, params map[string]string) string {
	return requestAPI(api, "GET", params, false)
}

func GetSignedAPI(api string, params map[string]string) string {
	return requestAPI(api, "GET", params, true)
}

func PostSignedAPI(api string, params map[string]string) string {
	return requestAPI(api, "POST", params, true)
}

func requestAPI(api string, action string, params map[string]string, isSign bool) string {
	host := "https://api.binance.com/api/v3"
	requestUrl := fmt.Sprintf("%s/%s", host, api)
	var req *http.Request
	if action == "GET" {
		req = getAPI(requestUrl, params, isSign)
	} else if action == "POST" {
		req = postAPI(requestUrl, params, isSign)
	}

	if isSign {
		req.Header.Set("X-MBX-APIKEY", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := resp.Body.Close() // we must close anyway
		client.CloseIdleConnections()
		if err == nil { // we must not overwrite the actual error if it is happened, and we did all the best to cleanup anyway
			err = errors.Wrap(err, "close")
		}
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	return string(bodyBytes)
}

func getAPI(requestUrl string, params map[string]string, isSign bool) *http.Request {
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		panic(err)
	}
	req.URL.RawQuery = getParams(params, isSign)
	return req
}

func postAPI(requestUrl string, params map[string]string, isSign bool) *http.Request {
	requestBody := getParams(params, isSign)
	req, err := http.NewRequest("POST", requestUrl, bytes.NewBufferString(requestBody))
	if err != nil {
		panic(err)
	}
	return req
}

func getParams(params map[string]string, isSign bool) string {
	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}

	if isSign {
		q.Add("recvWindow", "5000")
		q.Add("timestamp", strconv.FormatInt(time.Now().UTC().UnixMilli(), 10))
		signature := sign(q.Encode())
		q.Add("signature", signature)
	}
	return q.Encode()
}

func sign(data string) string {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secretKey))

	// Write Data to it
	h.Write([]byte(data))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha
}
