package Reports

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"compress/gzip"
)

func BuildReport(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	names, ok := q["name"]
	if !ok {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	name := names[0]
	t, err := template.ParseFiles("front-end/templates/reportbase.tmpl", "front-end/templates/"+strings.ToLower(name)+".tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		Code  string
	}{name, strings.ToLower(name) + ".js"}

	s := t.Lookup("reportbase.tmpl")
	err = s.ExecuteTemplate(w, "reportbase", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetReportData(rep report) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sq := bytes.NewBufferString("SELECT ")
		for _, f := range rep.Datasource.Viewtable.Fields.Fieldnames {
			sq.WriteString(`"` + f + `", `)
		}
		sq.Truncate(sq.Len() - 2)
		sq.WriteString(` FROM ( SELECT * FROM "`)
		sq.WriteString(rep.Datasource.Viewtable.Tablename)
		sq.WriteString(`" WHERE "scenario" = 'Future') AS r`)

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
	}
}

func EncodingWrapper(w http.ResponseWriter, r *http.Request) *json.Encoder {
	var encoder *json.Encoder
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		encoder = json.NewEncoder(gz)
		defer gz.Close()
	} else {
		encoder = json.NewEncoder(w)
	}
	return encoder
}
