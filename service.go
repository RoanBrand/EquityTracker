package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/RoanBrand/EquityTracker/scheduler"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type stockDetails struct {
	Name  string `json:"StockSymbol"`
	Id    string `json:"ID"`
	Index string `json:"Index"`
	Price string `json:"LastTradePrice"`
}

type chartData [][2]float64

func main() {
	scheduler := scheduler.NewScheduler(time.Minute*15, func(now time.Time) {
		results := getStockPrices([]string{"EOH"})
		first := results[0]
		numPrice, err := strconv.ParseFloat(strings.Replace(first.Price, ",", "", -1), 64)
		if err != nil {
			log.Fatal(err)
		}
		if err := insertRecord(first.Name, now, first.Id, numPrice); err != nil {
			log.Fatal(err)
		}
	})
	defer scheduler.Stop()

	http.HandleFunc("/getstockprices", api_getStockPrices)
	http.Handle("/", http.FileServer(http.Dir("./front-end")))
	log.Println("Started HTTP Server")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func api_getStockPrices(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	stockName, ok := params["name"]
	if !ok {
		http.Error(w, "No stock name provided", http.StatusInternalServerError)
		return
	}
	data, err := getStockHistory(stockName[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Vary", "Accept-Encoding")
	w.Header().Set("Content-Type", "application/json")

	var encoder *json.Encoder
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		encoder = json.NewEncoder(gz)
		defer gz.Close()
	} else {
		encoder = json.NewEncoder(w)
	}
	err = encoder.Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func getStockHistory(tableName string) (chartData, error) {
	db, err := sql.Open("postgres", "user=imqs password=1mq5p@55w0rd dbname=stocks sslmode=disable")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(fmt.Sprintf(`SELECT "TIMESTAMP", "PRICE" FROM "%s"`, tableName))
	var h chartData
	for rows.Next() {
		var price float64
		var ts time.Time
		rows.Scan(&ts, &price)
		h = append(h, [2]float64{float64(ts.Unix() * 1000), price})
	}
	return h, nil
}

func insertRecord(tableName string, ts time.Time, id string, price float64) error {
	db, err := sql.Open("postgres", "user=imqs password=1mq5p@55w0rd dbname=stocks sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()
	cleanTs := time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, ts.Location())
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO "%s" ("TIMESTAMP", "ID", "PRICE") VALUES ($1, $2, $3)`, tableName), cleanTs.Format(time.RFC3339), id, price)
	return err
}
