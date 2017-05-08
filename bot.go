package main

import (
    "encoding/json"
    "net/http"
    "bytes"
    "strings"
    "time"
    "log"
    "errors"
)

type currencyExchangeProvider interface {
	rate(currencySymbol1 string, currencySymbol2 string) (float64, error)
}

type fixerio struct {}

type multiCurrencyExchangeProvider []currencyExchangeProvider


func main() {
    
    mcep := multiCurrencyExchangeProvider{
        fixerio{},
    }

    http.HandleFunc("/rates/", func(w http.ResponseWriter, r *http.Request) {
        begin := time.Now()
        symbols := strings.SplitN(r.URL.Path, "/", 3)[2]
        sym := strings.SplitN(symbols, "_", 2)
        rate, err := mcep.rate(strings.ToUpper(sym[0]), strings.ToUpper(sym[1]))
        
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "rate": rate,
            "took": time.Since(begin).String(),
        })

    })

    http.ListenAndServe(":9090", nil)

}

func (mcep multiCurrencyExchangeProvider) rate(currencySymbol1 string, currencySymbol2 string) (float64, error){
    rates := make(chan float64, len(mcep))
    errs := make(chan error, len(mcep))

    for _, provider := range mcep {
        go func(cp currencyExchangeProvider) {
            r, err := cp.rate(currencySymbol1, currencySymbol2)
            if err != nil {
                errs <- err
                return
            }
            rates <- r
        }(provider)
    }

    theRate := 0.0
    for i := 0; i < len(mcep); i++ {
        select {
            case rate := <-rates:
                theRate = rate
            case err := <-errs:
                return 0, err
        }
    }
    return theRate, nil
}

func (c fixerio) rate(currencySymbol1 string, currencySymbol2 string) (float64, error) {
	resp, err := http.Get("http://api.fixer.io/latest?base="+currencySymbol1+"&symbols="+currencySymbol2)
    if err != nil {
        return 0, err
    }

    defer resp.Body.Close()

    var data map[string]interface{}
    buf := new(bytes.Buffer)
    buf.ReadFrom(resp.Body)
    b := buf.Bytes()

    if err := json.Unmarshal(b, &data); err != nil {
        log.Println("json unmarshal error")
    }
    for rates, ok := (data["rates"]).(map[string]interface{});; {
        if ok {
            return rates[currencySymbol2].(float64), nil
        }
        log.Printf("%+v\n", data["rates"])
        log.Println("type assertion error")
        return 0, errors.New("rate limit exceeded")
    }
}