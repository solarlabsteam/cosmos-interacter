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

type AscendexResponse struct {
	data []AscendexBarhist
}

type AscendexBarhist struct {
	data AscendexBarhistData
}

type AscendexBarhistData struct {
	c string
}

func getAscendexRate() (float64, error) {
	url := fmt.Sprintf("https://ascendex.com/api/pro/v1/barhist?symbol=%s/USDT&interval=1&n=1", strings.ToUpper(AscendexCurrency))
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

	response := AscendexResponse{}
	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		return 0, err
	}

	if len(response.data) == 0 {
		return 0, fmt.Errorf("empty response from Ascendex")
	}

	return strconv.ParseFloat(response.data[0].data.c, 64)
}
