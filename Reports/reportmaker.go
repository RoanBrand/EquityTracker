package Reports

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type reportList map[string]report

type report struct {
	Name       string     `xml:"name"`
	Title      string     `xml:"title"`
	Module     string     `xml:"module"`
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
	TableName string     `xml:"tablename"`
	Fields    fields     `xml:"fields"`
	Filters   conditions `xml:"filter"`
	Groupings fields     `xml:"groupings"`
	Orderings fields     `xml:"ordering"`
}

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

type reportServer struct {
	Reports reportList
	Modules []string
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
			log.Println(r.Datasource.Viewtable)
		}
		return nil
	})
	return reportServer{Reports: rl}, err
}
