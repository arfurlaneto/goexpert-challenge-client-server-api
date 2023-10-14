package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UsdBrlExchangeRateApiResponse struct {
	UsdBrl ExchangeRateApiResponse `json:"USDBRL"`
}

type ExchangeRateApiResponse struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PpctChange string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type ExchangeRate struct {
	ID         int `goorm:"primaryKey"`
	Code       string
	CodeIn     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PpctChange string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
	gorm.Model
}

type ApiResponse struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := CreateDatabaseConnection()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		apiCallCtx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
		defer cancel()
		apiResponse, err := GetUsdBrlExchangeRateFromApi(apiCallCtx)
		if err != nil {
			fmt.Printf("[ERROR] could not fetch exchange rate from api: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dbSaveCtx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
		defer cancel()
		err = SaveExchangeRateToDatabase(dbSaveCtx, db, &apiResponse.UsdBrl)
		if err != nil {
			fmt.Printf("[ERROR] could not save exchange rate in database: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(ApiResponse{Bid: apiResponse.UsdBrl.Bid})
	})

	http.ListenAndServe(":8080", nil)
}

func CreateDatabaseConnection() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("db.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&ExchangeRate{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetUsdBrlExchangeRateFromApi(ctx context.Context) (*UsdBrlExchangeRateApiResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
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
		return nil, errors.New("GetUsdBrlExchangeRate: invalid response from provider")
	}

	var exchangeRateData UsdBrlExchangeRateApiResponse
	err = json.NewDecoder(res.Body).Decode(&exchangeRateData)
	if err != nil {
		return nil, err
	}

	return &exchangeRateData, nil
}

func SaveExchangeRateToDatabase(ctx context.Context, db *gorm.DB, exchangeRateData *ExchangeRateApiResponse) error {
	return db.WithContext(ctx).Create(&ExchangeRate{
		Code:       exchangeRateData.Code,
		CodeIn:     exchangeRateData.CodeIn,
		Name:       exchangeRateData.Name,
		High:       exchangeRateData.High,
		Low:        exchangeRateData.Low,
		VarBid:     exchangeRateData.VarBid,
		PpctChange: exchangeRateData.PpctChange,
		Bid:        exchangeRateData.Bid,
		Ask:        exchangeRateData.Ask,
		Timestamp:  exchangeRateData.Timestamp,
		CreateDate: exchangeRateData.CreateDate,
	}).Error
}
