package Reports

import (
	"encoding/xml"
	"io/ioutil"
)

type reportList map[string]report

type report struct {
	Name       string     `xml:"name"`
	Title      string     `xml:"title"`
	Parameters parameters `xml:"parameters"`
	Datasource datasource `xml:"datasource"`
}

type parameters struct {
	Parameters []string `xml:"parameter"`
}

type datasource struct {
	DBName    string    `xml:"dbname"`
	Viewtable viewtable `xml:"viewtable"`
}

type viewtable struct {
	Tablename string     `xml:"tablename"`
	Fields    fields     `xml:"fields"`
	Filter    conditions `xml:"filter"`
}

type fields struct {
	Fieldnames []string `xml:"field"`
}

type conditions struct {
	Conditions []string `xml:"condition"`
}

func LoadReport(path string) (reportList, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := report{}
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}
	rl := make(reportList)
	rl[r.Name] = r
	return rl, nil
}
