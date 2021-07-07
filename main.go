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
	productCode = flag.String(
		"productCode",
		"",
		"use productCode to find gene.list, override -database for standard and outside",
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
		filepath.Join(dbPath, "Y150_gene_disease_list.xlsx"),
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

var orl map[string]map[string]string
var afList []string
var lof = make(map[string]bool)
var pp10 = make(map[string]bool)
var pp159 = make(map[string]bool)
var (
	reportArray []string
	reportMap   = make(map[string]*report)

	geneDiseaseDb = make(map[string]map[string]string)
	geneDb        = make(map[string][]map[string]string)

	geneListDb, inDb      map[string]bool
	extraCols, addDisCols []string
)

var t0 = time.Now()

func init() {
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

	afList = simpleUtil.File2Array(*alleleFrequencyList)

	LoFArray := simpleUtil.File2Array(*LoFList)
	for _, function := range LoFArray {
		lof[function] = true
	}

	for _, gene := range simpleUtil.File2Array(*PP159GeneList) {
		pp159[gene] = true
	}
	for _, gene := range simpleUtil.File2Array(*PP10GeneList) {
		pp10[gene] = true
	}
	if *productCode != "" {
		geneListDb, inDb = buildDatabaseGeneList(*productCode)
	} else {
		geneListDb, inDb = buildDatabaseGeneList(*database)
		if *database == "" {
			log.Println("empty database")
		}
	}
	extraCols = simpleUtil.File2Array(*extraColumnList)
	addDisCols = simpleUtil.File2Array(*additionalDiseaseColumnList)
}

func decodeDb() map[string]map[string]string {
	var err error
	// get code2 to decode db.aes to db
	if *username == "" {
		User, err := user.Current()
		simpleUtil.CheckErr(err)
		*username = User.Username
	}
	log.Printf("Username:\t%s\n", *username)
	codeKeyByte, err = hex.DecodeString(*codeKey)
	simpleUtil.CheckErr(err)
	log.Printf("CodeKey:\t%x************************************************%x\n", codeKeyByte[0:4], codeKeyByte[len(codeKeyByte)-4:])
	log.Printf("Code:\t%s************************%s\n", string(code[0:4]), string(code[len(code)-4:]))

	code3, err = AES.Encode([]byte(*username), code1)
	simpleUtil.CheckErr(err)
	var md5sum = md5.Sum(code3)
	code3fix := hex.EncodeToString(md5sum[:])

	code2, err = AES.Decode(codeKeyByte, []byte(code3fix))
	simpleUtil.CheckErr(err)
	return simpleUtil.Json2MapMap(
		simpleUtil.File2Decode(
			*aes,
			code2,
		),
	)
}

func loadMapArray(path string) (db []map[string]string, title []string) {
	if isGz.MatchString(path) {
		db, title = simpleUtil.Gz2MapArray(*varAnnos, "\t", skip)
	} else {
		db, title = simpleUtil.File2MapArray(*varAnnos, "\t", skip)
	}
	return
}

func main() {
	var db = decodeDb()
	var anno, title = loadMapArray(*varAnnos)

	createReports(append(title, extraCols...))

	for _, item := range anno {
		loadVar(item, title, db, ";")
	}

	// save report
	for _, tag := range reportArray {
		reportMap[tag].save()
	}

	log.Printf("time elapsed:\t%s\n", time.Now().Sub(t0).String())
}
