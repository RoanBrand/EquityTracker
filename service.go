package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	"github.com/RoanBrand/EquityTracker/Reports"
	"html/template"
)

const reportView = `SELECT
	"Water Model Summaries - Elements"."Capacity (kL,kL/d)",
	"Water Model Summaries - Elements"."Count",
	"Water Model Summaries - Elements"."Elements",
	"Water Model Summaries - Elements"."Length (m)",
	"Water Model Summaries - Elements"."OrderIndex",
	"Water Model Summaries - Elements"."Output (l/s)",
	"Water Model Summaries - Elements"."Replace Value (R)",
	"Water Model Summaries - Elements"."SecondOrder",
	"Water Model Summaries - Elements"."scenario"
FROM (
	SELECT * FROM "water_model_summaries_elements" where "scenario" = 'Future'
) AS "Water Model Summaries - Elements"`

const reportQuery = `SELECT
	"r"."Elements",
	"r"."Count",
	"r"."Length (m)",
	"r"."Capacity (kL,kL/d)",
	"r"."Output (l/s)",
	"r"."Replace Value (R)"
FROM ( SELECT * FROM "water_model_summaries_elements" where "scenario" = 'Future' ) AS "r"`

type watermodelsummariesbyelements struct {
	Capacity     string
	Count        int64
	Elements     string
	Length       float64
	OrderIndex   int
	Output       float64
	ReplaceValue float64
	SecondOrder  int
	scenario     string
}

type series struct {
	Name string    `json:"name"`
	Data []float64 `json:"data"`
}

type chartData_DiscreteColumn struct {
	Categories []string
	Series     []series
}

