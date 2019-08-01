package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/simple-util"
	"log"
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
	dbPath = filepath.Join(exPath, "db")
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
		"databases to use for StandardReport,join with '+'",
	)
	aes = flag.String(
		"aes",
		filepath.Join(dbPath, "db.lite.json.aes"),
		"db.aes",
	)
	alleleFrequencyList = flag.String(
		"afl",
		filepath.Join(dbPath, "AF.list"),
		"allele frequency list for checkout outside variants",
	)
	alleleFrequencyThreshold = flag.Float64(
		"aft",
		0.01,
		"allele frequency threshold for checkout outside variants",
	)
	LoFList = flag.String(
		"lof",
		filepath.Join(dbPath, "LoF.list"),
		"LoF list for checkout outside variiants",
	)
	officialReportList = flag.String(
		"orl",
		filepath.Join(dbPath, "OfficialReport.list"),
		"official report mutation list",
	)
	PP100GeneList = flag.String(
		"PP100",
		filepath.Join(dbPath, "PP100.gene.list"),
		"Supplementary Report PP100 gene list",
	)
	PP10GeneList = flag.String(
		"PP10",
		filepath.Join(dbPath, "PP10.gene.list"),
		"Supplementary Report PP10 gene list",
	)
	extraColumnList = flag.String(
		"extraCols",
		filepath.Join(dbPath, "extraColumn.list"),
		"extra columns add to annotation output",
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
		"if output OutsideReport",
	)
)

var (
	skip = regexp.MustCompile(`^##`)
	isGz = regexp.MustCompile(`\.gz(ip)?$`)
)

var code1 = []byte("118b09d39a5d3ecd56f9bd4f351dd6d6")
var code2, code3, codeKeyByte []byte

var officialReport *Report
var PP100Report *Report
var PP100OutsideReport *Report
var PP10Report *Report
var PP10OutsideReport *Report
var allReport *Report
var standardReport *Report
var outsideReport *Report

var orl map[string]map[string]string
var AFList []string
var LoF = make(map[string]bool)
var PP10 = make(map[string]bool)
var PP100 = make(map[string]bool)

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

	// parser config
	// load ORL
	orl = simple_util.File2MapMap(*officialReportList, "Transcript:cHGVS", "\t")

	AFList = simple_util.File2Array(*alleleFrequencyList)

	LoFArray := simple_util.File2Array(*LoFList)
	for _, function := range LoFArray {
		LoF[function] = true
	}

	PP100Gene := simple_util.File2Array(*PP100GeneList)
	for _, gene := range PP100Gene {
		PP100[gene] = true
	}
	PP10Gene := simple_util.File2Array(*PP10GeneList)
	for _, gene := range PP10Gene {
		PP10[gene] = true
	}

	var inDb = make(map[string]bool)
	tag := strings.Split(*database, "+")
	for _, k := range tag {
		inDb[k] = true
	}

	var extraCols = simple_util.File2Array(*extraColumnList)

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

	var header = append(title, extraCols...)

	// create Report
	officialReport = createReport("OfficialReport", *sheetName, *prefix)
	officialReport.addArray(header)
	officialReport.count--
	PP100Report = createReport("PP100", *sheetName, *prefix)
	PP100Report.addArray(header)
	PP100Report.count--
	PP100OutsideReport = createReport("PP100.Outside", *sheetName, *prefix)
	PP100OutsideReport.addArray(header)
	PP100OutsideReport.count--
	PP10Report = createReport("PP10", *sheetName, *prefix)
	PP10Report.addArray(header)
	PP10Report.count--
	PP10OutsideReport = createReport("PP10.Outside", *sheetName, *prefix)
	PP10OutsideReport.addArray(header)
	PP10OutsideReport.count--

	standardReport = createReport("Standard", *sheetName, *prefix)
	standardReport.addArray(header)
	standardReport.count--
	if *all {
		allReport = createReport("all", *sheetName, *prefix)
		allReport.addArray(header)
		allReport.count--
	}
	if *outside {
		outsideReport = createReport("Outside", *sheetName, *prefix)
		outsideReport.addArray(header)
		outsideReport.count--
	}

	dbSep := ";"
	for _, item := range anno {
		gene := item["Gene Symbol"]
		key := item["Transcript"] + ":" + item["cHGVS"]
		format(item)

		target, ok := db[key]
		var line []string
		for _, k := range title {
			line = append(line, item[k])
		}
		for _, k := range extraCols {
			line = append(line, target[k])
		}

		var isOutside bool
		if !ok && LoF[item["Function"]] && simple_util.CheckAFAllLowThen(item, AFList, *alleleFrequencyThreshold, true) {
			isOutside = true
		}

		_, inORL := orl[key]
		if inORL {
			officialReport.addArray(line)
		}
		if PP100[gene] {
			if ok {
				PP100Report.addArray(line)
			} else if isOutside {
				PP100OutsideReport.addArray(line)
			}
		}
		if PP10[gene] {
			if ok {
				PP10Report.addArray(line)
			} else if isOutside {
				PP10OutsideReport.addArray(line)
			}
		}
		if *all {
			allReport.addArray(line)
		}

		// check if in given db
		var skip = true
		tags := strings.Split(target["Database"], dbSep)
		for _, t := range tags {
			if inDb[t] {
				skip = false
			}
		}
		if ok {
			if !skip {
				standardReport.addArray(line)
			}
		} else {
			if *outside && isOutside {
				outsideReport.addArray(line)
			}
		}
	}

	// save report
	officialReport.save()
	PP100Report.save()
	PP100OutsideReport.save()
	PP10Report.save()
	PP10OutsideReport.save()

	standardReport.save()
	if *all {
		allReport.save()
	}
	if *outside {
		outsideReport.save()
	}

	log.Printf("time elapsed:\t%s\n", time.Now().Sub(t0).String())
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
