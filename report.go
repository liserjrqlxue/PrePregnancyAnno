package main

import (
	"fmt"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"strings"
)

type report struct {
	Tag, Prefix string
	Sheet       *xlsx.Sheet
	Tsv         *os.File
	sheetName   string
	row         *xlsx.Row
	err         error
	count       int64
}

func (r *report) checkError(msg ...string) {
	simple_util.CheckErr(r.err, msg...)
}

func (r *report) addArray(array []string) {
	row := r.Sheet.AddRow()
	for _, str := range array {
		row.AddCell().SetString(str)
	}
	_, r.err = fmt.Fprintln(r.Tsv, escapeLF(strings.Join(array, "\t")))
	r.checkError()
	r.count++
}

func (r *report) save() {
	simple_util.CheckErr(r.Tsv.Close())
	simple_util.CheckErr(r.Sheet.File.Save(r.Prefix + ".xlsx"))
	log.Printf("output %s:%d records\n", r.Tag, r.count)
}

func createReport(tag, sheetName, prefix string) (r *report) {
	r = &report{
		Tag:       tag,
		sheetName: sheetName,
	}
	r.Prefix = strings.Join([]string{prefix, tag}, ".")
	r.Tsv, r.err = os.Create(r.Prefix + ".tsv")
	r.checkError()
	r.Sheet, r.err = xlsx.NewFile().AddSheet(r.sheetName)
	r.checkError()
	return
}

func escapeLF(str string) string {
	return strings.Replace(str, "\n", "[n]", -1)
}

func createReports(header []string) {
	var tags = []string{"OfficialReport", "PP159", "PP159.Outside", "PP10", "PP10.Outside", "Standard"}
	if *outside {
		tags = append(tags, "Outside")
	}
	if *all {
		tags = append(tags, "all")
	}
	// create report
	for _, tag := range tags {
		reportMap[tag] = createReport(tag, *sheetName, *prefix)
		reportMap[tag].addArray(header)
		reportMap[tag].count--
		reportArray = append(reportArray, tag)
	}
}
