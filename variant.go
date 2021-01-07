package main

import (
	"regexp"
	"strings"

	simpleUtil "github.com/liserjrqlxue/simple-util"
)

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

func addDiseaseColumns(ok bool, addDisCols []string, item map[string]string, geneDisDb map[string]map[string]string, geneDb map[string][]map[string]string) {
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

func updateDisease(item, db map[string]string, inDb bool) {
	var gene = item["Gene Symbol"]
	for _, k := range extraCols {
		var diseaseName []string
		if k == "Chinese disease name" {
			if inDb {
				for _, diseaseNameEN := range strings.Split(db["English disease name"], "\n") {
					diseaseName = append(diseaseName, geneDiseaseDb[diseaseNameEN+":"+gene]["疾病名称-亚型"])
				}
			} else {
				for _, info := range geneDb[gene] {
					diseaseName = append(diseaseName, info["疾病名称-亚型"])
				}
			}
			item[k] = strings.Join(diseaseName, "\n")
		} else if !inDb && k == "English disease name" {
			for _, info := range geneDb[gene] {
				diseaseName = append(diseaseName, info["Disease Name(Sub-phenotype)-位点疾病"])
			}
			item[k] = strings.Join(diseaseName, "\n")
		} else {
			item[k] = db[k]
		}
	}
}

func isOutside(item map[string]string, inDb bool) bool {
	if !inDb && lof[item["Function"]] && simpleUtil.CheckAFAllLowThen(item, afList, *alleleFrequencyThreshold, true) {
		return true
	}
	return false
}

func isStandard(db map[string]string, dbSep string) bool {
	for _, t := range strings.Split(db["Database"], dbSep) {
		if inDb[t] {
			return true
		}
	}
	return false
}

func getVals(item map[string]string, keys []string) (vals []string) {
	for i := range keys {
		vals = append(vals, item[keys[i]])
	}
	return
}

func loadVar(item map[string]string, title []string, db map[string]map[string]string, dbSep string) {
	var gene = item["Gene Symbol"]
	var key = item["Transcript"] + ":" + item["cHGVS"]
	format(item)

	target, ok := db[key]
	updateDisease(item, target, ok)
	addDiseaseColumns(ok, addDisCols, item, geneDiseaseDb, geneDb)

	var line = getVals(item, append(title, extraCols...))
	var isOutside = isOutside(item, ok)

	if _, inORL := orl[key]; inORL {
		reportMap["OfficialReport"].addArray(line)
	}
	if ok {
		if pp159[gene] {
			reportMap["PP159"].addArray(line)
		}
		if pp10[gene] {
			reportMap["PP10"].addArray(line)
		}
		if isStandard(target, dbSep) {
			reportMap["Standard"].addArray(line)
		}
	} else if isOutside {
		if pp159[gene] {
			reportMap["PP159.Outside"].addArray(line)
		}
		if pp10[gene] {
			reportMap["PP10.Outside"].addArray(line)
		}
		if *outside && geneListDb[gene] {
			reportMap["Outside"].addArray(line)
		}
	}
	if *all {
		reportMap["all"].addArray(line)
	}
}
