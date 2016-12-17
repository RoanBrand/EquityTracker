package main

import (
	"bytes"
	"encoding/json"
	"github.com/RoanBrand/EquityTracker/scheduler"
	"log"
	"os"
	"os/exec"
	"time"
)

type stockDetails struct {
	Name  string `json:"StockSymbol"`
	Id    string `json:"ID"`
	Index string `json:"Index"`
	Price string `json:"LastTradePrice"`
}

func main() {
	query := exec.Command("python", "query.py", "EOH")
	query.Stderr = os.Stderr
	queryOutput := &bytes.Buffer{}
	query.Stdout = queryOutput
	err := query.Run()
	if err != nil {
		panic(err)
	}

	var results []stockDetails
	json.Unmarshal(queryOutput.Bytes(), &results)

	log.Printf("%#v", results)
	log.Println(results)

	scheduler := scheduler.NewScheduler(time.Second*5, func(now time.Time) {
		log.Println("Scheduler callback triggered @ ", now)
	})
	defer scheduler.Stop()

	for {
		time.Sleep(time.Second)
	}
}
