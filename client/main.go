package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ApiResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)
	defer cancel()

	apiResponse, err := GetExchangeRateFromApi(ctx)
	if err != nil {
		fmt.Printf("[ERROR] could not fetch exchange rate from api: %s\n", err.Error())
		return
	}

	err = SaveExchangeRateToFile(apiResponse.Bid)
	if err != nil {
		fmt.Printf("[ERROR] could not save exchange rate to file: %s\n", err.Error())
		return
	}
}

func GetExchangeRateFromApi(ctx context.Context) (*ApiResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("GetExchangeRateFromApi: invalid response from server")
	}

	var exchangeRateData ApiResponse
	err = json.NewDecoder(res.Body).Decode(&exchangeRateData)
	if err != nil {
		return nil, err
	}

	return &exchangeRateData, nil
}

func SaveExchangeRateToFile(exchangeRate string) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar:%s", exchangeRate))
	return err
}
