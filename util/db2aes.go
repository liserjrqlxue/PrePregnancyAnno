package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	simple_util "github.com/liserjrqlxue/simple-util"

	"github.com/liserjrqlxue/crypto/aes"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	dbPath = filepath.Join(exPath, "..", "db")
)

var (
	prefix = flag.String(
		"prefix",
		"",
		"output prefix.[all.lite].json, default is same to -excel")
	excel = flag.String(
		"excel",
		"",
		"PrePregnancy excel",
	)
	sheet = flag.String(
		"sheet",
		"database",
		"PrePregnancy excel sheet name for database",
	)
	keys = flag.String(
		"keys",
		"Transcript:cHGVS",
		"keys joint as map key",
	)
	sep = flag.String(
		"sep",
		":",
		"sep of keys",
	)
	codeKey = flag.String(
		"codeKey",
		"0e0760259f0826d18eb6e22988804617",
		"codeKey for encode db",
	)
	all = flag.Bool(
		"all",
		false,
		"if output all db",
	)
	json = flag.Bool(
		"json",
		false,
		"if output json file",
	)
	liteColumnList = flag.String(
		"liteCols",
		filepath.Join(dbPath, "extraColumn.list"),
		"columns of lite db",
	)
	extract = flag.String(
		"extract",
		"",
		"extract tsv, column names split by comma",
	)
)

func main() {
	flag.Parse()
	if *excel == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = *excel
	}
	key := strings.Split(*keys, *sep)
	_, db := simple_util.Sheet2MapArray(*excel, *sheet)
	var allDb = make(map[string]map[string]string)
	var liteDb = make(map[string]map[string]string)
	var liteCols = textUtil.File2Array(*liteColumnList)

	var extractFile *os.File
	var extractCols []string
	if *extract != "" {
		extractFile = osUtil.Create(*prefix + ".mut.tsv")
		extractCols = strings.Split(*extract, ",")
		fmtUtil.FprintStringArray(extractFile, extractCols, "\t")
	}
	defer func() {
		if *extract != "" {
			simpleUtil.DeferClose(extractFile)
		}
	}()

	for _, item := range db {
		var keyValues []string
		for _, k := range key {
			keyValues = append(keyValues, item[k])
		}
		mainKey := strings.Join(keyValues, *sep)
		allDb[mainKey] = item
		var lite = make(map[string]string)
		for _, k := range liteCols {
			lite[k] = item[k]
		}
		liteDb[mainKey] = lite
		if *extract != "" {
			var strArray []string
			for _, col := range extractCols {
				strArray = append(strArray, item[col])
			}
			fmtUtil.FprintStringArray(extractFile, strArray, "\t")
		}
	}

	var liteB = simpleUtil.HandleError(jsonUtil.JsonIndent(liteDb, "", "\t")).([]byte)
	if *json {
		simpleUtil.CheckErr(jsonUtil.Json2file(liteB, *prefix+".lite.json"))
	}
	AES.Encode2File(*prefix+".lite.json.aes", liteB, []byte(*codeKey))

	if *all {
		var allB = simpleUtil.HandleError(jsonUtil.JsonIndent(allDb, "", "\t")).([]byte)
		if *json {
			simpleUtil.CheckErr(jsonUtil.Json2file(allB, *prefix+".all.json"))
		}
		AES.Encode2File(*prefix+".all.json.aes", allB, []byte(*codeKey))
	}

}
