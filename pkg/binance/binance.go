package binance

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const baseURL = "https://api.binance.com/api/v3/ticker/price?symbol="

type Price struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func GetPrice(symbol string) (string, error) {
	resp, err := http.Get(baseURL + symbol) // дефолтный клиент не имеет таймаута и может зависнуть навсегда
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var price Price
	err = json.Unmarshal(body, &price)
	if err != nil {
		return "", err
	}
	return price.Price, nil
}