func main() {
	rl, err := Reports.LoadReport("Reports/Water/Model Summaries/by Elements.xml")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(rl)

	for k, report := range rl {
		if k == "" {
			continue
		}
		http.HandleFunc("/" + k + "/data", func(w http.ResponseWriter, r *http.Request) {
			sq := bytes.NewBufferString("SELECT ")
			for _, f := range report.Datasource.Viewtable.Fields.Fieldnames {
				sq.WriteString(`"`+f+`", `)
			}
			sq.Truncate(sq.Len()-2)
			sq.WriteString(` FROM ( SELECT * FROM "`)
			sq.WriteString(report.Datasource.Viewtable.Tablename)
			sq.WriteString(`") AS r`)

			db, err := sql.Open("postgres", "dbname=reports user=imqs password=1mq5p@55w0rd host=imqsrc sslmode=disable")
			if err != nil {
				return
			}
			defer db.Close()

			rows, err := db.Query(sq.String())
			if err != nil {
				return
			}
			defer rows.Close()

			colNames, err := rows.Columns()
			if err != nil {
				log.Fatalf("Error retrieving columns: %v", err)
			}
			dynRows := NewStringStringScan(colNames)
			jsonColumns, _ := json.Marshal(colNames)

			out := bytes.NewBufferString(`{"cols":`)
			out.Write(jsonColumns)
			out.WriteString(`, "rows":[`)
			for rows.Next() {
				err = dynRows.Update(rows)
				if err != nil {
					log.Fatalf("Row scan Error: %v", err)
				}
				jsonString, _ := json.Marshal(dynRows.Get())
				out.Write(jsonString)
				out.WriteString(",")
			}
			out.Truncate((out.Len() - 1))
			out.WriteString("]}")
			w.Header().Set("Content-Type", "application/json")
			w.Write(out.Bytes())
		})
		http.HandleFunc("/" + k, getGenericReport)
	}
/*
	var allFiles []string
	files, err := ioutil.ReadDir("./front-end/templates")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".tmpl") {
			allFiles = append(allFiles, "./front-end/templates/"+filename)
		}
	}

	templates, err := template.ParseFiles(allFiles...)
	if err != nil {
		log.Fatal(err)
	}

	fakeTitle := struct {
		Title string
		Code string
	}{"FAKE TITLE", "water-modelsummaries-byelements.js"}



	http.HandleFunc("/templatereport", func(w http.ResponseWriter, r* http.Request) {
		s1 := templates.Lookup("reportbase.tmpl")
		err = s1.ExecuteTemplate(w, "reportbase", fakeTitle)
		if err != nil {
			log.Fatal(err)
		}
	})
*/
	http.HandleFunc("/report-water-modelsummaries-byelements", report_water_modelsummaries)
	http.HandleFunc("/view-Water-ModelSummariesbyElements", getReport)
	http.Handle("/", http.FileServer(http.Dir("./front-end")))

	http.HandleFunc("/getreport", getGenericReport)

	log.Println("Started HTTP Server")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func report_water_modelsummaries(w http.ResponseWriter, r *http.Request) {
	data, err := getChartData()
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

func getChartData() (chartData_DiscreteColumn, error) {
	var h chartData_DiscreteColumn
	db, err := sql.Open("postgres", "dbname=reports user=imqs password=1mq5p@55w0rd host=imqsrc sslmode=disable")
	if err != nil {
		return h, err
	}
	defer db.Close()

	rows, err := db.Query(reportView)
	if err != nil {
		return h, err
	}
	defer rows.Close()

	ser := series{Name: "Elements"}
	for rows.Next() {
		var cap sql.NullFloat64
		var count int64
		var elem string
		var length sql.NullFloat64
		var orderindex int
		var output sql.NullFloat64
		var repval sql.NullFloat64
		var secondorder int
		var scenario string

		err := rows.Scan(&cap, &count, &elem, &length, &orderindex, &output, &repval, &secondorder, &scenario)
		if err != nil {
			log.Fatal(err)
		}
		h.Categories = append(h.Categories, elem)
		ser.Data = append(ser.Data, repval.Float64)
	}

	h.Series = append(h.Series, ser)
	return h, nil
}

func getGenericReport(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	names, ok := q["report"]
	if !ok {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	name := names[0]
	t, err := template.ParseFiles("front-end/templates/reportbase.tmpl", "front-end/templates/" + strings.ToLower(name) + ".tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		Code string
	}{name, strings.ToLower(name) + ".js"}

	s := t.Lookup("reportbase.tmpl")
	err = s.ExecuteTemplate(w, "reportbase", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getReport(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", "dbname=reports user=imqs password=1mq5p@55w0rd host=imqsrc sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(reportQuery)
	if err != nil {
		log.Fatalf("Query Error: %v", err)
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		log.Fatalf("Error retrieving columns: %v", err)
	}
	dynRows := NewStringStringScan(colNames)

	out := bytes.NewBufferString(`{"cols":`)
	jsonColumns, _ := json.Marshal(colNames)
	out.Write(jsonColumns)
	out.WriteString(`, "rows":[`)

	for rows.Next() {
		err = dynRows.Update(rows)
		if err != nil {
			log.Fatalf("Row scan Error: %v", err)
		}
		jsonString, _ := json.Marshal(dynRows.Get())
		out.Write(jsonString)
		out.WriteString(",")
	}
	out.Truncate((out.Len() - 1))
	out.WriteString("]}")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out.Bytes())
}

// SQL stuff
type stringStringScan struct {
	// cp are the column pointers
	cp []interface{}
	// row contains the final result
	row      []string
	colCount int
	colNames []string
}

func NewStringStringScan(columnNames []string) *stringStringScan {
	lenCN := len(columnNames)
	s := &stringStringScan{
		cp:       make([]interface{}, lenCN),
		row:      make([]string, lenCN),
		colCount: lenCN,
		colNames: columnNames,
	}
	for i := 0; i < lenCN; i++ {
		s.cp[i] = new(sql.RawBytes)
	}
	return s
}

func (s *stringStringScan) Update(rows *sql.Rows) error {
	if err := rows.Scan(s.cp...); err != nil {
		return err
	}
	for i := 0; i < s.colCount; i++ {
		if rb, ok := s.cp[i].(*sql.RawBytes); ok {
			s.row[i] = string(*rb)
			*rb = nil // reset pointer to discard current value to avoid a bug
		} else {
			return fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, s.colNames[i])
		}
	}
	return nil
}

func (s *stringStringScan) Get() []string {
	return s.row
}
