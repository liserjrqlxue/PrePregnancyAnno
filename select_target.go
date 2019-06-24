package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	username = flag.String(
		"usr",
		"",
		"username for -codeKey,default is user.Current().Username")
	codeKey = flag.String(
		"code",
		"",
		"code key for decode",
	)
	database = flag.String(
		"database",
		"PP100+F8",
		"databases to use,join with '+'",
	)
	aes = flag.String(
		"aes",
		dbPath+"db.lite.json.aes",
		"db.aes",
	)
	prefix = flag.String(
		"prefix",
		"",
		"output prefix.[xlsx,tsv], default is basename of -var",
	)
	sheetName = flag.String(
		"sheetName",
		"annotation",
		"output sheetName",
	)
	all = flag.Bool(
		"all",
		false,
		"if output all var",
	)
	outside = flag.Bool(
		"outside",
		false,
		"if output outside var",
	)
)

var (
	skip = regexp.MustCompile(`^##`)
	isGz = regexp.MustCompile(`\.gz(ip)?$`)
)

var code1 = []byte("118b09d39a5d3ecd56f9bd4f351dd6d6")
var code2, code3, codeKeyByte []byte

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

var file, allFile, outsideFile *os.File
var sheet, allSheet, outsideSheet *xlsx.Sheet

var err error

func main() {
	t0 := time.Now()
	flag.Parse()
	if *varAnnos == "" || *codeKey == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = *varAnnos
	}

	var inDb = make(map[string]bool)
	tag := strings.Split(*database, "+")
	for _, k := range tag {
		inDb[k] = true
	}

	var db = make(map[string]map[string]string)

	// get code2 to decode db.aes to db
	if *username == "" {
		User, err := user.Current()
		simple_util.CheckErr(err)
		*username = User.Username
	}
	log.Printf("Username:\t%s\n", *username)
	codeKeyByte, err = hex.DecodeString(*codeKey)
	simple_util.CheckErr(err)
	log.Printf("CodeKey:\t%x\n", codeKeyByte)

	code3, err = AES.Encode([]byte(*username), code1)
	simple_util.CheckErr(err)
	md5sum := md5.Sum([]byte(code3))
	code3fix := hex.EncodeToString(md5sum[:])

	code2, err = AES.Decode(codeKeyByte, []byte(code3fix))
	simple_util.CheckErr(err)
	b := simple_util.File2Decode(*aes, []byte(code2))
	db = simple_util.Json2MapMap(b)

	var anno []map[string]string
	var title []string

	if isGz.MatchString(*varAnnos) {
		anno, title = simple_util.Gz2MapArray(*varAnnos, "\t", skip)
	} else {
		anno, title = simple_util.File2MapArray(*varAnnos, "\t", skip)
	}

	// output
	file, err = os.Create(*prefix + ".tsv")
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(file)

	sheet, err = xlsx.NewFile().AddSheet(*sheetName)
	simple_util.CheckErr(err)

	row := sheet.AddRow()
	for _, str := range append(title, addHeader...) {
		row.AddCell().SetString(str)
	}
	_, err = fmt.Fprintln(file, strings.Join(append(title, addHeader...), "\t"))
	simple_util.CheckErr(err)

	if *all {
		allFile, err = os.Create(*prefix + ".all.tsv")
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(allFile)

		allSheet, err = xlsx.NewFile().AddSheet(*sheetName)
		simple_util.CheckErr(err)

		row := allSheet.AddRow()
		for _, str := range append(title, addHeader...) {
			row.AddCell().SetString(str)
		}
		_, err = fmt.Fprintln(allFile, strings.Join(append(title, addHeader...), "\t"))
		simple_util.CheckErr(err)
	}

	if *outside {
		outsideFile, err = os.Create(*prefix + ".outside.tsv")
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(outsideFile)

		outsideSheet, err = xlsx.NewFile().AddSheet(*sheetName)
		simple_util.CheckErr(err)

		row := outsideSheet.AddRow()
		for _, str := range append(title, addHeader...) {
			row.AddCell().SetString(str)
		}
		_, err = fmt.Fprintln(outsideFile, strings.Join(append(title, addHeader...), "\t"))
		simple_util.CheckErr(err)
	}

	dbSep := ";"
	for _, item := range anno {
		key := item["Transcript"] + ":" + item["cHGVS"]
		format(item)

		target, ok := db[key]
		var line []string
		var skip = true
		tags := strings.Split(target["Database"], dbSep)
		for _, t := range tags {
			if inDb[t] {
				skip = false
			}
		}
		for _, k := range title {
			line = append(line, item[k])
		}
		for _, k := range addHeader {
			line = append(line, target[k])
		}

		if *all {
			row := allSheet.AddRow()
			for _, str := range line {
				row.AddCell().SetString(str)
			}
			_, err = fmt.Fprintln(allFile, escapeLF(strings.Join(line, "\t")))
			simple_util.CheckErr(err)
		}
		if ok && !skip {
			row := sheet.AddRow()
			for _, str := range line {
				row.AddCell().SetString(str)
			}
			_, err = fmt.Fprintln(file, escapeLF(strings.Join(line, "\t")))
			simple_util.CheckErr(err)
		} else if *outside && outsideCheck(item) {
			row := outsideSheet.AddRow()
			for _, str := range line {
				row.AddCell().SetString(str)
			}
			_, err = fmt.Fprintln(outsideFile, escapeLF(strings.Join(line, "\t")))
			simple_util.CheckErr(err)
		}
	}
	simple_util.CheckErr(sheet.File.Save(*prefix + ".xlsx"))
	log.Printf("Output tsv:\t%s\n", *prefix+".tsv")
	log.Printf("Output excel:\t%s\n", *prefix+".xlsx")
	if *all {
		simple_util.CheckErr(allSheet.File.Save(*prefix + ".all.xlsx"))
		log.Printf("Output tsv:\t%s\n", *prefix+".all.tsv")
		log.Printf("Output excel:\t%s\n", *prefix+".all.xlsx")
	}
	if *outside {
		simple_util.CheckErr(outsideSheet.File.Save(*prefix + ".outside.xlsx"))
		log.Printf("Output tsv:\t%s\n", *prefix+".outside.tsv")
		log.Printf("Output excel:\t%s\n", *prefix+".outside.xlsx")
	}
	log.Printf("time elapsed:\t%s\n", time.Now().Sub(t0).String())
}

