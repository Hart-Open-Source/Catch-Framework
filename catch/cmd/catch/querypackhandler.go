package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func loadquerypacks(globalconfig *Config, platform string) error {
	var querypackjsonpath string
	globalconfig.querypacks = nil
	if platform == "workstations" {
		querypackjsonpath = "/catch/osquery_packs/workstations"
	} else if platform == "servers" {
		querypackjsonpath = "/catch/osquery_packs/servers"
	}
	var files []string

	err := filepath.Walk(querypackjsonpath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		getquerypack(globalconfig, file)
	}
	return nil
}

func getquerypack(globalconfig *Config, querypackpath string) error {
	// Load workstation or server osquery querypacks into globalconfig's map
	jsonfile, err := os.Open(querypackpath)
	if err != nil {
		panic(err)
	}
	defer jsonfile.Close()

	jsonbytes, _ := ioutil.ReadAll(jsonfile)

	var querypack map[string]interface{}
	queriesmap := make(map[string]Query)
	err = json.Unmarshal(jsonbytes, &querypack)
	if err != nil {
		panic(err)
	}

	for _, querylist := range querypack {
		for queryname, queryvalue := range querylist.(map[string]interface{}) {
			var query Query
			var specifications []string
			var implementations []string
			var hitrustcontrols []string
			var matches []string
			var successconditions []string

			for _, v := range queryvalue.(map[string]interface{})["specifications"].([]interface{}) {
				specifications = append(specifications, v.(string))
			}
			query.Specifications = specifications

			for _, v := range queryvalue.(map[string]interface{})["implementations"].([]interface{}) {
				implementations = append(implementations, v.(string))
			}
			query.Implementations = implementations

			for _, v := range queryvalue.(map[string]interface{})["hitrust_controls"].([]interface{}) {
				hitrustcontrols = append(hitrustcontrols, v.(string))
			}
			query.Hitrustcontrols = hitrustcontrols

			for _, v := range queryvalue.(map[string]interface{})["matches"].([]interface{}) {
				matches = append(matches, v.(string))
			}
			query.Matches = matches

			for _, v := range queryvalue.(map[string]interface{})["success_conditions"].([]interface{}) {
				successconditions = append(successconditions, v.(string))
			}
			query.Successconditions = successconditions
			queriesmap[queryname] = query
		}
	}
	globalconfig.querypacks = append(globalconfig.querypacks, queriesmap)
	printout("Query pack file loaded: " + querypackpath)
	return nil
}

type Query struct {
	Specifications    []string
	Implementations   []string
	Matches           []string
	Successconditions []string
	Hitrustcontrols   []string
}
