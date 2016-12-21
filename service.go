package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/RoanBrand/EquityTracker/scheduler"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	scheduler := scheduler.NewScheduler(time.Minute*15, func(now time.Time) {
		results := getStockPrices([]string{"EOH"})
		first := results[0]
		//log.Printf("%v @ %v\n", first, now)
		numPrice, err := strconv.ParseFloat(strings.Replace(first.Price, ",", "", -1), 64)
		if err != nil {
			log.Fatal(err)
		}
		//log.Println(numPrice)
		if err := insertRecord(first.Name, now, first.Id, numPrice); err != nil {
			log.Fatal(err)
		}
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

func checkTable() error {
	db, err := sql.Open("postgres", "user=imqs password=1mq5p@55w0rd dbname=stocks sslmode=disable")
	if err != nil {
		return err
	}
	rows, err := db.Query(`SELECT "TIMESTAMP", "PRICE" FROM "EOH" LIMIT 1`)
	var ts time.Time
	var price float64
	for rows.Next() {
		rows.Scan(&ts, &price)
		log.Println(ts, price)
	}

	return nil
}

func insertRecord(tableName string, ts time.Time, id string, price float64) error {
	db, err := sql.Open("postgres", "user=imqs password=1mq5p@55w0rd dbname=stocks sslmode=disable")
	if err != nil {
		return err
	}
	cleanTs := time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, ts.Location())
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO "%s" ("TIMESTAMP", "ID", "PRICE") VALUES ($1, $2, $3)`, tableName), cleanTs.Format(time.RFC3339), id, price)
	return err
}

// CREATE TABLE "EOH"
// (
// rowid serial NOT NULL,
// "TIMESTAMP" timestamp without time zone,
// "ID" character varying,
// "PRICE" double precision
// )
// WITH (
// OIDS=FALSE
// );
// ALTER TABLE "EOH"
// OWNER TO imqs;
