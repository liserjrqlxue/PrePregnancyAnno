package main

import (
	"crypto/md5"
	"encoding/hex"
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
)

var (
	skip = regexp.MustCompile(`^##`)
	isGz = regexp.MustCompile(`\.gz(ip)?$`)
)

var code1 = []byte("118b09d39a5d3ecd56f9bd4f351dd6d6")

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
	t0 := time.Now()
	flag.Parse()
	if *varAnnos == "" || *codeKey == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = *varAnnos
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

	// get code2 to decode db.aes to db
	User, err := user.Current()
	simple_util.CheckErr(err)
	usr := User.Username
	fmt.Printf("Username:\t%s\n", usr)
	codeKeyByte, err := hex.DecodeString(*codeKey)
	simple_util.CheckErr(err)
	fmt.Printf("CodeKey:\t%x\n", codeKeyByte)

	code3, err := AES.Encode([]byte(usr), code1)
	simple_util.CheckErr(err)
	md5sum := md5.Sum([]byte(code3))
	code3fix := hex.EncodeToString(md5sum[:])

	code2, err := AES.Decode(codeKeyByte, []byte(code3fix))
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

	row := sheet.AddRow()
	for _, str := range append(title, addHeader...) {
		row.AddCell().SetString(str)
	}
	_, err = fmt.Fprintln(file, strings.Join(append(title, addHeader...), "\t"))
	simple_util.CheckErr(err)

	dbSep := ";"
	for _, item := range anno {
		key := item["Transcript"] + ":" + item["cHGVS"]
		target, ok := db[key]
		format(item)
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

		if *all || (ok && !skip) {
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
	fmt.Printf("Output tsv:\t%s\n,", *prefix+".tsv")
	fmt.Printf("Output excel:\t%s\n", *prefix+".xlsx")
	fmt.Printf("time elapsed:\t%s\n", time.Now().Sub(t0).String())
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
