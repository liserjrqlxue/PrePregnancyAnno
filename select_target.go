package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/liserjrqlxue/crypto/aes"
	simpleUtil "github.com/liserjrqlxue/simple-util"
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
		"PP159+F8",
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
	PP159GeneList = flag.String(
		"PP159",
		filepath.Join(dbPath, "PP159.gene.list"),
		"Supplementary Report PP159 gene list",
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
	additionalDiseaseColumnList = flag.String(
		"addDisCols",
		filepath.Join(dbPath, "additionalDiseaseColumn.list"),
		"additional disease column to be filled up",
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
	diseaseInfo = flag.String(
		"diseaseInfo",
		filepath.Join(dbPath, "孕150基因-疾病 -发病率-干预list-20200302-V3-To韩瑞-修改后N-补充库外配置文件-0421.xlsx"),
		"diseaseInfo from excel",
	)
	diseaseInDb = flag.String(
		"diseaseInDb",
		"孕150库内位点配置",
		"diseaseInfo in db sheet name",
	)
	diseaseOutDb = flag.String(
		"diseaseOutDb",
		"库外位点配置文件",
		"diseaseInfo out db sheet name",
	)
)

var (
	skip = regexp.MustCompile(`^##`)
	isGz = regexp.MustCompile(`\.gz(ip)?$`)
)

var code = "118b09d39a5d3ecd56f9bd4f351dd6d6"
var code1 = []byte(code)
var code2, code3, codeKeyByte []byte

var reportArray []string
var reportMap = make(map[string]*Report)

var orl map[string]map[string]string
var AFList []string
var LoF = make(map[string]bool)
var PP10 = make(map[string]bool)
var PP159 = make(map[string]bool)
var err error

var (
	geneDiseaseDb = make(map[string]map[string]string)
	geneDb        = make(map[string][]map[string]string)
)

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
	// load Disease Info
	loadDiseaseInfo(*diseaseInfo, *diseaseInDb, *diseaseOutDb)
	// load ORL
	orl = simpleUtil.File2MapMap(*officialReportList, "Transcript:cHGVS", "\t")

	AFList = simpleUtil.File2Array(*alleleFrequencyList)

	LoFArray := simpleUtil.File2Array(*LoFList)
	for _, function := range LoFArray {
		LoF[function] = true
	}

	for _, gene := range simpleUtil.File2Array(*PP159GeneList) {
		PP159[gene] = true
	}
	for _, gene := range simpleUtil.File2Array(*PP10GeneList) {
		PP10[gene] = true
	}

	if *database == "" {
		log.Println("empty database")
	}

	var geneListDb, inDb = buildDatabaseGeneList(*database)

	var extraCols = simpleUtil.File2Array(*extraColumnList)
	var addDisCols = simpleUtil.File2Array(*additionalDiseaseColumnList)

	var db = make(map[string]map[string]string)

	// get code2 to decode db.aes to db
	if *username == "" {
		User, err := user.Current()
		simpleUtil.CheckErr(err)
		*username = User.Username
	}
	log.Printf("Username:\t%s\n", *username)
	codeKeyByte, err = hex.DecodeString(*codeKey)
	simpleUtil.CheckErr(err)
	log.Printf("CodeKey:\t%x\n", codeKeyByte)

	code3, err = AES.Encode([]byte(*username), code1)
	simpleUtil.CheckErr(err)
	var md5sum = md5.Sum(code3)
	code3fix := hex.EncodeToString(md5sum[:])

	code2, err = AES.Decode(codeKeyByte, []byte(code3fix))
	simpleUtil.CheckErr(err)
	b := simpleUtil.File2Decode(*aes, code2)
	db = simpleUtil.Json2MapMap(b)

	var anno []map[string]string
	var title []string

	if isGz.MatchString(*varAnnos) {
		anno, title = simpleUtil.Gz2MapArray(*varAnnos, "\t", skip)
	} else {
		anno, title = simpleUtil.File2MapArray(*varAnnos, "\t", skip)
	}

	var header = append(title, extraCols...)

	// create Report
	for _, tag := range []string{"OfficialReport", "PP159", "PP159.Outside", "PP10", "PP10.Outside", "Standard"} {
		reportMap[tag] = createReport(tag, *sheetName, *prefix)
		reportMap[tag].addArray(header)
		reportMap[tag].count--
		reportArray = append(reportArray, tag)
	}
	if *outside {
		tag := "Outside"
		reportMap[tag] = createReport(tag, *sheetName, *prefix)
		reportMap[tag].addArray(header)
		reportMap[tag].count--
		reportArray = append(reportArray, tag)
	}
	if *all {
		tag := "all"
		reportMap[tag] = createReport(tag, *sheetName, *prefix)
		reportMap[tag].addArray(header)
		reportMap[tag].count--
		reportArray = append(reportArray, tag)
	}

	dbSep := ";"
	for _, item := range anno {
		gene := item["Gene Symbol"]
		transcript := item["Transcript"]
		key := transcript + ":" + item["cHGVS"]
		format(item)

		target, ok := db[key]
		for _, k := range extraCols {
			var diseaseName []string
			if k == "Chinese disease name" {
				if ok {
					for _, diseaseNameEN := range strings.Split(target["English disease name"], "\n") {
						diseaseName = append(diseaseName, geneDiseaseDb[diseaseNameEN+":"+gene]["疾病名称-亚型"])
					}
				} else {
					for _, info := range geneDb[gene] {
						diseaseName = append(diseaseName, info["疾病名称-亚型"])
					}
				}
				item[k] = strings.Join(diseaseName, "\n")
			} else if !ok && k == "English disease name" {
				for _, info := range geneDb[gene] {
					diseaseName = append(diseaseName, info["Disease Name(Sub-phenotype)-位点疾病"])
				}
				item[k] = strings.Join(diseaseName, "\n")
			} else {
				item[k] = target[k]
			}
		}
		updateDisease(ok, addDisCols, item, geneDiseaseDb, geneDb)

		var line []string
		for _, k := range title {
			line = append(line, item[k])
		}
		for _, k := range extraCols {
			line = append(line, item[k])
		}

		var isOutside bool
		if !ok && LoF[item["Function"]] && simpleUtil.CheckAFAllLowThen(item, AFList, *alleleFrequencyThreshold, true) {
			isOutside = true
		}

		_, inORL := orl[key]
		if inORL {
			reportMap["OfficialReport"].addArray(line)
		}
		if PP159[gene] {
			if ok {
				reportMap["PP159"].addArray(line)
			} else if isOutside {
				reportMap["PP159.Outside"].addArray(line)
			}
		}
		if PP10[gene] {
			if ok {
				reportMap["PP10"].addArray(line)
			} else if isOutside {
				reportMap["PP10.Outside"].addArray(line)
			}
		}
		if *all {
			reportMap["all"].addArray(line)
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
				reportMap["Standard"].addArray(line)
			}
		} else {
			if *outside && isOutside && geneListDb[gene] {
				reportMap["Outside"].addArray(line)
			}
		}
	}

	// save report
	for _, tag := range reportArray {
		reportMap[tag].save()
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

func updateDisease(ok bool, addDisCols []string, item map[string]string, geneDisDb map[string]map[string]string, geneDb map[string][]map[string]string) {
	var info map[string][]string
	var gene = item["Gene Symbol"]
	if ok {
		var diseases = item["English disease name"]
		info = getDisInfoInDb(gene, diseases, addDisCols, geneDisDb)
	} else {
		info = getDisInfoOutDb(gene, addDisCols, geneDb)
	}
	for _, key := range addDisCols {
		item[key] = strings.Join(info[key], "\n")
	}
}

func getDisInfoInDb(gene, diseases string, addDisCols []string, geneDisDb map[string]map[string]string) map[string][]string {
	var info = make(map[string][]string)
	for _, disease := range strings.Split(diseases, "\n") {
		var diseaseInfo = geneDisDb[disease+":"+gene]
		for _, key := range addDisCols {
			info[key] = append(info[key], diseaseInfo[key])
		}
	}
	return info
}

func getDisInfoOutDb(gene string, addDisCols []string, geneDb map[string][]map[string]string) map[string][]string {
	var info = make(map[string][]string)
	for _, diseaseInfo := range geneDb[gene] {
		for _, key := range addDisCols {
			info[key] = append(info[key], diseaseInfo[key])
		}
	}
	return info
}
