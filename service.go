package main

import (
	"github.com/RoanBrand/EquityTracker/Reports"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	serv, err := Reports.NewReportServer("Reports")
	if err != nil {
		log.Fatal(err)
	}
	for k, report := range serv.Reports {
		if k == "" {
			continue
		}
		http.HandleFunc("/"+k+"/data", Reports.GetReportData(report)) // report content data
	}
	http.HandleFunc("/reportlist", Reports.GetReportList(serv.Reports)) // list of reports
	http.HandleFunc("/landingpage", Reports.ListPage(serv.Reports))

	http.HandleFunc("/report", Reports.BuildReport)            // report query
	http.Handle("/", http.FileServer(http.Dir("./front-end"))) // static files

	log.Println("Started HTTP Server")
	log.Fatal(http.ListenAndServe(":2016", nil))
}
