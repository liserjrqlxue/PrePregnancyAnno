package main

import (
	"flag"
	"github.com/liserjrqlxue/simple-util"
	"os"
	"strings"
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
)

var headerMap = map[string]string{
	"变异ID":                 "MutationID",
	"ClinVar Significance": "ClinVar Significance",
	"dbscSNV_ADA_SCORE":    "dbscSNV_ADA_SCORE",
	"dbscSNV_RF_SCORE":     "dbscSNV_RF_SCORE",
	"GERP++_RS":            "GERP++_RS",
	"GWASdb_or":            "GWASdb_or",
	"SIFT Pred":            "SIFT Pred",
	"Polyphen2 HDIV Pred":  "Polyphen2 HDIV Pred",
	"Polyphen2 HVAR Pred":  "Polyphen2 HVAR Pred",
	"MutationTaster Pred":  "MutationTaster Pred",
	"PP_disGroup":          "PP_disGroup",
	"中文-疾病名称":              "Chinese desease name",
	"遗传模式":                 "Inheritance",
	"中文-突变详情":              "Chinese mutation information",
	"中文-疾病简介":              "Chinese desease introduction",
	"英文-疾病名称":              "English desease name",
	"英文-突变详情":              "English mutation information",
	"英文-疾病简介":              "English desease introduction",
	"Evidence New + Check": "Evidence New + Check",
	"Auto ACMG + Check":    "Auto ACMG + Check",
	"Reference-final":      "Reference-final",
	"Reference-final-Info": "Reference-final-Info",
	"Database":             "Database",
}

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

	for _, item := range db {
		var keyValues []string
		for _, k := range key {
			keyValues = append(keyValues, item[k])
		}
		mainKey := strings.Join(keyValues, *sep)
		allDb[mainKey] = item
		var lite = make(map[string]string)
		for k, v := range headerMap {
			lite[v] = item[k]
		}
		liteDb[mainKey] = lite
	}

	liteB, err := simple_util.JsonIndent(liteDb, "", "\t")
	simple_util.CheckErr(err)
	err = simple_util.Json2file(liteB, *prefix+".lite.json")
	simple_util.CheckErr(err)
	liteF, err := os.Create(*prefix + ".lite.json.aes")
	defer simple_util.DeferClose(liteF)
	simple_util.CheckErr(err)
	simple_util.Encode2File(liteF, liteB, []byte(*codeKey))

	allB, err := simple_util.JsonIndent(allDb, "", "\t")
	simple_util.CheckErr(err)
	err = simple_util.Json2file(allB, *prefix+".all.json")
	simple_util.CheckErr(err)
	allF, err := os.Create(*prefix + ".all.json.aes")
	defer simple_util.DeferClose(allF)
	simple_util.CheckErr(err)
	simple_util.Encode2File(allF, allB, []byte(*codeKey))
}
