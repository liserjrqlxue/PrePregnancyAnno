package main

import (
	"fmt"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"strings"
)

type Report struct {
	Tag, Prefix string
	Sheet       *xlsx.Sheet
	Tsv         *os.File
	sheetName   string
	row         *xlsx.Row
	err         error
}

func (report *Report) checkError(msg ...string) {
	simple_util.CheckErr(report.err, msg...)
}

func (report *Report) addArray(array []string) {
	row := report.Sheet.AddRow()
	for _, str := range array {
		row.AddCell().SetString(str)
	}
	_, report.err = fmt.Fprintln(report.Tsv, escapeLF(strings.Join(array, "\t")))
	report.checkError()
}

func (report *Report) save() {
	simple_util.CheckErr(report.Tsv.Close())
	simple_util.CheckErr(report.Sheet.File.Save(report.Prefix + ".xlsx"))
	log.Printf("output %s:\n\t[%s.tsv]\n\t[%s.xlsx]\n", report.Tag, report.Prefix, report.Prefix)
}

func createReport(tag, sheetName, prefix string) (report *Report) {
	report = &Report{
		Tag:       tag,
		sheetName: sheetName,
	}
	report.Prefix = strings.Join([]string{prefix, tag}, ".")
	report.Tsv, report.err = os.Create(report.Prefix + ".tsv")
	report.checkError()
	report.Sheet, report.err = xlsx.NewFile().AddSheet(report.sheetName)
	report.checkError()
	return
}
