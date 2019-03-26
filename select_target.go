package main

import (
	"flag"
	"fmt"
	"github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	pSep   = string(os.PathSeparator)
	dbPath = exPath + pSep + "db" + pSep
)

// flag
var (
	varAnnos = flag.String(
		"var",
		"",
		"input annotation tsv(.gz)",
	)
	code = flag.String(
		"code",
		"",
		"code key for decode",
	)
	database = flag.String(
		"database",
		"",
		"databases to use,join with '+'",
	)
	aes = flag.String(
		"aes",
		"",
		"db.aes",
	)
	json = flag.String(
		"json",
		dbPath+"final.20181229.fix.tsv.xlsx.lite.json",
		"db.json")
	prefix = flag.String(
		"prefix",
		"",
		"output prefix.[xlsx,tsv]",
	)
	sheetName = flag.String(
		"sheetName",
		"annotation",
		"output sheetName",
	)
)

var (
	skip = regexp.MustCompile(`^##`)
	isGz = regexp.MustCompile(`\.gz(ip)?$`)
)

var code1 = []byte("118b09d39a5d3ecd56f9bd4f351dd6d6")

//var code2=[]byte("0e0760259f0826d18eb6e22988804617")

var addHeader = []string{
	"MutationID",
	"ClinVar Significance",
	"dbscSNV_ADA_SCORE",
	"dbscSNV_RF_SCORE",
	"GERP++_RS",
	"GWASdb_or",
	"SIFT Pred",
	"Polyphen2 HDIV Pred",
	"Polyphen2 HVAR Pred",
	"MutationTaster Pred",
	"PP_disGroup",
	"Chinese desease name",
	"Inheritance",
	"Chinese mutation information",
	"Chinese desease introduction",
	"English desease name",
	"English mutation information",
	"English desease introduction",
	"Evidence New + Check",
	"Auto ACMG + Check",
	"Reference-final",
	"Reference-final-Info",
	"Database",
}

func main() {
	flag.Parse()
	if *varAnnos == "" || *prefix == "" {
		flag.Usage()
		os.Exit(1)
	}

	file, err := os.Create(*prefix + ".tsv")
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(file)

	excel := xlsx.NewFile()
	sheet, err := excel.AddSheet(*sheetName)
	simple_util.CheckErr(err)

	var inDb = make(map[string]bool)
	tag := strings.Split(*database, "+")
	for _, k := range tag {
		inDb[k] = true
	}

	var db = make(map[string]map[string]string)
	if *aes != "" {
		// get code2 to decode db.aes to db
		User, err := user.Current()
		simple_util.CheckErr(err)
		usr := User.Username
		code3, err := AES.Encode([]byte(usr), code1)
		simple_util.CheckErr(err)
		code2, err := AES.Decode([]byte(*code), code3)
		b := simple_util.File2Decode(*aes, code2)
		db = simple_util.Json2MapMap(b)
	} else {
		db = simple_util.JsonFile2MapMap(*json)
	}

	var anno []map[string]string
	var title []string

	if isGz.MatchString(*varAnnos) {
		anno, title = simple_util.Gz2MapArray(*varAnnos, "\t", skip)
	} else {
		anno, title = simple_util.File2MapArray(*varAnnos, "\t", skip)
	}

	row := sheet.AddRow()
	for _, str := range append(title, addHeader...) {
		row.AddCell().SetString(str)
	}
	_, err = fmt.Fprintln(file, strings.Join(append(title, addHeader...), "\t"))
	simple_util.CheckErr(err)
	for _, item := range anno {
		key := item["Transcript"] + ":" + item["cHGVS"]
		target, ok := db[key]
		if ok {
			tags := strings.Split(target["Database"], ";")
			var skip = true
			for _, t := range tags {
				if inDb[t] {
					skip = false
				}
			}
			if skip {
				continue
			}
			var line []string
			for _, k := range title {
				line = append(line, item[k])
			}
			for _, k := range addHeader {
				line = append(line, target[k])
			}
			row := sheet.AddRow()
			for _, str := range line {
				row.AddCell().SetString(str)
			}
			_, err = fmt.Fprintln(file, escapeLF(strings.Join(line, "\t")))
			simple_util.CheckErr(err)
		}
	}
	err = excel.Save(*prefix + ".xlsx")
	simple_util.CheckErr(err)
}

func escapeLF(str string) string {
	return strings.Replace(str, "\n", "[n]", -1)
}
