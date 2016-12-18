package main

import (
	"bytes"
	"encoding/json"
	"github.com/RoanBrand/EquityTracker/scheduler"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type stockDetails struct {
	Name  string `json:"StockSymbol"`
	Id    string `json:"ID"`
	Index string `json:"Index"`
	Price string `json:"LastTradePrice"`
}

func main() {
	w := sync.WaitGroup{}
	scheduler := scheduler.NewScheduler(time.Minute, func(now time.Time) {
		results := getStockPrices([]string{"EOH"})
		log.Printf("%v @ %v", results, now)
	})
	defer scheduler.Stop()

	w.Add(1)
	w.Wait()
}

func getStockPrices(names []string) []stockDetails {
	pyArgs := append([]string{"query.py"}, names...)
	query := exec.Command("python", pyArgs...)
	query.Stderr = os.Stderr
	queryOutput := &bytes.Buffer{}
	query.Stdout = queryOutput
	err := query.Run()
	if err != nil {
		panic(err)
	}

	var results []stockDetails
	json.Unmarshal(queryOutput.Bytes(), &results)
	return results
}
