package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"encoding/json"
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
}
