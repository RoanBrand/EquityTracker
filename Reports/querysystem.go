package Reports

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func (s *reportServer) ListPage() func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		page := bytes.NewBufferString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">`)
		page.WriteString(`<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.6/css/bootstrap.min.css" integrity="sha384-rwoIResjU2yc3z8GV/NPeZWAv56rSmLldC3R/AZzGRnGxQQKnKkoFVhFQhNUwEyJ" crossorigin="anonymous">`)
		page.WriteString(`<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js"></script>`)
		page.WriteString(`<script src="https://cdnjs.cloudflare.com/ajax/libs/tether/1.4.0/js/tether.min.js"></script>`)
		page.WriteString(`<script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.6/js/bootstrap.min.js" integrity="sha384-vBWWzlZJ8ea9aCX4pEW3rVHjgjt7zpkNpZk+02D9phzyeVkE+jo0ieGizqPLForn" crossorigin="anonymous"></script>`)
		page.WriteString(`</head><body><div class="container-fluid">`)

		q := r.URL.Query()
		modules, ok := q["m"]
		if !ok {
			http.Error(w, "Invalid request, need to provide IMQS module", http.StatusBadRequest)
			return
		}
		module := modules[0]
		scenarios, ok := q["s"]
		if !ok {
			http.Error(w, "Invalid request, need to provide IMQS scenario", http.StatusBadRequest)
			return
		}
		sessionCookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Invalid request, need to provide IMQS module", http.StatusBadGateway)
			return
		}
		s.Sessions[sessionCookie.Value] = sessionInfo{module, scenarios[0]}

		page.WriteString(`<div class="row"><div class="col"><h1 class="display-4">` + module + ` Reports</h1></div></div><br>`)
		page.WriteString(`<dl class="row">`)
		cat := ""
		for _, v := range s.Reports {
			if v.Module == module {
				if v.Category != cat {
					cat = v.Category
					page.WriteString(`<dt class="col-3">` + cat + `</dt>`)
					page.WriteString(`<dd class="col-9"><a href="report?name=` + v.Name + `">` + v.Title + `</a></dd>`)
				} else {
					page.WriteString(`<dd class="col-9 offset-3"><a href="report?name=` + v.Name + `">` + v.Title + `</a></dd>`)
				}
			}
		}
		page.WriteString(`</dl></div></body></html>`)
		w.Header().Set("Content-Type", "text/html")
		w.Write(page.Bytes())
	}
}

func BuildReport(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	names, ok := q["name"]
	if !ok {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	name := names[0]
	flatOptions, ok := q["flat"]
	flat := false
	if ok && flatOptions[0] == "t" {
		flat = true
	}

	reportName := strings.ToLower(name)
	templates := []string{
		"front-end/templates/reportbase.tmpl",
		"front-end/templates/" + reportName + "-content.tmpl",
	}
	if flat {
		templates = append(templates, "front-end/templates/"+reportName+"-flat.tmpl")
	} else {
		templates = append(templates, "front-end/templates/"+reportName+".tmpl")
	}
	t, err := template.ParseFiles(templates...)
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

func (s *reportServer) GetReportData(rep report) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		query := bytes.NewBufferString("SELECT ")
		for _, f := range rep.DataSource.ViewTable.Fields.FieldNames {
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
		query.WriteString(rep.DataSource.ViewTable.TableName)
		//query.WriteString(`" WHERE "scenario" = 'Future') AS r`)
		query.WriteString(`" `)

		// FILTERS
		if len(rep.DataSource.ViewTable.Filters.Conditions) > 0 {
			query.WriteString(`WHERE `)

			sessionCookie, err := r.Cookie("session")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			session, ok := s.Sessions[sessionCookie.Value]
			if !ok {
				log.Println("No session found")
				return
			}

			for _, c := range rep.DataSource.ViewTable.Filters.Conditions {
				if strings.Contains(c, "{") {
					var filter string
					for _, p := range rep.Parameters.Parameters {
						if strings.Contains(c, p) {
							switch p {
							case "IMQS_scenario":
								filter = strings.Replace(c, `{`+p+`}`, `'`+session.IMQS_Scenario+`'`, 1)
							default:
								filter = ""
							}
							break
						}
					}
					query.WriteString(filter)
					query.WriteString(` AND `)
				}
			}
			query.Truncate(query.Len() - 5)
			query.WriteString(`) AS r`)
		}

		// GROUP BY
		if len(rep.DataSource.ViewTable.Groupings.FieldNames) > 0 {
			query.WriteString(` GROUP BY `)
			for _, g := range rep.DataSource.ViewTable.Groupings.FieldNames {
				query.WriteString(`"` + g.Name + `", `)
			}
			query.Truncate(query.Len() - 2)
		}

		// ORDER BY
		if len(rep.DataSource.ViewTable.Orderings.FieldNames) > 0 {
			query.WriteString(` ORDER BY `)
			for _, g := range rep.DataSource.ViewTable.Orderings.FieldNames {
				query.WriteString(`"` + g.Name + `", `)
			}
			query.Truncate(query.Len() - 2)
		}
		query.WriteString(` LIMIT 100000`)

		db, err := sql.Open("postgres", "dbname=reports user=imqs password=1mq5p@55w0rd host=localhost sslmode=disable")
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

func GeneratePDF(rep report) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		orientations, ok := q["o"]
		if !ok {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		orientation := orientations[0]
		sizes, ok := q["s"]
		if !ok {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		size := sizes[0]
		flatOptions, ok := q["f"]

		inputHTML := "http://127.0.0.1/reports/report"
		if ok && flatOptions[0] == "t" {
			inputHTML += "?flat=t&name=" + rep.Name
		} else {
			inputHTML += "?name=" + rep.Name
		}
		pdfArgs := []string{
			"--print-media-type",
			"--javascript-delay",
			"1500",
		}
		if orientation == "Landscape" {
			pdfArgs = append(pdfArgs, "--orientation", "Landscape") // default Portrait
		}
		if size == "A3" || size == "A2" {
			pdfArgs = append(pdfArgs, "--page-size", size) // default A4
		}
		pdfArgs = append(pdfArgs, inputHTML, rep.Name+".pdf")
		cmd := exec.Command("wkhtmltopdf", pdfArgs...)
		err := cmd.Run()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pdf, err := ioutil.ReadFile(rep.Name + ".pdf")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dlName := rep.Module + " - " + rep.Category + " - " + rep.Title + ".pdf"
		w.Header().Set("Content-Disposition", `attachment; filename="`+dlName+`"`)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdf)
	}

}
