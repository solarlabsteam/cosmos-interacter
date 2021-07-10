package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MxcResponse struct {
	Data []MxcInfo `json:"data"`
}

type MxcInfo struct {
	Last string `json:"last"`
}

func getMxcRate() (float64, error) {
	url := fmt.Sprintf("https://www.mxc.com/open/api/v2/market/ticker?symbol=%s_USDT", strings.ToUpper(MxcCurrency))
	client := http.Client{Timeout: time.Second * 2}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "cosmos-interacter")

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	response := MxcResponse{}
	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		return 0, err
	}

	if len(response.Data) == 0 {
		return 0, fmt.Errorf("empty response from Mxc")
	}

	return strconv.ParseFloat(response.Data[0].Last, 64)
}
