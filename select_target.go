package main

import (
	"encoding/csv"
	"flag"
	"github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/simple-util"
	"log"
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
	output = flag.String(
		"output",
		"",
		"output file")
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
	if *varAnnos == "" || *output == "" {
		flag.Usage()
		os.Exit(1)
	}

	file, err := os.Create(*output)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(file)

	w := csv.NewWriter(file)
	w.Comma = '\t'

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

	err = w.Write(append(title, addHeader...))
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
			err = w.Write(line)
			simple_util.CheckErr(err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}
