package Reports

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Root Level
type report struct {
	Name       string     `xml:"name"`
	Module     string     `xml:"module"`
	Category   string     `xml:"category"`
	Title      string     `xml:"title"`
	Parameters parameters `xml:"parameters"`
	DataSource datasource `xml:"datasource"`
	Layout     layout     `xml:"layout"`
}

// 1st Level
type parameters struct {
	Parameters []string `xml:"parameter"`
}

type datasource struct {
	DBName    string    `xml:"dbname"`
	ViewTable viewTable `xml:"viewtable"`
}

type layout struct {
	Toolbar string `xml:"toolbar"`
	Export  string `xml:"export"`
	Tabs    []tab  `xml:"tab"`
}

// 2nd Level
type viewTable struct {
	TableName string     `xml:"tablename"`
	Fields    fields     `xml:"fields"`
	Filters   conditions `xml:"filter"`
	Groupings fields     `xml:"groupings"`
	Orderings fields     `xml:"ordering"`
}

// 3rd level and beyond
type fields struct {
	FieldNames []field `xml:"field"`
}

type field struct {
	Aggregate   string `xml:"aggregate,attr"`
	DisplayName string `xml:"displayname,attr"`
	Name        string `xml:",innerxml"`
}

type conditions struct {
	Conditions []string `xml:"condition"`
}

/*
type tabs struct {
	Tabs []tab `xml:"tab"`
}*/

type tab struct {
	Name     string `xml:"name"`
	Template string `xml:"template"`
}

type reportList map[string]report

type reportServer struct {
	Reports reportList
	//Modules []string
}

func NewReportServer(rootPath string) (reportServer, error) {
	rl := make(reportList)
	err := filepath.Walk(rootPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".xml" {
			data, er := ioutil.ReadFile(path)
			if er != nil {
				return er
			}
			r := report{}
			er = xml.Unmarshal(data, &r)
			if er != nil {
				return er
			}
			rl[r.Name] = r
		}
		return nil
	})
	return reportServer{Reports: rl}, err
}
