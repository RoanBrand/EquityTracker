package main

import (
	"github.com/RoanBrand/EquityTracker/Reports"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	rl, err := Reports.LoadReport("Reports/Water/Model Summaries/by Elements.xml")
	if err != nil {
		log.Fatal(err)
	}
	for k, report := range rl {
		if k == "" {
			continue
		}
		http.HandleFunc("/"+k+"/data", Reports.GetReportData(report))
	}
	http.HandleFunc("/report", Reports.BuildReport)
	http.Handle("/", http.FileServer(http.Dir("./front-end")))

	log.Println("Started HTTP Server")
	log.Fatal(http.ListenAndServe(":80", nil))
}
