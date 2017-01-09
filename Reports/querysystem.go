package Reports

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
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
		query := bytes.NewBufferString("SELECT ")
		for _, f := range rep.Datasource.Viewtable.Fields.FieldNames {
			switch f.Aggregate {
			case "sum":
				query.WriteString(`SUM("` + f.Name + `")`)
				if f.DisplayName != "" {
					query.WriteString(` AS "` + f.DisplayName + `"`)
				} else {
					query.WriteString(` AS "` + f.Name + `"`)
				}
				query.WriteString(`, `)
				break
			default:
				query.WriteString(`"` + f.Name + `"`)
				if f.DisplayName != "" {
					query.WriteString(` AS "` + f.DisplayName + `"`)
				}
				query.WriteString(`, `)
			}

		}
		query.Truncate(query.Len() - 2)
		query.WriteString(` FROM ( SELECT * FROM "`)
		query.WriteString(rep.Datasource.Viewtable.TableName)
		query.WriteString(`" WHERE "scenario" = 'Future') AS r`)

		// GROUP BY
		if len(rep.Datasource.Viewtable.Groupings.FieldNames) > 0 {
			query.WriteString(` GROUP BY `)
			for _, g := range rep.Datasource.Viewtable.Groupings.FieldNames {
				query.WriteString(`"` + g.Name + `", `)
			}
			query.Truncate(query.Len() - 2)
		}

		// ORDER BY
		if len(rep.Datasource.Viewtable.Orderings.FieldNames) > 0 {
			query.WriteString(` ORDER BY `)
			for _, g := range rep.Datasource.Viewtable.Orderings.FieldNames {
				query.WriteString(`"` + g.Name + `", `)
			}
			query.Truncate(query.Len() - 2)
		}
		query.WriteString(` LIMIT 5000`)

		db, err := sql.Open("postgres", "dbname=reports user=imqs password=1mq5p@55w0rd host=imqsrc sslmode=disable")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
		log.Println(query.String())
		rows, err := db.Query(query.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

// wont gzip get closed before we write to it?
func encodingWrapper(w http.ResponseWriter, r *http.Request) *json.Encoder {
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


// cannot use close
func gzipWrapper(w http.ResponseWriter, r *http.Request) io.Writer {
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		return gzip.NewWriter(w)
	}
	return w
}
