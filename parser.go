package main

import simple_util "github.com/liserjrqlxue/simple-util"

func loadDiseaseInfo(excel, sheet string) {
	_, mapArray := simple_util.Sheet2MapArray(excel, sheet)
	for _, item := range mapArray {
		geneDiseaseDb[item["Disease Name(Sub-phenotype)-位点疾病"]+":"+item["*基因名称"]] = item
		geneDb[item["*基因名称"]] = append(geneDb[item["*基因名称"]], item)
	}
	return
}
