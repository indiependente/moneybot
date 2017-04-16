package main

import (
    "encoding/json"
    "net/http"
    "strings"
    "strconv"
    "log"
    "time"
)

type currencyExchangeProvider interface {
	rate(currencySymbol1 string, currencySymbol2 string) (float64, error)
}

func (c currencyExchangeProvider) rate(currencySymbol1 string, currencySymbol2 string) (float64, error) {
	resp, err := http.Get("http://api.fixer.io/latest?base="+currencySymbol1+"&symbols="+currencySymbol2)
    if err != nil {
        return 0, err
    }

    defer resp.Body.Close()

    var d struct {
	Base string `json:"base"`
	Date string `json:"date"`
	Rates struct {
		EUR float64 `json:"EUR"`
	} `json:"rates"`
}

    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
        return 0, err
    }

    return EUR, nil
}