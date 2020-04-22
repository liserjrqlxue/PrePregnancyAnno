package main

import (
	"log"
	"path/filepath"
	"strings"

	simpleUtil "github.com/liserjrqlxue/simple-util"
)

func loadDiseaseInfo(excel, inDb, outDb string) {
	_, mapArray := simpleUtil.Sheet2MapArray(excel, inDb)
	for _, item := range mapArray {
		geneDiseaseDb[item["Disease Name(Sub-phenotype)-位点疾病"]+":"+item["*基因名称"]] = item
	}
	_, mapArray = simpleUtil.Sheet2MapArray(excel, outDb)
	for _, item := range mapArray {
		geneDb[item["*基因名称"]] = append(geneDb[item["*基因名称"]], item)
	}
	return
}

func buildDatabaseGeneList(database string) (geneListDb, inDb map[string]bool) {
	geneListDb = make(map[string]bool)
	inDb = make(map[string]bool)
	tag := strings.Split(database, "+")
	for _, k := range tag {
		inDb[k] = true
		geneListFile := filepath.Join(dbPath, strings.Join([]string{k, "gene.list"}, "."))
		if simpleUtil.FileExists(geneListFile) {
			geneList := simpleUtil.File2Array(geneListFile)
			for _, gene := range geneList {
				geneListDb[gene] = true
			}
		} else {
			log.Printf("can not find gene list of [%s]", geneListFile)
		}
	}
	return
}
