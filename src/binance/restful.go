package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var apiKey = "crhlAbsbBAC9j9WAwyic8WFhUWSLP0TJIScir1ny5HxtTehq2G19sKE0tCFqho2s"
var secretKey = "Ou7EtmQ5sfBMe2zV8Sm0sNuxihp5UyZVIYMWboRbpQ8FhTCwgqH0S6t5bV66Oc7Y"

func GetAPI(api string, params map[string]string) string {
	return getAPI(api, params, false)
}

func GetSignedAPI(api string, params map[string]string) string {
	return getAPI(api, params, true)
}

func getAPI(api string, params map[string]string, isSign bool) string {
	host := "https://api.binance.com/api/v3"
	url := fmt.Sprintf("%s/%s", host, api)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}

	if isSign {
		q.Add("recvWindow", "5000")
		q.Add("timestamp", strconv.FormatInt(time.Now().UTC().UnixMilli(), 10))
		signature := sign(q.Encode())
		q.Add("signature", signature)
		req.Header.Set("X-MBX-APIKEY", apiKey)
	}

	req.URL.RawQuery = q.Encode()

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

func sign(data string) string {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secretKey))

	// Write Data to it
	h.Write([]byte(data))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha
}
