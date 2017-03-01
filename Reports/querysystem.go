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

func ListPage(reports reportList) func(w http.ResponseWriter, r *http.Request) {

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

		page.WriteString(`<div class="row"><div class="col"><h1 class="display-4">` + module + ` Reports</h1></div></div><br>`)
		page.WriteString(`<dl class="row">`)
		cat := ""
		for _, v := range reports {
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
	reportName := strings.ToLower(names[0])

	tmplBundle := []string{
		"Reports/templates/htmlbase.tmpl",
		"Reports/templates/modal-pdf.tmpl",
		"Reports/templates/" + reportName + "-content.tmpl",
	}
	tmplData := struct {
		Title string
		Code  string
	}{names[0], reportName + ".js"}

	flatOptions, ok := q["flat"]
	if ok && flatOptions[0] == "t" {
		tmplBundle = append(tmplBundle, "Reports/templates/"+reportName+"-flat.tmpl")
	} else {
		tmplBundle = append(tmplBundle, "Reports/templates/"+reportName+".tmpl")
	}

	t, err := template.ParseFiles(tmplBundle...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "base", tmplData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetReportData(rep report) func(w http.ResponseWriter, r *http.Request) {

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
		query.WriteString(`" WHERE "scenario" = 'Future') AS r`)

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
		query.WriteString(` LIMIT 10000`)

		db, err := sql.Open("postgres", "dbname=reports user=imqs password=1mq5p@55w0rd host=localhost sslmode=disable")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
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