func escapeLF(str string) string {
	return strings.Replace(str, "\n", "[n]", -1)
}

var (
	chr    = regexp.MustCompile(`^chr`)
	repeat = regexp.MustCompile(`dup|trf|;`)
)

func format(item map[string]string) {
	item["#Chr"] = chr.ReplaceAllString(item["#Chr"], "")
	item["RepeatTag"] = repeat.ReplaceAllString(item["RepeatTag"], "")
	if item["RepeatTag"] == "" {
		item["RepeatTag"] = "."
	}
	item["Zygosity"] = strings.Split(item["Zygosity"], "-")[0]
}

// Tier1 >1
// LoF 3
var FuncInfo = map[string]int{
	"splice-3":     3,
	"splice-5":     3,
	"init-loss":    3,
	"alt-start":    3,
	"frameshift":   3,
	"nonsense":     3,
	"stop-gain":    3,
	"span":         3,
	"missense":     2,
	"cds-del":      2,
	"cds-indel":    2,
	"cds-ins":      2,
	"splice-10":    2,
	"splice+10":    2,
	"coding-synon": 1,
	"splice-20":    1,
	"splice+20":    1,
}

var AFlist = []string{
	"ESP6500 AF",
	"1000G AF",
	"ExAC EAS AF",
	"ExAC AF",
	"GnomAD EAS AF",
	"GnomAD AF",
}
var threshold = 0.01

func outsideCheck(item map[string]string) bool {
	if FuncInfo[item["Function"]] < 3 {
		return false
	}
	if CheckAFAllLowThen(item, AFlist, threshold, true) {
		return true
	}
	return false
}

func CheckAFAllLowThen(item map[string]string, AFlist []string, threshold float64, includeEqual bool) bool {
	for _, key := range AFlist {
		af := item[key]
		if af == "" || af == "." || af == "0" {
			continue
		}
		AF, err := strconv.ParseFloat(af, 64)
		simple_util.CheckErr(err)
		if includeEqual {
			if AF > threshold {
				return false
			}
		} else {
			if AF >= threshold {
				return false
			}
		}
	}
	return true
}
