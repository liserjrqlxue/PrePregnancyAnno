package main

import (
	"flag"
	"github.com/liserjrqlxue/simple-util"
	"os"
	"path/filepath"
	"strings"
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
	var liteCols = simple_util.File2Array(*liteColumnList)

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
	}

	liteB, err := simple_util.JsonIndent(liteDb, "", "\t")
	simple_util.CheckErr(err)
	if *json {
		err = simple_util.Json2file(liteB, *prefix+".lite.json")
		simple_util.CheckErr(err)
	}
	liteF, err := os.Create(*prefix + ".lite.json.aes")
	defer simple_util.DeferClose(liteF)
	simple_util.CheckErr(err)
	simple_util.Encode2File(liteF, liteB, []byte(*codeKey))

	if *all {
		allB, err := simple_util.JsonIndent(allDb, "", "\t")
		simple_util.CheckErr(err)
		if *json {
			err = simple_util.Json2file(allB, *prefix+".all.json")
			simple_util.CheckErr(err)
		}
		allF, err := os.Create(*prefix + ".all.json.aes")
		defer simple_util.DeferClose(allF)
		simple_util.CheckErr(err)
		simple_util.Encode2File(allF, allB, []byte(*codeKey))
	}

}
