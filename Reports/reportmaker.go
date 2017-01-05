package Reports

import (
	"encoding/xml"
	"io/ioutil"
)

type report struct {
	Title      string      `xml:title`
	Parameters []parameter `xml:"parameters"`
	Datasource datasource  `xml:"datasource"`
	Tables     []table     `xml:"table"`
}

type parameter struct {
	Parameter string `xml:"parameter"`
}

type datasource struct {
	DBName    string    `xml:"dbname"`
	Viewtable viewtable `xml:"viewtable"`
}

type viewtable struct {
	Tablename string `xml:"tablename"`
}

type fields struct {
	Fieldnames []string `xml:"field"`
}

type conditions struct {
	Conditions []string `xml:"condition"`
}

type table struct {
	Heading string     `xml:"heading"`
	Fields  fields     `xml:"fields"`
	Filter  conditions `xml:"filter"`
}

func LoadReport(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	r := report{}
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return err
	}
	return nil
}
